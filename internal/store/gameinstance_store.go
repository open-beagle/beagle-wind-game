package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// GameInstanceStore 游戏实例存储接口
type GameInstanceStore interface {
	// List 获取所有实例
	List() ([]models.GameInstance, error)
	// Get 获取指定ID的实例
	Get(id string) (models.GameInstance, error)
	// Add 添加实例
	Add(instance models.GameInstance) error
	// Update 更新实例信息
	Update(instance models.GameInstance) error
	// Delete 删除实例
	Delete(id string) error
	// FindByNodeID 根据节点ID查找实例
	FindByNodeID(nodeID string) ([]models.GameInstance, error)
	// FindByCardID 根据卡片ID查找实例
	FindByCardID(cardID string) ([]models.GameInstance, error)
	// Cleanup 清理测试文件
	Cleanup() error
}

// YAMLGameInstanceStore YAML文件存储实现
type YAMLGameInstanceStore struct {
	dataFile  string
	instances []models.GameInstance
	mu        sync.RWMutex
}

// NewGameInstanceStore 创建游戏实例存储
func NewGameInstanceStore(dataFile string) (GameInstanceStore, error) {
	store := &YAMLGameInstanceStore{
		dataFile: dataFile,
	}

	// 初始化加载数据
	err := store.Load()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Load 加载实例数据
func (s *YAMLGameInstanceStore) Load() error {
	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，创建空文件
			s.instances = make([]models.GameInstance, 0)
			return s.Save()
		}
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		return fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	// 读取数据文件
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return fmt.Errorf("读取实例数据文件失败: %w", err)
	}

	// 解析YAML
	var instances []models.GameInstance
	err = yaml.Unmarshal(data, &instances)
	if err != nil {
		return fmt.Errorf("解析实例数据文件失败: %w", err)
	}

	s.instances = instances
	return nil
}

// Save 保存实例数据
func (s *YAMLGameInstanceStore) Save() error {
	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err == nil && fileInfo.IsDir() {
		return fmt.Errorf("写入文件失败: 目标是一个目录")
	}

	// 序列化为YAML
	data, err := yaml.Marshal(s.instances)
	if err != nil {
		return fmt.Errorf("序列化实例数据失败: %w", err)
	}

	// 写入文件
	err = os.WriteFile(s.dataFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入实例数据文件失败: %w", err)
	}

	return nil
}

// List 获取所有实例
func (s *YAMLGameInstanceStore) List() ([]models.GameInstance, error) {
	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.GameInstance{}, nil
		}
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	// 创建副本避免修改原始数据
	instances := make([]models.GameInstance, len(s.instances))
	copy(instances, s.instances)
	return instances, nil
}

// Get 获取指定ID的实例
func (s *YAMLGameInstanceStore) Get(id string) (models.GameInstance, error) {
	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return models.GameInstance{}, nil
		}
		return models.GameInstance{}, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		return models.GameInstance{}, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	for _, instance := range s.instances {
		if instance.ID == id {
			return instance, nil
		}
	}
	return models.GameInstance{}, fmt.Errorf("实例不存在: %s", id)
}

// Add 添加实例
func (s *YAMLGameInstanceStore) Add(instance models.GameInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err == nil && fileInfo.IsDir() {
		return fmt.Errorf("写入文件失败: 目标是一个目录")
	}

	// 检查ID是否已存在
	for _, i := range s.instances {
		if i.ID == instance.ID {
			return fmt.Errorf("实例ID已存在: %s", instance.ID)
		}
	}

	// 添加实例
	s.instances = append(s.instances, instance)

	// 保存数据
	return s.Save()
}

// Update 更新实例
func (s *YAMLGameInstanceStore) Update(instance models.GameInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err == nil && fileInfo.IsDir() {
		return fmt.Errorf("写入文件失败: 目标是一个目录")
	}

	// 查找并更新实例
	for i, inst := range s.instances {
		if inst.ID == instance.ID {
			s.instances[i] = instance
			return s.Save()
		}
	}

	return fmt.Errorf("实例不存在: %s", instance.ID)
}

// Delete 删除实例
func (s *YAMLGameInstanceStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err == nil && fileInfo.IsDir() {
		return fmt.Errorf("写入文件失败: 目标是一个目录")
	}

	// 查找并删除实例
	for i, instance := range s.instances {
		if instance.ID == id {
			// 从切片中删除实例
			s.instances = append(s.instances[:i], s.instances[i+1:]...)
			return s.Save()
		}
	}

	return fmt.Errorf("实例不存在: %s", id)
}

// FindByNodeID 根据节点ID查找实例
func (s *YAMLGameInstanceStore) FindByNodeID(nodeID string) ([]models.GameInstance, error) {
	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.GameInstance{}, nil
		}
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	var result []models.GameInstance
	for _, instance := range s.instances {
		if instance.NodeID == nodeID {
			result = append(result, instance)
		}
	}
	return result, nil
}

// FindByCardID 根据卡片ID查找实例
func (s *YAMLGameInstanceStore) FindByCardID(cardID string) ([]models.GameInstance, error) {
	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.GameInstance{}, nil
		}
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	var result []models.GameInstance
	for _, instance := range s.instances {
		if instance.CardID == cardID {
			result = append(result, instance)
		}
	}
	return result, nil
}

// Cleanup 清理测试文件
func (s *YAMLGameInstanceStore) Cleanup() error {
	return os.Remove(s.dataFile)
}
