package gamenode

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// GameNodeServerOptions 包含服务器配置选项
type GameNodeServerOptions struct {
	ListenAddr  string
	TLSCertFile string
	TLSKeyFile  string
}

// GameNodeServer 实现游戏节点服务器
type GameNodeServer struct {
	pb.UnimplementedGameNodeServiceServer
	opts        GameNodeServerOptions
	nodes       map[string]*nodeConnection
	pipelines   map[string]*GameNodePipeline
	nodeManager GameNodeManager
	mu          sync.RWMutex
}

type nodeConnection struct {
	info      *pb.NodeInfo
	lastSeen  time.Time
	status    string
	metrics   *pb.NodeMetrics
	sessionID string
}

// NewGameNodeServer 创建新的游戏节点服务器实例
func NewGameNodeServer(opts GameNodeServerOptions, manager GameNodeManager) *GameNodeServer {
	return &GameNodeServer{
		opts:        opts,
		nodes:       make(map[string]*nodeConnection),
		pipelines:   make(map[string]*GameNodePipeline),
		nodeManager: manager,
	}
}

// Start 启动服务器
func (s *GameNodeServer) Start() error {
	lis, err := net.Listen("tcp", s.opts.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// 配置gRPC服务器选项
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
	}

	// TODO: 如果提供了TLS证书，添加TLS选项

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGameNodeServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	// 启动节点状态监控
	go s.monitorNodeStatus()

	return grpcServer.Serve(lis)
}

// Register 实现节点注册
func (s *GameNodeServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证节点ID
	if req.NodeId == "" {
		return &pb.RegisterResponse{
			Success: false,
			Message: "node ID is required",
		}, nil
	}

	// 创建或更新节点连接信息
	sessionID := fmt.Sprintf("session-%s-%d", req.NodeId, time.Now().Unix())
	s.nodes[req.NodeId] = &nodeConnection{
		info:      req.NodeInfo,
		lastSeen:  time.Now(),
		status:    "ACTIVE",
		sessionID: sessionID,
	}

	// 通知节点管理器
	if err := s.nodeManager.RegisterNode(ctx, req.NodeId, req.NodeInfo); err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("failed to register node: %v", err),
		}, nil
	}

	return &pb.RegisterResponse{
		Success:   true,
		Message:   "registration successful",
		SessionId: sessionID,
	}, nil
}

// Heartbeat 实现心跳检测
func (s *GameNodeServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, exists := s.nodes[req.NodeId]
	if !exists {
		return nil, fmt.Errorf("node not registered")
	}

	if node.sessionID != req.SessionId {
		return nil, fmt.Errorf("invalid session ID")
	}

	node.lastSeen = time.Now()
	node.metrics = req.Metrics

	// 更新节点状态
	if err := s.nodeManager.UpdateNodeStatus(ctx, req.NodeId, "ACTIVE"); err != nil {
		return nil, fmt.Errorf("failed to update node status: %v", err)
	}

	// 更新节点指标
	if err := s.nodeManager.UpdateNodeMetrics(ctx, req.NodeId, req.Metrics); err != nil {
		return nil, fmt.Errorf("failed to update node metrics: %v", err)
	}

	return &pb.HeartbeatResponse{
		Success:    true,
		ServerTime: timestamppb.Now(),
	}, nil
}

// ExecutePipeline 执行流水线
func (s *GameNodeServer) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) (*pb.ExecutePipelineResponse, error) {
	s.mu.RLock()
	node, exists := s.nodes[req.NodeId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node not registered")
	}

	// 创建流水线
	pipeline := NewGameNodePipeline(req.Pipeline, node.dockerClient)
	s.mu.Lock()
	s.pipelines[req.PipelineId] = pipeline
	s.mu.Unlock()

	// 执行流水线
	go func() {
		if err := pipeline.Execute(ctx); err != nil {
			// TODO: 处理错误
			return
		}
	}()

	return &pb.ExecutePipelineResponse{
		ExecutionId: req.PipelineId,
		Accepted:    true,
		Message:     "pipeline execution started",
	}, nil
}

// GetPipelineStatus 获取流水线状态
func (s *GameNodeServer) GetPipelineStatus(ctx context.Context, req *pb.PipelineStatusRequest) (*pb.PipelineStatusResponse, error) {
	s.mu.RLock()
	pipeline, exists := s.pipelines[req.ExecutionId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pipeline not found")
	}

	return pipeline.GetStatus(), nil
}

// CancelPipeline 取消流水线执行
func (s *GameNodeServer) CancelPipeline(ctx context.Context, req *pb.PipelineCancelRequest) (*pb.PipelineCancelResponse, error) {
	s.mu.RLock()
	pipeline, exists := s.pipelines[req.ExecutionId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pipeline not found")
	}

	if err := pipeline.Cancel(); err != nil {
		return &pb.PipelineCancelResponse{
			Success: false,
			Message: fmt.Sprintf("failed to cancel pipeline: %v", err),
		}, nil
	}

	return &pb.PipelineCancelResponse{
		Success: true,
		Message: "pipeline cancelled successfully",
	}, nil
}

// GetNodeMetrics 获取节点指标
func (s *GameNodeServer) GetNodeMetrics(ctx context.Context, req *pb.NodeMetricsRequest) (*pb.NodeMetricsResponse, error) {
	s.mu.RLock()
	node, exists := s.nodes[req.NodeId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node not found")
	}

	return &pb.NodeMetricsResponse{
		NodeId:           req.NodeId,
		Metrics:          node.metrics,
		ContainerMetrics: nil, // TODO: 实现容器指标收集
	}, nil
}

// StreamNodeLogs 流式获取节点日志
func (s *GameNodeServer) StreamNodeLogs(req *pb.NodeLogsRequest, stream pb.GameNodeService_StreamNodeLogsServer) error {
	// TODO: 实现节点日志流
	return fmt.Errorf("not implemented")
}

// StreamContainerLogs 流式获取容器日志
func (s *GameNodeServer) StreamContainerLogs(req *pb.ContainerLogsRequest, stream pb.GameNodeService_StreamContainerLogsServer) error {
	// TODO: 实现容器日志流
	return fmt.Errorf("not implemented")
}

// SubscribeEvents 订阅事件
func (s *GameNodeServer) SubscribeEvents(req *pb.EventSubscriptionRequest, stream pb.GameNodeService_SubscribeEventsServer) error {
	// TODO: 实现事件订阅
	return fmt.Errorf("not implemented")
}

// monitorNodeStatus 监控节点状态
func (s *GameNodeServer) monitorNodeStatus() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for nodeID, conn := range s.nodes {
			if now.Sub(conn.lastSeen) > 2*time.Minute {
				conn.status = "OFFLINE"
				// TODO: 通知节点管理器节点离线
			}
		}
		s.mu.Unlock()
	}
}
