package gamenode

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/open-beagle/beagle-wind-game/internal/event"
	"github.com/open-beagle/beagle-wind-game/internal/log"
	"github.com/open-beagle/beagle-wind-game/internal/models"
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
	pb.UnimplementedGameNodeGRPCServiceServer

	// 服务依赖
	nodeService     NodeService
	pipelineService PipelineService

	// 事件管理
	eventManager event.EventManager

	// 日志管理
	logManager log.LogManager

	// 配置
	config *ServerConfig

	// 连接管理
	connections     map[string]*AgentConnection
	connectionMutex sync.RWMutex
}

// NodeService 节点服务接口
type NodeService interface {
	Create(node models.GameNode) error
	UpdateStatusOnlineStatus(id string, online bool) error
	UpdateStatusMetrics(id string, metrics models.MetricsReport) error
	UpdateStatusResource(id string, resource models.ResourceInfo) error
	Get(id string) (models.GameNode, error)
}

// PipelineService Pipeline服务接口
type PipelineService interface {
	CreatePipeline(ctx context.Context, pipeline *pb.ExecutePipelineRequest) error
	ExecutePipeline(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status models.PipelineState) error
	UpdateStepStatus(ctx context.Context, pipelineID string, stepID string, status models.StepState) error
	SaveStepLogs(ctx context.Context, pipelineID string, stepID string, logs []byte) error
	CancelPipeline(ctx context.Context, id string) error
}

// AgentConnection Agent连接
type AgentConnection struct {
	client pb.GameNodeGRPCServiceClient
	conn   *grpc.ClientConn
	nodeID string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	MaxConnections  int
	HeartbeatPeriod time.Duration
}

// NewGameNodeServer 创建新的游戏节点服务器
func NewGameNodeServer(
	nodeService NodeService,
	pipelineService PipelineService,
	eventManager event.EventManager,
	logManager log.LogManager,
	config *ServerConfig,
) (*GameNodeServer, error) {
	if config == nil {
		config = &ServerConfig{
			MaxConnections:  100,
			HeartbeatPeriod: 30 * time.Second,
		}
	}

	server := &GameNodeServer{
		nodeService:     nodeService,
		pipelineService: pipelineService,
		eventManager:    eventManager,
		logManager:      logManager,
		connections:     make(map[string]*AgentConnection),
		config:          config,
	}

	return server, nil
}

// Register 注册节点
func (s *GameNodeServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// 检查连接数量限制
	s.connectionMutex.RLock()
	if len(s.connections) >= s.config.MaxConnections {
		s.connectionMutex.RUnlock()
		return nil, status.Error(codes.ResourceExhausted, "maximum number of connections reached")
	}
	s.connectionMutex.RUnlock()

	// 转换节点类型
	nodeType := models.GameNodeType(req.Type)
	if nodeType != models.GameNodeTypePhysical && nodeType != models.GameNodeTypeVirtual {
		return nil, status.Error(codes.InvalidArgument, "invalid node type")
	}

	// 通过 nodeService 处理节点注册
	err := s.nodeService.Create(models.GameNode{
		ID:       req.Id,
		Alias:    req.Alias,
		Model:    req.Model,
		Type:     nodeType,
		Location: req.Location,
		Labels:   req.Labels,
		Status: models.GameNodeStatus{
			State:      models.GameNodeStateOnline,
			Online:     true,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
			Resource: models.ResourceInfo{
				ID:        req.Id,
				Timestamp: time.Now().Unix(),
				Hardware:  models.HardwareInfo{},
				Software:  models.SoftwareInfo{},
				Network:   models.NetworkInfo{},
			},
			Metrics: models.MetricsReport{
				ID:        req.Id,
				Timestamp: time.Now().Unix(),
				Metrics:   []models.Metric{},
			},
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to register node: %v", err))
	}

	// 创建连接
	conn := &AgentConnection{
		nodeID: req.Id,
	}
	s.connectionMutex.Lock()
	s.connections[req.Id] = conn
	s.connectionMutex.Unlock()

	// 发布节点注册事件
	s.eventManager.Publish(event.NewNodeEvent(req.Id, "registered", "Node registered successfully"))

	return &pb.RegisterResponse{
		SessionId: fmt.Sprintf("session-%d", time.Now().UnixNano()),
		Status:    "success",
		Message:   "Node registered successfully",
	}, nil
}

// Heartbeat 心跳检测
func (s *GameNodeServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	// 更新节点在线状态
	err := s.nodeService.UpdateStatusOnlineStatus(req.Id, true)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update node status: %v", err))
	}

	// 发布心跳事件
	s.eventManager.Publish(event.NewNodeEvent(req.Id, "heartbeat", "Heartbeat received"))

	return &pb.HeartbeatResponse{
		Status:  "success",
		Message: "Heartbeat received",
	}, nil
}

// ReportMetrics 上报节点指标
func (s *GameNodeServer) ReportMetrics(ctx context.Context, req *pb.MetricsReport) (*pb.ReportResponse, error) {
	// 转换指标数据
	metrics := models.MetricsReport{
		ID:        req.Id,
		Timestamp: req.Timestamp,
		Metrics:   make([]models.Metric, len(req.Metrics)),
	}

	// 转换指标
	for i, m := range req.Metrics {
		metrics.Metrics[i] = models.Metric{
			Name:   m.Name,
			Type:   m.Type,
			Value:  m.Value,
			Labels: m.Labels,
		}
	}

	// 更新节点指标
	err := s.nodeService.UpdateStatusMetrics(req.Id, metrics)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update metrics: %v", err))
	}

	// 发布指标更新事件
	s.eventManager.Publish(event.NewNodeEvent(req.Id, "metrics_updated", "Node metrics updated"))

	return &pb.ReportResponse{
		Status:  "success",
		Message: "Metrics reported successfully",
	}, nil
}

// UpdateResourceInfo 更新资源信息
func (s *GameNodeServer) UpdateResourceInfo(ctx context.Context, req *pb.ResourceInfo) (*pb.UpdateResponse, error) {
	// 转换资源信息
	resource := models.ResourceInfo{
		ID:        req.Id,
		Timestamp: req.Timestamp,
		Hardware: models.HardwareInfo{
			CPU: models.CPUInfo{
				Model:       req.Hardware.Cpu.Model,
				Cores:       req.Hardware.Cpu.Cores,
				Threads:     req.Hardware.Cpu.Threads,
				Frequency:   req.Hardware.Cpu.Frequency,
				Temperature: req.Hardware.Cpu.Temperature,
				Usage:       req.Hardware.Cpu.Usage,
				Cache:       req.Hardware.Cpu.Cache,
			},
			Memory: models.MemoryInfo{
				Total:     req.Hardware.Memory.Total,
				Available: req.Hardware.Memory.Available,
				Used:      req.Hardware.Memory.Used,
				Usage:     req.Hardware.Memory.Usage,
				Type:      req.Hardware.Memory.Type,
				Frequency: req.Hardware.Memory.Frequency,
				Channels:  req.Hardware.Memory.Channels,
			},
			GPU: models.GPUInfo{
				Model:       req.Hardware.Gpu.Model,
				MemoryTotal: req.Hardware.Gpu.MemoryTotal,
				MemoryUsed:  req.Hardware.Gpu.MemoryUsed,
				MemoryFree:  req.Hardware.Gpu.MemoryFree,
				MemoryUsage: req.Hardware.Gpu.MemoryUsage,
				Usage:       req.Hardware.Gpu.Usage,
				Temperature: req.Hardware.Gpu.Temperature,
				Power:       req.Hardware.Gpu.Power,
				CUDACores:   req.Hardware.Gpu.CudaCores,
			},
			Disk: models.DiskInfo{
				Model:      req.Hardware.Disk.Model,
				Capacity:   req.Hardware.Disk.Capacity,
				Used:       req.Hardware.Disk.Used,
				Free:       req.Hardware.Disk.Free,
				Usage:      req.Hardware.Disk.Usage,
				Type:       req.Hardware.Disk.Type,
				Interface:  req.Hardware.Disk.Interface,
				ReadSpeed:  req.Hardware.Disk.ReadSpeed,
				WriteSpeed: req.Hardware.Disk.WriteSpeed,
				IOPS:       req.Hardware.Disk.Iops,
			},
		},
		Software: models.SoftwareInfo{
			OSDistribution:    req.Software.OsDistribution,
			OSVersion:         req.Software.OsVersion,
			OSArchitecture:    req.Software.OsArchitecture,
			KernelVersion:     req.Software.KernelVersion,
			GPUDriverVersion:  req.Software.GpuDriverVersion,
			CUDAVersion:       req.Software.CudaVersion,
			DockerVersion:     req.Software.DockerVersion,
			ContainerdVersion: req.Software.ContainerdVersion,
			RuncVersion:       req.Software.RuncVersion,
		},
		Network: models.NetworkInfo{
			Bandwidth:   req.Network.Bandwidth,
			Latency:     req.Network.Latency,
			Connections: req.Network.Connections,
			PacketLoss:  req.Network.PacketLoss,
		},
	}

	// 更新节点资源信息
	err := s.nodeService.UpdateStatusResource(req.Id, resource)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update resource info: %v", err))
	}

	// 发布资源更新事件
	s.eventManager.Publish(event.NewNodeEvent(req.Id, "resource_updated", "Node resource info updated"))

	return &pb.UpdateResponse{
		Status:  "success",
		Message: "Resource info updated successfully",
	}, nil
}

// ExecutePipeline 执行流水线
func (s *GameNodeServer) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) (*pb.ExecutePipelineResponse, error) {
	// 创建流水线
	err := s.pipelineService.CreatePipeline(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create pipeline: %v", err))
	}

	// 执行流水线
	err = s.pipelineService.ExecutePipeline(ctx, req.PipelineId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to execute pipeline: %v", err))
	}

	// 发布流水线执行事件
	s.eventManager.Publish(event.NewPipelineEvent(req.Id, req.PipelineId, "started", "Pipeline execution started"))

	return &pb.ExecutePipelineResponse{
		Status:  "success",
		Message: "Pipeline execution started",
	}, nil
}

// UpdatePipelineStatus 更新流水线状态
func (s *GameNodeServer) UpdatePipelineStatus(ctx context.Context, req *pb.PipelineStatusUpdate) (*pb.UpdateResponse, error) {
	// 转换状态
	var state models.PipelineState
	switch req.Status {
	case "pending":
		state = models.PipelineStatePending
	case "running":
		state = models.PipelineStateRunning
	case "completed":
		state = models.PipelineStateCompleted
	case "failed":
		state = models.PipelineStateFailed
	case "canceled":
		state = models.PipelineStateCanceled
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid pipeline status")
	}

	// 更新流水线状态
	err := s.pipelineService.UpdateStatus(ctx, req.PipelineId, state)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update pipeline status: %v", err))
	}

	// 发布流水线状态更新事件
	s.eventManager.Publish(event.NewPipelineEvent(req.Id, req.PipelineId, string(state), req.ErrorMessage))

	return &pb.UpdateResponse{
		Status:  "success",
		Message: "Pipeline status updated successfully",
	}, nil
}

// UpdateStepStatus 更新步骤状态
func (s *GameNodeServer) UpdateStepStatus(ctx context.Context, req *pb.StepStatusUpdate) (*pb.UpdateResponse, error) {
	// 转换状态
	var state models.StepState
	switch req.Status {
	case pb.StepStatus_PENDING:
		state = models.StepStatePending
	case pb.StepStatus_RUNNING:
		state = models.StepStateRunning
	case pb.StepStatus_COMPLETED:
		state = models.StepStateCompleted
	case pb.StepStatus_FAILED:
		state = models.StepStateFailed
	case pb.StepStatus_CANCELLED:
		state = models.StepStateSkipped
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid step status")
	}

	// 更新步骤状态
	err := s.pipelineService.UpdateStepStatus(ctx, req.PipelineId, req.StepId, state)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update step status: %v", err))
	}

	// 保存步骤日志
	if len(req.Logs) > 0 {
		err = s.pipelineService.SaveStepLogs(ctx, req.PipelineId, req.StepId, req.Logs)
		if err != nil {
			stdlog.Printf("failed to save step logs: %v", err)
		}
	}

	// 发布步骤状态更新事件
	s.eventManager.Publish(event.NewContainerEvent("", req.StepId, string(state), req.ErrorMessage))

	return &pb.UpdateResponse{
		Status:  "success",
		Message: "Step status updated successfully",
	}, nil
}

// CancelPipeline 取消流水线
func (s *GameNodeServer) CancelPipeline(ctx context.Context, req *pb.PipelineCancelRequest) (*pb.CancelResponse, error) {
	// 取消流水线
	err := s.pipelineService.CancelPipeline(ctx, req.PipelineId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to cancel pipeline: %v", err))
	}

	// 发布流水线取消事件
	s.eventManager.Publish(event.NewPipelineEvent("", req.PipelineId, "canceled", req.Reason))

	return &pb.CancelResponse{
		Status:  "success",
		Message: "Pipeline canceled successfully",
	}, nil
}

// StreamLogs 流式获取日志
func (s *GameNodeServer) StreamLogs(req *pb.LogRequest, stream pb.GameNodeGRPCService_StreamLogsServer) error {
	// 获取日志流
	logCh := s.logManager.StreamLogs(stream.Context(), req.PipelineId, time.Unix(req.StartTime, 0))

	// 发送日志
	for logEntry := range logCh {
		err := stream.Send(&pb.LogEntry{
			PipelineId: logEntry.PipelineId,
			StepId:     logEntry.StepId,
			Level:      logEntry.Level,
			Message:    logEntry.Message,
			Timestamp:  logEntry.Timestamp,
		})
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to send log entry: %v", err))
		}
	}

	return nil
}

// Stop 停止服务器
func (s *GameNodeServer) Stop() {
	// 关闭所有连接
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()

	for _, conn := range s.connections {
		if conn.conn != nil {
			conn.conn.Close()
		}
	}
	s.connections = make(map[string]*AgentConnection)
}

// Start 启动服务器
func (s *GameNodeServer) Start(ctx context.Context, opts *GameNodeServerOptions) error {
	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()
	pb.RegisterGameNodeGRPCServiceServer(grpcServer, s)

	// 创建监听器
	listener, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// 启动心跳检查
	go s.startHeartbeatCheck(ctx)

	// 启动服务器
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			stdlog.Printf("failed to serve: %v", err)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()

	// 优雅关闭
	grpcServer.GracefulStop()
	return nil
}

// startHeartbeatCheck 启动心跳检查
func (s *GameNodeServer) startHeartbeatCheck(ctx context.Context) {
	ticker := time.NewTicker(s.config.HeartbeatPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkHeartbeats()
		}
	}
}

// checkHeartbeats 检查心跳
func (s *GameNodeServer) checkHeartbeats() {
	s.connectionMutex.RLock()
	defer s.connectionMutex.RUnlock()

	now := time.Now()
	for nodeID, conn := range s.connections {
		// 获取节点状态
		node, err := s.nodeService.Get(nodeID)
		if err != nil {
			stdlog.Printf("failed to get node status: %v", err)
			continue
		}

		// 检查最后在线时间
		if now.Sub(node.Status.LastOnline) > s.config.HeartbeatPeriod*2 {
			// 更新节点状态为离线
			if err := s.nodeService.UpdateStatusOnlineStatus(nodeID, false); err != nil {
				stdlog.Printf("failed to update node status: %v", err)
				continue
			}

			// 发布节点离线事件
			s.eventManager.Publish(event.NewNodeEvent(nodeID, "offline", "Node heartbeat timeout"))

			// 关闭连接
			if conn.conn != nil {
				conn.conn.Close()
			}
		}
	}
}
