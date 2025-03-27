package service

import (
	"context"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/types"
)

// PipelineStore Pipeline 存储接口
type PipelineStore interface {
	List(params types.PipelineListParams) (*types.PipelineListResult, error)
	Get(id string) (*models.GameNodePipeline, error)
	Add(pipeline *models.GameNodePipeline) error
	Update(pipeline *models.GameNodePipeline) error
	UpdateStatus(id string, status string) error
	Delete(id string, force bool) error
	Cleanup() error
}

// GameNodePipelineService 游戏节点流水线服务
type GameNodePipelineService struct {
	store PipelineStore
}

// NewGameNodePipelineService 创建新的游戏节点流水线服务
func NewGameNodePipelineService(store PipelineStore) *GameNodePipelineService {
	return &GameNodePipelineService{
		store: store,
	}
}

// ListPipelines 获取流水线列表
func (s *GameNodePipelineService) ListPipelines(params types.PipelineListParams) (*types.PipelineListResult, error) {
	return s.store.List(params)
}

// GetPipeline 获取流水线详情
func (s *GameNodePipelineService) GetPipeline(id string) (*models.GameNodePipeline, error) {
	return s.store.Get(id)
}

// CreatePipeline 创建流水线
func (s *GameNodePipelineService) CreatePipeline(ctx context.Context, pipeline *proto.ExecutePipelineRequest) error {
	// 创建流水线
	p := &models.GameNodePipeline{
		ID: pipeline.Id,
		Status: &models.PipelineStatus{
			ID:     pipeline.Id,
			State:  models.PipelineStatePending,
			NodeID: pipeline.PipelineId,
		},
	}
	return s.store.Add(p)
}

// ExecutePipeline 执行流水线
func (s *GameNodePipelineService) ExecutePipeline(ctx context.Context, id string) error {
	// 获取流水线
	pipeline, err := s.store.Get(id)
	if err != nil {
		return err
	}

	// 更新状态
	pipeline.Status.State = models.PipelineStateRunning
	return s.store.Update(pipeline)
}

// UpdateStatus 更新流水线状态
func (s *GameNodePipelineService) UpdateStatus(ctx context.Context, id string, status models.PipelineState) error {
	return s.store.UpdateStatus(id, string(status))
}

// UpdateStepStatus 更新步骤状态
func (s *GameNodePipelineService) UpdateStepStatus(ctx context.Context, pipelineID string, stepID string, status models.StepState) error {
	// 获取流水线
	pipeline, err := s.store.Get(pipelineID)
	if err != nil {
		return err
	}

	// 更新步骤状态
	for i := range pipeline.Status.Steps {
		if pipeline.Status.Steps[i].ID == stepID {
			pipeline.Status.Steps[i].State = status
			break
		}
	}

	return s.store.Update(pipeline)
}

// SaveStepLogs 保存步骤日志
func (s *GameNodePipelineService) SaveStepLogs(ctx context.Context, pipelineID string, stepID string, logs []byte) error {
	// 获取流水线
	pipeline, err := s.store.Get(pipelineID)
	if err != nil {
		return err
	}

	// 更新步骤日志
	for i := range pipeline.Status.Steps {
		if pipeline.Status.Steps[i].ID == stepID {
			pipeline.Status.Steps[i].Logs = logs
			break
		}
	}

	return s.store.Update(pipeline)
}

// CancelPipeline 取消流水线
func (s *GameNodePipelineService) CancelPipeline(ctx context.Context, id string) error {
	return s.store.UpdateStatus(id, string(models.PipelineStateCanceled))
}

// DeletePipeline 删除流水线
func (s *GameNodePipelineService) DeletePipeline(id string, force bool) error {
	return s.store.Delete(id, force)
}
