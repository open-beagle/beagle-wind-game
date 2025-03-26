package service

import (
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// GamePlatformService 游戏平台服务
type GamePlatformService struct {
	platformStore store.GamePlatformStore
}

// NewGamePlatformService 创建游戏平台服务
func NewGamePlatformService(platformStore store.GamePlatformStore) *GamePlatformService {
	return &GamePlatformService{
		platformStore: platformStore,
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
func (s *GamePlatformService) List(params GamePlatformListParams) (*GamePlatformListResult, error) {
	platforms, err := s.platformStore.List()
	if err != nil {
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
		return &GamePlatformListResult{
			Total: total,
			Items: []models.GamePlatform{},
		}, nil
	}
	if end > total {
		end = total
	}

	return &GamePlatformListResult{
		Total: total,
		Items: filteredPlatforms[start:end],
	}, nil
}

// Get 获取指定ID的平台
func (s *GamePlatformService) Get(id string) (*models.GamePlatform, error) {
	platform, err := s.platformStore.Get(id)
	if err != nil {
		return nil, fmt.Errorf("存储层错误: %w", err)
	}
	if platform.ID == "" {
		return nil, fmt.Errorf("平台不存在: %s", id)
	}
	return &platform, nil
}

// GamePlatformAccessResult 平台远程访问结果
type GamePlatformAccessResult struct {
	Link      string    `json:"link"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetAccess 获取平台远程访问链接
func (s *GamePlatformService) GetAccess(id string) (GamePlatformAccessResult, error) {
	// 检查平台是否存在
	platform, err := s.platformStore.Get(id)
	if err != nil {
		return GamePlatformAccessResult{}, fmt.Errorf("存储层错误: %w", err)
	}
	if platform.ID == "" {
		return GamePlatformAccessResult{}, fmt.Errorf("平台不存在: %s", id)
	}

	// 生成访问链接
	expiresAt := time.Now().Add(24 * time.Hour)
	return GamePlatformAccessResult{
		Link:      "https://vnc.example.com/platform/" + id,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshAccess 刷新平台远程访问链接
func (s *GamePlatformService) RefreshAccess(id string) (GamePlatformAccessResult, error) {
	// 检查平台是否存在
	platform, err := s.platformStore.Get(id)
	if err != nil {
		return GamePlatformAccessResult{}, fmt.Errorf("存储层错误: %w", err)
	}
	if platform.ID == "" {
		return GamePlatformAccessResult{}, fmt.Errorf("平台不存在: %s", id)
	}

	// 刷新访问链接
	expiresAt := time.Now().Add(24 * time.Hour)
	return GamePlatformAccessResult{
		Link:      "https://vnc.example.com/platform/" + id + "?refresh=" + time.Now().String(),
		ExpiresAt: expiresAt,
	}, nil
}

// Update 更新平台信息
func (s *GamePlatformService) Update(id string, platformData models.GamePlatform) error {
	// 检查平台是否存在
	existingPlatform, err := s.platformStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingPlatform.ID == "" {
		return fmt.Errorf("平台不存在: %s", id)
	}

	// 确保ID不变
	platformData.ID = existingPlatform.ID

	// 设置更新时间
	platformData.UpdatedAt = time.Now()

	// 保留创建时间
	platformData.CreatedAt = existingPlatform.CreatedAt

	// 更新平台信息
	err = s.platformStore.Update(platformData)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}

// Create 创建新平台
func (s *GamePlatformService) Create(platformData models.GamePlatform) (string, error) {
	// 检查平台是否已存在
	existingPlatform, err := s.platformStore.Get(platformData.ID)
	if err != nil {
		return "", fmt.Errorf("存储层错误: %w", err)
	}
	if existingPlatform.ID != "" {
		return "", fmt.Errorf("平台ID已存在: %s", platformData.ID)
	}

	// 设置时间戳
	now := time.Now()
	platformData.CreatedAt = now
	platformData.UpdatedAt = now

	// 添加到存储
	err = s.platformStore.Add(platformData)
	if err != nil {
		return "", fmt.Errorf("存储层错误: %w", err)
	}

	return platformData.ID, nil
}

// Delete 删除平台
func (s *GamePlatformService) Delete(id string) error {
	// 检查平台是否存在
	existingPlatform, err := s.platformStore.Get(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingPlatform.ID == "" {
		return fmt.Errorf("平台不存在: %s", id)
	}

	// 删除平台
	err = s.platformStore.Delete(id)
	if err != nil {
		return fmt.Errorf("存储层错误: %w", err)
	}
	return nil
}
