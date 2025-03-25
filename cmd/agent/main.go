package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/open-beagle/beagle-wind-game/internal/agent"
)

func main() {
	// 解析命令行参数
	serverAddr := flag.String("server", "localhost:50051", "服务器地址")
	flag.Parse()

	// 创建 Agent 实例
	a, err := agent.NewAgent(*serverAddr)
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 Agent
	if err := a.Start(ctx); err != nil {
		log.Fatalf("启动 Agent 失败: %v", err)
	}

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 优雅关闭
	a.Stop()
}
