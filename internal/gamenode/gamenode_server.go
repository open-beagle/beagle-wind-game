package gamenode

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// GameNodeServerOptions 包含服务器配置选项
type GameNodeServerOptions struct {
	ListenAddr  string
	TLSCertFile string
	TLSKeyFile  string
}

// GameNodeServer 游戏节点服务器
type GameNodeServer struct {
	pb.UnimplementedGameNodeServiceServer

	// 节点管理
	nodes     map[string]*pb.NodeInfo
	nodeMutex sync.RWMutex

	// 会话管理
	sessions     map[string]*NodeSession
	sessionMutex sync.RWMutex

	// 事件管理
	eventManager *GameNodeEventManager

	// 日志管理
	logManager *GameNodeLogManager

	// 配置
	config *ServerConfig

	// 节点状态
	nodeStatus map[string]*NodeStatus
}

// NodeSession 节点会话
type NodeSession struct {
	NodeID     string
	LastActive time.Time
	Expired    bool
}

// ServerConfig 服务器配置
type ServerConfig struct {
	MaxNodes        int
	SessionTimeout  time.Duration
	HeartbeatPeriod time.Duration
}

// NodeStatus 节点状态
type NodeStatus struct {
	LastSeen  time.Time
	State     string
	NodeID    string
	Connected bool
}

// NewGameNodeServer 创建新的游戏节点服务器
func NewGameNodeServer(config *ServerConfig) (*GameNodeServer, error) {
	if config == nil {
		config = &ServerConfig{
			MaxNodes:        100,
			SessionTimeout:  5 * time.Minute,
			HeartbeatPeriod: 30 * time.Second,
		}
	}

	server := &GameNodeServer{
		nodes:        make(map[string]*pb.NodeInfo),
		sessions:     make(map[string]*NodeSession),
		nodeStatus:   make(map[string]*NodeStatus),
		eventManager: NewGameNodeEventManager(),
		logManager:   NewGameNodeLogManager(),
		config:       config,
	}

	// 启动会话清理任务
	go server.cleanupSessions()
	// 启动节点清理任务
	go server.cleanupNodes()

	return server, nil
}

// Register 注册节点
func (s *GameNodeServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.nodeMutex.Lock()
	defer s.nodeMutex.Unlock()

	// 检查节点数量限制
	if len(s.nodes) >= s.config.MaxNodes {
		return nil, status.Error(codes.ResourceExhausted, "maximum number of nodes reached")
	}

	// 检查节点是否已存在
	if _, exists := s.nodes[req.NodeId]; exists {
		return nil, status.Error(codes.AlreadyExists, "node already registered")
	}

	// 保存节点信息
	s.nodes[req.NodeId] = req.NodeInfo

	// 创建节点状态
	s.nodeStatus[req.NodeId] = &NodeStatus{
		LastSeen:  time.Now(),
		State:     "online",
		NodeID:    req.NodeId,
		Connected: true,
	}

	// 创建会话
	sessionID := generateSessionID()
	session := &NodeSession{
		NodeID:     req.NodeId,
		LastActive: time.Now(),
		Expired:    false,
	}

	s.sessionMutex.Lock()
	s.sessions[sessionID] = session
	s.sessionMutex.Unlock()

	// 发布节点注册事件
	s.eventManager.Publish(NewGameNodeNodeEvent(req.NodeId, "registered", "Node registered successfully"))

	return &pb.RegisterResponse{
		SessionId: sessionID,
		Success:   true,
		Message:   "Node registered successfully",
	}, nil
}

// Heartbeat 心跳检测
func (s *GameNodeServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	// 验证会话
	session, exists := s.sessions[req.SessionId]
	if !exists || session.Expired {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
	}

	// 更新会话状态
	session.LastActive = time.Now()
	session.Expired = false

	// 更新节点状态
	s.nodeMutex.Lock()
	if _, exists := s.nodes[req.NodeId]; exists {
		if status, ok := s.nodeStatus[req.NodeId]; ok {
			status.LastSeen = time.Now()
			status.Connected = true
		}
	}
	s.nodeMutex.Unlock()

	return &pb.HeartbeatResponse{
		Success:    true,
		ServerTime: timestamppb.Now(),
	}, nil
}

// ExecutePipeline 执行Pipeline
func (s *GameNodeServer) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) (*pb.ExecutePipelineResponse, error) {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理 Pipeline 执行
	return nil, status.Error(codes.Unimplemented, "pipeline execution not implemented")
}

// GetPipelineStatus 获取Pipeline状态
func (s *GameNodeServer) GetPipelineStatus(ctx context.Context, req *pb.PipelineStatusRequest) (*pb.PipelineStatusResponse, error) {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理 Pipeline 状态查询
	return nil, status.Error(codes.Unimplemented, "pipeline status not implemented")
}

// CancelPipeline 取消Pipeline
func (s *GameNodeServer) CancelPipeline(ctx context.Context, req *pb.PipelineCancelRequest) (*pb.PipelineCancelResponse, error) {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理 Pipeline 取消
	return nil, status.Error(codes.Unimplemented, "pipeline cancellation not implemented")
}

// StreamNodeLogs 流式获取节点日志
func (s *GameNodeServer) StreamNodeLogs(req *pb.NodeLogsRequest, stream pb.GameNodeService_StreamNodeLogsServer) error {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理日志流
	return status.Error(codes.Unimplemented, "node logs not implemented")
}

// StreamContainerLogs 流式获取容器日志
func (s *GameNodeServer) StreamContainerLogs(req *pb.ContainerLogsRequest, stream pb.GameNodeService_StreamContainerLogsServer) error {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理容器日志流
	return status.Error(codes.Unimplemented, "container logs not implemented")
}

// StartContainer 启动容器
func (s *GameNodeServer) StartContainer(ctx context.Context, req *pb.StartContainerRequest) (*pb.StartContainerResponse, error) {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理容器启动
	return nil, status.Error(codes.Unimplemented, "container start not implemented")
}

// StopContainer 停止容器
func (s *GameNodeServer) StopContainer(ctx context.Context, req *pb.StopContainerRequest) (*pb.StopContainerResponse, error) {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理容器停止
	return nil, status.Error(codes.Unimplemented, "container stop not implemented")
}

// RestartContainer 重启容器
func (s *GameNodeServer) RestartContainer(ctx context.Context, req *pb.RestartContainerRequest) (*pb.RestartContainerResponse, error) {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// TODO: 通过 nodeService 处理容器重启
	return nil, status.Error(codes.Unimplemented, "container restart not implemented")
}

// GetNodeMetrics 获取节点指标
func (s *GameNodeServer) GetNodeMetrics(ctx context.Context, req *pb.NodeMetricsRequest) (*pb.NodeMetricsResponse, error) {
	// 验证节点是否存在
	s.nodeMutex.RLock()
	_, exists := s.nodes[req.NodeId]
	s.nodeMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// 获取节点状态
	nodeStatus, ok := s.nodeStatus[req.NodeId]
	if !ok {
		return nil, status.Error(codes.Internal, "node status not found")
	}

	// 构建响应
	return &pb.NodeMetricsResponse{
		NodeId: req.NodeId,
		Metrics: &pb.NodeMetrics{
			CpuUsage:    0, // TODO: 从实际监控数据获取
			MemoryUsage: 0, // TODO: 从实际监控数据获取
			DiskUsage:   0, // TODO: 从实际监控数据获取
			CollectedAt: timestamppb.New(nodeStatus.LastSeen),
		},
	}, nil
}

// Stop 停止服务器
func (s *GameNodeServer) Stop() {
	s.nodeMutex.Lock()
	defer s.nodeMutex.Unlock()

	// 清空节点和会话
	s.nodes = make(map[string]*pb.NodeInfo)
	s.sessions = make(map[string]*NodeSession)
	s.nodeStatus = make(map[string]*NodeStatus)
}

// cleanupSessions 清理过期会话
func (s *GameNodeServer) cleanupSessions() {
	ticker := time.NewTicker(s.config.SessionTimeout)
	defer ticker.Stop()

	for range ticker.C {
		s.sessionMutex.Lock()
		now := time.Now()

		for sessionID, session := range s.sessions {
			if now.Sub(session.LastActive) > s.config.SessionTimeout {
				session.Expired = true
				delete(s.sessions, sessionID)

				// 更新节点状态
				s.nodeMutex.Lock()
				if status, ok := s.nodeStatus[session.NodeID]; ok {
					status.Connected = false
					status.State = "offline"
				}
				s.nodeMutex.Unlock()
			}
		}
		s.sessionMutex.Unlock()
	}
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

// monitorNodeStatus 监控节点状态
func (s *GameNodeServer) monitorNodeStatus() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.nodeMutex.Lock()
		now := time.Now()
		for _, agent := range s.nodes {
			if now.Sub(agent.lastSeen) > 2*time.Minute {
				agent.status = "OFFLINE"
				// 发布节点离线事件
				s.eventManager.Publish(NewGameNodeNodeEvent(agent.nodeID, "offline", "Node is offline"))
			}
		}
		s.nodeMutex.Unlock()
	}
}

// GetNodeStatus 获取节点状态
func (s *GameNodeServer) GetNodeStatus(nodeID string) (*NodeStatus, bool) {
	s.nodeMutex.RLock()
	defer s.nodeMutex.RUnlock()

	status, ok := s.nodeStatus[nodeID]
	return status, ok
}

// cleanupNodes 清理离线节点
func (s *GameNodeServer) cleanupNodes() {
	ticker := time.NewTicker(s.config.HeartbeatPeriod * 3)
	defer ticker.Stop()

	for range ticker.C {
		s.nodeMutex.Lock()
		now := time.Now()

		for _, status := range s.nodeStatus {
			if now.Sub(status.LastSeen) > s.config.HeartbeatPeriod*3 {
				status.Connected = false
				status.State = "offline"
			}
		}
		s.nodeMutex.Unlock()
	}
}
