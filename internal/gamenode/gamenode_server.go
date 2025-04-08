package gamenode

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"strconv"
	"strings"
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

// ReportMetrics 报告节点指标
func (s *GameNodeServer) ReportMetrics(ctx context.Context, req *pb.MetricsReport) (*pb.ReportResponse, error) {
	// 验证请求
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "node id is required")
	}

	// 验证节点存在
	_, err := s.nodeService.Get(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("node not found: %v", err))
	}

	// 更新最后在线时间
	err = s.nodeService.UpdateStatusOnlineStatus(req.Id, true)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update online status: %v", err))
	}

	// 创建一个新的指标信息实例
	metricsInfo := models.MetricsInfo{
		CPUs:     []models.CPUMetrics{},
		Memory:   models.MemoryMetrics{},
		GPUs:     []models.GPUMetrics{},
		Storages: []models.StorageMetrics{},
		Network:  models.NetworkMetrics{},
	}

	// 遍历请求中的指标，并将其转换为我们的结构体
	for _, metric := range req.Metrics {
		// 基于指标名称分类
		switch {
		case strings.HasPrefix(metric.Name, "cpu."):
			// 如果是CPU指标
			if metric.Name == "cpu.usage" {
				// 确保已有CPU设备，如果没有则创建
				if len(metricsInfo.CPUs) == 0 {
					metricsInfo.CPUs = append(metricsInfo.CPUs, models.CPUMetrics{})
				}
				// 更新使用率
				metricsInfo.CPUs[0].Usage = metric.Value
			} else if metric.Name == "cpu.temperature" {
				// 注意：在新的模型中没有CPU温度字段，记录到日志
				s.eventManager.Publish(event.NewNodeEvent(req.Id, "warn", "CPU temperature metric ignored - field not in model"))
			}
		case strings.HasPrefix(metric.Name, "memory."):
			// 如果是内存指标
			if metric.Name == "memory.usage" {
				metricsInfo.Memory.Usage = metric.Value
			} else if metric.Name == "memory.used" {
				metricsInfo.Memory.Used = int64(metric.Value)
			} else if metric.Name == "memory.available" {
				metricsInfo.Memory.Available = int64(metric.Value)
			} else if metric.Name == "memory.total" {
				metricsInfo.Memory.Total = int64(metric.Value)
			}
		case strings.HasPrefix(metric.Name, "gpu."):
			// 如果是GPU指标
			if metric.Name == "gpu.usage" {
				// 确保已有GPU设备，如果没有则创建
				if len(metricsInfo.GPUs) == 0 {
					metricsInfo.GPUs = append(metricsInfo.GPUs, models.GPUMetrics{})
				}
				// 更新使用率
				metricsInfo.GPUs[0].Usage = metric.Value
			} else if metric.Name == "gpu.memory_usage" {
				// 确保已有GPU设备，如果没有则创建
				if len(metricsInfo.GPUs) == 0 {
					metricsInfo.GPUs = append(metricsInfo.GPUs, models.GPUMetrics{})
				}
				// 更新显存使用率
				metricsInfo.GPUs[0].MemoryUsage = metric.Value
			} else if metric.Name == "gpu.temperature" {
				// 注意：在新的模型中没有GPU温度字段，记录到日志
				s.eventManager.Publish(event.NewNodeEvent(req.Id, "warn", "GPU temperature metric ignored - field not in model"))
			} else if metric.Name == "gpu.memory_used" {
				// 确保已有GPU设备，如果没有则创建
				if len(metricsInfo.GPUs) == 0 {
					metricsInfo.GPUs = append(metricsInfo.GPUs, models.GPUMetrics{})
				}
				// 更新已用显存
				metricsInfo.GPUs[0].MemoryUsed = int64(metric.Value)
			} else if metric.Name == "gpu.memory_free" {
				// 确保已有GPU设备，如果没有则创建
				if len(metricsInfo.GPUs) == 0 {
					metricsInfo.GPUs = append(metricsInfo.GPUs, models.GPUMetrics{})
				}
				// 更新可用显存
				metricsInfo.GPUs[0].MemoryFree = int64(metric.Value)
			} else if metric.Name == "gpu.power" {
				// 注意：在新的模型中没有GPU功耗字段，记录到日志
				s.eventManager.Publish(event.NewNodeEvent(req.Id, "warn", "GPU power metric ignored - field not in model"))
			}
		case strings.HasPrefix(metric.Name, "storage."):
			// 如果是存储指标
			parts := strings.Split(metric.Name, ".")
			if len(parts) >= 3 {
				index, err := strconv.Atoi(parts[1])
				if err == nil {
					// 确保存储列表有足够的容量
					for len(metricsInfo.Storages) <= index {
						metricsInfo.Storages = append(metricsInfo.Storages, models.StorageMetrics{})
					}
					// 更新对应索引的存储指标
					if parts[2] == "usage" {
						metricsInfo.Storages[index].Usage = metric.Value
					} else if parts[2] == "used" {
						metricsInfo.Storages[index].Used = int64(metric.Value)
					} else if parts[2] == "free" {
						metricsInfo.Storages[index].Free = int64(metric.Value)
					} else if parts[2] == "capacity" {
						metricsInfo.Storages[index].Capacity = int64(metric.Value)
					}
				}
			}
		case strings.HasPrefix(metric.Name, "network."):
			// 如果是网络指标
			if metric.Name == "network.inbound" {
				metricsInfo.Network.InboundTraffic = metric.Value
			} else if metric.Name == "network.outbound" {
				metricsInfo.Network.OutboundTraffic = metric.Value
			} else if metric.Name == "network.connections" {
				metricsInfo.Network.Connections = int32(metric.Value)
			}
		}
	}

	// 如果没有存储设备，添加一个默认的
	if len(metricsInfo.Storages) == 0 {
		metricsInfo.Storages = append(metricsInfo.Storages, models.StorageMetrics{})
	}

	// 更新节点指标
	err = s.nodeService.UpdateStatusMetrics(req.Id, metricsInfo)
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
	// 转换为硬件信息 - 使用扁平结构
	hardwareInfo := models.HardwareInfo{
		CPUs:     []models.CPUDevice{},
		Memories: []models.MemoryDevice{},
		GPUs:     []models.GPUDevice{},
		Storages: []models.StorageDevice{},
		Networks: []models.NetworkDevice{},
	}

	// CPU信息 - 从扁平结构获取
	for _, cpu := range req.Hardware.Cpus {
		cpuDevice := models.CPUDevice{
			Model:        cpu.Model,
			Cores:        cpu.Cores,
			Threads:      cpu.Threads,
			Frequency:    cpu.Frequency,
			Cache:        cpu.Cache,
			Architecture: cpu.Architecture,
		}
		hardwareInfo.CPUs = append(hardwareInfo.CPUs, cpuDevice)
	}

	// 内存信息 - 从扁平结构获取
	for _, memory := range req.Hardware.Memories {
		memoryDevice := models.MemoryDevice{
			Size:      memory.Size,
			Type:      memory.Type,
			Frequency: memory.Frequency,
		}
		hardwareInfo.Memories = append(hardwareInfo.Memories, memoryDevice)
	}

	// GPU信息 - 从扁平结构获取
	for _, gpu := range req.Hardware.Gpus {
		gpuDevice := models.GPUDevice{
			Model:             gpu.Model,
			MemoryTotal:       gpu.MemoryTotal,
			Architecture:      gpu.Architecture,
			DriverVersion:     gpu.DriverVersion,
			ComputeCapability: gpu.ComputeCapability,
			TDP:               gpu.Tdp,
		}
		hardwareInfo.GPUs = append(hardwareInfo.GPUs, gpuDevice)
	}

	// 存储设备信息 - 从扁平结构获取
	for _, storage := range req.Hardware.Storages {
		storageDevice := models.StorageDevice{
			Type:     storage.Type,
			Model:    storage.Model,
			Capacity: storage.Capacity,
			Path:     storage.Path,
		}
		hardwareInfo.Storages = append(hardwareInfo.Storages, storageDevice)
	}

	// 网络设备信息 - 从扁平结构获取
	for _, network := range req.Hardware.Networks {
		networkDevice := models.NetworkDevice{
			Name:       network.Name,
			MacAddress: network.MacAddress,
			IpAddress:  network.IpAddress,
			Speed:      network.Speed,
		}
		hardwareInfo.Networks = append(hardwareInfo.Networks, networkDevice)
	}

	// 创建系统信息
	systemInfo := models.SystemInfo{
		OSDistribution:       req.System.OsDistribution,
		OSVersion:            req.System.OsVersion,
		OSArchitecture:       req.System.OsArchitecture,
		KernelVersion:        req.System.KernelVersion,
		GPUDriverVersion:     req.System.GpuDriverVersion,
		GPUComputeAPIVersion: req.System.GpuComputeApiVersion,
		DockerVersion:        req.System.DockerVersion,
		ContainerdVersion:    req.System.ContainerdVersion,
		RuncVersion:          req.System.RuncVersion,
	}

	// 更新节点硬件和系统信息
	err := s.nodeService.UpdateHardwareAndSystem(req.NodeId, hardwareInfo, systemInfo)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update hardware and system info: %v", err))
	}

	// 发布资源更新事件
	s.eventManager.Publish(event.NewNodeEvent(req.NodeId, "resource_updated", "Node hardware and system info updated"))

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
