package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// GameCardStore 游戏卡片存储接口
type GameCardStore interface {
	// List 获取所有游戏卡片
	List(ctx context.Context) ([]models.GameCard, error)
	// Get 获取指定ID的游戏卡片
	Get(ctx context.Context, id string) (models.GameCard, error)
	// Add 添加游戏卡片
	Add(ctx context.Context, card models.GameCard) error
	// Update 更新游戏卡片信息
	Update(ctx context.Context, card models.GameCard) error
	// Delete 删除游戏卡片
	Delete(ctx context.Context, id string) error
	// Load 加载游戏卡片数据
	Load(ctx context.Context) error
	// Save 保存游戏卡片数据
	Save(ctx context.Context) error
	// Cleanup 清理测试文件
	Cleanup(ctx context.Context) error
}

// YAMLGameCardStore YAML文件存储实现
type YAMLGameCardStore struct {
	dataFile  string
	cards     []models.GameCard
	mu        sync.RWMutex
	logger    utils.Logger
	yamlSaver *utils.YAMLSaver
}

// NewGameCardStore 创建游戏卡片存储
func NewGameCardStore(ctx context.Context, dataFile string) (GameCardStore, error) {
	logger := utils.New("GameCardStore")

	store := &YAMLGameCardStore{
		dataFile: dataFile,
		cards:    []models.GameCard{},
		logger:   logger,
	}

	// 创建YAML保存器，使用1秒的延迟保存
	store.yamlSaver = utils.NewYAMLSaver(
		dataFile,
		func() interface{} {
			// 这个函数返回当前的卡片数据，在保存时调用
			store.mu.RLock()
			defer store.mu.RUnlock()
			return store.cards
		},
		logger,
		utils.WithDelay(time.Second),
	)

	// 初始化加载数据
	logger.Info("初始化游戏卡片存储，数据文件: %s", dataFile)
	err := store.Load(ctx)
	if err != nil {
		logger.Error("加载游戏卡片数据失败: %v", err)
		return nil, err
	}

	logger.Info("成功加载游戏卡片数据，共%d张卡片", len(store.cards))
	return store, nil
}

// Load 加载游戏卡片数据
func (s *YAMLGameCardStore) Load(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保目录存在
	dir := filepath.Dir(s.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		s.logger.Error("创建目录失败: %v", err)
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 读取数据文件
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，创建一个空的卡片列表
			s.logger.Info("数据文件不存在，创建空列表: %s", s.dataFile)
			s.cards = []models.GameCard{}
			// 直接调用Save，不需要加锁
			return s.Save(ctx)
		}
		s.logger.Error("读取数据文件失败: %v", err)
		return fmt.Errorf("读取数据文件失败: %w", err)
	}

	// 解析YAML
	var cards []models.GameCard
	err = yaml.Unmarshal(data, &cards)
	if err != nil {
		s.logger.Error("解析游戏卡片数据文件失败: %v", err)
		return fmt.Errorf("解析游戏卡片数据文件失败: %w", err)
	}

	s.cards = cards
	s.logger.Debug("成功加载%d张游戏卡片", len(cards))
	return nil
}

// Save 保存游戏卡片数据到文件
func (s *YAMLGameCardStore) Save(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	// 使用延迟保存器进行保存
	return s.yamlSaver.Save(ctx)
}

// List 获取所有卡片
func (s *YAMLGameCardStore) List(ctx context.Context) ([]models.GameCard, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("数据文件不存在，返回空列表: %s", s.dataFile)
			return []models.GameCard{}, nil
		}
		s.logger.Error("读取文件失败: %v", err)
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("读取文件失败: 目标是一个目录: %s", s.dataFile)
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	// 创建副本避免修改原始数据
	cards := make([]models.GameCard, len(s.cards))
	copy(cards, s.cards)
	s.logger.Debug("返回%d张游戏卡片", len(cards))
	return cards, nil
}

// Get 获取指定ID的卡片
func (s *YAMLGameCardStore) Get(ctx context.Context, id string) (models.GameCard, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return models.GameCard{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("数据文件不存在，返回空卡片: %s", s.dataFile)
			return models.GameCard{}, nil
		}
		s.logger.Error("读取文件失败: %v", err)
		return models.GameCard{}, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("读取文件失败: 目标是一个目录: %s", s.dataFile)
		return models.GameCard{}, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	for _, card := range s.cards {
		if card.ID == id {
			s.logger.Debug("找到游戏卡片: %s", id)
			return card, nil
		}
	}
	s.logger.Warn("未找到游戏卡片: %s", id)
	return models.GameCard{}, nil
}

// Add 添加卡片
func (s *YAMLGameCardStore) Add(ctx context.Context, card models.GameCard) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查ID是否已存在
	for _, c := range s.cards {
		if c.ID == card.ID {
			s.logger.Error("卡片ID已存在: %s", card.ID)
			return fmt.Errorf("卡片ID已存在: %s", card.ID)
		}
	}

	// 添加卡片
	s.cards = append(s.cards, card)
	s.logger.Info("添加新卡片: %s (%s)", card.ID, card.Name)

	// 保存数据
	return s.Save(ctx)
}

// Update 更新卡片
func (s *YAMLGameCardStore) Update(ctx context.Context, card models.GameCard) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找卡片
	found := false
	for i, c := range s.cards {
		if c.ID == card.ID {
			s.cards[i] = card
			found = true
			s.logger.Info("更新卡片: %s (%s)", card.ID, card.Name)
			break
		}
	}

	if !found {
		s.logger.Error("卡片不存在，无法更新: %s", card.ID)
		return fmt.Errorf("卡片不存在: %s", card.ID)
	}

	// 保存数据
	return s.Save(ctx)
}

// Delete 删除卡片
func (s *YAMLGameCardStore) Delete(ctx context.Context, id string) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找卡片
	found := false
	for i, card := range s.cards {
		if card.ID == id {
			// 删除卡片
			s.cards = append(s.cards[:i], s.cards[i+1:]...)
			found = true
			s.logger.Info("删除卡片: %s", id)
			break
		}
	}

	if !found {
		s.logger.Error("卡片不存在，无法删除: %s", id)
		return fmt.Errorf("卡片不存在: %s", id)
	}

	// 保存数据
	return s.Save(ctx)
}

// Cleanup 清理存储文件
func (s *YAMLGameCardStore) Cleanup(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.logger.Info("清理存储文件: %s", s.dataFile)
	err := os.Remove(s.dataFile)
	if err != nil {
		s.logger.Error("清理存储文件失败: %v", err)
		return err
	}
	s.logger.Info("存储文件清理成功")
	return nil
}

// Close 关闭存储，确保所有待处理的保存操作完成
func (s *YAMLGameCardStore) Close() {
	s.logger.Info("关闭GameCardStore，确保数据保存...")
	if s.yamlSaver != nil {
		s.yamlSaver.Close()
	}
}
