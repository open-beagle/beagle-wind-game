package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
	"github.com/open-beagle/beagle-wind-game/internal/types"
	"gopkg.in/yaml.v3"
)

// GameNodePipelineStore Pipeline 存储接口
type GameNodePipelineStore interface {
	Load() error
	Save() error
	List(params types.PipelineListParams) (*types.PipelineListResult, error)
	Get(id string) (*gamenode.GameNodePipeline, error)
	Add(pipeline *gamenode.GameNodePipeline) error
	Update(pipeline *gamenode.GameNodePipeline) error
	Delete(id string, force bool) error
	UpdateStatus(id string, status string) error
	Cleanup() error
}

// YAMLGameNodePipelineStore YAML文件存储实现
type YAMLGameNodePipelineStore struct {
	dataFile  string
	pipelines map[string]*gamenode.GameNodePipeline
	mu        sync.RWMutex
}

// NewGameNodePipelineStore 创建新的 Pipeline 存储实例
func NewGameNodePipelineStore(dataFile string) (GameNodePipelineStore, error) {
	store := &YAMLGameNodePipelineStore{
		dataFile:  dataFile,
		pipelines: make(map[string]*gamenode.GameNodePipeline),
	}

	// 初始化加载数据
	err := store.Load()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Load 从文件加载 Pipeline 数据
func (s *YAMLGameNodePipelineStore) Load() error {
	// 读取文件内容
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read pipeline data file: %v", err)
	}

	// 解析 YAML 数据
	var pipelines map[string]*gamenode.GameNodePipeline
	if err := yaml.Unmarshal(data, &pipelines); err != nil {
		return fmt.Errorf("failed to unmarshal pipeline data: %v", err)
	}

	s.pipelines = pipelines
	return nil
}

// Save 保存 Pipeline 数据到文件
func (s *YAMLGameNodePipelineStore) Save() error {
	// 序列化数据
	data, err := yaml.Marshal(s.pipelines)
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline data: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(s.dataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write pipeline data file: %v", err)
	}

	return nil
}

// List 获取流水线列表
func (s *YAMLGameNodePipelineStore) List(params types.PipelineListParams) (*types.PipelineListResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: 实现过滤和排序
	result := &types.PipelineListResult{
		Total: int64(len(s.pipelines)),
		Items: make([]*gamenode.GameNodePipeline, 0, len(s.pipelines)),
	}

	for _, pipeline := range s.pipelines {
		result.Items = append(result.Items, pipeline)
	}

	return result, nil
}

// Get 获取流水线详情
func (s *YAMLGameNodePipelineStore) Get(id string) (*gamenode.GameNodePipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipeline, exists := s.pipelines[id]
	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", id)
	}

	return pipeline, nil
}

// Add 添加新的流水线
func (s *YAMLGameNodePipelineStore) Add(pipeline *gamenode.GameNodePipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelines[pipeline.GetName()]; exists {
		return fmt.Errorf("pipeline already exists: %s", pipeline.GetName())
	}

	s.pipelines[pipeline.GetName()] = pipeline
	return s.Save()
}

// Update 更新流水线
func (s *YAMLGameNodePipelineStore) Update(pipeline *gamenode.GameNodePipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelines[pipeline.GetName()]; !exists {
		return fmt.Errorf("pipeline not found: %s", pipeline.GetName())
	}

	s.pipelines[pipeline.GetName()] = pipeline
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

	// 更新状态
	pipeline.UpdateStatus(gamenode.PipelineState(status))
	return s.Save()
}

// Delete 删除流水线
func (s *YAMLGameNodePipelineStore) Delete(id string, force bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pipeline, exists := s.pipelines[id]
	if !exists {
		return fmt.Errorf("pipeline not found: %s", id)
	}

	// 检查是否可以删除
	if !force && pipeline.GetStatus().State != gamenode.PipelineStateCompleted {
		return fmt.Errorf("cannot delete pipeline in state: %s", pipeline.GetStatus().State)
	}

	delete(s.pipelines, id)
	return s.Save()
}

// Cleanup 清理过期的流水线数据
func (s *YAMLGameNodePipelineStore) Cleanup() error {
	return os.Remove(s.dataFile)
}
