package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// GameInstanceStore 游戏实例存储接口
type GameInstanceStore interface {
	// List 获取实例列表
	List(ctx context.Context) ([]models.GameInstance, error)
	// Get 获取实例详情
	Get(ctx context.Context, id string) (models.GameInstance, error)
	// Add 添加实例
	Add(ctx context.Context, instance models.GameInstance) error
	// Update 更新实例
	Update(ctx context.Context, instance models.GameInstance) error
	// Delete 删除实例
	Delete(ctx context.Context, id string) error
	// FindByNodeID 根据节点ID查找实例
	FindByNodeID(ctx context.Context, nodeID string) ([]models.GameInstance, error)
	// FindByCardID 根据卡片ID查找实例
	FindByCardID(ctx context.Context, cardID string) ([]models.GameInstance, error)
	// Load 从存储中加载实例
	Load(ctx context.Context) error
	// Save 保存实例到存储
	Save(ctx context.Context) error
	// Cleanup 清理存储文件
	Cleanup(ctx context.Context) error
	// Close 关闭存储
	Close()
}

// YAMLGameInstanceStore YAML游戏实例存储实现
type YAMLGameInstanceStore struct {
	dataFile  string
	instances []models.GameInstance
	mu        sync.RWMutex
	logger    utils.Logger
	yamlSaver *utils.YAMLSaver
	ctx       context.Context    // 存储的独立上下文
	cancel    context.CancelFunc // 用于取消存储的上下文
}

// NewGameInstanceStore 创建游戏实例存储
func NewGameInstanceStore(ctx context.Context, dataFile string, logger utils.Logger) (GameInstanceStore, error) {
	// 创建存储的独立上下文
	storeCtx, cancel := context.WithCancel(context.Background())

	// 创建存储
	store := &YAMLGameInstanceStore{
		dataFile:  dataFile,
		instances: []models.GameInstance{},
		logger:    logger,
		ctx:       storeCtx,
		cancel:    cancel,
	}

	// 创建YAML保存器，使用1秒的延迟保存
	store.yamlSaver = utils.NewYAMLSaver(
		dataFile,
		func() interface{} {
			// 这个函数返回当前的实例数据，在保存时调用
			store.mu.RLock()
			defer store.mu.RUnlock()
			return store.instances
		},
		logger,
		utils.WithDelay(time.Second),
	)

	// 确保目录存在
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("创建数据目录失败: %v", err)
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 加载数据
	err := store.Load(ctx)
	if err != nil {
		logger.Warn("加载游戏实例数据失败: %v", err)
	}

	logger.Info("初始化游戏实例存储成功，数据文件: %s", dataFile)
	return store, nil
}

// Load 加载数据
func (s *YAMLGameInstanceStore) Load(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件是否存在
	_, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("数据文件不存在，初始化空数据: %s", s.dataFile)
			s.instances = []models.GameInstance{}
			return nil
		}
		s.logger.Error("检查数据文件失败: %v", err)
		return fmt.Errorf("检查数据文件失败: %w", err)
	}

	// 读取文件
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		s.logger.Error("读取数据文件失败: %v", err)
		return fmt.Errorf("读取数据文件失败: %w", err)
	}

	// 解析YAML
	var instances []models.GameInstance
	err = yaml.Unmarshal(data, &instances)
	if err != nil {
		// 尝试解析JSON
		jsonErr := json.Unmarshal(data, &instances)
		if jsonErr != nil {
			s.logger.Error("解析YAML/JSON数据失败: %v / %v", err, jsonErr)
			return fmt.Errorf("解析数据失败: %w", err)
		}
	}

	s.instances = instances
	s.logger.Info("加载了 %d 个游戏实例", len(instances))
	return nil
}

// Save 保存数据
func (s *YAMLGameInstanceStore) Save(ctx context.Context) error {
	// 使用存储的独立上下文进行保存
	return s.yamlSaver.Save(s.ctx)
}

// List 列出所有实例
func (s *YAMLGameInstanceStore) List(ctx context.Context) ([]models.GameInstance, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("数据文件不存在，返回空列表: %s", s.dataFile)
			return []models.GameInstance{}, nil
		}
		s.logger.Error("读取文件失败: %v", err)
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("读取文件失败: 目标是一个目录: %s", s.dataFile)
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	// 复制实例列表
	instances := make([]models.GameInstance, len(s.instances))
	copy(instances, s.instances)

	s.logger.Debug("列出 %d 个游戏实例", len(instances))
	return instances, nil
}

// Get 获取指定实例
func (s *YAMLGameInstanceStore) Get(ctx context.Context, id string) (models.GameInstance, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return models.GameInstance{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Error("数据文件不存在: %s", s.dataFile)
			return models.GameInstance{}, fmt.Errorf("数据文件不存在")
		}
		s.logger.Error("读取文件失败: %v", err)
		return models.GameInstance{}, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("读取文件失败: 目标是一个目录: %s", s.dataFile)
		return models.GameInstance{}, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	for _, instance := range s.instances {
		if instance.ID == id {
			s.logger.Debug("找到游戏实例: %s", id)
			return instance, nil
		}
	}
	s.logger.Error("实例不存在: %s", id)
	return models.GameInstance{}, fmt.Errorf("实例不存在: %s", id)
}

// Add 添加实例
func (s *YAMLGameInstanceStore) Add(ctx context.Context, instance models.GameInstance) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err == nil && fileInfo.IsDir() {
		s.logger.Error("写入文件失败: 目标是一个目录: %s", s.dataFile)
		return fmt.Errorf("写入文件失败: 目标是一个目录")
	}

	// 检查ID是否已存在
	for _, existing := range s.instances {
		if existing.ID == instance.ID {
			s.logger.Error("实例ID已存在: %s", instance.ID)
			return fmt.Errorf("实例ID已存在: %s", instance.ID)
		}
	}

	// 添加实例
	s.instances = append(s.instances, instance)
	s.logger.Info("添加新实例: %s", instance.ID)

	// 保存更改
	if err := s.yamlSaver.Save(ctx); err != nil {
		s.logger.Error("保存实例数据失败: %v", err)
		return fmt.Errorf("保存实例数据失败: %w", err)
	}

	s.logger.Info("添加实例成功: %s - %s", instance.ID, instance.NodeID)
	return nil
}

// Update 更新实例
func (s *YAMLGameInstanceStore) Update(ctx context.Context, instance models.GameInstance) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err == nil && fileInfo.IsDir() {
		s.logger.Error("写入文件失败: 目标是一个目录: %s", s.dataFile)
		return fmt.Errorf("写入文件失败: 目标是一个目录")
	}

	// 查找实例
	found := false
	for i, existing := range s.instances {
		if existing.ID == instance.ID {
			// 更新实例
			s.instances[i] = instance
			found = true
			s.logger.Info("更新实例: %s", instance.ID)
			break
		}
	}

	if !found {
		s.logger.Error("实例不存在，无法更新: %s", instance.ID)
		return fmt.Errorf("实例不存在: %s", instance.ID)
	}

	// 保存更改
	if err := s.yamlSaver.Save(ctx); err != nil {
		s.logger.Error("保存实例数据失败: %v", err)
		return fmt.Errorf("保存实例数据失败: %w", err)
	}

	s.logger.Info("更新实例成功: %s - %s", instance.ID, instance.NodeID)
	return nil
}

// Delete 删除实例
func (s *YAMLGameInstanceStore) Delete(ctx context.Context, id string) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err == nil && fileInfo.IsDir() {
		s.logger.Error("写入文件失败: 目标是一个目录: %s", s.dataFile)
		return fmt.Errorf("写入文件失败: 目标是一个目录")
	}

	// 查找实例
	found := false
	for i, instance := range s.instances {
		if instance.ID == id {
			// 删除实例
			s.instances = append(s.instances[:i], s.instances[i+1:]...)
			found = true
			s.logger.Info("删除实例: %s", id)
			break
		}
	}

	if !found {
		s.logger.Error("实例不存在，无法删除: %s", id)
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 保存更改
	if err := s.yamlSaver.Save(ctx); err != nil {
		s.logger.Error("保存实例数据失败: %v", err)
		return fmt.Errorf("保存实例数据失败: %w", err)
	}

	s.logger.Info("删除实例成功: %s", id)
	return nil
}

// FindByNodeID 根据节点ID查找实例
func (s *YAMLGameInstanceStore) FindByNodeID(ctx context.Context, nodeID string) ([]models.GameInstance, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("数据文件不存在，返回空列表: %s", s.dataFile)
			return []models.GameInstance{}, nil
		}
		s.logger.Error("读取文件失败: %v", err)
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("读取文件失败: 目标是一个目录: %s", s.dataFile)
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	// 查找指定节点ID的实例
	var result []models.GameInstance
	for _, instance := range s.instances {
		if instance.NodeID == nodeID {
			result = append(result, instance)
		}
	}

	s.logger.Debug("根据节点ID '%s' 找到 %d 个实例", nodeID, len(result))
	return result, nil
}

// FindByCardID 根据卡片ID查找实例
func (s *YAMLGameInstanceStore) FindByCardID(ctx context.Context, cardID string) ([]models.GameInstance, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("数据文件不存在，返回空列表: %s", s.dataFile)
			return []models.GameInstance{}, nil
		}
		s.logger.Error("读取文件失败: %v", err)
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("读取文件失败: 目标是一个目录: %s", s.dataFile)
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	// 查找指定卡片ID的实例
	var result []models.GameInstance
	for _, instance := range s.instances {
		if instance.CardID == cardID {
			result = append(result, instance)
		}
	}

	s.logger.Debug("根据卡片ID '%s' 找到 %d 个实例", cardID, len(result))
	return result, nil
}

// Cleanup 清理存储文件
func (s *YAMLGameInstanceStore) Cleanup(ctx context.Context) error {
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

// Close 关闭存储
func (s *YAMLGameInstanceStore) Close() {
	s.logger.Info("关闭GameInstanceStore，确保数据保存...")
	if s.yamlSaver != nil {
		s.yamlSaver.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}
}
