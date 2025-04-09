package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// GameNodeService 游戏节点服务
type GameNodeService struct {
	store  store.GameNodeStore
	logger utils.Logger
}

// NewGameNodeService 创建游戏节点服务
func NewGameNodeService(store store.GameNodeStore) *GameNodeService {
	logger := utils.New("GameNodeService")
	return &GameNodeService{
		store:  store,
		logger: logger,
	}
}

// GameNodeListParams 节点列表查询参数
type GameNodeListParams struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword   string `form:"keyword" binding:"omitempty"`
	Status    string `form:"status" binding:"omitempty,oneof=online offline maintenance"`
	Type      string `form:"type" binding:"omitempty,oneof=physical virtual"`
	Region    string `form:"region" binding:"omitempty"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at status"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
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
func (s *GameNodeService) List(ctx context.Context, params GameNodeListParams) (*GameNodeListResult, error) {
	s.logger.Debug("获取游戏节点列表，参数: page=%d, size=%d, keyword=%s, status=%s, type=%s, region=%s, sortBy=%s, sortOrder=%s",
		params.Page, params.PageSize, params.Keyword, params.Status, params.Type, params.Region, params.SortBy, params.SortOrder)

	// 从存储获取节点列表
	nodes, err := s.store.List(ctx)
	if err != nil {
		s.logger.Error("获取节点列表失败: %v", err)
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
		s.logger.Debug("分页超出范围，返回空列表: start=%d, total=%d", start, total)
		return &GameNodeListResult{
			Total: total,
			Items: []models.GameNode{},
		}, nil
	}
	if end > total {
		end = total
	}

	s.logger.Debug("返回游戏节点列表: total=%d, filtered=%d, page=%d, size=%d",
		len(nodes), total, params.Page, params.PageSize)
	return &GameNodeListResult{
		Total: total,
		Items: filteredNodes[start:end],
	}, nil
}

// Get 获取节点信息
func (s *GameNodeService) Get(ctx context.Context, id string) (models.GameNode, error) {
	s.logger.Debug("获取游戏节点信息: %s", id)

	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return models.GameNode{}, fmt.Errorf("存储层错误: %w", err)
	}

	// 如果节点不存在，返回错误
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return models.GameNode{}, fmt.Errorf("节点不存在: %s", id)
	}

	s.logger.Debug("成功获取游戏节点信息: %s", id)
	return node, nil
}

// Create 创建游戏节点
func (s *GameNodeService) Create(ctx context.Context, node models.GameNode) error {
	s.logger.Debug("创建游戏节点: %s", node.ID)

	// 检查节点是否已存在
	existingNode, err := s.store.Get(ctx, node.ID)
	if err != nil && !strings.Contains(err.Error(), "节点不存在") {
		s.logger.Error("检查节点是否存在失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingNode.ID != "" {
		s.logger.Error("节点ID已存在: %s", node.ID)
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
		Hardware:   models.HardwareInfo{},
		System:     models.SystemInfo{},
		Metrics:    models.MetricsInfo{},
	}

	err = s.store.Add(ctx, node)
	if err != nil {
		s.logger.Error("添加节点失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功创建游戏节点: %s", node.ID)
	return nil
}

// Update 更新游戏节点
func (s *GameNodeService) Update(ctx context.Context, node models.GameNode) error {
	s.logger.Debug("更新游戏节点: %s", node.ID)

	// 检查节点是否存在
	existingNode, err := s.store.Get(ctx, node.ID)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingNode.ID == "" {
		s.logger.Error("节点不存在: %s", node.ID)
		return fmt.Errorf("节点不存在: %s", node.ID)
	}

	// 保留创建时间
	node.CreatedAt = existingNode.CreatedAt
	// 更新更新时间
	node.UpdatedAt = time.Now()

	err = s.store.Update(ctx, node)
	if err != nil {
		s.logger.Error("更新节点失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功更新游戏节点: %s", node.ID)
	return nil
}

// Delete 删除游戏节点
func (s *GameNodeService) Delete(ctx context.Context, id string, force bool) error {
	s.logger.Debug("删除游戏节点: %s, force=%v", id, force)

	// 检查节点是否存在
	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return fmt.Errorf("节点不存在: %s", id)
	}

	// 检查节点状态
	if !force && node.Status.State != models.GameNodeStateOffline {
		s.logger.Error("节点状态不允许删除: id=%s, 状态=%s", id, node.Status.State)
		return fmt.Errorf("节点状态不允许删除: %s", node.Status.State)
	}

	err = s.store.Delete(ctx, id)
	if err != nil {
		s.logger.Error("删除节点失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功删除游戏节点: %s", id)
	return nil
}

// UpdateStatusState 更新节点状态
func (s *GameNodeService) UpdateStatusState(ctx context.Context, id string, state string) error {
	s.logger.Debug("更新游戏节点状态: id=%s, state=%s", id, state)

	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.State = models.GameNodeState(state)
	node.Status.UpdatedAt = time.Now()

	err = s.store.Update(ctx, node)
	if err != nil {
		s.logger.Error("更新节点状态失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功更新游戏节点状态: id=%s, state=%s", id, state)
	return nil
}

// UpdateStatusMetrics 更新节点指标
func (s *GameNodeService) UpdateStatusMetrics(ctx context.Context, id string, metrics models.MetricsInfo) error {
	s.logger.Debug("更新游戏节点指标: %s", id)

	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.Metrics = metrics
	node.Status.UpdatedAt = time.Now()

	err = s.store.Update(ctx, node)
	if err != nil {
		s.logger.Error("更新节点指标失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功更新游戏节点指标: %s", id)
	return nil
}

// UpdateHardwareAndSystem 更新节点硬件和系统信息
func (s *GameNodeService) UpdateHardwareAndSystem(ctx context.Context, id string, hardware models.HardwareInfo, system models.SystemInfo) error {
	s.logger.Debug("更新游戏节点硬件和系统信息: %s", id)

	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.Hardware = hardware
	node.Status.System = system
	node.Status.UpdatedAt = time.Now()
	node.UpdatedAt = time.Now()

	err = s.store.Update(ctx, node)
	if err != nil {
		s.logger.Error("更新节点硬件和系统信息失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功更新游戏节点硬件和系统信息: %s", id)
	return nil
}

// UpdateStatusOnlineStatus 更新节点在线状态
func (s *GameNodeService) UpdateStatusOnlineStatus(ctx context.Context, id string, online bool) error {
	s.logger.Debug("更新游戏节点在线状态: id=%s, online=%v", id, online)

	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return fmt.Errorf("节点不存在: %s", id)
	}

	node.Status.Online = online
	if online {
		node.Status.LastOnline = time.Now()
		node.Status.State = models.GameNodeStateOnline
	} else {
		node.Status.State = models.GameNodeStateOffline
	}
	node.Status.UpdatedAt = time.Now()

	err = s.store.Update(ctx, node)
	if err != nil {
		s.logger.Error("更新节点在线状态失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功更新游戏节点在线状态: id=%s, online=%v", id, online)
	return nil
}

// GetAccess 获取节点访问链接
func (s *GameNodeService) GetAccess(ctx context.Context, id string) (NodeAccessResult, error) {
	s.logger.Debug("获取游戏节点访问链接: %s", id)

	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return NodeAccessResult{}, fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return NodeAccessResult{}, fmt.Errorf("节点不存在: %s", id)
	}

	// TODO: 实现访问链接生成逻辑
	result := NodeAccessResult{
		Link:      fmt.Sprintf("https://node-%s.example.com", id),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.logger.Debug("成功获取游戏节点访问链接: id=%s, link=%s, expires=%v", id, result.Link, result.ExpiresAt)
	return result, nil
}

// RefreshAccess 刷新节点访问链接
func (s *GameNodeService) RefreshAccess(ctx context.Context, id string) (NodeAccessResult, error) {
	s.logger.Debug("刷新游戏节点访问链接: %s", id)

	node, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取节点信息失败: %v", err)
		return NodeAccessResult{}, fmt.Errorf("存储层错误: %w", err)
	}
	if node.ID == "" {
		s.logger.Error("节点不存在: %s", id)
		return NodeAccessResult{}, fmt.Errorf("节点不存在: %s", id)
	}

	// TODO: 实现访问链接刷新逻辑
	result := NodeAccessResult{
		Link:      fmt.Sprintf("https://node-%s.example.com", id),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.logger.Debug("成功刷新游戏节点访问链接: id=%s, link=%s, expires=%v", id, result.Link, result.ExpiresAt)
	return result, nil
}
