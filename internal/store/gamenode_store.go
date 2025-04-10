package store

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// GameNodeStore 游戏节点存储接口
type GameNodeStore interface {
	// Load 从文件加载数据
	Load(ctx context.Context) error
	// Save 保存数据到文件
	Save(ctx context.Context) error
	// List 获取所有节点
	List(ctx context.Context) ([]models.GameNode, error)
	// Get 获取指定ID的节点
	Get(ctx context.Context, id string) (models.GameNode, error)
	// Add 添加节点
	Add(ctx context.Context, node models.GameNode) error
	// Update 更新节点
	Update(ctx context.Context, node models.GameNode) error
	// Delete 删除节点
	Delete(ctx context.Context, id string) error
	// Cleanup 清理存储文件
	Cleanup(ctx context.Context) error
}

// YAMLGameNodeStore YAML格式的游戏节点存储
type YAMLGameNodeStore struct {
	dataFile  string
	nodes     []models.GameNode
	mu        sync.RWMutex
	logger    utils.Logger
	yamlSaver *utils.YAMLSaver
	ctx       context.Context    // 存储的独立上下文
	cancel    context.CancelFunc // 用于取消存储的上下文
}

// NewGameNodeStore 创建YAML格式的游戏节点存储
func NewGameNodeStore(ctx context.Context, dataFile string) (GameNodeStore, error) {
	logger := utils.New("GameNodeStore")

	// 创建存储的独立上下文
	storeCtx, cancel := context.WithCancel(context.Background())

	store := &YAMLGameNodeStore{
		dataFile: dataFile,
		nodes:    []models.GameNode{},
		logger:   logger,
		ctx:      storeCtx,
		cancel:   cancel,
	}

	// 创建YAML保存器，使用1秒的延迟保存
	store.yamlSaver = utils.NewYAMLSaver(
		dataFile,
		func() interface{} {
			// 这个函数返回当前的节点数据，在保存时调用
			store.mu.RLock()
			defer store.mu.RUnlock()
			return store.nodes
		},
		logger,
		utils.WithDelay(time.Second),
	)

	// 尝试加载现有数据
	err := store.Load(ctx)
	if err != nil {
		logger.Warn("加载节点数据失败: %v，将使用空数据开始", err)
	}

	return store, nil
}

// Load 从文件加载节点数据
func (s *YAMLGameNodeStore) Load(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("从文件加载节点数据: %s", s.dataFile)

	// 检查文件是否存在
	_, err := os.Stat(s.dataFile)
	if os.IsNotExist(err) {
		s.logger.Info("数据文件不存在，将使用空数据: %s", s.dataFile)
		return nil
	}

	// 读取文件内容
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		s.logger.Error("读取文件失败: %v", err)
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 反序列化YAML
	var nodes []models.GameNode
	err = yaml.Unmarshal(data, &nodes)
	if err != nil {
		s.logger.Error("解析YAML失败: %v", err)
		return fmt.Errorf("解析YAML失败: %w", err)
	}

	s.nodes = nodes
	s.logger.Info("成功加载 %d 个节点数据", len(nodes))
	return nil
}

// Save 保存节点数据到文件
func (s *YAMLGameNodeStore) Save(ctx context.Context) error {
	// 使用存储的独立上下文进行保存
	return s.yamlSaver.Save(s.ctx)
}

// List 获取所有节点
func (s *YAMLGameNodeStore) List(ctx context.Context) ([]models.GameNode, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Debug("获取所有节点列表，当前共 %d 个节点", len(s.nodes))
	// 创建副本避免修改原始数据
	nodes := make([]models.GameNode, len(s.nodes))
	copy(nodes, s.nodes)
	return nodes, nil
}

// Get 获取指定ID的节点
func (s *YAMLGameNodeStore) Get(ctx context.Context, id string) (models.GameNode, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return models.GameNode{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Debug("获取节点，ID: %s", id)
	for _, node := range s.nodes {
		if node.ID == id {
			s.logger.Debug("找到节点: %s", id)
			return node, nil
		}
	}
	s.logger.Warn("未找到节点: %s", id)
	return models.GameNode{}, fmt.Errorf("节点不存在: %s", id)
}

// Add 添加节点
func (s *YAMLGameNodeStore) Add(ctx context.Context, node models.GameNode) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("添加节点，ID: %s", node.ID)

	// 检查ID是否已存在
	for _, existing := range s.nodes {
		if existing.ID == node.ID {
			s.logger.Warn("节点已存在，无法添加: %s", node.ID)
			return fmt.Errorf("节点已存在: %s", node.ID)
		}
	}

	s.nodes = append(s.nodes, node)
	err := s.Save(ctx)
	if err != nil {
		s.logger.Error("保存节点失败: %v", err)
		return err
	}

	s.logger.Info("节点添加成功，ID: %s", node.ID)
	return nil
}

// Update 更新节点
func (s *YAMLGameNodeStore) Update(ctx context.Context, node models.GameNode) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("更新节点，ID: %s", node.ID)

	// 查找并更新节点
	for i, existing := range s.nodes {
		if existing.ID == node.ID {
			s.nodes[i] = node
			err := s.Save(ctx)
			if err != nil {
				s.logger.Error("保存节点失败: %v", err)
				return err
			}
			s.logger.Info("节点更新成功，ID: %s", node.ID)
			return nil
		}
	}

	s.logger.Warn("节点不存在，无法更新: %s", node.ID)
	return fmt.Errorf("节点不存在: %s", node.ID)
}

// Delete 删除节点
func (s *YAMLGameNodeStore) Delete(ctx context.Context, id string) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("删除节点，ID: %s", id)

	// 查找并删除节点
	for i, node := range s.nodes {
		if node.ID == id {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			err := s.Save(ctx)
			if err != nil {
				s.logger.Error("保存节点失败: %v", err)
				return err
			}
			s.logger.Info("节点删除成功，ID: %s", id)
			return nil
		}
	}

	s.logger.Warn("节点不存在，无法删除: %s", id)
	return fmt.Errorf("节点不存在: %s", id)
}

// Cleanup 清理存储文件
func (s *YAMLGameNodeStore) Cleanup(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.logger.Info("清理存储文件: %s", s.dataFile)
	err := os.Remove(s.dataFile)
	if err != nil {
		s.logger.Error("清理存储文件失败: %v", err)
		return err
	}
	s.logger.Info("存储文件清理成功")
	return nil
}

// Close 关闭存储，确保所有待处理的保存操作完成
func (s *YAMLGameNodeStore) Close() {
	s.logger.Info("关闭GameNodeStore，确保数据保存...")
	if s.yamlSaver != nil {
		s.yamlSaver.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}
}
