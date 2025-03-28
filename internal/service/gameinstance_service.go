package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// 错误常量定义
var (
	ErrInstanceNotFound       = fmt.Errorf("实例不存在")
	ErrInstanceAlreadyRunning = fmt.Errorf("实例已经在运行中")
	ErrInstanceNotRunning     = fmt.Errorf("实例未在运行中")
	ErrNodeNotFound           = fmt.Errorf("节点不存在")
	ErrNodeNotReady           = fmt.Errorf("节点未就绪")
)

// GameInstanceService 游戏实例服务
type GameInstanceService struct {
	GameInstanceStore store.GameInstanceStore
}

// NewGameInstanceService 创建游戏实例服务
func NewGameInstanceService(GameInstanceStore store.GameInstanceStore) *GameInstanceService {
	return &GameInstanceService{
		GameInstanceStore: GameInstanceStore,
	}
}

// GameInstanceListParams 实例列表查询参数
type GameInstanceListParams struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword    string `form:"keyword" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty,oneof=running stopped paused error"`
	NodeID     string `form:"node_id" binding:"omitempty"`
	CardID     string `form:"card_id" binding:"omitempty"`
	PlatformID string `form:"platform_id" binding:"omitempty"`
}

// InstanceListResult 实例列表结果
type InstanceListResult struct {
	Total int                   `json:"total"`
	Items []models.GameInstance `json:"items"`
}

// List 获取游戏实例列表
func (s *GameInstanceService) List(params GameInstanceListParams) (*InstanceListResult, error) {
	// 从存储获取实例列表
	instances, err := s.GameInstanceStore.List()
	if err != nil {
		return nil, fmt.Errorf("存储层错误: %w", err)
	}

	// 过滤和分页
	var filteredInstances []models.GameInstance
	for _, instance := range instances {
		// 关键词过滤
		if params.Keyword != "" {
			if !strings.Contains(instance.ID, params.Keyword) &&
				!strings.Contains(instance.NodeID, params.Keyword) &&
				!strings.Contains(instance.CardID, params.Keyword) {
				continue
			}
		}

		// 状态过滤
		if params.Status != "" && instance.Status != params.Status {
			continue
		}

		// 节点ID过滤
		if params.NodeID != "" && instance.NodeID != params.NodeID {
			continue
		}

		// 游戏卡片ID过滤
		if params.CardID != "" && instance.CardID != params.CardID {
			continue
		}

		// 平台ID过滤
		if params.PlatformID != "" && instance.PlatformID != params.PlatformID {
			continue
		}

		// 如果满足所有条件，添加到结果中
		filteredInstances = append(filteredInstances, instance)
	}

	// 处理分页
	total := len(filteredInstances)
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
		return &InstanceListResult{
			Total: total,
			Items: []models.GameInstance{},
		}, nil
	}
	if end > total {
		end = total
	}

	return &InstanceListResult{
		Total: total,
		Items: filteredInstances[start:end],
	}, nil
}

// Get 获取实例详情
func (s *GameInstanceService) Get(id string) (models.GameInstance, error) {
	// 从存储获取实例详情
	instance, err := s.GameInstanceStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return models.GameInstance{}, fmt.Errorf("存储层错误")
		}
		if strings.Contains(err.Error(), "实例不存在") {
			return models.GameInstance{}, fmt.Errorf("实例不存在: %s", id)
		}
		return models.GameInstance{}, err
	}

	if instance.ID == "" {
		return models.GameInstance{}, fmt.Errorf("实例不存在: %s", id)
	}

	return instance, nil
}

// CreateInstanceParams 创建实例参数
type CreateInstanceParams struct {
	NodeID     string `json:"node_id" binding:"required"`
	PlatformID string `json:"platform_id" binding:"required"`
	CardID     string `json:"card_id" binding:"required"`
	Config     string `json:"config,omitempty"` // 自定义配置
}

// Create 创建游戏实例
func (s *GameInstanceService) Create(params CreateInstanceParams) (string, error) {
	// 创建实例
	now := time.Now()
	instance := models.GameInstance{
		ID:         params.NodeID + "-" + params.CardID,
		NodeID:     params.NodeID,
		PlatformID: params.PlatformID,
		CardID:     params.CardID,
		Status:     "starting", // 初始状态为starting
		Resources:  "",         // 待填充
		Config:     params.Config,
		CreatedAt:  now,
		UpdatedAt:  now,
		StartedAt:  now,
	}

	// 检查实例是否已存在
	existingInstance, err := s.GameInstanceStore.Get(instance.ID)
	if err != nil && !strings.Contains(err.Error(), "实例不存在") {
		return "", fmt.Errorf("存储层错误: %w", err)
	}
	if existingInstance.ID != "" {
		return "", fmt.Errorf("实例ID已存在: %s", instance.ID)
	}

	// 保存实例
	err = s.GameInstanceStore.Add(instance)
	if err != nil {
		return "", fmt.Errorf("存储层错误: %w", err)
	}

	return instance.ID, nil
}

// UpdateInstanceParams 更新实例参数
type UpdateInstanceParams struct {
	Status      string `json:"status,omitempty"`
	Resources   string `json:"resources,omitempty"`
	Performance string `json:"performance,omitempty"`
	SaveData    string `json:"save_data,omitempty"`
	Config      string `json:"config,omitempty"`
	Backup      string `json:"backup,omitempty"`
}

// Update 更新游戏实例
func (s *GameInstanceService) Update(id string, instance models.GameInstance) error {
	// 检查实例是否存在
	existingInstance, err := s.GameInstanceStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}
	if existingInstance.ID == "" {
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 保留创建时间
	instance.CreatedAt = existingInstance.CreatedAt
	// 更新更新时间
	instance.UpdatedAt = time.Now()
	// 确保ID一致
	instance.ID = id

	// 更新实例
	err = s.GameInstanceStore.Update(instance)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}

	return nil
}

// Delete 删除游戏实例
func (s *GameInstanceService) Delete(id string) error {
	// 检查实例是否存在
	existingInstance, err := s.GameInstanceStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}
	if existingInstance.ID == "" {
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 删除实例
	err = s.GameInstanceStore.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}

	return nil
}

// Start 启动游戏实例
func (s *GameInstanceService) Start(id string) error {
	// 获取实例
	instance, err := s.GameInstanceStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		if strings.Contains(err.Error(), "实例不存在") {
			return fmt.Errorf("实例不存在: %s", id)
		}
		return err
	}

	if instance.ID == "" {
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 检查实例状态
	if instance.Status == "running" {
		return ErrInstanceAlreadyRunning
	}

	// 更新实例状态
	instance.Status = "running"
	instance.StartedAt = time.Now()
	instance.UpdatedAt = time.Now()

	// 保存更新
	err = s.GameInstanceStore.Update(instance)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}

	return nil
}

// Stop 停止游戏实例
func (s *GameInstanceService) Stop(id string) error {
	// 获取实例
	instance, err := s.GameInstanceStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		if strings.Contains(err.Error(), "实例不存在") {
			return fmt.Errorf("实例不存在: %s", id)
		}
		return err
	}

	if instance.ID == "" {
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 检查实例状态
	if instance.Status != "running" {
		return ErrInstanceNotRunning
	}

	// 更新实例状态
	instance.Status = "stopped"
	instance.UpdatedAt = time.Now()

	// 保存更新
	err = s.GameInstanceStore.Update(instance)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}

	return nil
}
