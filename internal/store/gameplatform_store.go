package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"gopkg.in/yaml.v3"
)

// GamePlatformStore 游戏平台存储接口
type GamePlatformStore interface {
	// List 获取所有平台
	List() ([]models.GamePlatform, error)
	// Get 获取指定ID的平台
	Get(id string) (models.GamePlatform, error)
	// Add 添加平台
	Add(platform models.GamePlatform) error
	// Update 更新平台信息
	Update(platform models.GamePlatform) error
	// Delete 删除平台
	Delete(id string) error
	// Cleanup 清理测试文件
	Cleanup() error
}

// YAMLGamePlatformStore YAML文件存储实现
type YAMLGamePlatformStore struct {
	configFile string
	platforms  []models.GamePlatform
	mu         sync.RWMutex
}

// NewGamePlatformStore 创建游戏平台存储
func NewGamePlatformStore(configFile string) (GamePlatformStore, error) {
	store := &YAMLGamePlatformStore{
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
func (s *YAMLGamePlatformStore) Load() error {
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

// Save 保存平台配置到文件
func (s *YAMLGamePlatformStore) Save() error {
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

// List 获取所有平台
func (s *YAMLGamePlatformStore) List() ([]models.GamePlatform, error) {
	// 检查是否已加载数据
	if s.platforms == nil {
		return nil, fmt.Errorf("存储层错误")
	}
	// 创建副本避免修改原始数据
	platforms := make([]models.GamePlatform, len(s.platforms))
	copy(platforms, s.platforms)
	return platforms, nil
}

// Get 获取指定ID的平台
func (s *YAMLGamePlatformStore) Get(id string) (models.GamePlatform, error) {
	for _, platform := range s.platforms {
		if platform.ID == id {
			return platform, nil
		}
	}
	return models.GamePlatform{}, nil
}

// Add 添加平台
func (s *YAMLGamePlatformStore) Add(platform models.GamePlatform) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查ID是否已存在
	for _, p := range s.platforms {
		if p.ID == platform.ID {
			return fmt.Errorf("存储层错误")
		}
	}

	// 添加平台
	s.platforms = append(s.platforms, platform)

	// 保存更改到文件
	return s.Save()
}

// Update 更新平台
func (s *YAMLGamePlatformStore) Update(platform models.GamePlatform) error {
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
		return fmt.Errorf("存储层错误")
	}

	// 保存更改到文件
	return s.Save()
}

// Delete 删除平台
func (s *YAMLGamePlatformStore) Delete(id string) error {
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
		return fmt.Errorf("存储层错误")
	}

	// 保存更改到文件
	return s.Save()
}

// Cleanup 清理测试文件
func (s *YAMLGamePlatformStore) Cleanup() error {
	return os.Remove(s.configFile)
}
