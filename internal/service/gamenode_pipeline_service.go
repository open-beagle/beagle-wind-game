package service

import (
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
	"github.com/open-beagle/beagle-wind-game/internal/types"
)

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
func (s *GameNodePipelineService) GetPipeline(id string) (*gamenode.GameNodePipeline, error) {
	return s.store.Get(id)
}

// CancelPipeline 取消流水线
func (s *GameNodePipelineService) CancelPipeline(id string, reason string) error {
	return s.store.UpdateStatus(id, "canceled")
}

// DeletePipeline 删除流水线
func (s *GameNodePipelineService) DeletePipeline(id string, force bool) error {
	return s.store.Delete(id, force)
}

// PipelineStore 流水线存储接口
type PipelineStore interface {
	// List 获取流水线列表
	List(params types.PipelineListParams) (*types.PipelineListResult, error)
	// Get 获取流水线详情
	Get(id string) (*gamenode.GameNodePipeline, error)
	// UpdateStatus 更新流水线状态
	UpdateStatus(id string, status string) error
	// Delete 删除流水线
	Delete(id string, force bool) error
}
