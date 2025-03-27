package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
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
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 验证必要参数
	if *nodeID == "" {
		log.Fatal("必须指定节点ID (--node-id)")
	}
	if *nodeName == "" {
		log.Fatal("必须指定节点名称 (--node-name)")
	}

	// 创建 Docker 客户端
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("警告: 无法创建 Docker 客户端: %v", err)
		dockerClient = nil
	}

	// 创建 GameNodeAgent 实例
	agent := gamenode.NewGameNodeAgent(*nodeID, dockerClient)
	if agent == nil {
		log.Fatal("创建 GameNodeAgent 失败")
	}

	// 创建错误通道
	errCh := make(chan error, 1)

	// 启动 Agent
	go func() {
		if err := agent.Start(); err != nil {
			errCh <- fmt.Errorf("启动 Agent 失败: %v", err)
		}
	}()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case sig := <-sigCh:
		log.Printf("收到信号: %v\n", sig)
	case err := <-errCh:
		log.Printf("Agent 错误: %v\n", err)
	}

	// 优雅关闭
	log.Println("正在关闭 Agent...")

	// 停止 Agent
	if err := agent.Stop(); err != nil {
		log.Printf("关闭 Agent 时出错: %v\n", err)
	}

	log.Println("Agent 已关闭")
}
