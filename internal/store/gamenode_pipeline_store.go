package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/types"
)

// GameNodePipelineStore 游戏节点流水线存储接口
type GameNodePipelineStore interface {
	// Get 获取指定ID的流水线
	Get(ctx context.Context, id string) (*models.GameNodePipeline, error)
	// List 获取所有流水线
	List(ctx context.Context) ([]*models.GameNodePipeline, error)
	// Add 添加新的流水线
	Add(ctx context.Context, pipeline *models.GameNodePipeline) error
	// Update 更新流水线
	Update(ctx context.Context, pipeline *models.GameNodePipeline) error
	// Delete 删除流水线
	Delete(ctx context.Context, id string) error
	// Save 保存所有流水线到文件
	Save() error
	// Load 从文件加载所有流水线
	Load() error
	// Cleanup 清理资源
	Cleanup() error
}

// YAMLGameNodePipelineStore 基于YAML文件的流水线存储实现
type YAMLGameNodePipelineStore struct {
	filepath  string
	pipelines map[string]*models.GameNodePipeline
	mu        sync.RWMutex
}

// NewYAMLGameNodePipelineStore 创建新的YAML流水线存储
func NewYAMLGameNodePipelineStore(filepath string) *YAMLGameNodePipelineStore {
	return &YAMLGameNodePipelineStore{
		filepath:  filepath,
		pipelines: make(map[string]*models.GameNodePipeline),
	}
}

// Get 获取指定ID的流水线
func (s *YAMLGameNodePipelineStore) Get(id string) (*models.GameNodePipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipeline, exists := s.pipelines[id]
	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", id)
	}

	return pipeline, nil
}

// List 获取所有流水线
func (s *YAMLGameNodePipelineStore) List(params types.PipelineListParams) (*types.PipelineListResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipelines := make([]*models.GameNodePipeline, 0, len(s.pipelines))
	for _, pipeline := range s.pipelines {
		pipelines = append(pipelines, pipeline)
	}

	return &types.PipelineListResult{
		Total: int64(len(pipelines)),
		Items: pipelines,
	}, nil
}

// Add 添加新的流水线
func (s *YAMLGameNodePipelineStore) Add(pipeline *models.GameNodePipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pipeline.Status == nil {
		return fmt.Errorf("pipeline status is nil")
	}

	if _, exists := s.pipelines[pipeline.Status.ID]; exists {
		return fmt.Errorf("pipeline already exists: %s", pipeline.Status.ID)
	}

	s.pipelines[pipeline.Status.ID] = pipeline
	return s.Save()
}

// Update 更新流水线
func (s *YAMLGameNodePipelineStore) Update(pipeline *models.GameNodePipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pipeline.Status == nil {
		return fmt.Errorf("pipeline status is nil")
	}

	if _, exists := s.pipelines[pipeline.Status.ID]; !exists {
		return fmt.Errorf("pipeline not found: %s", pipeline.Status.ID)
	}

	s.pipelines[pipeline.Status.ID] = pipeline
	return s.Save()
}

// UpdateStatus 更新流水线状态
func (s *YAMLGameNodePipelineStore) UpdateStatus(id string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pipeline, exists := s.pipelines[id]
	if !exists {
		return fmt.Errorf("pipeline not found: %s", id)
	}

	pipeline.Status.State = models.PipelineState(status)
	return s.Save()
}

// Delete 删除流水线
func (s *YAMLGameNodePipelineStore) Delete(id string, force bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelines[id]; !exists {
		return fmt.Errorf("pipeline not found: %s", id)
	}

	delete(s.pipelines, id)
	return s.Save()
}

// Save 保存所有流水线到文件
func (s *YAMLGameNodePipelineStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(s.filepath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 将流水线转换为JSON
	data, err := json.MarshalIndent(s.pipelines, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pipelines: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(s.filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Load 从文件加载所有流水线
func (s *YAMLGameNodePipelineStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(s.filepath); os.IsNotExist(err) {
		return nil
	}

	// 读取文件
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 解析JSON
	if err := json.Unmarshal(data, &s.pipelines); err != nil {
		return fmt.Errorf("failed to unmarshal pipelines: %w", err)
	}

	return nil
}

// Cleanup 清理资源
func (s *YAMLGameNodePipelineStore) Cleanup() error {
	return nil
}
