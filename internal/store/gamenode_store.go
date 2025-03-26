package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// GameNodeStore 节点存储接口
type GameNodeStore interface {
	Load() error
	Save() error
	List() ([]models.GameNode, error)
	Get(id string) (models.GameNode, error)
	Add(node models.GameNode) error
	Update(node models.GameNode) error
	Delete(id string) error
	Cleanup() error
}

// YAMLGameNodeStore YAML文件存储实现
type YAMLGameNodeStore struct {
	dataFile string
	nodes    []models.GameNode
	mu       sync.RWMutex
}

// NewGameNodeStore 创建新的节点存储实例
func NewGameNodeStore(dataFile string) (GameNodeStore, error) {
	store := &YAMLGameNodeStore{
		dataFile: dataFile,
	}

	// 初始化加载数据
	err := store.Load()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Load 从文件加载节点数据
func (s *YAMLGameNodeStore) Load() error {
	// 读取文件内容
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析YAML
	var nodes []models.GameNode
	err = yaml.Unmarshal(data, &nodes)
	if err != nil {
		return fmt.Errorf("解析YAML失败: %w", err)
	}

	s.nodes = nodes
	return nil
}

// Save 保存节点数据到文件
func (s *YAMLGameNodeStore) Save() error {
	// 序列化为YAML
	data, err := yaml.Marshal(s.nodes)
	if err != nil {
		return fmt.Errorf("序列化YAML失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(s.dataFile, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// List 获取所有节点
func (s *YAMLGameNodeStore) List() ([]models.GameNode, error) {
	// 创建副本避免修改原始数据
	nodes := make([]models.GameNode, len(s.nodes))
	copy(nodes, s.nodes)
	return nodes, nil
}

// Get 获取指定ID的节点
func (s *YAMLGameNodeStore) Get(id string) (models.GameNode, error) {
	for _, node := range s.nodes {
		if node.ID == id {
			return node, nil
		}
	}
	return models.GameNode{}, fmt.Errorf("节点不存在: %s", id)
}

// Add 添加节点
func (s *YAMLGameNodeStore) Add(node models.GameNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查ID是否已存在
	for _, existing := range s.nodes {
		if existing.ID == node.ID {
			return fmt.Errorf("节点已存在: %s", node.ID)
		}
	}

	s.nodes = append(s.nodes, node)
	return s.Save()
}

// Update 更新节点
func (s *YAMLGameNodeStore) Update(node models.GameNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并更新节点
	for i, existing := range s.nodes {
		if existing.ID == node.ID {
			s.nodes[i] = node
			return s.Save()
		}
	}
	return fmt.Errorf("节点不存在: %s", node.ID)
}

// Delete 删除节点
func (s *YAMLGameNodeStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除节点
	for i, node := range s.nodes {
		if node.ID == id {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			return s.Save()
		}
	}
	return fmt.Errorf("节点不存在: %s", id)
}

// Cleanup 清理存储文件
func (s *YAMLGameNodeStore) Cleanup() error {
	return os.Remove(s.dataFile)
}
