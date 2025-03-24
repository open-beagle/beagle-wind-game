package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// PlatformStore 游戏平台存储，从配置文件读取静态数据
type PlatformStore struct {
	configFile string
	platforms  []models.GamePlatform
	mu         sync.RWMutex
}

// NewPlatformStore 创建游戏平台存储
func NewPlatformStore(configFile string) (*PlatformStore, error) {
	store := &PlatformStore{
		configFile: configFile,
	}

	// 初始化加载数据
	err := store.Load()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Load 加载平台数据
func (s *PlatformStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		return fmt.Errorf("读取平台配置文件失败: %w", err)
	}

	// 解析YAML
	var platforms []models.GamePlatform
	err = yaml.Unmarshal(data, &platforms)
	if err != nil {
		return fmt.Errorf("解析平台配置文件失败: %w", err)
	}

	s.platforms = platforms
	return nil
}

// List 获取所有平台
func (s *PlatformStore) List() ([]models.GamePlatform, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 创建副本避免修改原始数据
	platforms := make([]models.GamePlatform, len(s.platforms))
	copy(platforms, s.platforms)

	return platforms, nil
}

// Get 获取指定ID的平台
func (s *PlatformStore) Get(id string) (models.GamePlatform, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, platform := range s.platforms {
		if platform.ID == id {
			return platform, nil
		}
	}

	return models.GamePlatform{}, fmt.Errorf("平台不存在: %s", id)
}

// Save 保存平台配置到文件
func (s *PlatformStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 将平台数据序列化为YAML
	data, err := yaml.Marshal(s.platforms)
	if err != nil {
		return fmt.Errorf("序列化平台数据失败: %w", err)
	}

	// 写入文件
	err = os.WriteFile(s.configFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入平台配置文件失败: %w", err)
	}

	return nil
}

// Update 更新平台信息
func (s *PlatformStore) Update(platform models.GamePlatform) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并更新平台
	found := false
	for i, p := range s.platforms {
		if p.ID == platform.ID {
			s.platforms[i] = platform
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("平台不存在: %s", platform.ID)
	}

	// 保存更改到文件
	return s.Save()
}

// Add 添加平台
func (s *PlatformStore) Add(platform models.GamePlatform) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查ID是否已存在
	for _, p := range s.platforms {
		if p.ID == platform.ID {
			return fmt.Errorf("平台ID已存在: %s", platform.ID)
		}
	}

	// 添加平台
	s.platforms = append(s.platforms, platform)

	// 保存更改到文件
	return s.Save()
}

// Delete 删除平台
func (s *PlatformStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除平台
	found := false
	for i, p := range s.platforms {
		if p.ID == id {
			// 从切片中移除元素
			s.platforms = append(s.platforms[:i], s.platforms[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("平台不存在: %s", id)
	}

	// 保存更改到文件
	return s.Save()
}
