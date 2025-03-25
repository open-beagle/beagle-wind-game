package service

import (
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// GameCardService 游戏卡片服务
type GameCardService struct {
	cardStore     store.GameCardStore
	platformStore store.PlatformStore
	instanceStore store.InstanceStore
}

// NewGameCardService 创建游戏卡片服务
func NewGameCardService(cardStore store.GameCardStore, platformStore store.PlatformStore, instanceStore store.InstanceStore) *GameCardService {
	return &GameCardService{
		cardStore:     cardStore,
		platformStore: platformStore,
		instanceStore: instanceStore,
	}
}

// GameCardListParams 游戏卡片列表查询参数
type GameCardListParams struct {
	Page     int    `form:"page"`
	PageSize int    `form:"size"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
	Category string `form:"category"`
}

// GameCardListResult 游戏卡片列表结果
type GameCardListResult struct {
	Total int               `json:"total"`
	Items []models.GameCard `json:"items"`
}

// ListGameCards 获取游戏卡片列表
func (s *GameCardService) ListGameCards(params GameCardListParams) (GameCardListResult, error) {
	// 从存储获取卡片列表
	cards, err := s.cardStore.List()
	if err != nil {
		return GameCardListResult{}, err
	}

	// 过滤和分页
	var filteredCards []models.GameCard
	for _, card := range cards {
		// 关键词过滤
		if params.Keyword != "" {
			// TODO: 实现关键词过滤
		}

		// 类型过滤
		if params.Type != "" && card.Type != params.Type {
			continue
		}

		// 分类过滤
		if params.Category != "" && card.Category != params.Category {
			continue
		}

		// 如果满足条件，添加到结果中
		filteredCards = append(filteredCards, card)
	}

	// 处理分页
	total := len(filteredCards)
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
		return GameCardListResult{
			Total: total,
			Items: []models.GameCard{},
		}, nil
	}
	if end > total {
		end = total
	}

	return GameCardListResult{
		Total: total,
		Items: filteredCards[start:end],
	}, nil
}

// GetGameCard 获取游戏卡片详情
func (s *GameCardService) GetGameCard(id string) (models.GameCard, error) {
	// 从存储获取卡片详情
	card, err := s.cardStore.Get(id)
	if err != nil {
		return models.GameCard{}, err
	}

	return card, nil
}

// CreateGameCard 创建游戏卡片
func (s *GameCardService) CreateGameCard(card models.GameCard) (string, error) {
	// 检查关联的平台是否存在
	if card.PlatformID != "" {
		_, err := s.platformStore.Get(card.PlatformID)
		if err != nil {
			return "", fmt.Errorf("关联的平台不存在: %w", err)
		}
	}

	// 设置创建时间和更新时间
	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	// 保存卡片
	err := s.cardStore.Add(card)
	if err != nil {
		return "", err
	}

	return card.ID, nil
}

// UpdateGameCard 更新游戏卡片
func (s *GameCardService) UpdateGameCard(id string, card models.GameCard) error {
	// 检查卡片是否存在
	existingCard, err := s.cardStore.Get(id)
	if err != nil {
		return err
	}

	// 检查关联的平台是否存在
	if card.PlatformID != "" && card.PlatformID != existingCard.PlatformID {
		_, err := s.platformStore.Get(card.PlatformID)
		if err != nil {
			return fmt.Errorf("关联的平台不存在: %w", err)
		}
	}

	// 保留创建时间
	card.CreatedAt = existingCard.CreatedAt
	// 更新更新时间
	card.UpdatedAt = time.Now()
	// 确保ID一致
	card.ID = id

	// 更新卡片
	return s.cardStore.Update(card)
}

// DeleteGameCard 删除游戏卡片
func (s *GameCardService) DeleteGameCard(id string) error {
	// 检查卡片是否存在
	_, err := s.cardStore.Get(id)
	if err != nil {
		return err
	}

	// 检查是否有实例使用该卡片
	instances, err := s.instanceStore.FindByCardID(id)
	if err != nil {
		return fmt.Errorf("检查实例失败: %w", err)
	}

	if len(instances) > 0 {
		return fmt.Errorf("有%d个实例正在使用该卡片，无法删除", len(instances))
	}

	// 删除卡片
	return s.cardStore.Delete(id)
}
