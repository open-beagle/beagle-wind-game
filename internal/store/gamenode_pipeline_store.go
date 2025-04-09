package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
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
	Save(ctx context.Context) error
	// Load 从文件加载所有流水线
	Load(ctx context.Context) error
	// Cleanup 清理资源
	Cleanup(ctx context.Context) error
	// Close 关闭存储
	Close()
}

// YAMLGameNodePipelineStore 基于YAML文件的流水线存储实现
type YAMLGameNodePipelineStore struct {
	filepath  string
	pipelines map[string]*models.GameNodePipeline
	mu        sync.RWMutex
	logger    utils.Logger
	yamlSaver *utils.YAMLSaver
}

// NewYAMLGameNodePipelineStore 创建新的YAML流水线存储
func NewYAMLGameNodePipelineStore(ctx context.Context, filepath string) *YAMLGameNodePipelineStore {
	logger := utils.New("GameNodePipelineStore")

	store := &YAMLGameNodePipelineStore{
		filepath:  filepath,
		pipelines: make(map[string]*models.GameNodePipeline),
		logger:    logger,
	}

	// 创建YAML保存器，使用1秒的延迟保存
	store.yamlSaver = utils.NewYAMLSaver(
		filepath,
		func() interface{} {
			// 这个函数返回当前的流水线数据，在保存时调用
			store.mu.RLock()
			defer store.mu.RUnlock()
			return store.pipelines
		},
		logger,
		utils.WithDelay(time.Second),
	)

	// 初始化加载数据
	logger.Info("初始化流水线存储，数据文件: %s", filepath)
	if err := store.Load(ctx); err != nil {
		logger.Error("加载流水线数据失败: %v", err)
	}

	logger.Info("成功加载流水线数据，共%d个流水线", len(store.pipelines))
	return store
}

// Get 获取指定ID的流水线
func (s *YAMLGameNodePipelineStore) Get(ctx context.Context, id string) (*models.GameNodePipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipeline, exists := s.pipelines[id]
	if !exists {
		s.logger.Error("流水线不存在: %s", id)
		return nil, fmt.Errorf("pipeline not found: %s", id)
	}

	s.logger.Debug("获取流水线: %s", id)
	return pipeline, nil
}

// List 获取所有流水线
func (s *YAMLGameNodePipelineStore) List(ctx context.Context) ([]*models.GameNodePipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipelines := make([]*models.GameNodePipeline, 0, len(s.pipelines))
	for _, pipeline := range s.pipelines {
		pipelines = append(pipelines, pipeline)
	}

	s.logger.Debug("返回%d个流水线", len(pipelines))
	return pipelines, nil
}

// Add 添加新的流水线
func (s *YAMLGameNodePipelineStore) Add(ctx context.Context, pipeline *models.GameNodePipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pipeline.Status == nil {
		s.logger.Error("流水线状态为空")
		return fmt.Errorf("pipeline status is nil")
	}

	if _, exists := s.pipelines[pipeline.Status.ID]; exists {
		s.logger.Error("流水线已存在: %s", pipeline.Status.ID)
		return fmt.Errorf("pipeline already exists: %s", pipeline.Status.ID)
	}

	s.pipelines[pipeline.Status.ID] = pipeline
	s.logger.Info("添加流水线: %s", pipeline.Status.ID)
	return s.Save(ctx)
}

// Update 更新流水线
func (s *YAMLGameNodePipelineStore) Update(ctx context.Context, pipeline *models.GameNodePipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pipeline.Status == nil {
		s.logger.Error("流水线状态为空")
		return fmt.Errorf("pipeline status is nil")
	}

	if _, exists := s.pipelines[pipeline.Status.ID]; !exists {
		s.logger.Error("流水线不存在: %s", pipeline.Status.ID)
		return fmt.Errorf("pipeline not found: %s", pipeline.Status.ID)
	}

	s.pipelines[pipeline.Status.ID] = pipeline
	s.logger.Info("更新流水线: %s", pipeline.Status.ID)
	return s.Save(ctx)
}

// Delete 删除流水线
func (s *YAMLGameNodePipelineStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelines[id]; !exists {
		s.logger.Error("流水线不存在: %s", id)
		return fmt.Errorf("pipeline not found: %s", id)
	}

	delete(s.pipelines, id)
	s.logger.Info("删除流水线: %s", id)
	return s.Save(ctx)
}

// Save 保存所有流水线到文件
func (s *YAMLGameNodePipelineStore) Save(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	// 使用延迟保存器进行保存
	return s.yamlSaver.Save(ctx)
}

// Load 从文件加载所有流水线
func (s *YAMLGameNodePipelineStore) Load(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(s.filepath); os.IsNotExist(err) {
		s.logger.Info("数据文件不存在，使用空数据: %s", s.filepath)
		return nil
	}

	// 读取文件
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		s.logger.Error("读取文件失败: %v", err)
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 解析JSON
	if err := json.Unmarshal(data, &s.pipelines); err != nil {
		s.logger.Error("解析流水线数据失败: %v", err)
		return fmt.Errorf("failed to unmarshal pipelines: %w", err)
	}

	s.logger.Debug("成功从文件加载%d个流水线", len(s.pipelines))
	return nil
}

// Cleanup 清理资源
func (s *YAMLGameNodePipelineStore) Cleanup(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空流水线数据
	s.pipelines = make(map[string]*models.GameNodePipeline)

	// 尝试删除文件
	if err := os.Remove(s.filepath); err != nil && !os.IsNotExist(err) {
		s.logger.Error("删除文件失败: %v", err)
		return fmt.Errorf("failed to remove file: %w", err)
	}

	s.logger.Info("已清理流水线数据和文件")
	return nil
}

// Close 关闭存储，确保所有待处理的保存操作完成
func (s *YAMLGameNodePipelineStore) Close() {
	s.logger.Info("关闭GameNodePipelineStore，确保数据保存...")
	if s.yamlSaver != nil {
		s.yamlSaver.Close()
	}
}
