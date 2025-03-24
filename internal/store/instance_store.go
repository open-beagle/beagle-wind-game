package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// InstanceStore 游戏实例存储
type InstanceStore struct {
	dataFile  string
	instances []models.GameInstance
	mu        sync.RWMutex
}

// NewInstanceStore 创建游戏实例存储
func NewInstanceStore(dataFile string) (*InstanceStore, error) {
	store := &InstanceStore{
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
func (s *InstanceStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 读取数据文件
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		// 如果文件不存在，创建一个空的实例列表
		s.instances = []models.GameInstance{}
		return s.Save()
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
func (s *InstanceStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
func (s *InstanceStore) List() ([]models.GameInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 创建副本避免修改原始数据
	instances := make([]models.GameInstance, len(s.instances))
	copy(instances, s.instances)

	return instances, nil
}

// Get 获取指定ID的实例
func (s *InstanceStore) Get(id string) (models.GameInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, instance := range s.instances {
		if instance.ID == id {
			return instance, nil
		}
	}

	return models.GameInstance{}, fmt.Errorf("实例不存在: %s", id)
}

// Add 添加实例
func (s *InstanceStore) Add(instance models.GameInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
func (s *InstanceStore) Update(instance models.GameInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
func (s *InstanceStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
func (s *InstanceStore) FindByNodeID(nodeID string) ([]models.GameInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.GameInstance
	for _, instance := range s.instances {
		if instance.NodeID == nodeID {
			result = append(result, instance)
		}
	}

	return result, nil
}

// FindByCardID 根据卡片ID查找实例
func (s *InstanceStore) FindByCardID(cardID string) ([]models.GameInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.GameInstance
	for _, instance := range s.instances {
		if instance.CardID == cardID {
			result = append(result, instance)
		}
	}

	return result, nil
}
