package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// ContainerManager 容器管理器
type ContainerManager struct {
	cli    *client.Client
	logger utils.Logger
}

// NewContainerManager 创建新的容器管理器
func NewContainerManager() (*ContainerManager, error) {
	// 创建日志器
	logger, err := utils.NewWithConfig(utils.LoggerConfig{
		Level:  utils.DEBUG,
		Output: utils.CONSOLE,
		Module: "ContainerManager",
	})
	if err != nil {
		return nil, fmt.Errorf("创建日志器失败: %w", err)
	}

	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(), // 启用 API 版本协商
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Docker 客户端失败: %w", err)
	}

	// 验证连接
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("Docker 连接测试失败: %w", err)
	}

	// 获取服务器 API 版本
	version, err := cli.ServerVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 Docker 服务器版本失败: %w", err)
	}
	logger.Info("Docker 服务器版本: %s, API 版本: %s", version.Version, version.APIVersion)

	return &ContainerManager{
		logger: logger,
		cli:    cli,
	}, nil
}

// generateContainerName 生成容器名称
func generateContainerName() string {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())
	// 生成6位随机数字
	randomNum := rand.Intn(1000000)
	// 格式化为6位数字字符串
	return fmt.Sprintf("BWG%06d", randomNum)
}

// RunContainer 运行容器
func (m *ContainerManager) RunContainer(ctx context.Context, step *models.PipelineStep) error {
	m.logger.Debug("准备运行容器步骤: %s, 镜像: %s", step.Name, step.Container.Image)

	// 确保镜像存在
	if err := m.ensureImageExists(ctx, step.Container.Image); err != nil {
		return fmt.Errorf("准备镜像失败: %w", err)
	}

	// 准备容器配置
	config := &container.Config{
		Image:        step.Container.Image,
		Cmd:          []string{"sh", "-c", joinCommands(step.Container.Commands)},
		Env:          convertMapToSlice(step.Container.Environment),
		Hostname:     step.Container.Hostname,
		AttachStdout: true,
		AttachStderr: true,
	}

	// 准备主机配置
	hostConfig := &container.HostConfig{
		Privileged:  step.Container.Privileged,
		SecurityOpt: step.Container.SecurityOpt,
		CapAdd:      step.Container.CapAdd,
		Tmpfs:       make(map[string]string),
		Binds:       step.Container.Volumes,
	}

	// 设置 Tmpfs
	for _, tmpfs := range step.Container.Tmpfs {
		hostConfig.Tmpfs[tmpfs] = ""
	}

	// 生成容器名称
	containerName := generateContainerName()
	m.logger.Debug("生成容器名称: %s", containerName)

	// 创建容器
	m.logger.Debug("开始创建容器...")
	resp, err := m.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		m.logger.Error("创建容器失败: %v", err)
		return fmt.Errorf("创建容器失败: %w", err)
	}
	m.logger.Debug("容器创建成功，ID: %s", resp.ID)

	m.logger.Debug("开始启动容器: %s", resp.ID)
	// 启动容器
	err = m.cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		m.logger.Error("启动容器失败: %v", err)
		// 尝试清理容器
		if removeErr := m.RemoveContainer(ctx, resp.ID); removeErr != nil {
			m.logger.Error("清理失败的容器失败: %v", removeErr)
		}
		return fmt.Errorf("启动容器失败: %w", err)
	}
	m.logger.Debug("容器启动成功: %s", resp.ID)

	m.logger.Debug("开始获取容器日志流: %s", resp.ID)
	// 获取容器日志流
	logs, err := m.cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	})
	if err != nil {
		m.logger.Error("获取容器日志失败: %v", err)
		// 尝试清理容器
		if removeErr := m.RemoveContainer(ctx, resp.ID); removeErr != nil {
			m.logger.Error("清理失败的容器失败: %v", removeErr)
		}
		return fmt.Errorf("获取容器日志失败: %w", err)
	}
	defer logs.Close()
	m.logger.Debug("容器日志流获取成功: %s", resp.ID)

	// 创建一个通道用于等待容器完成
	done := make(chan error, 1)
	go func() {
		m.logger.Debug("开始等待容器完成: %s", resp.ID)
		// 等待容器完成
		statusCh, errCh := m.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			m.logger.Error("等待容器完成时发生错误: %v", err)
			done <- err
		case status := <-statusCh:
			m.logger.Debug("容器状态更新: %+v", status)
			done <- nil
		}
	}()

	// 创建一个缓冲区用于读取日志
	buf := make([]byte, 1024)
	for {
		select {
		case err := <-done:
			if err != nil {
				m.logger.Error("等待容器完成失败: %v", err)
				return fmt.Errorf("等待容器完成失败: %w", err)
			}
			// 检查容器退出状态
			m.logger.Debug("开始检查容器退出状态: %s", resp.ID)
			inspect, err := m.cli.ContainerInspect(ctx, resp.ID)
			if err != nil {
				m.logger.Error("检查容器状态失败: %v", err)
				return fmt.Errorf("检查容器状态失败: %w", err)
			}
			m.logger.Debug("容器退出状态: %+v", inspect.State)
			if inspect.State.ExitCode != 0 {
				// 删除容器
				m.logger.Debug("容器执行失败，开始删除容器: %s", resp.ID)
				if err := m.RemoveContainer(ctx, resp.ID); err != nil {
					m.logger.Error("删除容器失败: %v", err)
					// 不返回错误，因为容器已经执行失败
				} else {
					m.logger.Debug("容器删除成功: %s", resp.ID)
				}
				return fmt.Errorf("容器执行失败，退出码: %d", inspect.State.ExitCode)
			}
			m.logger.Debug("容器执行完成: %s", resp.ID)

			// 删除容器
			m.logger.Debug("开始删除容器: %s", resp.ID)
			if err := m.RemoveContainer(ctx, resp.ID); err != nil {
				m.logger.Error("删除容器失败: %v", err)
				// 不返回错误，因为容器已经成功执行完成
			} else {
				m.logger.Debug("容器删除成功: %s", resp.ID)
			}

			return nil
		default:
			// 读取日志
			n, err := logs.Read(buf)
			if err != nil {
				if err == io.EOF {
					continue
				}
				return fmt.Errorf("读取容器日志失败: %w", err)
			}
			if n > 0 {
				// 发送日志事件
				m.logger.Debug("[%s] %s", step.Name, string(buf[:n]))
			}
		}
	}
}

// StopContainer 停止容器
func (m *ContainerManager) StopContainer(ctx context.Context, containerID string) error {
	timeout := 10
	err := m.cli.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("停止容器失败: %w", err)
	}
	return nil
}

// RemoveContainer 删除容器
func (m *ContainerManager) RemoveContainer(ctx context.Context, containerID string) error {
	err := m.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("删除容器失败: %w", err)
	}
	return nil
}

// Close 关闭容器管理器
func (m *ContainerManager) Close() error {
	return m.cli.Close()
}

// convertMapToSlice 将 map[string]string 转换为 []string
func convertMapToSlice(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// ensureImageExists 确保镜像存在，如果不存在则拉取
func (m *ContainerManager) ensureImageExists(ctx context.Context, imageName string) error {
	// 检查镜像是否存在
	_, err := m.cli.ImageInspect(ctx, imageName)
	if err == nil {
		m.logger.Debug("镜像已存在: %s", imageName)
		return nil
	}

	// 如果错误不是"镜像不存在"，则返回错误
	if !client.IsErrNotFound(err) {
		return fmt.Errorf("检查镜像状态失败: %w", err)
	}

	// 镜像不存在，开始拉取
	m.logger.Info("开始拉取镜像: %s", imageName)

	// 设置重试次数和间隔
	maxRetries := 3
	retryInterval := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			m.logger.Info("第 %d 次重试拉取镜像: %s", retry+1, imageName)
			select {
			case <-ctx.Done():
				return fmt.Errorf("拉取镜像超时")
			case <-time.After(retryInterval):
			}
		}

		// 拉取镜像
		resp, err := m.cli.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			if retry == maxRetries-1 {
				return fmt.Errorf("拉取镜像失败: %w", err)
			}
			m.logger.Warn("拉取镜像失败，准备重试: %v", err)
			continue
		}

		// 读取拉取进度
		decoder := json.NewDecoder(resp)
		var lastStatus string
		var pullError error
		var shouldBreak bool

		for {
			select {
			case <-ctx.Done():
				resp.Close()
				return fmt.Errorf("拉取镜像超时")
			default:
				var pullResult struct {
					Status   string `json:"status"`
					Error    string `json:"error"`
					Progress string `json:"progress"`
				}
				if err := decoder.Decode(&pullResult); err != nil {
					if err == io.EOF {
						shouldBreak = true
						break
					}
					resp.Close()
					return fmt.Errorf("解析拉取进度失败: %w", err)
				}

				if pullResult.Error != "" {
					pullError = fmt.Errorf("拉取镜像失败: %s", pullResult.Error)
					shouldBreak = true
					break
				}

				// 只在状态变化时记录日志
				if pullResult.Status != lastStatus {
					lastStatus = pullResult.Status
					if pullResult.Progress != "" {
						m.logger.Debug("拉取进度: %s - %s", pullResult.Status, pullResult.Progress)
					} else {
						m.logger.Debug("拉取进度: %s", pullResult.Status)
					}
				}
			}
			if shouldBreak {
				break
			}
		}
		resp.Close()

		if pullError != nil {
			if retry == maxRetries-1 {
				return pullError
			}
			continue
		}

		// 验证镜像是否成功拉取
		_, err = m.cli.ImageInspect(ctx, imageName)
		if err == nil {
			m.logger.Info("镜像拉取成功: %s", imageName)
			return nil
		}

		if retry == maxRetries-1 {
			return fmt.Errorf("镜像拉取完成但验证失败: %w", err)
		}
	}

	return fmt.Errorf("拉取镜像失败，已达到最大重试次数")
}

// joinCommands 将命令列表连接成一个 shell 命令
func joinCommands(commands []string) string {
	if len(commands) == 0 {
		return ""
	}

	// 如果只有一个命令，直接返回
	if len(commands) == 1 {
		return commands[0]
	}

	// 将多个命令用 && 连接
	var result strings.Builder
	for i, cmd := range commands {
		if i > 0 {
			result.WriteString(" && ")
		}
		// 如果命令包含特殊字符，需要加引号
		if strings.ContainsAny(cmd, "|&;<>()$`\\\"' \t\n") {
			result.WriteString("'")
			result.WriteString(strings.ReplaceAll(cmd, "'", "'\\''"))
			result.WriteString("'")
		} else {
			result.WriteString(cmd)
		}
	}
	return result.String()
}
