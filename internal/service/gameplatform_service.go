package service

import (
	"context"
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// GamePlatformService 游戏平台服务
type GamePlatformService struct {
	platformStore store.GamePlatformStore
	logger        utils.Logger
}

// NewGamePlatformService 创建游戏平台服务
func NewGamePlatformService(platformStore store.GamePlatformStore) *GamePlatformService {
	logger := utils.New("GamePlatformService")
	return &GamePlatformService{
		platformStore: platformStore,
		logger:        logger,
	}
}

// GamePlatformListParams 平台列表查询参数
type GamePlatformListParams struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty"`
	Status   string `form:"status" binding:"omitempty,oneof=active maintenance inactive"`
}

// GamePlatformListResult 平台列表查询结果
type GamePlatformListResult struct {
	Total int                   `json:"total"`
	Items []models.GamePlatform `json:"items"`
}

// List 获取游戏平台列表
func (s *GamePlatformService) List(ctx context.Context, params GamePlatformListParams) (*GamePlatformListResult, error) {
	s.logger.Debug("获取游戏平台列表，参数: page=%d, size=%d, keyword=%s, status=%s",
		params.Page, params.PageSize, params.Keyword, params.Status)

	platforms, err := s.platformStore.List(ctx)
	if err != nil {
		s.logger.Error("获取平台列表失败: %v", err)
		return nil, fmt.Errorf("存储层错误: %w", err)
	}

	// 过滤和分页
	var filteredPlatforms []models.GamePlatform
	for _, platform := range platforms {
		// 关键词过滤
		if params.Keyword != "" {
			// TODO: 实现关键词过滤
		}

		// 状态过滤
		if params.Status != "" {
			// TODO: 实现状态过滤
		}

		// 如果没有过滤条件或者满足过滤条件，添加到结果中
		if params.Keyword == "" && params.Status == "" {
			filteredPlatforms = append(filteredPlatforms, platform)
		}
	}

	// 处理分页
	total := len(filteredPlatforms)
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
		return &GamePlatformListResult{
			Total: total,
			Items: []models.GamePlatform{},
		}, nil
	}
	if end > total {
		end = total
	}

	s.logger.Debug("返回游戏平台列表: total=%d, filtered=%d, page=%d, size=%d",
		len(platforms), total, params.Page, params.PageSize)
	return &GamePlatformListResult{
		Total: total,
		Items: filteredPlatforms[start:end],
	}, nil
}

// Get 获取指定ID的平台
func (s *GamePlatformService) Get(ctx context.Context, id string) (*models.GamePlatform, error) {
	s.logger.Debug("获取游戏平台详情: %s", id)

	platform, err := s.platformStore.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取平台详情失败: %v", err)
		return nil, fmt.Errorf("存储层错误: %w", err)
	}
	if platform.ID == "" {
		s.logger.Error("平台不存在: %s", id)
		return nil, fmt.Errorf("平台不存在: %s", id)
	}

	s.logger.Debug("成功获取游戏平台详情: %s", id)
	return &platform, nil
}

// GamePlatformAccessResult 平台远程访问结果
type GamePlatformAccessResult struct {
	Link      string    `json:"link"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetAccess 获取平台远程访问链接
func (s *GamePlatformService) GetAccess(ctx context.Context, id string) (GamePlatformAccessResult, error) {
	s.logger.Debug("获取游戏平台访问链接: %s", id)

	// 检查平台是否存在
	platform, err := s.platformStore.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取平台详情失败: %v", err)
		return GamePlatformAccessResult{}, fmt.Errorf("存储层错误: %w", err)
	}
	if platform.ID == "" {
		s.logger.Error("平台不存在: %s", id)
		return GamePlatformAccessResult{}, fmt.Errorf("平台不存在: %s", id)
	}

	// 生成访问链接
	expiresAt := time.Now().Add(24 * time.Hour)
	link := "https://vnc.example.com/platform/" + id

	s.logger.Debug("成功生成平台访问链接: id=%s, link=%s, expires=%v", id, link, expiresAt)
	return GamePlatformAccessResult{
		Link:      link,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshAccess 刷新平台远程访问链接
func (s *GamePlatformService) RefreshAccess(ctx context.Context, id string) (GamePlatformAccessResult, error) {
	s.logger.Debug("刷新游戏平台访问链接: %s", id)

	// 检查平台是否存在
	platform, err := s.platformStore.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取平台详情失败: %v", err)
		return GamePlatformAccessResult{}, fmt.Errorf("存储层错误: %w", err)
	}
	if platform.ID == "" {
		s.logger.Error("平台不存在: %s", id)
		return GamePlatformAccessResult{}, fmt.Errorf("平台不存在: %s", id)
	}

	// 刷新访问链接
	expiresAt := time.Now().Add(24 * time.Hour)
	link := "https://vnc.example.com/platform/" + id + "?refresh=" + time.Now().String()

	s.logger.Debug("成功刷新平台访问链接: id=%s, link=%s, expires=%v", id, link, expiresAt)
	return GamePlatformAccessResult{
		Link:      link,
		ExpiresAt: expiresAt,
	}, nil
}

// Update 更新平台信息
func (s *GamePlatformService) Update(ctx context.Context, id string, platformData models.GamePlatform) error {
	s.logger.Debug("更新游戏平台: %s", id)

	// 检查平台是否存在
	existingPlatform, err := s.platformStore.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取平台详情失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingPlatform.ID == "" {
		s.logger.Error("平台不存在: %s", id)
		return fmt.Errorf("平台不存在: %s", id)
	}

	// 确保ID不变
	platformData.ID = existingPlatform.ID

	// 设置更新时间
	platformData.UpdatedAt = time.Now()

	// 保留创建时间
	platformData.CreatedAt = existingPlatform.CreatedAt

	// 更新平台信息
	err = s.platformStore.Update(ctx, platformData)
	if err != nil {
		s.logger.Error("更新平台失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功更新游戏平台: %s", id)
	return nil
}

// Create 创建新平台
func (s *GamePlatformService) Create(ctx context.Context, platformData models.GamePlatform) (string, error) {
	s.logger.Debug("创建游戏平台: %s", platformData.ID)

	// 检查平台是否已存在
	existingPlatform, err := s.platformStore.Get(ctx, platformData.ID)
	if err != nil {
		s.logger.Error("检查平台是否存在失败: %v", err)
		return "", fmt.Errorf("存储层错误: %w", err)
	}
	if existingPlatform.ID != "" {
		s.logger.Error("平台ID已存在: %s", platformData.ID)
		return "", fmt.Errorf("平台ID已存在: %s", platformData.ID)
	}

	// 设置时间戳
	now := time.Now()
	platformData.CreatedAt = now
	platformData.UpdatedAt = now

	// 添加到存储
	err = s.platformStore.Add(ctx, platformData)
	if err != nil {
		s.logger.Error("添加平台失败: %v", err)
		return "", fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功创建游戏平台: %s", platformData.ID)
	return platformData.ID, nil
}

// Delete 删除平台
func (s *GamePlatformService) Delete(ctx context.Context, id string) error {
	s.logger.Debug("删除游戏平台: %s", id)

	// 检查平台是否存在
	existingPlatform, err := s.platformStore.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取平台详情失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingPlatform.ID == "" {
		s.logger.Error("平台不存在: %s", id)
		return fmt.Errorf("平台不存在: %s", id)
	}

	// 删除平台
	err = s.platformStore.Delete(ctx, id)
	if err != nil {
		s.logger.Error("删除平台失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功删除游戏平台: %s", id)
	return nil
}
