package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/open-beagle/beagle-wind-game/internal/api"
	"github.com/open-beagle/beagle-wind-game/internal/config"
	"github.com/open-beagle/beagle-wind-game/internal/service"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/open-beagle/beagle-wind-game/pkg/middleware"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化存储
	platformStore, nodeStore, cardStore, instanceStore, err := initStores(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize stores: %v", err)
	}

	// 初始化服务
	platformService := service.NewPlatformService(platformStore)
	nodeService := service.NewNodeService(nodeStore, instanceStore)
	gameCardService := service.NewGameCardService(cardStore, platformStore, instanceStore)
	instanceService := service.NewInstanceService(instanceStore, nodeStore, cardStore, platformStore)

	// 设置路由
	r := api.SetupRouter(platformService, nodeService, gameCardService, instanceService)

	// 添加中间件
	r.Use(middleware.Logger())

	// 启动服务器
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initStores 初始化所有存储
func initStores(cfg *config.Config) (*store.PlatformStore, *store.NodeStore, *store.GameCardStore, *store.InstanceStore, error) {
	// 初始化平台存储
	platformStore, err := store.NewPlatformStore(filepath.Join("config", "platforms.yaml"))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("初始化平台存储失败: %w", err)
	}

	// 初始化节点存储
	nodeStore, err := store.NewNodeStore(filepath.Join("data", "game-nodes.yaml"))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("初始化节点存储失败: %w", err)
	}

	// 初始化游戏卡片存储
	cardStore, err := store.NewGameCardStore(filepath.Join("data", "game-cards.yaml"))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("初始化游戏卡片存储失败: %w", err)
	}

	// 初始化游戏实例存储
	instanceStore, err := store.NewInstanceStore(filepath.Join("data", "game-instances.yaml"))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("初始化游戏实例存储失败: %w", err)
	}

	return platformStore, nodeStore, cardStore, instanceStore, nil
}
