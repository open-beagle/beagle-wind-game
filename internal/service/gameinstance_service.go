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
	logger            utils.Logger
}

// NewGameInstanceService 创建游戏实例服务
func NewGameInstanceService(GameInstanceStore store.GameInstanceStore) *GameInstanceService {
	logger := utils.New("GameInstanceService")
	return &GameInstanceService{
		GameInstanceStore: GameInstanceStore,
		logger:            logger,
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
func (s *GameInstanceService) List(ctx context.Context) ([]models.GameInstance, error) {
	s.logger.Debug("获取所有游戏实例")

	instances, err := s.GameInstanceStore.List(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return nil, fmt.Errorf("存储层错误")
		}
		s.logger.Error("获取实例列表失败: %v", err)
		return nil, err
	}

	s.logger.Debug("获取到 %d 个游戏实例", len(instances))
	return instances, nil
}

// Get 获取实例详情
func (s *GameInstanceService) Get(ctx context.Context, id string) (models.GameInstance, error) {
	s.logger.Debug("获取游戏实例: %s", id)

	// 从存储获取实例详情
	instance, err := s.GameInstanceStore.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return models.GameInstance{}, fmt.Errorf("存储层错误")
		}
		if strings.Contains(err.Error(), "实例不存在") {
			s.logger.Error("实例不存在: %s", id)
			return models.GameInstance{}, fmt.Errorf("实例不存在: %s", id)
		}
		s.logger.Error("获取实例失败: %v", err)
		return models.GameInstance{}, err
	}

	if instance.ID == "" {
		s.logger.Error("实例不存在: %s", id)
		return models.GameInstance{}, fmt.Errorf("实例不存在: %s", id)
	}

	s.logger.Debug("成功获取游戏实例: %s", id)
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
func (s *GameInstanceService) Create(ctx context.Context, params CreateInstanceParams) (string, error) {
	s.logger.Debug("创建游戏实例: 节点=%s, 平台=%s, 卡片=%s", params.NodeID, params.PlatformID, params.CardID)

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
	existingInstance, err := s.GameInstanceStore.Get(ctx, instance.ID)
	if err != nil && !strings.Contains(err.Error(), "实例不存在") {
		s.logger.Error("检查实例是否存在失败: %v", err)
		return "", fmt.Errorf("存储层错误: %w", err)
	}
	if existingInstance.ID != "" {
		s.logger.Error("实例ID已存在: %s", instance.ID)
		return "", fmt.Errorf("实例ID已存在: %s", instance.ID)
	}

	// 保存实例
	err = s.GameInstanceStore.Add(ctx, instance)
	if err != nil {
		s.logger.Error("添加实例失败: %v", err)
		return "", fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功创建游戏实例: %s", instance.ID)
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
func (s *GameInstanceService) Update(ctx context.Context, instance models.GameInstance) error {
	s.logger.Debug("更新游戏实例: %s", instance.ID)

	// 检查实例是否存在
	existingInstance, err := s.GameInstanceStore.Get(ctx, instance.ID)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("获取实例失败: %v", err)
		return err
	}
	if existingInstance.ID == "" {
		s.logger.Error("实例不存在: %s", instance.ID)
		return fmt.Errorf("实例不存在: %s", instance.ID)
	}

	// 保留创建时间
	instance.CreatedAt = existingInstance.CreatedAt
	// 更新更新时间
	instance.UpdatedAt = time.Now()
	// 确保ID一致
	instance.ID = existingInstance.ID

	// 更新实例
	err = s.GameInstanceStore.Update(ctx, instance)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("更新实例失败: %v", err)
		return err
	}

	s.logger.Info("成功更新游戏实例: %s", instance.ID)
	return nil
}

// Delete 删除游戏实例
func (s *GameInstanceService) Delete(ctx context.Context, id string) error {
	s.logger.Debug("删除游戏实例: %s", id)

	// 检查实例是否存在
	instance, err := s.GameInstanceStore.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		if strings.Contains(err.Error(), "实例不存在") {
			s.logger.Error("实例不存在: %s", id)
			return fmt.Errorf("实例不存在: %s", id)
		}
		s.logger.Error("获取实例失败: %v", err)
		return err
	}

	if instance.ID == "" {
		s.logger.Error("实例不存在: %s", id)
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 删除实例
	err = s.GameInstanceStore.Delete(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("删除实例失败: %v", err)
		return err
	}

	s.logger.Info("成功删除游戏实例: %s", id)
	return nil
}

// Start 启动游戏实例
func (s *GameInstanceService) Start(ctx context.Context, id string) error {
	s.logger.Debug("启动游戏实例: %s", id)

	// 获取实例
	instance, err := s.GameInstanceStore.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("获取实例失败: %v", err)
		return err
	}

	if instance.ID == "" {
		s.logger.Error("实例不存在: %s", id)
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 检查实例状态
	if instance.Status == "running" {
		s.logger.Error("实例已经在运行中: %s", id)
		return ErrInstanceAlreadyRunning
	}

	// 更新实例状态
	instance.Status = "running"
	instance.StartedAt = time.Now()
	instance.UpdatedAt = time.Now()

	// 保存更新
	err = s.GameInstanceStore.Update(ctx, instance)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("更新实例状态失败: %v", err)
		return err
	}

	s.logger.Info("成功启动游戏实例: %s", id)
	return nil
}

// Stop 停止游戏实例
func (s *GameInstanceService) Stop(ctx context.Context, id string) error {
	s.logger.Debug("停止游戏实例: %s", id)

	// 获取实例
	instance, err := s.GameInstanceStore.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		if strings.Contains(err.Error(), "实例不存在") {
			s.logger.Error("实例不存在: %s", id)
			return fmt.Errorf("实例不存在: %s", id)
		}
		s.logger.Error("获取实例失败: %v", err)
		return err
	}

	if instance.ID == "" {
		s.logger.Error("实例不存在: %s", id)
		return fmt.Errorf("实例不存在: %s", id)
	}

	// 检查实例状态
	if instance.Status != "running" {
		s.logger.Error("实例未在运行中: %s", id)
		return ErrInstanceNotRunning
	}

	// 更新实例状态
	instance.Status = "stopped"
	instance.UpdatedAt = time.Now()

	// 保存更新
	err = s.GameInstanceStore.Update(ctx, instance)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("更新实例状态失败: %v", err)
		return err
	}

	s.logger.Info("成功停止游戏实例: %s", id)
	return nil
}
