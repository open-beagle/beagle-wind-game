package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	ggrpc "google.golang.org/grpc"

	"github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// GRPCServer 统一的 gRPC 服务器管理类
type GRPCServer struct {
	// 服务器配置
	config *ServerConfig

	// 日志框架
	logger utils.Logger

	// 子服务器
	nodeServer     *GameNodeServer
	pipelineServer *GamePipelineServer

	// 服务器实例
	server *ggrpc.Server

	// 同步原语
	mu   sync.RWMutex
	done chan struct{}
}

// NewGRPCServer 创建新的 gRPC 服务器
func NewGRPCServer(
	nodeService GameNodeServiceInterface,
	pipelineService GamePipelineServiceInterface,
	logger utils.Logger,
	config *ServerConfig,
) *GRPCServer {
	// 创建 gRPC 服务器实例
	server := ggrpc.NewServer()

	// 创建节点服务器
	nodeServer := NewGameNodeServer(
		nodeService,
		pipelineService,
		logger,
		config,
	)

	// 创建 Pipeline 服务器
	pipelineServer := NewGamePipelineServer(logger)

	// 注册服务
	proto.RegisterGameNodeGRPCServiceServer(server, nodeServer)
	proto.RegisterGamePipelineGRPCServiceServer(server, pipelineServer)

	return &GRPCServer{
		config:         config,
		logger:         logger,
		nodeServer:     nodeServer,
		pipelineServer: pipelineServer,
		server:         server,
		done:           make(chan struct{}),
	}
}

// Start 启动 gRPC 服务器
func (s *GRPCServer) Start(ctx context.Context, listenAddr string) error {
	// 创建监听器
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("监听地址失败: %v", err)
	}

	s.logger.Info("gRPC 服务器开始监听地址: %s", listenAddr)

	// 启动服务器
	go func() {
		if err := s.server.Serve(listener); err != nil {
			s.logger.Error("gRPC 服务器运行失败: %v", err)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()
	s.logger.Info("收到停止信号，正在关闭服务器")

	// 优雅关闭
	s.server.GracefulStop()
	return nil
}

// Stop 停止 gRPC 服务器
func (s *GRPCServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 关闭子服务器
	if s.nodeServer != nil {
		s.nodeServer.Stop()
	}
	if s.pipelineServer != nil {
		s.pipelineServer.Stop()
	}

	// 关闭主服务器
	if s.server != nil {
		s.server.GracefulStop()
	}

	close(s.done)
	return nil
}

// GetNodeServer 获取节点服务器实例
func (s *GRPCServer) GetNodeServer() *GameNodeServer {
	return s.nodeServer
}

// GetPipelineServer 获取 Pipeline 服务器实例
func (s *GRPCServer) GetPipelineServer() *GamePipelineServer {
	return s.pipelineServer
}
