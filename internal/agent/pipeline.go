package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/agent/proto"
)

// Pipeline 代表一个正在执行的流水线
type Pipeline struct {
	sync.RWMutex

	id          string
	name        string
	description string
	steps       []*pb.PipelineStep
	envs        map[string]string
	args        map[string]string

	currentStep int
	status      string
	startTime   *timestamppb.Timestamp
	endTime     *timestamppb.Timestamp
	error       string

	containerStatuses map[string]*pb.ContainerStatus
	dockerClient      *client.Client
	eventManager      *EventManager
	nodeID            string
}

// NewPipeline 创建一个新的 Pipeline 实例
func NewPipeline(req *pb.ExecutePipelineRequest, dockerClient *client.Client) *Pipeline {
	now := timestamppb.Now()
	return &Pipeline{
		id:                req.PipelineId,
		name:              req.Pipeline.Name,
		description:       req.Pipeline.Description,
		steps:             req.Pipeline.Steps,
		envs:              req.Envs,
		args:              req.Args,
		status:            "pending",
		startTime:         now,
		containerStatuses: make(map[string]*pb.ContainerStatus),
		dockerClient:      dockerClient,
		eventManager:      NewEventManager(),
		nodeID:            req.NodeId,
	}
}

// Execute 执行 Pipeline
func (p *Pipeline) Execute(ctx context.Context) error {
	p.Lock()
	if p.status == "running" {
		p.Unlock()
		return fmt.Errorf("pipeline 已在运行中")
	}
	p.status = "running"
	p.Unlock()

	defer func() {
		p.Lock()
		if p.status == "running" {
			p.status = "completed"
		}
		p.Unlock()
	}()

	// 创建执行上下文
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 执行每个步骤
	for i, step := range p.steps {
		// 检查上下文是否已取消
		select {
		case <-execCtx.Done():
			return execCtx.Err()
		default:
		}

		// 更新当前步骤
		p.Lock()
		p.currentStep = i
		p.Unlock()

		// 执行步骤
		err := Retry(execCtx, func() error {
			return p.executeStep(execCtx, step)
		}, DefaultRetryConfig)

		if err != nil {
			p.Lock()
			p.status = "failed"
			p.error = err.Error()
			p.endTime = timestamppb.Now()
			p.Unlock()
			return fmt.Errorf("步骤 %d 执行失败: %v", i+1, err)
		}
	}

	p.Lock()
	p.endTime = timestamppb.Now()
	p.Unlock()

	return nil
}

// executeStep 执行单个步骤
func (p *Pipeline) executeStep(ctx context.Context, step *pb.PipelineStep) error {
	// 创建容器配置
	config := &container.Config{
		Image: step.Container.Image,
		Cmd:   step.Container.Command,
		Env:   formatEnvs(step.Container.Environment),
	}

	// 创建主机配置
	hostConfig := &container.HostConfig{
		Binds:       formatVolumes(step.Container.Volumes),
		NetworkMode: container.NetworkMode("bridge"), // 默认使用 bridge 网络
		Resources: container.Resources{
			CPUPeriod: int64(step.Container.Resources.Cpu * 100000), // 将 CPU 核心数转换为 CPU 周期
			CPUQuota:  int64(step.Container.Resources.Cpu * 100000),
			Memory:    step.Container.Resources.Memory,
		},
	}

	// 创建容器
	containerResp, err := p.dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, step.Container.ContainerName)
	if err != nil {
		return WrapError(fmt.Errorf("创建容器失败: %v", err), true)
	}

	// 启动容器
	if err := p.dockerClient.ContainerStart(ctx, containerResp.ID, container.StartOptions{}); err != nil {
		return WrapError(fmt.Errorf("启动容器失败: %v", err), true)
	}

	// 更新容器状态
	p.updateContainerStatus(containerResp.ID, "running", nil)

	// 等待容器完成
	statusCh, errCh := p.dockerClient.ContainerWait(ctx, containerResp.ID, "not-running")
	select {
	case err := <-errCh:
		return WrapError(fmt.Errorf("等待容器失败: %v", err), true)
	case status := <-statusCh:
		if status.StatusCode != 0 {
			// 获取容器日志以帮助诊断问题
			logs, err := p.dockerClient.ContainerLogs(ctx, containerResp.ID, container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
			})
			if err != nil {
				return WrapError(fmt.Errorf("容器执行失败，退出码: %d", status.StatusCode), false)
			}
			defer logs.Close()

			// 处理日志
			if err := p.processLogs(logs); err != nil {
				return WrapError(fmt.Errorf("处理容器日志失败: %v", err), true)
			}

			return WrapError(fmt.Errorf("容器执行失败，退出码: %d", status.StatusCode), false)
		}
	}

	// 更新容器状态为完成
	p.updateContainerStatus(containerResp.ID, "completed", nil)

	return nil
}

// updateContainerStatus 更新容器状态
func (p *Pipeline) updateContainerStatus(containerID string, status string, err error) {
	containerStatus := &pb.ContainerStatus{
		ContainerId: containerID,
		Status:      status,
	}
	if err != nil {
		containerStatus.ExitMessage = err.Error()
		p.eventManager.Publish(NewEvent(EventTypeContainer, p.nodeID, containerID, status, err.Error()))
	} else {
		p.eventManager.Publish(NewEvent(EventTypeContainer, p.nodeID, containerID, status, ""))
	}
}

// processLogs 处理容器日志
func (p *Pipeline) processLogs(logs io.ReadCloser) error {
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Text()
		// 发送日志事件
		p.eventManager.Publish(NewEvent(EventTypeContainer, p.nodeID, p.id, "log", line))
	}
	return scanner.Err()
}

// GetStatus 获取 Pipeline 状态
func (p *Pipeline) GetStatus() *pb.PipelineStatusResponse {
	return &pb.PipelineStatusResponse{
		ExecutionId: p.id,
		Status:      p.status,
		CurrentStep: int32(p.currentStep),
		TotalSteps:  int32(len(p.steps)),
		Progress:    float32(p.currentStep) / float32(len(p.steps)),
	}
}

// Cancel 取消 Pipeline 执行
func (p *Pipeline) Cancel() error {
	p.Lock()
	defer p.Unlock()

	if p.status != "running" {
		return fmt.Errorf("Pipeline 未在运行状态")
	}

	p.status = "canceled"
	for _, containerID := range p.containerStatuses {
		if err := p.dockerClient.ContainerStop(context.Background(), containerID.ContainerId, container.StopOptions{}); err != nil {
			p.updateContainerStatus(containerID.ContainerId, "failed", err)
			continue
		}
		p.updateContainerStatus(containerID.ContainerId, "stopped", nil)
	}

	return nil
}

// formatEnvs 格式化环境变量
func formatEnvs(envs map[string]string) []string {
	result := make([]string, 0, len(envs))
	for k, v := range envs {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// formatVolumes 格式化卷映射
func formatVolumes(volumes []*pb.VolumeMapping) []string {
	var result []string
	for _, vol := range volumes {
		if vol.Readonly {
			result = append(result, fmt.Sprintf("%s:%s:ro", vol.HostPath, vol.ContainerPath))
		} else {
			result = append(result, fmt.Sprintf("%s:%s", vol.HostPath, vol.ContainerPath))
		}
	}
	return result
}
