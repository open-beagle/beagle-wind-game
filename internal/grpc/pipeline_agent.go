package grpc

import (
	"context"
	"io"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// PipelineAgent 表示 Pipeline Agent
type PipelineAgent struct {
	agent   *Agent
	logger  utils.Logger
	sources map[string]*models.GamePipeline
	mu      sync.RWMutex
}

// NewPipelineAgent 创建一个新的 Pipeline Agent
func NewPipelineAgent(agent *Agent) *PipelineAgent {
	return &PipelineAgent{
		agent:   agent,
		logger:  utils.New("PipelineAgent"),
		sources: make(map[string]*models.GamePipeline),
	}
}

// GetSourceCount 获取当前 Pipeline 数量
func (a *PipelineAgent) GetSourceCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.sources)
}

// Start 启动 Pipeline Agent
func (a *PipelineAgent) Start(ctx context.Context) error {
	a.logger.Info("启动 Pipeline Agent...")

	// 启动 Pipeline 流
	go a.runPipelineStream(ctx)

	return nil
}

// Stop 停止 Pipeline Agent
func (a *PipelineAgent) Stop(ctx context.Context) {
	a.logger.Info("停止 Pipeline Agent...")
}

// runPipelineStream 运行 Pipeline 流
func (a *PipelineAgent) runPipelineStream(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := a.handlePipelineStream(ctx); err != nil {
				a.logger.Error("Pipeline 流处理失败: %v", err)
				time.Sleep(5 * time.Second) // 重试延迟
				continue
			}
		}
	}
}

// handlePipelineStream 处理 Pipeline 流
func (a *PipelineAgent) handlePipelineStream(ctx context.Context) error {
	client := a.agent.GetPipelineClient()

	// 创建流
	stream, err := client.PipelineStream(ctx)
	if err != nil {
		return err
	}

	// 发送初始心跳
	req := &proto.PipelineStreamRequest{
		Heartbeat: &proto.Heartbeat{
			NodeId:    a.agent.id,
			Timestamp: timestamppb.Now(),
		},
	}
	if err := stream.Send(req); err != nil {
		return err
	}

	// 处理响应
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// 处理不同类型的响应
		switch {
		case resp.GetHeartbeatAck() != nil:
			// 处理心跳确认
			a.logger.Debug("收到心跳确认")
		case resp.GetPipeline() != nil:
			// 处理 Pipeline 任务
			if err := a.handlePipeline(ctx, resp.GetPipeline()); err != nil {
				a.logger.Error("处理 Pipeline 任务失败: %v", err)
			}
		case resp.GetCancel() != nil:
			// 处理取消命令
			a.logger.Info("收到取消命令: %s", resp.GetCancel().Reason)
		}
	}
}

// convertProtoToModelPipeline 将 proto.GamePipeline 转换为 models.GamePipeline
func (a *PipelineAgent) convertProtoToModelPipeline(pipeline *proto.GamePipeline) *models.GamePipeline {
	// 初始化步骤状态
	now := time.Now()

	modelPipeline := &models.GamePipeline{
		ID:          pipeline.Id,
		Model:       models.PipelineModel(pipeline.Model),
		Name:        pipeline.Name,
		Description: pipeline.Description,
		Envs:        pipeline.Envs,
		Args:        pipeline.Args,
		Steps:       make([]models.PipelineStep, len(pipeline.Steps)),
		Status: &models.PipelineStatus{
			NodeID:      a.agent.id,
			State:       models.PipelineState(models.PipelineStatePending),
			CurrentStep: 0,
			TotalSteps:  int32(len(pipeline.Steps)),
			Steps:       make([]models.StepStatus, len(pipeline.Steps)),
			UpdatedAt:   &now,
		},
	}

	// 转换步骤
	for i, step := range pipeline.Steps {
		modelPipeline.Steps[i] = models.PipelineStep{
			Name: step.Name,
			Type: step.Type,
			Container: models.ContainerConfig{
				Image:         step.Container.Image,
				ContainerName: step.Container.ContainerName,
				Hostname:      step.Container.Hostname,
				Privileged:    step.Container.Privileged,
				Deploy: models.DeployConfig{
					Resources: models.ResourcesConfig{
						Reservations: models.ReservationsConfig{
							Devices: make([]models.DeviceConfig, len(step.Container.Deploy.Resources.Reservations.Devices)),
						},
					},
				},
				SecurityOpt: step.Container.SecurityOpt,
				CapAdd:      step.Container.CapAdd,
				Tmpfs:       step.Container.Tmpfs,
				Devices:     step.Container.Devices,
				Volumes:     step.Container.Volumes,
				Ports:       step.Container.Ports,
				Environment: step.Container.Environment,
				Command:     step.Container.Command,
			},
		}

		// 转换设备配置
		for j, device := range step.Container.Deploy.Resources.Reservations.Devices {
			modelPipeline.Steps[i].Container.Deploy.Resources.Reservations.Devices[j] = models.DeviceConfig{
				Capabilities: device.Capabilities,
			}
		}

		// 初始化步骤状态
		modelPipeline.Status.Steps[i] = models.StepStatus{
			ID:    step.Name,
			Name:  step.Name,
			State: models.StepStatePending,
		}
	}

	return modelPipeline
}

// handlePipeline 处理 Pipeline 任务
func (a *PipelineAgent) handlePipeline(ctx context.Context, pipeline *proto.GamePipeline) error {
	a.logger.Info("处理 Pipeline 任务: %s", pipeline.Id)

	// 将 proto.GamePipeline 转换为 models.GamePipeline
	modelPipeline := a.convertProtoToModelPipeline(pipeline)

	// 存储到 sources
	a.mu.Lock()
	a.sources[modelPipeline.ID] = modelPipeline
	a.mu.Unlock()

	// 获取 Pipeline 服务客户端
	client := a.agent.GetPipelineClient()

	// 更新 Pipeline 状态
	now := timestamppb.Now()
	status := &proto.PipelineStatus{
		NodeId:      a.agent.id,
		State:       proto.PipelineState_PIPELINE_STATE_RUNNING,
		CurrentStep: 0,
		TotalSteps:  int32(len(pipeline.Steps)),
		StartTime:   now,
		UpdatedAt:   now,
	}

	// 发送状态更新
	_, err := client.UpdatePipelineStatus(ctx, &proto.UpdatePipelineStatusRequest{
		PipelineId: pipeline.Id,
		Status:     status,
	})
	if err != nil {
		return err
	}

	// TODO: 实现 Pipeline 执行逻辑

	return nil
}

// UpdatePipelineStatus 更新 Pipeline 状态
func (a *PipelineAgent) UpdatePipelineStatus(ctx context.Context, pipelineId string, status *proto.PipelineStatus) error {
	client := a.agent.GetPipelineClient()
	_, err := client.UpdatePipelineStatus(ctx, &proto.UpdatePipelineStatusRequest{
		PipelineId: pipelineId,
		Status:     status,
	})
	return err
}

// UpdateStepStatus 更新步骤状态
func (a *PipelineAgent) UpdateStepStatus(ctx context.Context, pipelineId string, stepId string, status *proto.StepStatus) error {
	client := a.agent.GetPipelineClient()
	_, err := client.UpdateStepStatus(ctx, &proto.UpdateStepStatusRequest{
		PipelineId: pipelineId,
		StepId:     stepId,
		Status:     status,
	})
	return err
}
