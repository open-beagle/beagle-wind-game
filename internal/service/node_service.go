package service

import (
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// NodeService 游戏节点服务
type NodeService struct {
	nodeStore     *store.NodeStore
	instanceStore *store.InstanceStore
}

// NewNodeService 创建游戏节点服务
func NewNodeService(nodeStore *store.NodeStore, instanceStore *store.InstanceStore) *NodeService {
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
		}

		// 状态过滤
		if params.Status != "" && node.Status != params.Status {
			continue
		}

		// 如果没有过滤条件或者满足过滤条件，添加到结果中
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

// GetNode 获取节点详情
func (s *NodeService) GetNode(id string) (models.GameNode, error) {
	// 获取节点基本信息
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return models.GameNode{}, err
	}

	// 获取节点上的实例信息
	_, err = s.instanceStore.FindByNodeID(id)
	if err != nil {
		return models.GameNode{}, fmt.Errorf("获取实例信息失败: %w", err)
	}

	// TODO: 构建实例信息摘要，添加到节点对象中

	return node, nil
}

// CreateNode 创建节点
func (s *NodeService) CreateNode(node models.GameNode) (string, error) {
	// 设置创建时间和更新时间
	now := time.Now()
	node.CreatedAt = now
	node.UpdatedAt = now

	// 保存节点
	err := s.nodeStore.Add(node)
	if err != nil {
		return "", err
	}

	return node.ID, nil
}

// UpdateNode 更新节点
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

	// 更新节点
	return s.nodeStore.Update(node)
}

// DeleteNode 删除节点
func (s *NodeService) DeleteNode(id string) error {
	// 检查节点是否存在
	_, err := s.nodeStore.Get(id)
	if err != nil {
		return err
	}

	// 检查节点上是否有运行中的实例
	instances, err := s.instanceStore.FindByNodeID(id)
	if err != nil {
		return fmt.Errorf("检查实例失败: %w", err)
	}

	if len(instances) > 0 {
		return fmt.Errorf("节点上有%d个实例，无法删除", len(instances))
	}

	// 删除节点
	return s.nodeStore.Delete(id)
}

// UpdateNodeStatus 更新节点状态
func (s *NodeService) UpdateNodeStatus(id string, status string, reason string) error {
	// 获取节点
	node, err := s.nodeStore.Get(id)
	if err != nil {
		return err
	}

	// 更新状态
	node.Status = status
	// TODO: 记录状态变更原因
	node.UpdatedAt = time.Now()

	// 更新节点
	return s.nodeStore.Update(node)
}
