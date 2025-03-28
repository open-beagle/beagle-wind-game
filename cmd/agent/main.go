package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	stdlog "log"

	"github.com/docker/docker/client"
	"github.com/open-beagle/beagle-wind-game/internal/event"
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
	"github.com/open-beagle/beagle-wind-game/internal/log"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func checkEnvironment() error {
	// 检查 Docker 环境
	if _, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
		return fmt.Errorf("Docker 环境检查失败: %v", err)
	}
	return nil
}

func main() {
	// 解析命令行参数
	nodeID := flag.String("node-id", "", "节点ID")
	serverAddr := flag.String("server-addr", "", "服务器地址")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Beagle Wind Game Agent v%s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		return
	}

	// 验证必要参数
	if *nodeID == "" {
		stdlog.Fatal("必须指定节点ID (--node-id)")
	}
	if *serverAddr == "" {
		stdlog.Fatal("必须指定服务器地址 (--server-addr)")
	}

	// 环境检查
	if err := checkEnvironment(); err != nil {
		stdlog.Fatalf("环境检查失败: %v", err)
	}

	// 组件初始化
	eventManager := event.NewDefaultEventManager()
	logManager := log.NewDefaultLogManager()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		stdlog.Printf("警告: 无法创建 Docker 客户端: %v", err)
		dockerClient = nil
	}
	config := gamenode.NewDefaultAgentConfig()

	// 创建 Agent
	agent := gamenode.NewGameNodeAgent(
		*nodeID,
		*serverAddr,
		eventManager,
		logManager,
		dockerClient,
		config,
	)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建错误通道
	errCh := make(chan error, 1)

	// 启动 Agent
	go func() {
		if err := agent.Start(ctx); err != nil {
			errCh <- fmt.Errorf("启动 Agent 失败: %v", err)
		}
	}()

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 处理信号或错误
	select {
	case err := <-errCh:
		stdlog.Printf("Agent 错误: %v", err)
	case sig := <-sigCh:
		stdlog.Printf("收到信号: %v", sig)
	}

	// 执行清理
	cancel()
	agent.Stop()
}
