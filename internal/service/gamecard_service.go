package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// GameCardService 游戏卡片服务
type GameCardService struct {
	cardStore store.GameCardStore
}

// NewGameCardService 创建游戏卡片服务
func NewGameCardService(cardStore store.GameCardStore) *GameCardService {
	return &GameCardService{
		cardStore: cardStore,
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
		if os.IsNotExist(err) {
			return GameCardListResult{
				Total: 0,
				Items: []models.GameCard{},
			}, nil
		}
		return GameCardListResult{}, fmt.Errorf("存储层错误")
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

// GetGameCard 获取游戏卡详情
func (s *GameCardService) GetGameCard(id string) (*models.GameCard, error) {
	// 从存储获取游戏卡详情
	card, err := s.cardStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return nil, fmt.Errorf("存储层错误")
		}
		return nil, err
	}

	// 如果卡片不存在，返回错误
	if card.ID == "" {
		return nil, fmt.Errorf("卡片不存在: %s", id)
	}

	return &card, nil
}

// CreateGameCard 创建游戏卡
func (s *GameCardService) CreateGameCard(card models.GameCard) (string, error) {
	// 检查游戏卡是否已存在
	existingCard, err := s.cardStore.Get(card.ID)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return "", fmt.Errorf("存储层错误")
		}
		return "", err
	}
	if existingCard.ID != "" {
		return "", fmt.Errorf("卡片ID已存在: %s", card.ID)
	}

	// 保存游戏卡
	err = s.cardStore.Add(card)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return "", fmt.Errorf("存储层错误")
		}
		return "", err
	}

	return card.ID, nil
}

// UpdateGameCard 更新游戏卡
func (s *GameCardService) UpdateGameCard(id string, card models.GameCard) error {
	// 检查游戏卡是否存在
	existingCard, err := s.cardStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}
	if existingCard.ID == "" {
		return fmt.Errorf("卡片不存在: %s", id)
	}

	// 保留创建时间
	card.CreatedAt = existingCard.CreatedAt
	// 更新更新时间
	card.UpdatedAt = time.Now()
	// 确保ID一致
	card.ID = id

	// 更新游戏卡
	err = s.cardStore.Update(card)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}

	return nil
}

// DeleteGameCard 删除游戏卡
func (s *GameCardService) DeleteGameCard(id string) error {
	// 检查游戏卡是否存在
	existingCard, err := s.cardStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}
	if existingCard.ID == "" {
		return fmt.Errorf("卡片不存在: %s", id)
	}

	// 删除游戏卡
	err = s.cardStore.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			return fmt.Errorf("存储层错误")
		}
		return err
	}

	return nil
}
