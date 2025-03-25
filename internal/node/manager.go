package node

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// Manager 节点管理器
type Manager struct {
	sync.RWMutex
	nodes     map[string]*models.GameNode
	nodesFile string
}

// NewManager 创建新的节点管理器
func NewManager(nodesFile string) (*Manager, error) {
	m := &Manager{
		nodes:     make(map[string]*models.GameNode),
		nodesFile: nodesFile,
	}

	// 加载节点数据
	if err := m.loadNodes(); err != nil {
		return nil, fmt.Errorf("加载节点数据失败: %v", err)
	}

	return m, nil
}

// loadNodes 从文件加载节点数据
func (m *Manager) loadNodes() error {
	data, err := ioutil.ReadFile(m.nodesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var nodes []*models.GameNode
	if err := yaml.Unmarshal(data, &nodes); err != nil {
		return err
	}

	m.Lock()
	defer m.Unlock()

	for _, node := range nodes {
		m.nodes[node.ID] = node
	}

	return nil
}

// saveNodes 保存节点数据到文件
func (m *Manager) saveNodes() error {
	m.RLock()
	nodes := make([]*models.GameNode, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	m.RUnlock()

	data, err := yaml.Marshal(nodes)
	if err != nil {
		return err
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(m.nodesFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(m.nodesFile, data, 0644)
}

// GetNode 获取节点信息
func (m *Manager) GetNode(id string) (*models.GameNode, error) {
	m.RLock()
	defer m.RUnlock()

	node, exists := m.nodes[id]
	if !exists {
		return nil, fmt.Errorf("节点不存在: %s", id)
	}

	return node, nil
}

// UpdateNodeStatus 更新节点状态
func (m *Manager) UpdateNodeStatus(id string, status string) error {
	m.Lock()
	defer m.Unlock()

	node, exists := m.nodes[id]
	if !exists {
		return fmt.Errorf("节点不存在: %s", id)
	}

	// 更新状态
	node.Status.State = models.GameNodeState(status)
	node.Status.UpdatedAt = time.Now().UTC()
	node.UpdatedAt = time.Now().UTC()

	// 保存到文件
	if err := m.saveNodes(); err != nil {
		return fmt.Errorf("保存节点状态失败: %v", err)
	}

	return nil
}

// UpdateNodeMetrics 更新节点指标
func (m *Manager) UpdateNodeMetrics(id string, metrics map[string]interface{}) error {
	m.Lock()
	defer m.Unlock()

	node, exists := m.nodes[id]
	if !exists {
		return fmt.Errorf("节点不存在: %s", id)
	}

	// 更新指标
	node.Status.Metrics = metrics
	node.Status.UpdatedAt = time.Now().UTC()
	node.UpdatedAt = time.Now().UTC()

	// 保存到文件
	if err := m.saveNodes(); err != nil {
		return fmt.Errorf("保存节点指标失败: %v", err)
	}

	return nil
}

// UpdateNodeResources 更新节点资源使用情况
func (m *Manager) UpdateNodeResources(id string, resources map[string]interface{}) error {
	m.Lock()
	defer m.Unlock()

	node, exists := m.nodes[id]
	if !exists {
		return fmt.Errorf("节点不存在: %s", id)
	}

	// 转换资源使用情况为字符串格式
	stringResources := make(map[string]string)
	for k, v := range resources {
		if str, ok := v.(string); ok {
			stringResources[k] = str
		} else {
			stringResources[k] = fmt.Sprintf("%v", v)
		}
	}

	// 更新资源使用情况
	node.Status.Resources = stringResources
	node.Status.UpdatedAt = time.Now().UTC()
	node.UpdatedAt = time.Now().UTC()

	// 保存到文件
	if err := m.saveNodes(); err != nil {
		return fmt.Errorf("保存节点资源使用情况失败: %v", err)
	}

	return nil
}

// UpdateNodeOnlineStatus 更新节点在线状态
func (m *Manager) UpdateNodeOnlineStatus(id string, online bool) error {
	m.Lock()
	defer m.Unlock()

	node, exists := m.nodes[id]
	if !exists {
		return fmt.Errorf("节点不存在: %s", id)
	}

	// 更新在线状态
	node.Status.Online = online
	if online {
		node.Status.LastOnline = time.Now().UTC()
		node.Status.State = models.GameNodeStateOnline
	} else {
		node.Status.State = models.GameNodeStateOffline
	}
	node.Status.UpdatedAt = time.Now().UTC()
	node.UpdatedAt = time.Now().UTC()

	// 保存到文件
	if err := m.saveNodes(); err != nil {
		return fmt.Errorf("保存节点在线状态失败: %v", err)
	}

	return nil
}

// ListNodes 列出所有节点
func (m *Manager) ListNodes() []*models.GameNode {
	m.RLock()
	defer m.RUnlock()

	nodes := make([]*models.GameNode, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetOnlineNodes 获取所有在线节点
func (m *Manager) GetOnlineNodes() []*models.GameNode {
	m.RLock()
	defer m.RUnlock()

	nodes := make([]*models.GameNode, 0)
	for _, node := range m.nodes {
		if node.Status.Online {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetOfflineNodes 获取所有离线节点
func (m *Manager) GetOfflineNodes() []*models.GameNode {
	m.RLock()
	defer m.RUnlock()

	nodes := make([]*models.GameNode, 0)
	for _, node := range m.nodes {
		if !node.Status.Online {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// CheckNodesStatus 检查节点状态
func (m *Manager) CheckNodesStatus(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Lock()
			now := time.Now().UTC()
			for _, node := range m.nodes {
				if node.Status.Online && now.Sub(node.Status.LastOnline) > 2*time.Minute {
					node.Status.Online = false
					node.Status.State = models.GameNodeStateOffline
					node.Status.UpdatedAt = now
					node.UpdatedAt = now
				}
			}
			m.Unlock()

			// 保存更新后的状态
			if err := m.saveNodes(); err != nil {
				fmt.Printf("保存节点状态失败: %v\n", err)
			}
		}
	}
}
