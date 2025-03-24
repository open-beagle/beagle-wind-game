package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-beagle/beagle-wind-game/internal/agent/proto"
)

// AgentServer 实现节点Agent的gRPC服务端
type AgentServer struct {
	proto.UnimplementedAgentServiceServer
	opts             ServerOptions
	server           *grpc.Server
	nodeConnections  map[string]*nodeConnection
	connectionsMutex sync.RWMutex
}

// 节点连接信息
type nodeConnection struct {
	nodeID      string
	sessionID   string
	lastSeen    time.Time
	info        *proto.NodeInfo
	metrics     *proto.NodeMetrics
	eventStream map[string]proto.AgentService_SubscribeEventsServer
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

// NewAgentServer 创建新的Agent服务器实例
func NewAgentServer(opts ServerOptions) *AgentServer {
	return &AgentServer{
		opts:            opts,
		nodeConnections: make(map[string]*nodeConnection),
	}
}

// Start 启动gRPC服务
func (s *AgentServer) Start() error {
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
	proto.RegisterAgentServiceServer(s.server, s)

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
func (s *AgentServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// monitorNodeStatus 定期监控节点状态，清理超时节点
func (s *AgentServer) monitorNodeStatus() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.checkNodesStatus()
	}
}

// checkNodesStatus 检查所有节点的状态，清理超时节点
func (s *AgentServer) checkNodesStatus() {
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
func (s *AgentServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// 验证请求
	if req.NodeId == "" {
		return &proto.RegisterResponse{
			Success: false,
			Message: "节点ID不能为空",
		}, nil
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
			eventStream: make(map[string]proto.AgentService_SubscribeEventsServer),
		}
	}

	fmt.Printf("节点 %s 注册成功，会话ID: %s\n", req.NodeId, sessionID)

	return &proto.RegisterResponse{
		SessionId: sessionID,
		Success:   true,
		Message:   "注册成功",
	}, nil
}

// Heartbeat 实现心跳接口
func (s *AgentServer) Heartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	// 验证节点和会话
	s.connectionsMutex.RLock()
	conn, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &proto.HeartbeatResponse{
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

	return &proto.HeartbeatResponse{
		Success:    true,
		ServerTime: timestamppb.Now(),
	}, nil
}

// ExecutePipeline 实现Pipeline执行接口
func (s *AgentServer) ExecutePipeline(ctx context.Context, req *proto.ExecutePipelineRequest) (*proto.ExecutePipelineResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &proto.ExecutePipelineResponse{
			Accepted: false,
			Message:  "节点未注册",
		}, nil
	}

	// 生成执行ID
	executionID := fmt.Sprintf("%s-%s-%d", req.NodeId, req.PipelineId, time.Now().UnixNano())

	// TODO: 实现Pipeline执行逻辑

	return &proto.ExecutePipelineResponse{
		ExecutionId: executionID,
		Accepted:    true,
		Message:     "Pipeline执行请求已接受",
	}, nil
}

// GetPipelineStatus 实现获取Pipeline状态接口
func (s *AgentServer) GetPipelineStatus(ctx context.Context, req *proto.PipelineStatusRequest) (*proto.PipelineStatusResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("节点未注册: %s", req.NodeId)
	}

	// TODO: 实现获取Pipeline状态逻辑

	return &proto.PipelineStatusResponse{
		ExecutionId: req.ExecutionId,
		Status:      "pending",
		CurrentStep: 0,
		TotalSteps:  0,
		Progress:    0.0,
		StartTime:   timestamppb.Now(),
	}, nil
}

// CancelPipeline 实现取消Pipeline接口
func (s *AgentServer) CancelPipeline(ctx context.Context, req *proto.PipelineCancelRequest) (*proto.PipelineCancelResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &proto.PipelineCancelResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现取消Pipeline逻辑

	return &proto.PipelineCancelResponse{
		Success: true,
		Message: "Pipeline取消请求已发送",
	}, nil
}

// StartContainer 实现启动容器接口
func (s *AgentServer) StartContainer(ctx context.Context, req *proto.StartContainerRequest) (*proto.StartContainerResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &proto.StartContainerResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现启动容器逻辑

	return &proto.StartContainerResponse{
		Success:     true,
		ContainerId: req.ContainerId,
		Message:     "容器启动请求已发送",
	}, nil
}

// StopContainer 实现停止容器接口
func (s *AgentServer) StopContainer(ctx context.Context, req *proto.StopContainerRequest) (*proto.StopContainerResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &proto.StopContainerResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现停止容器逻辑

	return &proto.StopContainerResponse{
		Success: true,
		Message: "容器停止请求已发送",
	}, nil
}

// RestartContainer 实现重启容器接口
func (s *AgentServer) RestartContainer(ctx context.Context, req *proto.RestartContainerRequest) (*proto.RestartContainerResponse, error) {
	// 验证节点
	s.connectionsMutex.RLock()
	_, exists := s.nodeConnections[req.NodeId]
	s.connectionsMutex.RUnlock()

	if !exists {
		return &proto.RestartContainerResponse{
			Success: false,
			Message: "节点未注册",
		}, nil
	}

	// TODO: 实现重启容器逻辑

	return &proto.RestartContainerResponse{
		Success: true,
		Message: "容器重启请求已发送",
	}, nil
}

// GetNodeMetrics 实现获取节点指标接口
func (s *AgentServer) GetNodeMetrics(ctx context.Context, req *proto.NodeMetricsRequest) (*proto.NodeMetricsResponse, error) {
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
		return &proto.NodeMetricsResponse{
			NodeId: req.NodeId,
		}, nil
	}

	return &proto.NodeMetricsResponse{
		NodeId:  req.NodeId,
		Metrics: metrics,
		// TODO: 添加容器指标
	}, nil
}

// StreamNodeLogs 实现节点日志流接口
func (s *AgentServer) StreamNodeLogs(req *proto.NodeLogsRequest, stream proto.AgentService_StreamNodeLogsServer) error {
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
			err := stream.Send(&proto.LogEntry{
				Source:    "node",
				SourceId:  req.NodeId,
				Level:     "info",
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
func (s *AgentServer) StreamContainerLogs(req *proto.ContainerLogsRequest, stream proto.AgentService_StreamContainerLogsServer) error {
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
			err := stream.Send(&proto.LogEntry{
				Source:    "container",
				SourceId:  req.ContainerId,
				Level:     "info",
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
func (s *AgentServer) SubscribeEvents(req *proto.EventSubscriptionRequest, stream proto.AgentService_SubscribeEventsServer) error {
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
func (s *AgentServer) BroadcastEvent(nodeID string, event *proto.Event) error {
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
