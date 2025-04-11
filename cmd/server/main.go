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
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
	"github.com/open-beagle/beagle-wind-game/internal/service"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

// initStores 初始化所有存储
func initStores(ctx context.Context) (store.GameNodeStore, *store.YAMLGameNodePipelineStore, store.GamePlatformStore, store.GameCardStore, store.GameInstanceStore, error) {
	// 初始化游戏节点存储
	gamenodeStore, err := store.NewGameNodeStore(ctx, "data/gamenodes.yaml")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建节点存储失败: %v", err)
	}

	// 初始化游戏节点流水线存储
	gamenodePipelineStore := store.NewYAMLGameNodePipelineStore(ctx, "data/game-pipelines.yaml")

	// 初始化游戏平台存储
	gamePlatformStore, err := store.NewGamePlatformStore(ctx, "config/gameplatforms.yaml")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建平台存储失败: %v", err)
	}

	// 初始化游戏卡牌存储
	gameCardStore, err := store.NewGameCardStore(ctx, "data/gamecards.yaml")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建卡牌存储失败: %v", err)
	}

	// 创建logger
	logger := utils.New("GameInstanceStore")

	// 初始化游戏实例存储
	gameInstanceStore, err := store.NewGameInstanceStore(ctx, "data/gameinstances.yaml", logger)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("创建实例存储失败: %v", err)
	}

	return gamenodeStore, gamenodePipelineStore, gamePlatformStore, gameCardStore, gameInstanceStore, nil
}

func main() {
	// 解析命令行参数
	httpAddr := flag.String("http", ":8080", "HTTP服务器监听地址")
	grpcAddr := flag.String("grpc", ":50051", "gRPC服务器监听地址")
	logLevel := flag.String("log-level", "INFO", "日志级别: DEBUG, INFO, WARN, ERROR, FATAL (用于gRPC服务)")
	logFile := flag.String("log-file", "", "日志文件路径, 为空则只输出到控制台 (用于gRPC服务)")
	logBoth := flag.Bool("log-both", false, "是否同时输出到文件和控制台 (用于gRPC服务)")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Beagle Wind Game Server v%s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		return
	}

	// 初始化日志框架（仅用于gRPC服务）
	logLevelMap := map[string]utils.LogLevel{
		"DEBUG": utils.DEBUG,
		"INFO":  utils.INFO,
		"WARN":  utils.WARN,
		"ERROR": utils.ERROR,
		"FATAL": utils.FATAL,
	}
	level, ok := logLevelMap[*logLevel]
	if !ok {
		level = utils.INFO
	}
	utils.InitLogger(*logFile, level, *logBoth)
	logger := utils.New("gRPC")

	// 创建错误通道
	errCh := make(chan error, 1)

	// 初始化存储
	logger.Info("初始化存储...")
	gamenodeStore, gamenodePipelineStore, gamePlatformStore, gameCardStore, gameInstanceStore, err := initStores(context.Background())
	if err != nil {
		logger.Fatal("初始化存储失败: %v", err)
	}
	logger.Info("存储初始化完成")

	// 创建服务实例
	logger.Info("创建服务实例...")
	nodeService := service.NewGameNodeService(gamenodeStore)
	pipelineService := service.NewGameNodePipelineService(gamenodePipelineStore)
	platformService := service.NewGamePlatformService(gamePlatformStore)
	cardService := service.NewGameCardService(gameCardStore)
	instanceService := service.NewGameInstanceService(gameInstanceStore)

	// 创建 agent 服务器
	logger.Info("创建 gRPC 服务器...")
	gamenodeServer := gamenode.NewGameNodeServer(
		nodeService,
		pipelineService,
		logger,
		&gamenode.ServerConfig{
			MaxConnections:  100,
			HeartbeatPeriod: time.Second * 30,
		},
	)

	// 设置 HTTP 路由
	router := gin.Default()

	// 注册路由处理器
	gamenodeHandler := api.NewGameNodeHandler(nodeService)
	gamenodeHandler.RegisterRoutes(router)

	// TODO: 其他服务的路由处理器将在实现后添加
	_ = platformService // 避免未使用变量警告
	_ = cardService     // 避免未使用变量警告
	_ = instanceService // 避免未使用变量警告

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 HTTP 服务器
	go func() {
		logger.Info("HTTP服务器开始监听 %s", *httpAddr)
		if err := router.Run(*httpAddr); err != nil {
			errCh <- fmt.Errorf("HTTP服务器运行失败: %v", err)
		}
	}()

	// 启动 gRPC 服务器
	go func() {
		logger.Info("gRPC服务器开始监听 %s", *grpcAddr)
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
		logger.Error("服务器错误: %v", err)
		os.Exit(1)
	case sig := <-sigCh:
		logger.Info("收到信号: %v", sig)
	}

	// 优雅关闭
	logger.Info("正在关闭服务器...")

	// 取消上下文
	cancel()

	// 关闭存储层，确保数据保存
	logger.Info("正在关闭GameNodeStore...")
	if closer, ok := gamenodeStore.(interface{ Close() }); ok {
		closer.Close()
	}

	logger.Info("正在关闭GameCardStore...")
	if closer, ok := gameCardStore.(interface{ Close() }); ok {
		closer.Close()
	}

	logger.Info("正在关闭GameInstanceStore...")
	if closer, ok := gameInstanceStore.(interface{ Close() }); ok {
		closer.Close()
	}

	logger.Info("正在关闭GamePlatformStore...")
	if closer, ok := gamePlatformStore.(interface{ Close() }); ok {
		closer.Close()
	}

	logger.Info("正在关闭GameNodePipelineStore...")
	gamenodePipelineStore.Close()

	// 等待一段时间让服务器完成关闭
	time.Sleep(5 * time.Second)
	logger.Info("服务器已关闭")

	// 确保日志写入磁盘
	if logger, ok := logger.(interface{ Sync() error }); ok {
		logger.Sync()
	}
}
