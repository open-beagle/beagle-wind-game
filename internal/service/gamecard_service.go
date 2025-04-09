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

// GameCardService 游戏卡片服务
type GameCardService struct {
	cardStore store.GameCardStore
	logger    utils.Logger
}

// NewGameCardService 创建游戏卡片服务
func NewGameCardService(cardStore store.GameCardStore) *GameCardService {
	logger := utils.New("GameCardService")
	return &GameCardService{
		cardStore: cardStore,
		logger:    logger,
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

// List 获取游戏卡片列表
func (s *GameCardService) List(ctx context.Context, params GameCardListParams) (*GameCardListResult, error) {
	s.logger.Debug("获取游戏卡片列表，参数: page=%d, size=%d, keyword=%s, type=%s, category=%s",
		params.Page, params.PageSize, params.Keyword, params.Type, params.Category)

	// 从存储获取卡片列表
	cards, err := s.cardStore.List(ctx)
	if err != nil {
		s.logger.Error("获取卡片列表失败: %v", err)
		return nil, fmt.Errorf("存储层错误: %w", err)
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
		s.logger.Debug("分页超出范围，返回空列表: start=%d, total=%d", start, total)
		return &GameCardListResult{
			Total: total,
			Items: []models.GameCard{},
		}, nil
	}
	if end > total {
		end = total
	}

	s.logger.Debug("返回游戏卡片列表: total=%d, filtered=%d, page=%d, size=%d",
		len(cards), total, params.Page, params.PageSize)
	return &GameCardListResult{
		Total: total,
		Items: filteredCards[start:end],
	}, nil
}

// Get 获取游戏卡详情
func (s *GameCardService) Get(ctx context.Context, id string) (*models.GameCard, error) {
	s.logger.Debug("获取游戏卡详情: id=%s", id)

	// 从存储获取游戏卡详情
	card, err := s.cardStore.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return nil, fmt.Errorf("存储层错误")
		}
		s.logger.Error("获取游戏卡详情失败: %v", err)
		return nil, fmt.Errorf("存储层错误: %w", err)
	}

	// 如果卡片不存在，返回错误
	if card.ID == "" {
		s.logger.Error("卡片不存在: %s", id)
		return nil, fmt.Errorf("卡片不存在: %s", id)
	}

	s.logger.Debug("成功获取游戏卡详情: id=%s, name=%s", card.ID, card.Name)
	return &card, nil
}

// Create 创建游戏卡
func (s *GameCardService) Create(ctx context.Context, card models.GameCard) (string, error) {
	s.logger.Debug("创建游戏卡: id=%s, name=%s", card.ID, card.Name)

	// 检查游戏卡是否已存在
	existingCard, err := s.cardStore.Get(ctx, card.ID)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return "", fmt.Errorf("存储层错误")
		}
		s.logger.Error("检查卡片是否存在失败: %v", err)
		return "", fmt.Errorf("存储层错误: %w", err)
	}
	if existingCard.ID != "" {
		s.logger.Error("卡片ID已存在: %s", card.ID)
		return "", fmt.Errorf("卡片ID已存在: %s", card.ID)
	}

	// 设置创建时间和更新时间
	now := time.Now()
	if card.CreatedAt.IsZero() {
		card.CreatedAt = now
	}
	card.UpdatedAt = now

	// 保存游戏卡
	err = s.cardStore.Add(ctx, card)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return "", fmt.Errorf("存储层错误")
		}
		s.logger.Error("添加卡片失败: %v", err)
		return "", fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功创建游戏卡: id=%s, name=%s", card.ID, card.Name)
	return card.ID, nil
}

// Update 更新游戏卡
func (s *GameCardService) Update(ctx context.Context, id string, card models.GameCard) error {
	s.logger.Debug("更新游戏卡: id=%s, name=%s", id, card.Name)

	// 检查游戏卡是否存在
	existingCard, err := s.cardStore.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("检查卡片是否存在失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingCard.ID == "" {
		s.logger.Error("卡片不存在: %s", id)
		return fmt.Errorf("卡片不存在: %s", id)
	}

	// 保留创建时间
	card.CreatedAt = existingCard.CreatedAt
	// 更新更新时间
	card.UpdatedAt = time.Now()
	// 确保ID一致
	card.ID = id

	// 更新游戏卡
	err = s.cardStore.Update(ctx, card)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("更新卡片失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功更新游戏卡: id=%s, name=%s", id, card.Name)
	return nil
}

// Delete 删除游戏卡
func (s *GameCardService) Delete(ctx context.Context, id string) error {
	s.logger.Debug("删除游戏卡: id=%s", id)

	// 检查游戏卡是否存在
	existingCard, err := s.cardStore.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("检查卡片是否存在失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}
	if existingCard.ID == "" {
		s.logger.Error("卡片不存在: %s", id)
		return fmt.Errorf("卡片不存在: %s", id)
	}

	// 删除游戏卡
	err = s.cardStore.Delete(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "目标是一个目录") {
			s.logger.Error("存储层错误: 目标是一个目录")
			return fmt.Errorf("存储层错误")
		}
		s.logger.Error("删除卡片失败: %v", err)
		return fmt.Errorf("存储层错误: %w", err)
	}

	s.logger.Info("成功删除游戏卡: id=%s", id)
	return nil
}
