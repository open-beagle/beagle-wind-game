package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// NodeStore 游戏节点存储
type NodeStore struct {
	dataFile string
	nodes    []models.GameNode
	mu       sync.RWMutex
}

// NewNodeStore 创建游戏节点存储
func NewNodeStore(dataFile string) (*NodeStore, error) {
	store := &NodeStore{
		dataFile: dataFile,
	}

	// 初始化加载数据
	err := store.Load()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Load 加载节点数据
func (s *NodeStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 读取数据文件
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		// 如果文件不存在，创建一个空的节点列表
		s.nodes = []models.GameNode{}
		return s.Save()
	}

	// 解析YAML
	var nodes []models.GameNode
	err = yaml.Unmarshal(data, &nodes)
	if err != nil {
		return fmt.Errorf("解析节点数据文件失败: %w", err)
	}

	s.nodes = nodes
	return nil
}

// Save 保存节点数据
func (s *NodeStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 序列化为YAML
	data, err := yaml.Marshal(s.nodes)
	if err != nil {
		return fmt.Errorf("序列化节点数据失败: %w", err)
	}

	// 写入文件
	err = os.WriteFile(s.dataFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入节点数据文件失败: %w", err)
	}

	return nil
}

// List 获取所有节点
func (s *NodeStore) List() ([]models.GameNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 创建副本避免修改原始数据
	nodes := make([]models.GameNode, len(s.nodes))
	copy(nodes, s.nodes)

	return nodes, nil
}

// Get 获取指定ID的节点
func (s *NodeStore) Get(id string) (models.GameNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, node := range s.nodes {
		if node.ID == id {
			return node, nil
		}
	}

	return models.GameNode{}, fmt.Errorf("节点不存在: %s", id)
}

// Add 添加节点
func (s *NodeStore) Add(node models.GameNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查ID是否已存在
	for _, n := range s.nodes {
		if n.ID == node.ID {
			return fmt.Errorf("节点ID已存在: %s", node.ID)
		}
	}

	// 添加节点
	s.nodes = append(s.nodes, node)

	// 保存数据
	return s.Save()
}

// Update 更新节点
func (s *NodeStore) Update(node models.GameNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并更新节点
	for i, n := range s.nodes {
		if n.ID == node.ID {
			s.nodes[i] = node
			return s.Save()
		}
	}

	return fmt.Errorf("节点不存在: %s", node.ID)
}

// Delete 删除节点
func (s *NodeStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除节点
	for i, node := range s.nodes {
		if node.ID == id {
			// 从切片中删除节点
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			return s.Save()
		}
	}

	return fmt.Errorf("节点不存在: %s", id)
}
