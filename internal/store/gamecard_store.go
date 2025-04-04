package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// GameCardStore 游戏卡片存储接口
type GameCardStore interface {
	// List 获取所有游戏卡片
	List() ([]models.GameCard, error)
	// Get 获取指定ID的游戏卡片
	Get(id string) (models.GameCard, error)
	// Add 添加游戏卡片
	Add(card models.GameCard) error
	// Update 更新游戏卡片信息
	Update(card models.GameCard) error
	// Delete 删除游戏卡片
	Delete(id string) error
	// Cleanup 清理测试文件
	Cleanup() error
}

// YAMLGameCardStore YAML文件存储实现
type YAMLGameCardStore struct {
	dataFile string
	cards    []models.GameCard
	mu       sync.RWMutex
}

// NewGameCardStore 创建游戏卡片存储
func NewGameCardStore(dataFile string) (GameCardStore, error) {
	store := &YAMLGameCardStore{
		dataFile: dataFile,
	}

	// 初始化加载数据
	err := store.Load()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Load 加载游戏卡片数据
func (s *YAMLGameCardStore) Load() error {
	// 确保目录存在
	dir := filepath.Dir(s.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 读取数据文件
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，创建一个空的卡片列表
			s.cards = []models.GameCard{}
			// 直接调用Save，不需要加锁
			return s.Save()
		}
		return fmt.Errorf("读取数据文件失败: %w", err)
	}

	// 解析YAML
	var cards []models.GameCard
	err = yaml.Unmarshal(data, &cards)
	if err != nil {
		return fmt.Errorf("解析游戏卡片数据文件失败: %w", err)
	}

	s.cards = cards
	return nil
}

// Save 保存游戏卡片数据到文件
func (s *YAMLGameCardStore) Save() error {
	// 序列化为YAML
	data, err := yaml.Marshal(s.cards)
	if err != nil {
		return fmt.Errorf("序列化游戏卡片数据失败: %w", err)
	}

	// 写入文件
	err = os.WriteFile(s.dataFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入游戏卡片数据文件失败: %w", err)
	}

	return nil
}

// List 获取所有卡片
func (s *YAMLGameCardStore) List() ([]models.GameCard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.GameCard{}, nil
		}
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	// 创建副本避免修改原始数据
	cards := make([]models.GameCard, len(s.cards))
	copy(cards, s.cards)
	return cards, nil
}

// Get 获取指定ID的卡片
func (s *YAMLGameCardStore) Get(id string) (models.GameCard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件状态
	fileInfo, err := os.Stat(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return models.GameCard{}, nil
		}
		return models.GameCard{}, fmt.Errorf("读取文件失败: %w", err)
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		return models.GameCard{}, fmt.Errorf("读取文件失败: 目标是一个目录")
	}

	for _, card := range s.cards {
		if card.ID == id {
			return card, nil
		}
	}
	return models.GameCard{}, nil
}

// Add 添加卡片
func (s *YAMLGameCardStore) Add(card models.GameCard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查ID是否已存在
	for _, c := range s.cards {
		if c.ID == card.ID {
			return fmt.Errorf("卡片ID已存在: %s", card.ID)
		}
	}

	// 添加卡片
	s.cards = append(s.cards, card)

	// 保存数据
	return s.Save()
}

// Update 更新卡片
func (s *YAMLGameCardStore) Update(card models.GameCard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并更新卡片
	for i, c := range s.cards {
		if c.ID == card.ID {
			s.cards[i] = card
			return s.Save()
		}
	}

	return fmt.Errorf("卡片不存在: %s", card.ID)
}

// Delete 删除卡片
func (s *YAMLGameCardStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除卡片
	for i, card := range s.cards {
		if card.ID == id {
			// 从切片中删除卡片
			s.cards = append(s.cards[:i], s.cards[i+1:]...)
			return s.Save()
		}
	}

	return fmt.Errorf("卡片不存在: %s", id)
}

// Cleanup 清理测试文件
func (s *YAMLGameCardStore) Cleanup() error {
	return os.Remove(s.dataFile)
}
