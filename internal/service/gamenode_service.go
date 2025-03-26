package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// GameNodeService 游戏节点服务
type GameNodeService struct {
	store store.GameNodeStore
}

// NewGameNodeService 创建游戏节点服务
func NewGameNodeService(store store.GameNodeStore) *GameNodeService {
	return &GameNodeService{
		store: store,
	}
}

// GameNodeListParams 节点列表查询参数
type GameNodeListParams struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty"`
	Status   string `form:"status" binding:"omitempty,oneof=online offline maintenance"`
}

// GameNodeListResult 节点列表查询结果
type GameNodeListResult struct {
	Total int               `json:"total"`
	Items []models.GameNode `json:"items"`
}

// NodeAccessResult 节点访问链接结果
type NodeAccessResult struct {
	Link      string    `json:"link"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// List 获取游戏节点列表
func (s *GameNodeService) List(params GameNodeListParams) (*GameNodeListResult, error) {
	// 从存储获取节点列表
	nodes, err := s.store.List()
	if err != nil {
		return nil, fmt.Errorf("存储层错误: %w", err)
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
		return &GameNodeListResult{
			Total: total,
			Items: []models.GameNode{},
		}, nil
	}
	if end > total {
		end = total
	}

	return &GameNodeListResult{
		Total: total,
		Items: filteredNodes[start:end],
	}, nil
}

// Get 获取节点信息
func (s *GameNodeService) Get(id string) (*models.GameNode, error) {
	node, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("存储层错误: %w", err)
	}

	// 如果节点不存在，返回错误
	if node.ID == "" {
		return nil, fmt.Errorf("节点不存在: %s", id)
	}

	return &node, nil
}

// Create 创建游戏节点
func (s *GameNodeService) Create(node models.GameNode) error {
	// 检查节点是否已存在
	existingNode, err := s.store.Get(node.ID)
	if err != nil && !strings.Contains(err.Error(), "节点不存在") {
		return fmt.Errorf("存储层错误: %w", err)
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

	err = s.store.Add(node)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// Update 更新游戏节点
func (s *GameNodeService) Update(id string, node models.GameNode) error {
	// 检查节点是否存在
	existingNode, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
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

	err = s.store.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// Delete 删除游戏节点
func (s *GameNodeService) Delete(id string) error {
	// 检查节点是否存在
	node, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	err = s.store.Delete(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// UpdateStatusState 更新节点状态
func (s *GameNodeService) UpdateStatusState(id string, state string) error {
	node, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.State = models.GameNodeState(state)
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	err = s.store.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// UpdateStatusMetrics 更新节点指标
func (s *GameNodeService) UpdateStatusMetrics(id string, metrics map[string]interface{}) error {
	node, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.Metrics = metrics
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	err = s.store.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// UpdateStatusResources 更新节点资源
func (s *GameNodeService) UpdateStatusResources(id string, resources map[string]interface{}) error {
	node, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
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

	err = s.store.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// UpdateStatusOnlineStatus 更新节点在线状态
func (s *GameNodeService) UpdateStatusOnlineStatus(id string, online bool) error {
	node, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
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

	err = s.store.Update(node)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// GetAccess 获取节点访问链接
func (s *GameNodeService) GetAccess(id string) (NodeAccessResult, error) {
	node, err := s.store.Get(id)
	if err != nil && !strings.Contains(err.Error(), "节点不存在") {
		return NodeAccessResult{}, fmt.Errorf("存储层错误: %w", err)
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

// RefreshAccess 刷新节点访问链接
func (s *GameNodeService) RefreshAccess(id string) (NodeAccessResult, error) {
	node, err := s.store.Get(id)
	if err != nil && !strings.Contains(err.Error(), "节点不存在") {
		return NodeAccessResult{}, fmt.Errorf("存储层错误: %w", err)
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
