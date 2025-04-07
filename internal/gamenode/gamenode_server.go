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
	Update(node models.GameNode) error
	UpdateStatusOnlineStatus(id string, online bool) error
	UpdateStatusMetrics(id string, metrics models.MetricsInfo) error
	UpdateHardwareAndSystem(id string, hardware models.HardwareInfo, system models.SystemInfo) error
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
		return &pb.RegisterResponse{
			Success: false,
			Message: "达到最大连接数限制",
		}, nil
	}
	s.connectionMutex.RUnlock()

	// 转换节点类型
	nodeType := models.GameNodeType(req.Type)
	if nodeType != models.GameNodeTypePhysical && nodeType != models.GameNodeTypeVirtual {
		return &pb.RegisterResponse{
			Success: false,
			Message: "无效的节点类型",
		}, nil
	}

	// 检查节点是否存在
	node, err := s.nodeService.Get(req.Id)
	if err != nil {
		// 节点不存在，创建新节点
		node = models.GameNode{
			ID:       req.Id,
			Alias:    req.Alias,
			Model:    req.Model,
			Type:     nodeType,
			Location: req.Location,
			Hardware: req.Hardware,
			System:   req.System,
			Labels:   req.Labels,
			Status: models.GameNodeStatus{
				State:      models.GameNodeStateOnline,
				Online:     true,
				LastOnline: time.Now(),
				UpdatedAt:  time.Now(),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.nodeService.Create(node); err != nil {
			return &pb.RegisterResponse{
				Success: false,
				Message: fmt.Sprintf("创建节点失败: %v", err),
			}, nil
		}
	} else {
		// 节点存在，更新状态
		node.Status.State = models.GameNodeStateOnline
		node.Status.Online = true
		node.Status.LastOnline = time.Now()
		node.UpdatedAt = time.Now()

		// 更新节点信息
		node.Alias = req.Alias
		node.Model = req.Model
		node.Type = nodeType
		node.Location = req.Location
		node.Hardware = req.Hardware
		node.System = req.System
		node.Labels = req.Labels

		if err := s.nodeService.Update(node); err != nil {
			return &pb.RegisterResponse{
				Success: false,
				Message: fmt.Sprintf("更新节点失败: %v", err),
			}, nil
		}
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
		Success: true,
		Message: "节点注册成功",
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
	// 创建MetricsInfo结构
	metricsInfo := models.MetricsInfo{}

	// 初始化CPU信息
	cpuDevice := struct {
		Model       string  `json:"model" yaml:"model"`
		Cores       int32   `json:"cores" yaml:"cores"`
		Threads     int32   `json:"threads" yaml:"threads"`
		Usage       float64 `json:"usage" yaml:"usage"`
		Temperature float64 `json:"temperature" yaml:"temperature"`
	}{}

	// 初始化GPU信息
	gpuDevice := struct {
		Model       string  `json:"model" yaml:"model"`
		MemoryTotal int64   `json:"memory_total" yaml:"memory_total"`
		Usage       float64 `json:"usage" yaml:"usage"`
		MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`
		MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`
		MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"`
		Temperature float64 `json:"temperature" yaml:"temperature"`
		Power       float64 `json:"power" yaml:"power"`
	}{}

	// 初始化存储信息
	storageDevice := struct {
		Path       string  `json:"path" yaml:"path"`
		Type       string  `json:"type" yaml:"type"`
		Model      string  `json:"model" yaml:"model"`
		Capacity   int64   `json:"capacity" yaml:"capacity"`
		Used       int64   `json:"used" yaml:"used"`
		Free       int64   `json:"free" yaml:"free"`
		Usage      float64 `json:"usage" yaml:"usage"`
		ReadSpeed  float64 `json:"read_speed" yaml:"read_speed"`
		WriteSpeed float64 `json:"write_speed" yaml:"write_speed"`
	}{
		Path: "/",
	}

	// 处理指标数据
	for _, m := range req.Metrics {
		switch m.Name {
		case "cpu_usage":
			cpuDevice.Usage = m.Value
		case "cpu_temperature":
			cpuDevice.Temperature = m.Value
		case "memory_usage":
			metricsInfo.Memory.Usage = m.Value
		case "memory_used":
			metricsInfo.Memory.Used = int64(m.Value)
		case "memory_available":
			metricsInfo.Memory.Available = int64(m.Value)
		case "disk_usage":
			storageDevice.Usage = m.Value
		case "disk_used":
			storageDevice.Used = int64(m.Value)
		case "disk_free":
			storageDevice.Free = int64(m.Value)
		case "gpu_usage":
			gpuDevice.Usage = m.Value
		case "gpu_memory_used":
			gpuDevice.MemoryUsed = int64(m.Value)
		case "gpu_memory_free":
			gpuDevice.MemoryFree = int64(m.Value)
		case "gpu_memory_usage":
			gpuDevice.MemoryUsage = m.Value
		case "gpu_temperature":
			gpuDevice.Temperature = m.Value
		case "gpu_power":
			gpuDevice.Power = m.Value
		case "network_inbound":
			metricsInfo.Network.InboundTraffic = m.Value
		case "network_outbound":
			metricsInfo.Network.OutboundTraffic = m.Value
		case "network_connections":
			metricsInfo.Network.Connections = int32(m.Value)
		}
	}

	// 将指标加入到结构体中
	metricsInfo.CPU.Devices = append(metricsInfo.CPU.Devices, cpuDevice)
	metricsInfo.GPU.Devices = append(metricsInfo.GPU.Devices, gpuDevice)
	metricsInfo.Storage.Devices = append(metricsInfo.Storage.Devices, storageDevice)

	// 更新节点指标
	err := s.nodeService.UpdateStatusMetrics(req.Id, metricsInfo)
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
	// 转换为硬件信息
	hardwareInfo := models.HardwareInfo{}

	// CPU信息
	if req.Hardware.Cpu != nil {
		cpuDevice := struct {
			Model        string  `json:"model" yaml:"model"`
			Cores        int32   `json:"cores" yaml:"cores"`
			Threads      int32   `json:"threads" yaml:"threads"`
			Frequency    float64 `json:"frequency" yaml:"frequency"`
			Cache        int64   `json:"cache" yaml:"cache"`
			Socket       string  `json:"socket" yaml:"socket"`
			Manufacturer string  `json:"manufacturer" yaml:"manufacturer"`
			Architecture string  `json:"architecture" yaml:"architecture"`
		}{
			Model:     req.Hardware.Cpu.Model,
			Cores:     req.Hardware.Cpu.Cores,
			Threads:   req.Hardware.Cpu.Threads,
			Frequency: req.Hardware.Cpu.Frequency,
			Cache:     req.Hardware.Cpu.Cache,
		}
		hardwareInfo.CPU.Devices = append(hardwareInfo.CPU.Devices, cpuDevice)
	}

	// 内存信息
	if req.Hardware.Memory != nil {
		memoryDevice := struct {
			Size         int64   `json:"size" yaml:"size"`
			Type         string  `json:"type" yaml:"type"`
			Frequency    float64 `json:"frequency" yaml:"frequency"`
			Manufacturer string  `json:"manufacturer" yaml:"manufacturer"`
			Serial       string  `json:"serial" yaml:"serial"`
			Slot         string  `json:"slot" yaml:"slot"`
			PartNumber   string  `json:"part_number" yaml:"part_number"`
			FormFactor   string  `json:"form_factor" yaml:"form_factor"`
		}{
			Size:      req.Hardware.Memory.Total,
			Type:      req.Hardware.Memory.Type,
			Frequency: req.Hardware.Memory.Frequency,
		}
		hardwareInfo.Memory.Devices = append(hardwareInfo.Memory.Devices, memoryDevice)
	}

	// GPU信息
	if req.Hardware.Gpu != nil {
		gpuDevice := struct {
			Model        string `json:"model" yaml:"model"`
			MemoryTotal  int64  `json:"memory_total" yaml:"memory_total"`
			CudaCores    int32  `json:"cuda_cores" yaml:"cuda_cores"`
			Manufacturer string `json:"manufacturer" yaml:"manufacturer"`
			Bus          string `json:"bus" yaml:"bus"`
			PciSlot      string `json:"pci_slot" yaml:"pci_slot"`
			Serial       string `json:"serial" yaml:"serial"`
			Architecture string `json:"architecture" yaml:"architecture"`
			TDP          int32  `json:"tdp" yaml:"tdp"` // 功耗指标(W)
		}{
			Model:       req.Hardware.Gpu.Model,
			MemoryTotal: req.Hardware.Gpu.MemoryTotal,
			CudaCores:   req.Hardware.Gpu.CudaCores,
		}
		hardwareInfo.GPU.Devices = append(hardwareInfo.GPU.Devices, gpuDevice)
	}

	// 存储设备信息
	if req.Hardware.Storage != nil && len(req.Hardware.Storage.Devices) > 0 {
		for _, device := range req.Hardware.Storage.Devices {
			storageDevice := struct {
				Type         string `json:"type" yaml:"type"`
				Model        string `json:"model" yaml:"model"`
				Capacity     int64  `json:"capacity" yaml:"capacity"`
				Path         string `json:"path" yaml:"path"`
				Serial       string `json:"serial" yaml:"serial"`
				Interface    string `json:"interface" yaml:"interface"`
				Manufacturer string `json:"manufacturer" yaml:"manufacturer"`
				FormFactor   string `json:"form_factor" yaml:"form_factor"`
				Firmware     string `json:"firmware" yaml:"firmware"`
			}{
				Type:     device.Type,
				Capacity: device.Capacity,
			}
			hardwareInfo.Storage.Devices = append(hardwareInfo.Storage.Devices, storageDevice)
		}
	}

	// 创建系统信息（暂时为空，可以从其他字段中获取）
	systemInfo := models.SystemInfo{}

	// 更新节点硬件和系统信息
	err := s.nodeService.UpdateHardwareAndSystem(req.Id, hardwareInfo, systemInfo)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update hardware and system info: %v", err))
	}

	// 发布资源更新事件
	s.eventManager.Publish(event.NewNodeEvent(req.Id, "resource_updated", "Node hardware and system info updated"))

	return &pb.UpdateResponse{
		Status:  "success",
		Message: "Hardware and system info updated successfully",
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
