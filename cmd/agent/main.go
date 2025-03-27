package main

import (
	"context"
	"flag"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/open-beagle/beagle-wind-game/internal/event"
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
	"github.com/open-beagle/beagle-wind-game/internal/log"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// 解析命令行参数
	nodeID := flag.String("node-id", "", "节点ID")
	nodeName := flag.String("node-name", "", "节点名称")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Beagle Wind Game Agent v%s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		return
	}

	// 配置日志
	stdlog.SetFlags(stdlog.Ldate | stdlog.Ltime | stdlog.Lshortfile)

	// 验证必要参数
	if *nodeID == "" {
		stdlog.Fatal("必须指定节点ID (--node-id)")
	}
	if *nodeName == "" {
		stdlog.Fatal("必须指定节点名称 (--node-name)")
	}

	// 创建 Docker 客户端
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		stdlog.Printf("警告: 无法创建 Docker 客户端: %v", err)
		dockerClient = nil
	}

	// 创建 GameNodeAgent 实例
	agent := gamenode.NewGameNodeAgent(
		*nodeID,
		*nodeName,
		"", // serverAddr
		"", // grpcAddr
		"", // logAddr
		event.NewDefaultEventManager(),
		log.NewDefaultLogManager(),
		dockerClient,
		&gamenode.AgentConfig{},
	)
	if agent == nil {
		stdlog.Fatal("创建 GameNodeAgent 失败")
	}

	// 创建错误通道
	errCh := make(chan error, 1)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 Agent
	go func() {
		if err := agent.Start(ctx); err != nil {
			errCh <- fmt.Errorf("启动 Agent 失败: %v", err)
		}
	}()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case sig := <-sigCh:
		stdlog.Printf("收到信号: %v\n", sig)
	case err := <-errCh:
		stdlog.Printf("Agent 错误: %v\n", err)
	}

	// 优雅关闭
	stdlog.Println("正在关闭 Agent...")

	// 停止 Agent
	agent.Stop()

	stdlog.Println("Agent 已关闭")
}
