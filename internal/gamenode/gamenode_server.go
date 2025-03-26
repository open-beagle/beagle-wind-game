package gamenode

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// GameAgentServer 实现节点Agent的gRPC服务端
type GameAgentServer struct {
	pb.UnimplementedAgentServiceServer
	opts             ServerOptions
	server           *grpc.Server
	nodeConnections  map[string]*nodeConnection
	connectionsMutex sync.RWMutex
	manager          GameAgentManager
	pipelines        map[string]*Pipeline
	pipelinesMutex   sync.RWMutex
}

// 节点连接信息
type nodeConnection struct {
	nodeID      string
	sessionID   string
	lastSeen    time.Time
	info        *pb.NodeInfo
	metrics     *pb.NodeMetrics
	eventStream map[string]pb.AgentService_SubscribeEventsServer
	mutex       sync.RWMutex
}

// ServerOptions 服务器配置选项
type ServerOptions struct {
	ListenAddr   string
	TLSCertFile  string
	TLSKeyFile   string
	MaxHeartbeat time.Duration // 节点心跳超时时间
}

// DefaultServerOptions 默认服务器配置
var DefaultServerOptions = ServerOptions{
	ListenAddr:   ":50051",
	TLSCertFile:  "",
	TLSKeyFile:   "",
	MaxHeartbeat: 30 * time.Second,
}

// NewGameAgentServer 创建新的Agent服务器实例
func NewGameAgentServer(opts ServerOptions, manager GameAgentManager) *GameAgentServer {
	return &GameAgentServer{
		opts:            opts,
		manager:         manager,
		nodeConnections: make(map[string]*nodeConnection),
		pipelines:       make(map[string]*Pipeline),
	}
}

// Start 启动gRPC服务
func (s *GameAgentServer) Start() error {
	// 设置gRPC服务器选项
	var serverOpts []grpc.ServerOption

	// 添加TLS证书
	if s.opts.TLSCertFile != "" && s.opts.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(s.opts.TLSCertFile, s.opts.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("加载TLS证书失败: %w", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	// 添加保活策略
	kaParams := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveParams(kaParams))

	// 创建gRPC服务器
	s.server = grpc.NewServer(serverOpts...)
	pb.RegisterAgentServiceServer(s.server, s)

	// 启用反射服务
	reflection.Register(s.server)

	// 监听网络端口
	lis, err := net.Listen("tcp", s.opts.ListenAddr)
	if err != nil {
		return fmt.Errorf("网络监听失败: %w", err)
	}

	// 启动节点状态监控
	go s.monitorNodeStatus()

	// 启动服务
	fmt.Printf("Agent服务器启动，监听地址: %s\n", s.opts.ListenAddr)
	return s.server.Serve(lis)
}

// Stop 停止gRPC服务
func (s *GameAgentServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// monitorNodeStatus 定期监控节点状态，清理超时节点
func (s *GameAgentServer) monitorNodeStatus() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.checkNodesStatus()
	}
}

// checkNodesStatus 检查所有节点的状态，清理超时节点
func (s *GameAgentServer) checkNodesStatus() {
	now := time.Now()
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()

	for nodeID, conn := range s.nodeConnections {
		// 检查心跳超时
		if now.Sub(conn.lastSeen) > s.opts.MaxHeartbeat {
			fmt.Printf("节点 %s 心跳超时，移除连接\n", nodeID)
			delete(s.nodeConnections, nodeID)
		}
	}
}

// Register 实现节点注册接口
func (s *GameAgentServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// 验证请求
	if req.NodeId == "" {
		return &pb.RegisterResponse{
			Success: false,
			Message: "节点ID不能为空",
		}, fmt.Errorf("节点ID不能为空")
	}

	// 验证节点管理器
	if s.manager == nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: "节点管理器未初始化",
		}, fmt.Errorf("节点管理器未初始化")
	}

	// 获取节点信息
	_, err := s.manager.Get(req.NodeId)
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("获取节点信息失败: %v", err),
		}, fmt.Errorf("获取节点信息失败: %w", err)
	}

	// 更新节点状态
	err = s.manager.UpdateStatusState(req.NodeId, string(models.GameNodeStateReady))
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("更新节点状态失败: %v", err),
		}, fmt.Errorf("更新节点状态失败: %w", err)
	}

	// 生成会话ID
	sessionID := fmt.Sprintf("%s-%d", req.NodeId, time.Now().UnixNano())

	// 创建或更新节点连接
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()

	// 如果节点已存在，更新会话
	if conn, exists := s.nodeConnections[req.NodeId]; exists {
		conn.sessionID = sessionID
		conn.lastSeen = time.Now()
		conn.info = req.NodeInfo
	} else {
		// 创建新节点连接
		s.nodeConnections[req.NodeId] = &nodeConnection{
			nodeID:      req.NodeId,
			sessionID:   sessionID,
			lastSeen:    time.Now(),
			info:        req.NodeInfo,
			eventStream: make(map[string]pb.AgentService_SubscribeEventsServer),
		}
	}

	fmt.Printf("节点 %s 注册成功，会话ID: %s\n", req.NodeId, sessionID)

	return &pb.RegisterResponse{
		SessionId: sessionID,
		Success:   true,
		Message:   "注册成功",
	}, nil
}

// Heartbeat 实现心跳接口
func (s *GameAgentServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	// 验证节点和会话
	s.connectionsMutex.RLock()
	conn, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &pb.HeartbeatResponse{
			Success:    false,
			ServerTime: timestamppb.Now(),
		}, nil
	}

	// 更新节点状态
	conn.mutex.Lock()
	conn.lastSeen = time.Now()
	if req.Metrics != nil {
		conn.metrics = req.Metrics
	}
	conn.mutex.Unlock()

	return &pb.HeartbeatResponse{
		Success:    true,
		ServerTime: timestamppb.Now(),
	}, nil
}

// ExecutePipeline 实现Pipeline执行接口
func (s *GameAgentServer) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) (*pb.ExecutePipelineResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &pb.ExecutePipelineResponse{
			Accepted: false,
			Message:  "节点未注册",
		}, nil
	}

	// 生成执行ID
	executionID := fmt.Sprintf("%s-%s-%d", req.NodeId, req.PipelineId, time.Now().UnixNano())

	// 创建 Pipeline 实例
	pipeline := NewPipeline(req)
	if pipeline == nil {
		return &pb.ExecutePipelineResponse{
			Accepted: false,
			Message:  "创建 Pipeline 失败",
		}, fmt.Errorf("创建 Pipeline 失败")
	}

	// 保存 Pipeline 实例
	s.pipelinesMutex.Lock()
	s.pipelines[executionID] = pipeline
	s.pipelinesMutex.Unlock()

	// 执行 Pipeline
	go func() {
		if err := pipeline.Execute(ctx); err != nil {
			fmt.Printf("Pipeline 执行失败: %v\n", err)
		}
	}()

	return &pb.ExecutePipelineResponse{
		Accepted:    true,
		ExecutionId: executionID,
		Message:     "Pipeline 已接受",
	}, nil
}

// GetPipelineStatus 实现获取Pipeline状态接口
func (s *GameAgentServer) GetPipelineStatus(ctx context.Context, req *pb.PipelineStatusRequest) (*pb.PipelineStatusResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("节点未注册: %s", req.NodeId)
	}

	// TODO: 实现获取Pipeline状态逻辑

	return &pb.PipelineStatusResponse{
		ExecutionId: req.ExecutionId,
		Status:      "pending",
		CurrentStep: 0,
		TotalSteps:  0,
		Progress:    0.0,
		StartTime:   timestamppb.Now(),
	}, nil
}

// CancelPipeline 实现取消Pipeline接口
func (s *GameAgentServer) CancelPipeline(ctx context.Context, req *pb.PipelineCancelRequest) (*pb.PipelineCancelResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &pb.PipelineCancelResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现取消Pipeline逻辑

	return &pb.PipelineCancelResponse{
		Success: true,
		Message: "Pipeline取消请求已发送",
	}, nil
}

// StartContainer 实现启动容器接口
func (s *GameAgentServer) StartContainer(ctx context.Context, req *pb.StartContainerRequest) (*pb.StartContainerResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &pb.StartContainerResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现启动容器逻辑

	return &pb.StartContainerResponse{
		Success:     true,
		ContainerId: req.ContainerId,
		Message:     "容器启动请求已发送",
	}, nil
}

// StopContainer 实现停止容器接口
func (s *GameAgentServer) StopContainer(ctx context.Context, req *pb.StopContainerRequest) (*pb.StopContainerResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &pb.StopContainerResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现停止容器逻辑

	return &pb.StopContainerResponse{
		Success: true,
		Message: "容器停止请求已发送",
	}, nil
}

// RestartContainer 实现重启容器接口
func (s *GameAgentServer) RestartContainer(ctx context.Context, req *pb.RestartContainerRequest) (*pb.RestartContainerResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &pb.RestartContainerResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现重启容器逻辑

	return &pb.RestartContainerResponse{
		Success: true,
		Message: "容器重启请求已发送",
	}, nil
}

// GetNodeMetrics 实现获取节点指标接口
func (s *GameAgentServer) GetNodeMetrics(ctx context.Context, req *pb.NodeMetricsRequest) (*pb.NodeMetricsResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	conn, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("节点未注册: %s", req.NodeId)
	}

	// 获取指标
	conn.mutex.RLock()
	metrics := conn.metrics
	conn.mutex.RUnlock()

	if metrics == nil {
		return &pb.NodeMetricsResponse{
			NodeId: req.NodeId,
		}, nil
	}

	return &pb.NodeMetricsResponse{
		NodeId:  req.NodeId,
		Metrics: metrics,
		// TODO: 添加容器指标
	}, nil
}

// StreamNodeLogs 实现节点日志流接口
func (s *GameAgentServer) StreamNodeLogs(req *pb.NodeLogsRequest, stream pb.AgentService_StreamNodeLogsServer) error {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("节点未注册: %s", req.NodeId)
	}

	// TODO: 实现节点日志流逻辑
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// 发送示例日志
			err := stream.Send(&pb.LogEntry{
				Source:    "node",
				Content:   "示例节点日志",
				Timestamp: timestamppb.Now(),
			})
			if err != nil {
				return err
			}
		}
	}
}

// StreamContainerLogs 实现容器日志流接口
func (s *GameAgentServer) StreamContainerLogs(req *pb.ContainerLogsRequest, stream pb.AgentService_StreamContainerLogsServer) error {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("节点未注册: %s", req.NodeId)
	}

	// TODO: 实现容器日志流逻辑
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// 发送示例日志
			err := stream.Send(&pb.LogEntry{
				Source:    "container",
				Content:   "示例容器日志",
				Timestamp: timestamppb.Now(),
			})
			if err != nil {
				return err
			}
		}
	}
}

// SubscribeEvents 实现事件订阅接口
func (s *GameAgentServer) SubscribeEvents(req *pb.EventSubscriptionRequest, stream pb.AgentService_SubscribeEventsServer) error {
	// 验证节点
	s.connectionsMutex.RLock()
	conn, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("节点未注册: %s", req.NodeId)
	}

	// 生成订阅ID
	subscriptionID := fmt.Sprintf("%s-%d", req.NodeId, time.Now().UnixNano())

	// 注册事件流
	conn.mutex.Lock()
	conn.eventStream[subscriptionID] = stream
	conn.mutex.Unlock()

	// 在函数退出时取消订阅
	defer func() {
		conn.mutex.Lock()
		delete(conn.eventStream, subscriptionID)
		conn.mutex.Unlock()
	}()

	// 处理事件流
	ctx := stream.Context()
	<-ctx.Done()
	return ctx.Err()
}

// BroadcastEvent 向节点发送事件
func (s *GameAgentServer) BroadcastEvent(nodeID string, event *pb.Event) error {
	s.connectionsMutex.RLock()
	conn, exists := s.nodeConnections[nodeID]
	s.connectionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("节点未注册: %s", nodeID)
	}

	// 发送事件到所有订阅流
	conn.mutex.RLock()
	defer conn.mutex.RUnlock()

	for id, eventStream := range conn.eventStream {
		err := eventStream.Send(event)
		if err != nil {
			fmt.Printf("向节点 %s 订阅 %s 发送事件失败: %v\n", nodeID, id, err)
		}
	}

	return nil
}
