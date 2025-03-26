package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// NodeStore 游戏节点存储接口
type NodeStore interface {
	// List 获取所有节点
	List() ([]models.GameNode, error)
	// Get 获取指定ID的节点
	Get(id string) (models.GameNode, error)
	// Add 添加节点
	Add(node models.GameNode) error
	// Update 更新节点信息
	Update(node models.GameNode) error
	// Delete 删除节点
	Delete(id string) error
	// Cleanup 清理测试文件
	Cleanup() error
}

// YAMLNodeStore YAML文件存储实现
type YAMLNodeStore struct {
	dataFile string
	nodes    []models.GameNode
	mu       sync.RWMutex
}

// NewNodeStore 创建新的节点存储实例
func NewNodeStore(dataFile string) (NodeStore, error) {
	store := &YAMLNodeStore{
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
func (s *YAMLNodeStore) Load() error {
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
func (s *YAMLNodeStore) Save() error {
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
func (s *YAMLNodeStore) List() ([]models.GameNode, error) {
	// 创建副本避免修改原始数据
	nodes := make([]models.GameNode, len(s.nodes))
	copy(nodes, s.nodes)
	return nodes, nil
}

// Get 获取指定ID的节点
func (s *YAMLNodeStore) Get(id string) (models.GameNode, error) {
	for _, node := range s.nodes {
		if node.ID == id {
			return node, nil
		}
	}
	return models.GameNode{}, nil
}

// Add 添加节点
func (s *YAMLNodeStore) Add(node models.GameNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查ID是否已存在
	for _, existing := range s.nodes {
		if existing.ID == node.ID {
			return fmt.Errorf("存储层错误")
		}
	}

	s.nodes = append(s.nodes, node)
	return s.Save()
}

// Update 更新节点
func (s *YAMLNodeStore) Update(node models.GameNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并更新节点
	for i, existing := range s.nodes {
		if existing.ID == node.ID {
			s.nodes[i] = node
			return s.Save()
		}
	}
	return fmt.Errorf("存储层错误")
}

// Delete 删除节点
func (s *YAMLNodeStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除节点
	for i, node := range s.nodes {
		if node.ID == id {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			return s.Save()
		}
	}
	return fmt.Errorf("存储层错误")
}

// Cleanup 清理存储文件
func (s *YAMLNodeStore) Cleanup() error {
	return os.Remove(s.dataFile)
}
