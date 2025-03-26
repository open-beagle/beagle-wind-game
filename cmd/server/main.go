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

func main() {
	// 解析命令行参数
	grpcAddr := flag.String("grpc-addr", ":50051", "gRPC服务监听地址")
	httpAddr := flag.String("http-addr", ":8080", "HTTP服务监听地址")
	tlsCertFile := flag.String("tls-cert", "", "TLS证书文件路径")
	tlsKeyFile := flag.String("tls-key", "", "TLS密钥文件路径")
	flag.Parse()

	// 加载配置
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

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

	gameinstanceStore, err := store.NewGameInstanceStore("data/instances.yaml")
	if err != nil {
		log.Fatalf("创建实例存储失败: %v", err)
	}
	defer gameinstanceStore.Cleanup()

	// 创建服务实例
	gameplatformService := service.NewGamePlatformService(gameplatformStore)
	gamenodeService := service.NewGameNodeService(gamenodeStore)
	gameCardService := service.NewGameCardService(gameCardStore)
	gameinstanceService := service.NewGameInstanceService(gameinstanceStore)

	// 创建节点服务适配器
	nodeManager := gamenode.NewGameNodeManager(gamenodeService)

	// 创建服务器选项
	opts := gamenode.ServerOptions{
		ListenAddr:   *grpcAddr,
		TLSCertFile:  *tlsCertFile,
		TLSKeyFile:   *tlsKeyFile,
		MaxHeartbeat: 30 * time.Second,
	}

	// 创建 agent 服务器
	gamenodeServer := gamenode.NewGameNodeServer(opts, nodeManager)

	// 设置 HTTP 路由
	router := api.SetupRouter(gameplatformService, gamenodeService, gameCardService, gameinstanceService)

	// 启动 gRPC 服务器
	go func() {
		fmt.Printf("gRPC服务器正在监听 %s\n", *grpcAddr)
		if err := gamenodeServer.Start(); err != nil {
			log.Fatalf("gRPC服务器运行失败: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		fmt.Printf("HTTP服务器正在监听 %s\n", *httpAddr)
		if err := router.Run(*httpAddr); err != nil {
			log.Fatalf("HTTP服务器运行失败: %v", err)
		}
	}()

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 优雅关闭
	fmt.Println("正在关闭服务器...")
	gamenodeServer.Stop()
	fmt.Println("服务器已关闭")
}
