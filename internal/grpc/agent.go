package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	dockerclient "github.com/docker/docker/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// AgentOptions 配置选项
type AgentOptions struct {
	HeartbeatPeriod time.Duration
}

// Agent 表示基础 Agent
type Agent struct {
	mu     sync.RWMutex
	conn   *grpc.ClientConn
	logger utils.Logger

	// 基本信息
	id         string
	serverAddr string

	// Docker 客户端
	dockerClient *dockerclient.Client

	// 配置
	opts *AgentOptions

	stopChan  chan struct{}
	isRunning bool

	// 服务客户端
	gameNodeClient proto.GameNodeGRPCServiceClient
	pipelineClient proto.GamePipelineGRPCServiceClient
}

// NewAgent 创建一个新的 Agent
func NewAgent(
	ctx context.Context,
	id string,
	serverAddr string,
	dockerClient *dockerclient.Client,
	opts *AgentOptions,
) (*Agent, error) {
	agent := &Agent{
		id:           id,
		serverAddr:   serverAddr,
		dockerClient: dockerClient,
		opts:         opts,
		logger:       utils.New("Agent"),
		stopChan:     make(chan struct{}),
		isRunning:    false,
	}

	// 建立连接
	if err := agent.connect(ctx); err != nil {
		return nil, fmt.Errorf("connect failed: %v", err)
	}

	// 初始化服务客户端
	agent.initClients()

	return agent, nil
}

// initClients 初始化服务客户端
func (a *Agent) initClients() {
	a.gameNodeClient = proto.NewGameNodeGRPCServiceClient(a.conn)
	a.pipelineClient = proto.NewGamePipelineGRPCServiceClient(a.conn)
}

// connect 建立 gRPC 连接
func (a *Agent) connect(ctx context.Context) error {
	a.logger.Debug("开始连接到服务器: %s", a.serverAddr)

	// 使用 grpc.NewClient 创建连接
	conn, err := grpc.NewClient(
		a.serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		a.logger.Error("连接服务器失败: %v", err)
		return fmt.Errorf("连接服务器失败: %v", err)
	}

	// 检查连接状态
	state := conn.GetState()
	if state != connectivity.Ready {
		// 尝试连接
		conn.Connect()
		// 等待状态变化
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}

	a.conn = conn
	a.logger.Info("成功连接到服务器")

	return nil
}

// Start 启动 Agent
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return nil
	}

	a.isRunning = true

	// 启动主循环
	go a.run(ctx)

	return nil
}

// Stop 停止 Agent
func (a *Agent) Stop(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return
	}

	close(a.stopChan)
	if a.conn != nil {
		a.conn.Close()
	}
	a.isRunning = false
}

// run 运行 Agent 的主循环
func (a *Agent) run(ctx context.Context) {
	heartbeatPeriod := 5 * time.Second
	if a.opts != nil && a.opts.HeartbeatPeriod > 0 {
		heartbeatPeriod = a.opts.HeartbeatPeriod
	}

	ticker := time.NewTicker(heartbeatPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopChan:
			return
		case <-ticker.C:
			// 发送心跳
			if err := a.sendHeartbeat(ctx); err != nil {
				a.logger.Error("Failed to send heartbeat: %v", err)
			}
		}
	}
}

// sendHeartbeat 发送心跳消息
func (a *Agent) sendHeartbeat(ctx context.Context) error {
	// 基础 Agent 不实现具体的心跳逻辑
	return nil
}

// GetGameNodeClient 获取 GameNode 服务客户端
func (a *Agent) GetGameNodeClient() proto.GameNodeGRPCServiceClient {
	return a.gameNodeClient
}

// GetPipelineClient 获取 Pipeline 服务客户端
func (a *Agent) GetPipelineClient() proto.GamePipelineGRPCServiceClient {
	return a.pipelineClient
}

// GetLogger 获取日志记录器
func (a *Agent) GetLogger() utils.Logger {
	return a.logger
}

// GetDockerClient 获取 Docker 客户端
func (a *Agent) GetDockerClient() *dockerclient.Client {
	return a.dockerClient
}

// IsRunning 检查 Agent 是否在运行
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}
