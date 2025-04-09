package store

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// GamePlatformStore 游戏平台存储接口
type GamePlatformStore interface {
	// List 获取所有平台
	List(ctx context.Context) ([]models.GamePlatform, error)
	// Get 获取指定ID的平台
	Get(ctx context.Context, id string) (models.GamePlatform, error)
	// Add 添加平台
	Add(ctx context.Context, platform models.GamePlatform) error
	// Update 更新平台信息
	Update(ctx context.Context, platform models.GamePlatform) error
	// Delete 删除平台
	Delete(ctx context.Context, id string) error
	// Load 加载平台数据
	Load(ctx context.Context) error
	// Save 保存平台数据
	Save(ctx context.Context) error
	// Cleanup 清理测试文件
	Cleanup(ctx context.Context) error
	// Close 关闭存储，确保所有待处理的保存操作完成
	Close()
}

// YAMLGamePlatformStore YAML文件存储实现
type YAMLGamePlatformStore struct {
	configFile string
	platforms  []models.GamePlatform
	mu         sync.RWMutex
	logger     utils.Logger
	yamlSaver  *utils.YAMLSaver
}

// NewGamePlatformStore 创建游戏平台存储
func NewGamePlatformStore(ctx context.Context, configFile string) (GamePlatformStore, error) {
	logger := utils.New("GamePlatformStore")

	store := &YAMLGamePlatformStore{
		configFile: configFile,
		platforms:  []models.GamePlatform{},
		logger:     logger,
	}

	// 创建YAML保存器，使用1秒的延迟保存
	store.yamlSaver = utils.NewYAMLSaver(
		configFile,
		func() interface{} {
			// 这个函数返回当前平台数据，在保存时调用
			store.mu.RLock()
			defer store.mu.RUnlock()
			return store.platforms
		},
		logger,
		utils.WithDelay(time.Second),
	)

	// 初始化加载数据
	err := store.Load(ctx)
	if err != nil {
		logger.Error("加载平台数据失败: %v", err)
		return nil, err
	}

	logger.Info("成功加载平台数据，共%d个平台", len(store.platforms))
	return store, nil
}

// Load 加载平台数据
func (s *YAMLGamePlatformStore) Load(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("加载游戏平台数据: %s", s.configFile)

	// 检查文件是否存在
	fileInfo, err := os.Stat(s.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，创建一个空的平台列表
			s.logger.Info("平台配置文件不存在，创建空列表: %s", s.configFile)
			s.platforms = []models.GamePlatform{}
			return s.Save(ctx)
		}
		s.logger.Error("读取平台配置文件失败: %v", err)
		return fmt.Errorf("存储层错误")
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("平台配置文件是一个目录: %s", s.configFile)
		return fmt.Errorf("存储层错误")
	}

	// 读取文件
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		s.logger.Error("读取平台配置文件失败: %v", err)
		return fmt.Errorf("存储层错误")
	}

	// 解析YAML
	var platforms []models.GamePlatform
	if len(data) > 0 {
		err = yaml.Unmarshal(data, &platforms)
		if err != nil {
			s.logger.Error("解析平台配置失败: %v", err)
			return fmt.Errorf("存储层错误")
		}
	} else {
		platforms = []models.GamePlatform{}
	}

	s.platforms = platforms
	s.logger.Debug("成功加载%d个游戏平台", len(platforms))
	return nil
}

// Save 保存平台数据
func (s *YAMLGamePlatformStore) Save(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	// 使用延迟保存器进行保存
	return s.yamlSaver.Save(ctx)
}

// List 获取所有平台
func (s *YAMLGamePlatformStore) List(ctx context.Context) ([]models.GamePlatform, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Debug("获取所有游戏平台")

	// 检查文件是否是目录
	fileInfo, err := os.Stat(s.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("平台配置文件不存在，返回空列表: %s", s.configFile)
			return []models.GamePlatform{}, nil
		}
		s.logger.Error("读取平台配置文件失败: %v", err)
		return nil, fmt.Errorf("存储层错误")
	}

	// 如果是目录，返回错误
	if fileInfo.IsDir() {
		s.logger.Error("平台配置文件是一个目录: %s", s.configFile)
		return nil, fmt.Errorf("存储层错误")
	}
	// 创建副本避免修改原始数据
	platforms := make([]models.GamePlatform, len(s.platforms))
	copy(platforms, s.platforms)
	return platforms, nil
}

// Get 获取指定ID的平台
func (s *YAMLGamePlatformStore) Get(ctx context.Context, id string) (models.GamePlatform, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return models.GamePlatform{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Debug("获取平台，ID: %s", id)

	for _, platform := range s.platforms {
		if platform.ID == id {
			s.logger.Debug("找到平台: %s", id)
			return platform, nil
		}
	}

	s.logger.Warn("未找到平台: %s", id)
	return models.GamePlatform{}, nil
}

// Add 添加平台
func (s *YAMLGamePlatformStore) Add(ctx context.Context, platform models.GamePlatform) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("添加平台，ID: %s", platform.ID)

	// 检查ID是否已存在
	for _, p := range s.platforms {
		if p.ID == platform.ID {
			s.logger.Warn("平台已存在，无法添加: %s", platform.ID)
			return fmt.Errorf("存储层错误")
		}
	}

	// 添加平台
	s.platforms = append(s.platforms, platform)

	// 保存更改到文件
	err := s.Save(ctx)
	if err != nil {
		s.logger.Error("保存平台数据失败: %v", err)
		return err
	}

	s.logger.Info("平台添加成功，ID: %s", platform.ID)
	return nil
}

// Update 更新平台
func (s *YAMLGamePlatformStore) Update(ctx context.Context, platform models.GamePlatform) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("更新平台，ID: %s", platform.ID)

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
		s.logger.Warn("平台不存在，无法更新: %s", platform.ID)
		return fmt.Errorf("存储层错误")
	}

	// 保存更改到文件
	err := s.Save(ctx)
	if err != nil {
		s.logger.Error("保存平台数据失败: %v", err)
		return err
	}

	s.logger.Info("平台更新成功，ID: %s", platform.ID)
	return nil
}

// Delete 删除平台
func (s *YAMLGamePlatformStore) Delete(ctx context.Context, id string) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("删除平台，ID: %s", id)

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
		s.logger.Warn("平台不存在，无法删除: %s", id)
		return fmt.Errorf("存储层错误")
	}

	// 保存更改到文件
	err := s.Save(ctx)
	if err != nil {
		s.logger.Error("保存平台数据失败: %v", err)
		return err
	}

	s.logger.Info("平台删除成功，ID: %s", id)
	return nil
}

// Cleanup 清理测试文件
func (s *YAMLGamePlatformStore) Cleanup(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.logger.Info("清理存储文件: %s", s.configFile)
	err := os.Remove(s.configFile)
	if err != nil && !os.IsNotExist(err) {
		s.logger.Error("清理存储文件失败: %v", err)
		return err
	}
	s.logger.Info("存储文件清理成功")
	return nil
}

// Close 关闭存储，确保所有待处理的保存操作完成
func (s *YAMLGamePlatformStore) Close() {
	s.logger.Info("关闭GamePlatformStore，确保数据保存...")
	if s.yamlSaver != nil {
		s.yamlSaver.Close()
	}
}
