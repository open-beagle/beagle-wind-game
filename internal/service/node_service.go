package service

import (
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// NodeService 游戏节点服务
type NodeService struct {
	nodeStore     store.NodeStore
	instanceStore store.InstanceStore
}

// NewNodeService 创建游戏节点服务
func NewNodeService(nodeStore store.NodeStore, instanceStore store.InstanceStore) *NodeService {
	return &NodeService{
		nodeStore:     nodeStore,
		instanceStore: instanceStore,
	}
}

// NodeListParams 节点列表查询参数
type NodeListParams struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty"`
	Status   string `form:"status" binding:"omitempty,oneof=online offline maintenance"`
}

// NodeListResult 节点列表查询结果
type NodeListResult struct {
	Total int               `json:"total"`
	Items []models.GameNode `json:"items"`
}

// ListNodes 获取游戏节点列表
func (s *NodeService) ListNodes(params NodeListParams) (NodeListResult, error) {
	// 从存储获取节点列表
	nodes, err := s.nodeStore.List()
	if err != nil {
		return NodeListResult{}, err
	}

	// 过滤和分页
	var filteredNodes []models.GameNode
	for _, node := range nodes {
		// 关键词过滤
		if params.Keyword != "" {
			// TODO: 实现关键词过滤
			continue
		}

		// 状态过滤
		if params.Status != "" && string(node.Status.State) != params.Status {
			continue
		}

		// 如果满足所有条件，添加到结果中
		filteredNodes = append(filteredNodes, node)
	}

	// 处理分页
	total := len(filteredNodes)
	// 默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	// 计算分页范围
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start >= total {
		return NodeListResult{
			Total: total,
			Items: []models.GameNode{},
		}, nil
	}
	if end > total {
		end = total
	}

	return NodeListResult{
		Total: total,
		Items: filteredNodes[start:end],
	}, nil
}

// GetNode 获取节点信息
func (s *NodeService) GetNode(id string) (*models.GameNode, error) {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// CreateNode 创建游戏节点
func (s *NodeService) CreateNode(node models.GameNode) error {
	// 设置创建时间和更新时间
	now := time.Now()
	node.CreatedAt = now
	node.UpdatedAt = now

	// 设置初始状态
	node.Status = models.GameNodeStatus{
		State:      models.GameNodeStateOffline,
		Online:     false,
		LastOnline: now,
		UpdatedAt:  now,
		Resources:  make(map[string]string),
		Metrics:    make(map[string]interface{}),
	}

	return s.nodeStore.Add(node)
}

// UpdateNode 更新游戏节点
func (s *NodeService) UpdateNode(id string, node models.GameNode) error {
	// 检查节点是否存在
	existingNode, err := s.nodeStore.Get(id)
	if err != nil {
		return err
	}

	// 保留创建时间
	node.CreatedAt = existingNode.CreatedAt
	// 更新更新时间
	node.UpdatedAt = time.Now()
	// 确保ID一致
	node.ID = id

	return s.nodeStore.Update(node)
}

// DeleteNode 删除游戏节点
func (s *NodeService) DeleteNode(id string) error {
	// 检查节点是否存在
	_, err := s.nodeStore.Get(id)
	if err != nil {
		return err
	}

	// 检查是否有实例使用该节点
	instances, err := s.instanceStore.FindByNodeID(id)
	if err != nil {
		return fmt.Errorf("检查实例失败: %w", err)
	}

	if len(instances) > 0 {
		return fmt.Errorf("有%d个实例正在使用该节点，无法删除", len(instances))
	}

	return s.nodeStore.Delete(id)
}

// UpdateNodeStatus 更新节点状态
func (s *NodeService) UpdateNodeStatus(id string, status string) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return err
	}

	node.Status.State = models.GameNodeState(status)
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	return s.nodeStore.Update(node)
}

// UpdateNodeMetrics 更新节点指标
func (s *NodeService) UpdateNodeMetrics(id string, metrics map[string]interface{}) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return err
	}

	node.Status.Metrics = metrics
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	return s.nodeStore.Update(node)
}

// UpdateNodeResources 更新节点资源使用情况
func (s *NodeService) UpdateNodeResources(id string, resources map[string]interface{}) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return err
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

	node.Status.Resources = stringResources
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	return s.nodeStore.Update(node)
}

// UpdateNodeOnlineStatus 更新节点在线状态
func (s *NodeService) UpdateNodeOnlineStatus(id string, online bool) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return err
	}

	node.Status.Online = online
	if online {
		node.Status.LastOnline = time.Now()
		node.Status.State = models.GameNodeStateOnline
	} else {
		node.Status.State = models.GameNodeStateOffline
	}
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	return s.nodeStore.Update(node)
}
