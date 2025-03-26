package gamenode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/docker/docker/client"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// DockerClientInterface 定义 Docker 客户端接口
type DockerClientInterface interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
}

// GamePipeline 代表一个正在执行的流水线
type GamePipeline struct {
	sync.RWMutex

	id          string
	name        string
	description string
	steps       []*pb.GamePipelineStep
	envs        map[string]string
	args        map[string]string

	currentStep int
	status      string
	startTime   *timestamppb.Timestamp
	endTime     *timestamppb.Timestamp
	error       string

	containerStatuses map[string]*pb.ContainerStatus
	eventManager      *EventManager
	nodeID            string
	dockerClient      *client.Client
}

// NewGamePipeline 创建一个新的 GamePipeline 实例
func NewGamePipeline(req *pb.ExecuteGamePipelineRequest, dockerClient *client.Client) *GamePipeline {
	if dockerClient == nil {
		return nil
	}

	now := timestamppb.Now()
	return &GamePipeline{
		id:                req.GamePipelineId,
		name:              req.GamePipeline.Name,
		description:       req.GamePipeline.Description,
		steps:             req.GamePipeline.Steps,
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

// Execute 执行 GamePipeline
func (p *GamePipeline) Execute(ctx context.Context) error {
	p.Lock()
	if p.status == "running" {
		p.Unlock()
		return fmt.Errorf("GamePipeline 已在运行中")
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
func (p *GamePipeline) executeStep(ctx context.Context, step *pb.GamePipelineStep) error {
	// 检查步骤类型
	if step.Type == "docker" && step.Container != nil {
		return fmt.Errorf("Docker 步骤需要 Docker 客户端支持")
	}

	// 更新容器状态为完成
	p.updateContainerStatus("", "completed", nil)

	return nil
}

// updateContainerStatus 更新容器状态
func (p *GamePipeline) updateContainerStatus(containerID string, status string, err error) {
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
func (p *GamePipeline) processLogs(logs io.ReadCloser) error {
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Text()
		// 发送日志事件
		p.eventManager.Publish(NewEvent(EventTypeContainer, p.nodeID, p.id, "log", line))
	}
	return scanner.Err()
}

// GetStatus 获取 GamePipeline 状态
func (p *GamePipeline) GetStatus() *pb.GamePipelineStatusResponse {
	return &pb.GamePipelineStatusResponse{
		ExecutionId: p.id,
		Status:      p.status,
		CurrentStep: int32(p.currentStep),
		TotalSteps:  int32(len(p.steps)),
		Progress:    float32(p.currentStep) / float32(len(p.steps)),
	}
}

// Cancel 取消 GamePipeline 执行
func (p *GamePipeline) Cancel() error {
	p.Lock()
	defer p.Unlock()

	if p.status != "running" {
		return fmt.Errorf("GamePipeline 未在运行状态")
	}

	p.status = "canceled"
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
