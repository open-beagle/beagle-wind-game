package service

import (
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// NodeService 游戏节点服务
type NodeService struct {
	nodeStore store.NodeStore
}

// NewNodeService 创建游戏节点服务
func NewNodeService(nodeStore store.NodeStore) *NodeService {
	return &NodeService{
		nodeStore: nodeStore,
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

// NodeAccessResult 节点访问链接结果
type NodeAccessResult struct {
	Link      string    `json:"link"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ListNodes 获取游戏节点列表
func (s *NodeService) ListNodes(params NodeListParams) (NodeListResult, error) {
	// 从存储获取节点列表
	nodes, err := s.nodeStore.List()
	if err != nil {
		return NodeListResult{}, fmt.Errorf("存储层错误")
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
		return nil, fmt.Errorf("存储层错误")
	}

	// 如果节点不存在，返回错误
	if node.ID == "" {
		return nil, fmt.Errorf("节点不存在: %s", id)
	}

	return &node, nil
}

// CreateNode 创建游戏节点
func (s *NodeService) CreateNode(node models.GameNode) error {
	// 检查节点是否已存在
	existingNode, err := s.nodeStore.Get(node.ID)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	if existingNode.ID != "" {
		return fmt.Errorf("节点ID已存在: %s", node.ID)
	}

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

	err = s.nodeStore.Add(node)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	return nil
}

// UpdateNode 更新游戏节点
func (s *NodeService) UpdateNode(id string, node models.GameNode) error {
	// 检查节点是否存在
	existingNode, err := s.nodeStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	if existingNode.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	// 保留创建时间
	node.CreatedAt = existingNode.CreatedAt
	// 更新更新时间
	node.UpdatedAt = time.Now()
	// 确保ID一致
	node.ID = id

	err = s.nodeStore.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	return nil
}

// DeleteNode 删除游戏节点
func (s *NodeService) DeleteNode(id string) error {
	// 检查节点是否存在
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	err = s.nodeStore.Delete(id)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	return nil
}

// UpdateNodeStatus 更新节点状态
func (s *NodeService) UpdateNodeStatus(id string, status string) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.State = models.GameNodeState(status)
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	err = s.nodeStore.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	return nil
}

// UpdateNodeMetrics 更新节点指标
func (s *NodeService) UpdateNodeMetrics(id string, metrics map[string]interface{}) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.Metrics = metrics
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	err = s.nodeStore.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	return nil
}

// UpdateNodeResources 更新节点资源
func (s *NodeService) UpdateNodeResources(id string, resources map[string]interface{}) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	// 将 interface{} 转换为 string
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

	err = s.nodeStore.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	return nil
}

// UpdateNodeOnlineStatus 更新节点在线状态
func (s *NodeService) UpdateNodeOnlineStatus(id string, online bool) error {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.Online = online
	if online {
		node.Status.LastOnline = time.Now()
	}
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	err = s.nodeStore.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误")
	}
	return nil
}

// GetNodeAccess 获取节点访问链接
func (s *NodeService) GetNodeAccess(id string) (NodeAccessResult, error) {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return NodeAccessResult{}, fmt.Errorf("存储层错误")
	}
	if node.ID == "" {
		return NodeAccessResult{}, fmt.Errorf("节点不存在: %s", id)
	}

	// 生成访问链接和过期时间
	link := fmt.Sprintf("http://%s:%s", node.Network["ip"], node.Network["port"])
	expiresAt := time.Now().Add(24 * time.Hour)

	return NodeAccessResult{
		Link:      link,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshNodeAccess 刷新节点访问链接
func (s *NodeService) RefreshNodeAccess(id string) (NodeAccessResult, error) {
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return NodeAccessResult{}, fmt.Errorf("存储层错误")
	}
	if node.ID == "" {
		return NodeAccessResult{}, fmt.Errorf("节点不存在: %s", id)
	}

	// 生成新的访问链接和过期时间
	link := fmt.Sprintf("http://%s:%s", node.Network["ip"], node.Network["port"])
	expiresAt := time.Now().Add(24 * time.Hour)

	return NodeAccessResult{
		Link:      link,
		ExpiresAt: expiresAt,
	}, nil
}
