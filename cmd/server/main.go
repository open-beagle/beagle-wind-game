package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/open-beagle/beagle-wind-game/internal/api"
	"github.com/open-beagle/beagle-wind-game/internal/event"
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
	"github.com/open-beagle/beagle-wind-game/internal/log"
	"github.com/open-beagle/beagle-wind-game/internal/service"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

// initStores 初始化所有存储
func initStores() (store.GameNodeStore, *store.YAMLGameNodePipelineStore, store.GamePlatformStore, store.GameCardStore, store.GameInstanceStore, error) {
	// 初始化游戏节点存储
	gamenodeStore, err := store.NewGameNodeStore("data/game-nodes.yaml")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建节点存储失败: %v", err)
	}

	// 初始化游戏节点流水线存储
	gamenodePipelineStore := store.NewYAMLGameNodePipelineStore("data/game-pipelines.yaml")

	// 初始化游戏平台存储
	gamePlatformStore, err := store.NewGamePlatformStore("config/platforms.yaml")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建平台存储失败: %v", err)
	}

	// 初始化游戏卡牌存储
	gameCardStore, err := store.NewGameCardStore("data/game-cards.yaml")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建卡牌存储失败: %v", err)
	}

	// 初始化游戏实例存储
	gameInstanceStore, err := store.NewGameInstanceStore("data/game-instances.yaml")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建实例存储失败: %v", err)
	}

	return gamenodeStore, gamenodePipelineStore, gamePlatformStore, gameCardStore, gameInstanceStore, nil
}

func main() {
	// 解析命令行参数
	httpAddr := flag.String("http", ":8080", "HTTP服务器监听地址")
	grpcAddr := flag.String("grpc", ":50051", "gRPC服务器监听地址")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Beagle Wind Game Server v%s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		return
	}

	// 创建错误通道
	errCh := make(chan error, 1)

	// 初始化存储
	gamenodeStore, gamenodePipelineStore, gamePlatformStore, gameCardStore, gameInstanceStore, err := initStores()
	if err != nil {
		fmt.Printf("初始化存储失败: %v\n", err)
		os.Exit(1)
	}
	defer gamenodeStore.Cleanup()
	defer gamenodePipelineStore.Cleanup()
	defer gamePlatformStore.Cleanup()
	defer gameCardStore.Cleanup()
	defer gameInstanceStore.Cleanup()

	// 创建服务实例
	nodeService := service.NewGameNodeService(gamenodeStore)
	pipelineService := service.NewGameNodePipelineService(gamenodePipelineStore)
	// TODO: 这些服务将在后续功能中使用
	// platformService := service.NewGamePlatformService(gamePlatformStore)
	// cardService := service.NewGameCardService(gameCardStore)
	// instanceService := service.NewGameInstanceService(gameInstanceStore)

	// 创建事件管理器
	eventManager := event.NewDefaultEventManager()

	// 创建日志管理器
	logManager := log.NewDefaultLogManager()

	// 创建 agent 服务器
	gamenodeServer, err := gamenode.NewGameNodeServer(
		nodeService,
		pipelineService,
		eventManager,
		logManager,
		&gamenode.ServerConfig{
			MaxConnections:  100,
			HeartbeatPeriod: time.Second * 30,
		},
	)
	if err != nil {
		fmt.Printf("创建服务器失败: %v\n", err)
		os.Exit(1)
	}

	// 设置 HTTP 路由
	router := gin.Default()
	gamenodeHandler := api.NewGameNodeHandler(nodeService)
	gamenodeHandler.RegisterRoutes(router)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 HTTP 服务器
	go func() {
		fmt.Printf("HTTP服务器正在监听 %s\n", *httpAddr)
		if err := router.Run(*httpAddr); err != nil {
			errCh <- fmt.Errorf("HTTP服务器运行失败: %v", err)
		}
	}()

	// 启动 gRPC 服务器
	go func() {
		fmt.Printf("gRPC服务器正在监听 %s\n", *grpcAddr)
		opts := &gamenode.GameNodeServerOptions{
			ListenAddr: *grpcAddr,
		}
		if err := gamenodeServer.Start(ctx, opts); err != nil {
			errCh <- fmt.Errorf("gRPC服务器运行失败: %v", err)
		}
	}()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待错误或中断信号
	select {
	case err := <-errCh:
		fmt.Printf("服务器错误: %v\n", err)
		os.Exit(1)
	case sig := <-sigCh:
		fmt.Printf("收到信号: %v\n", sig)
	}

	// 优雅关闭
	fmt.Println("正在关闭服务器...")

	// 取消上下文
	cancel()

	// 等待一段时间让服务器完成关闭
	time.Sleep(5 * time.Second)
	fmt.Println("服务器已关闭")
}
