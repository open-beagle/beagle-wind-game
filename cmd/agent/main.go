package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/open-beagle/beagle-wind-game/internal/grpc"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

var (
	serverAddr = flag.String("server", "localhost:50051", "gRPC server address")
	nodeID     = flag.String("id", "", "node ID")
)

func main() {
	flag.Parse()

	// 1. 初始化日志
	logger := utils.New("Agent")
	logger.Info("启动 Agent...")

	// 2. 参数验证
	if *nodeID == "" {
		logger.Fatal("节点ID不能为空")
	}

	// 3. 创建 Docker 客户端
	dockerClient, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		logger.Fatal("创建 Docker 客户端失败: %v", err)
	}

	// 4. 创建基础 Agent
	baseAgent, err := grpc.NewAgent(
		context.Background(),
		*nodeID,
		*serverAddr,
		dockerClient,
		&grpc.AgentOptions{
			HeartbeatPeriod: 5 * time.Second,
		},
	)
	if err != nil {
		logger.Fatal("创建基础 Agent 失败: %v", err)
	}

	// 5. 创建业务 Agent
	gameNodeAgent, err := grpc.NewGameNodeAgent(baseAgent)
	if err != nil {
		logger.Fatal("创建 GameNode Agent 失败: %v", err)
	}

	pipelineAgent := grpc.NewPipelineAgent(baseAgent)

	// 6. 注册节点
	if err := gameNodeAgent.Register(context.Background()); err != nil {
		logger.Fatal("注册节点失败: %v", err)
	}

	// 7. 启动 Agent
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := baseAgent.Start(ctx); err != nil {
		logger.Fatal("启动 Agent 失败: %v", err)
	}

	// 启动 Pipeline Agent
	if err := pipelineAgent.Start(ctx); err != nil {
		logger.Fatal("启动 Pipeline Agent 失败: %v", err)
	}

	// 8. 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 9. 停止 Agent
	baseAgent.Stop(ctx)
	logger.Info("Agent 已停止")
}
