package service

import (
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// PlatformService 游戏平台服务
type PlatformService struct {
	platformStore store.PlatformStore
}

// NewPlatformService 创建游戏平台服务
func NewPlatformService(platformStore store.PlatformStore) *PlatformService {
	return &PlatformService{
		platformStore: platformStore,
	}
}

// PlatformListParams 平台列表查询参数
type PlatformListParams struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty"`
	Status   string `form:"status" binding:"omitempty,oneof=active maintenance inactive"`
}

// PlatformListResult 平台列表查询结果
type PlatformListResult struct {
	Total int                   `json:"total"`
	Items []models.GamePlatform `json:"items"`
}

// ListPlatforms 获取游戏平台列表
func (s *PlatformService) ListPlatforms(params PlatformListParams) (PlatformListResult, error) {
	platforms, err := s.platformStore.List()
	if err != nil {
		return PlatformListResult{}, err
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
		return PlatformListResult{
			Total: total,
			Items: []models.GamePlatform{},
		}, nil
	}
	if end > total {
		end = total
	}

	return PlatformListResult{
		Total: total,
		Items: filteredPlatforms[start:end],
	}, nil
}

// GetPlatform 获取游戏平台详情
func (s *PlatformService) GetPlatform(id string) (models.GamePlatform, error) {
	return s.platformStore.Get(id)
}

// PlatformAccessResult 平台远程访问结果
type PlatformAccessResult struct {
	Link      string    `json:"link"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetPlatformAccess 获取平台远程访问链接
func (s *PlatformService) GetPlatformAccess(id string) (PlatformAccessResult, error) {
	// 检查平台是否存在
	_, err := s.platformStore.Get(id)
	if err != nil {
		return PlatformAccessResult{}, err
	}

	// 生成访问链接
	expiresAt := time.Now().Add(24 * time.Hour)
	return PlatformAccessResult{
		Link:      "https://vnc.example.com/platform/" + id,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshPlatformAccess 刷新平台远程访问链接
func (s *PlatformService) RefreshPlatformAccess(id string) (PlatformAccessResult, error) {
	// 检查平台是否存在
	_, err := s.platformStore.Get(id)
	if err != nil {
		return PlatformAccessResult{}, err
	}

	// 刷新访问链接
	expiresAt := time.Now().Add(24 * time.Hour)
	return PlatformAccessResult{
		Link:      "https://vnc.example.com/platform/" + id + "?refresh=" + time.Now().String(),
		ExpiresAt: expiresAt,
	}, nil
}

// UpdatePlatform 更新平台信息
func (s *PlatformService) UpdatePlatform(id string, platformData models.GamePlatform) error {
	// 检查平台是否存在
	existingPlatform, err := s.platformStore.Get(id)
	if err != nil {
		return err
	}

	// 确保ID不变
	platformData.ID = existingPlatform.ID

	// 设置更新时间
	platformData.UpdatedAt = time.Now()

	// 保留创建时间
	platformData.CreatedAt = existingPlatform.CreatedAt

	// 更新平台信息
	return s.platformStore.Update(platformData)
}

// CreatePlatform 创建新平台
func (s *PlatformService) CreatePlatform(platformData models.GamePlatform) (string, error) {
	// 设置时间戳
	now := time.Now()
	platformData.CreatedAt = now
	platformData.UpdatedAt = now

	// 添加到存储
	err := s.platformStore.Add(platformData)
	if err != nil {
		return "", err
	}

	return platformData.ID, nil
}

// DeletePlatform 删除平台
func (s *PlatformService) DeletePlatform(id string) error {
	// 检查平台是否存在
	_, err := s.platformStore.Get(id)
	if err != nil {
		return err
	}

	// 删除平台
	return s.platformStore.Delete(id)
}
