package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/api"
	"github.com/open-beagle/beagle-wind-game/internal/config"
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
	"github.com/open-beagle/beagle-wind-game/internal/service"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// 解析命令行参数
	grpcAddr := flag.String("grpc-addr", ":50051", "gRPC服务监听地址")
	httpAddr := flag.String("http-addr", ":8080", "HTTP服务监听地址")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Beagle Wind Game Server v%s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		return
	}

	// 加载配置
	if _, err := config.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 配置日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 创建存储实例
	gameplatformStore, err := store.NewGamePlatformStore("config/platforms.yaml")
	if err != nil {
		log.Fatalf("创建平台存储失败: %v", err)
	}
	defer gameplatformStore.Cleanup()

	gamenodeStore, err := store.NewGameNodeStore("data/nodes.yaml")
	if err != nil {
		log.Fatalf("创建节点存储失败: %v", err)
	}
	defer gamenodeStore.Cleanup()

	gameCardStore, err := store.NewGameCardStore("data/game_cards.yaml")
	if err != nil {
		log.Fatalf("创建游戏卡片存储失败: %v", err)
	}
	defer gameCardStore.Cleanup()

	gameinstanceStore, err := store.NewGameInstanceStore("data/instances.yaml")
	if err != nil {
		log.Fatalf("创建实例存储失败: %v", err)
	}
	defer gameinstanceStore.Cleanup()

	gamenodePipelineStore, err := store.NewGameNodePipelineStore("data/pipelines.yaml")
	if err != nil {
		log.Fatalf("创建流水线存储失败: %v", err)
	}
	defer gamenodePipelineStore.Cleanup()

	// 创建服务实例
	gameplatformService := service.NewGamePlatformService(gameplatformStore)
	gamenodeService := service.NewGameNodeService(gamenodeStore)
	gameCardService := service.NewGameCardService(gameCardStore)
	gameinstanceService := service.NewGameInstanceService(gameinstanceStore)
	gamenodePipelineService := service.NewGameNodePipelineService(gamenodePipelineStore)

	// 创建服务器配置
	serverConfig := &gamenode.ServerConfig{
		MaxNodes:        100,
		SessionTimeout:  5 * time.Minute,
		HeartbeatPeriod: 30 * time.Second,
	}

	// 创建 agent 服务器
	gamenodeServer, err := gamenode.NewGameNodeServer(serverConfig)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 设置 HTTP 路由
	router := api.SetupRouter(gameplatformService, gamenodeService, gameCardService, gameinstanceService, gamenodePipelineService)

	// 创建错误通道
	errCh := make(chan error, 2)

	// 启动 gRPC 服务器
	go func() {
		log.Printf("gRPC服务器正在监听 %s\n", *grpcAddr)
		if err := gamenodeServer.Start(*grpcAddr); err != nil {
			errCh <- fmt.Errorf("gRPC服务器运行失败: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		log.Printf("HTTP服务器正在监听 %s\n", *httpAddr)
		if err := router.Run(*httpAddr); err != nil {
			errCh <- fmt.Errorf("HTTP服务器运行失败: %v", err)
		}
	}()

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case sig := <-sigCh:
		log.Printf("收到信号: %v\n", sig)
	case err := <-errCh:
		log.Printf("服务器错误: %v\n", err)
	}

	// 优雅关闭
	log.Println("正在关闭服务器...")

	// 关闭 gRPC 服务器
	gamenodeServer.Stop()

	// 关闭存储
	gameplatformStore.Cleanup()
	gamenodeStore.Cleanup()
	gameCardStore.Cleanup()
	gameinstanceStore.Cleanup()
	gamenodePipelineStore.Cleanup()

	log.Println("服务器已关闭")
}
