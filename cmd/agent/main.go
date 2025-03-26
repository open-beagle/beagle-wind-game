package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
)

func main() {
	// 解析命令行参数
	serverAddr := flag.String("server", "localhost:50051", "服务器地址")
	flag.Parse()

	// 创建 Docker 客户端
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Printf("警告: 无法创建 Docker 客户端: %v", err)
		dockerClient = nil
	}

	// 创建 GameNodeAgent 实例
	agent := gamenode.NewGameNodeAgent(*serverAddr, dockerClient)
	if agent == nil {
		log.Fatal("创建 GameNodeAgent 失败")
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 Agent
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("启动 Agent 失败: %v", err)
	}

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 优雅关闭
	log.Println("正在关闭 Agent...")
	agent.Stop()
}
