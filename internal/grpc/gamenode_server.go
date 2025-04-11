package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
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
	nodeService     GameNodeServiceInterface
	pipelineService GamePipelineServiceInterface

	// 日志框架
	logger utils.Logger

	// 配置
	config *ServerConfig

	// 连接管理
	mu          sync.RWMutex
	connections map[string]*AgentConnection

	// 服务器
	server *grpc.Server
	done   chan struct{}
}

// GameNodeServiceInterface 节点服务接口
type GameNodeServiceInterface interface {
	// 创建节点
	Create(ctx context.Context, node models.GameNode) error
	// 更新节点
	Update(ctx context.Context, node models.GameNode) error
	// 获取节点
	Get(ctx context.Context, id string) (models.GameNode, error)
	// 更新节点在线状态
	UpdateStatusOnlineStatus(ctx context.Context, id string, online bool) error
	// 更新节点指标
	UpdateStatusMetrics(ctx context.Context, id string, metrics models.MetricsInfo) error
	// 更新节点硬件和系统信息
	UpdateHardwareAndSystem(ctx context.Context, id string, hardware models.HardwareInfo, system models.SystemInfo) error
}

// GamePipelineServiceInterface Pipeline服务接口
type GamePipelineServiceInterface interface {
	// 创建Pipeline
	Create(ctx context.Context, pipeline *models.GamePipeline) error
	// 更新Pipeline
	Update(ctx context.Context, pipeline *models.GamePipeline) error
	// 获取Pipeline
	Get(ctx context.Context, id string) (*models.GamePipeline, error)
	// 更新Pipeline状态
	UpdateStatus(ctx context.Context, id string, status *models.PipelineStatus) error
	// 更新Step状态
	UpdateStepStatus(ctx context.Context, pipelineID string, stepID string, status *models.StepStatus) error
	// 取消Pipeline
	Cancel(ctx context.Context, id string) error
}

// AgentConnection Agent连接
type AgentConnection struct {
	nodeID     string
	lastActive time.Time
}

// ServerConfig 服务器配置
type ServerConfig struct {
	MaxConnections  int
	HeartbeatPeriod time.Duration
}

// NewGameNodeServer 创建新的游戏节点服务器
func NewGameNodeServer(
	nodeService GameNodeServiceInterface,
	pipelineService GamePipelineServiceInterface,
	logger utils.Logger,
	config *ServerConfig,
) *GameNodeServer {
	return &GameNodeServer{
		nodeService:     nodeService,
		pipelineService: pipelineService,
		logger:          logger,
		connections:     make(map[string]*AgentConnection),
		config:          config,
		server:          grpc.NewServer(),
		done:            make(chan struct{}),
	}
}

// Register 处理节点注册请求
func (s *GameNodeServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// 1. 验证节点信息完整性
	if req.Id == "" {
		return &pb.RegisterResponse{
			Success: false,
			Message: "节点ID不能为空",
		}, nil
	}
	if req.Type == "" {
		return &pb.RegisterResponse{
			Success: false,
			Message: "节点类型不能为空",
		}, nil
	}

	// 2. 检查节点是否已存在
	existingNode, err := s.nodeService.Get(ctx, req.Id)
	if err != nil && !strings.Contains(err.Error(), "节点不存在") {
		s.logger.Error("检查节点是否存在失败: %v", err)
		return &pb.RegisterResponse{
			Success: false,
			Message: "检查节点是否存在失败",
		}, nil
	}

	// 3. 获取节点维护状态
	var nodeState models.GameNodeStaticState
	if existingNode.ID != "" {
		nodeState = existingNode.State
	} else {
		nodeState = models.GameNodeStaticStateNormal // 新节点默认为正常状态
	}

	// 4. 根据维护状态决定后续行为
	switch nodeState {
	case models.GameNodeStaticStateNormal, models.GameNodeStaticStateMaintenance, models.GameNodeStaticStateDisabled:
		// 允许注册，继续处理
	default:
		s.logger.Error("无效的节点状态: %v", nodeState)
		return &pb.RegisterResponse{
			Success: false,
			Message: "无效的节点状态",
		}, nil
	}

	// 5. 构建或更新节点信息
	node := models.GameNode{
		ID:       req.Id,
		Alias:    req.Alias,
		Model:    req.Model,
		Type:     models.GameNodeType(req.Type),
		Location: req.Location,
		Labels:   req.Labels,
		State:    nodeState,
		Hardware: req.Hardware,
		System:   req.System,
		Status: models.GameNodeStatus{
			State:      models.GameNodeStateOnline,
			Online:     true,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
			Hardware:   models.HardwareInfo{},
			System:     models.SystemInfo{},
			Metrics:    models.MetricsInfo{},
		},
	}

	// 6. 创建或更新节点
	if existingNode.ID == "" {
		// 新节点
		node.CreatedAt = time.Now()
		node.UpdatedAt = time.Now()
		if err := s.nodeService.Create(ctx, node); err != nil {
			s.logger.Error("创建节点失败: %v", err)
			return &pb.RegisterResponse{
				Success: false,
				Message: "创建节点失败",
			}, nil
		}
		s.logger.Info("节点注册成功-新节点: %s", node.ID)
	} else {
		// 更新已有节点：只更新状态信息
		existingNode.Status.State = models.GameNodeStateOnline
		existingNode.Status.Online = true
		existingNode.Status.LastOnline = time.Now()

		if err := s.nodeService.Update(ctx, existingNode); err != nil {
			s.logger.Error("更新节点状态失败: %v", err)
			return &pb.RegisterResponse{
				Success: false,
				Message: "更新节点状态失败",
			}, nil
		}
		s.logger.Info("节点注册成功-已知节点: %s", existingNode.ID)
	}

	// 7. 返回注册响应
	return &pb.RegisterResponse{
		Success: true,
		Message: "",
		State:   convertToProtoStaticState(nodeState),
	}, nil
}

// convertToProtoStaticState 将 models.GameNodeStaticState 转换为 proto.GameNodeStaticState
func convertToProtoStaticState(state models.GameNodeStaticState) pb.GameNodeStaticState {
	switch state {
	case models.GameNodeStaticStateNormal:
		return pb.GameNodeStaticState_NODE_STATE_NORMAL
	case models.GameNodeStaticStateMaintenance:
		return pb.GameNodeStaticState_NODE_STATE_MAINTENANCE
	case models.GameNodeStaticStateDisabled:
		return pb.GameNodeStaticState_NODE_STATE_DISABLED
	default:
		return pb.GameNodeStaticState_NODE_STATE_NORMAL
	}
}

// Heartbeat 心跳检测
func (s *GameNodeServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	// 更新节点在线状态
	err := s.nodeService.UpdateStatusOnlineStatus(ctx, req.Id, true)
	if err != nil {
		s.logger.Error("更新节点在线状态失败: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update node status: %v", err))
	}

	s.logger.Debug("收到节点心跳: %s", req.Id)

	return &pb.HeartbeatResponse{
		Status:  "success",
		Message: "心跳接收成功",
	}, nil
}

// ReportMetrics 处理指标报告请求
func (s *GameNodeServer) ReportMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsResponse, error) {
	s.logger.Debug("收到指标报告请求", "node_id", req.Id)

	// 获取节点
	node, err := s.nodeService.Get(ctx, req.Id)
	if err != nil {
		s.logger.Error("获取节点失败", "node_id", req.Id, "error", err)
		return nil, status.Error(codes.NotFound, "节点不存在")
	}

	// 更新节点指标
	node.Status.Metrics = models.MetricsInfo{
		CPUs:     make([]models.CPUMetrics, len(req.Metrics.Cpus)),
		Memory:   models.MemoryMetrics{},
		GPUs:     make([]models.GPUMetrics, len(req.Metrics.Gpus)),
		Storages: make([]models.StorageMetrics, len(req.Metrics.Storages)),
		Network:  models.NetworkMetrics{},
	}

	// CPU指标
	for i, metric := range req.Metrics.Cpus {
		node.Status.Metrics.CPUs[i] = models.CPUMetrics{
			Model:   metric.Model,
			Cores:   metric.Cores,
			Threads: metric.Threads,
			Usage:   metric.Usage,
		}
	}

	// 内存指标
	node.Status.Metrics.Memory = models.MemoryMetrics{
		Total:     req.Metrics.Memory.Total,
		Available: req.Metrics.Memory.Available,
		Used:      req.Metrics.Memory.Used,
		Usage:     req.Metrics.Memory.Usage,
	}

	// GPU指标
	for i, metric := range req.Metrics.Gpus {
		node.Status.Metrics.GPUs[i] = models.GPUMetrics{
			Model:       metric.Model,
			MemoryTotal: metric.MemoryTotal,
			GPUUsage:    metric.GpuUsage,
			MemoryUsed:  metric.MemoryUsed,
			MemoryFree:  metric.MemoryFree,
			MemoryUsage: metric.MemoryUsage,
		}
	}

	// 存储指标
	for i, metric := range req.Metrics.Storages {
		node.Status.Metrics.Storages[i] = models.StorageMetrics{
			Path:     metric.Path,
			Type:     metric.Type,
			Model:    metric.Model,
			Capacity: metric.Capacity,
			Used:     metric.Used,
			Free:     metric.Free,
			Usage:    metric.Usage,
		}
	}

	// 网络指标
	node.Status.Metrics.Network = models.NetworkMetrics{
		InboundTraffic:  req.Metrics.Network.InboundTraffic,
		OutboundTraffic: req.Metrics.Network.OutboundTraffic,
		Connections:     req.Metrics.Network.Connections,
	}

	// 更新节点状态
	err = s.nodeService.UpdateStatusMetrics(ctx, req.Id, node.Status.Metrics)
	if err != nil {
		s.logger.Error("更新节点状态失败", "node_id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "更新节点状态失败")
	}

	s.logger.Debug("节点指标更新成功: %s", req.Id)

	return &pb.MetricsResponse{
		Success: true,
		Message: "指标更新成功",
	}, nil
}

// convertToModelsHardwareInfo 将 proto.HardwareInfo 转换为 models.HardwareInfo
func convertToModelsHardwareInfo(hardware *pb.HardwareInfo) models.HardwareInfo {
	info := models.HardwareInfo{
		CPUs:     make([]models.CPUDevice, len(hardware.Cpus)),
		Memories: make([]models.MemoryDevice, len(hardware.Memories)),
		GPUs:     make([]models.GPUDevice, len(hardware.Gpus)),
		Storages: make([]models.StorageDevice, len(hardware.Storages)),
		Networks: make([]models.NetworkDevice, len(hardware.Networks)),
	}

	// CPU信息
	for i, cpu := range hardware.Cpus {
		info.CPUs[i] = models.CPUDevice{
			Model:        cpu.Model,
			Cores:        cpu.Cores,
			Threads:      cpu.Threads,
			Frequency:    cpu.Frequency,
			Cache:        cpu.Cache,
			Architecture: cpu.Architecture,
		}
	}

	// 内存信息
	for i, memory := range hardware.Memories {
		info.Memories[i] = models.MemoryDevice{
			Size:      memory.Size,
			Type:      memory.Type,
			Frequency: memory.Frequency,
		}
	}

	// GPU信息
	for i, gpu := range hardware.Gpus {
		info.GPUs[i] = models.GPUDevice{
			Model:             gpu.Model,
			MemoryTotal:       gpu.MemoryTotal,
			Architecture:      gpu.Architecture,
			DriverVersion:     gpu.DriverVersion,
			ComputeCapability: gpu.ComputeCapability,
			TDP:               gpu.Tdp,
		}
	}

	// 存储设备信息
	for i, storage := range hardware.Storages {
		info.Storages[i] = models.StorageDevice{
			Type:     storage.Type,
			Model:    storage.Model,
			Capacity: storage.Capacity,
			Path:     storage.Path,
		}
	}

	// 网络设备信息
	for i, network := range hardware.Networks {
		info.Networks[i] = models.NetworkDevice{
			Name:       network.Name,
			MacAddress: network.MacAddress,
			IpAddress:  network.IpAddress,
			Speed:      network.Speed,
		}
	}

	return info
}

// convertToModelsSystemInfo 将 proto.SystemInfo 转换为 models.SystemInfo
func convertToModelsSystemInfo(system *pb.SystemInfo) models.SystemInfo {
	return models.SystemInfo{
		OSDistribution:       system.OsDistribution,
		OSVersion:            system.OsVersion,
		OSArchitecture:       system.OsArchitecture,
		KernelVersion:        system.KernelVersion,
		GPUDriverVersion:     system.GpuDriverVersion,
		GPUComputeAPIVersion: system.GpuComputeApiVersion,
		DockerVersion:        system.DockerVersion,
		ContainerdVersion:    system.ContainerdVersion,
		RuncVersion:          system.RuncVersion,
	}
}

// ReportResource 处理资源信息更新请求
func (s *GameNodeServer) ReportResource(ctx context.Context, req *pb.ResourceRequest) (*pb.ResourceResponse, error) {
	// 转换为硬件信息
	hardwareInfo := convertToModelsHardwareInfo(req.Hardware)
	systemInfo := convertToModelsSystemInfo(req.System)

	// 更新节点硬件和系统信息
	err := s.nodeService.UpdateHardwareAndSystem(ctx, req.NodeId, hardwareInfo, systemInfo)
	if err != nil {
		s.logger.Error("更新节点硬件和系统信息失败: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update hardware and system info: %v", err))
	}

	s.logger.Info("节点硬件和系统信息更新成功: %s", req.NodeId)

	return &pb.ResourceResponse{
		Success: true,
		Message: "Hardware and system info updated successfully",
	}, nil
}

// UpdateNodeState 处理节点状态变更请求
func (s *GameNodeServer) UpdateNodeState(ctx context.Context, req *pb.StateChangeRequest) (*pb.StateChangeResponse, error) {
	// 获取节点
	node, err := s.nodeService.Get(ctx, req.NodeId)
	if err != nil {
		s.logger.Error("获取节点失败: %v", err)
		return nil, status.Error(codes.NotFound, "节点不存在")
	}

	// 将 proto 的 GameNodeStaticState 转换为 models.GameNodeState
	var targetState models.GameNodeState
	switch req.TargetState {
	case pb.GameNodeStaticState_NODE_STATE_NORMAL:
		targetState = models.GameNodeStateOnline
	case pb.GameNodeStaticState_NODE_STATE_MAINTENANCE:
		targetState = models.GameNodeStateMaintenance
	case pb.GameNodeStaticState_NODE_STATE_DISABLED:
		targetState = models.GameNodeStateOffline
	default:
		s.logger.Error("无效的目标状态: %v", req.TargetState)
		return nil, status.Error(codes.InvalidArgument, "无效的目标状态")
	}

	// 更新节点状态
	node.Status.State = targetState
	err = s.nodeService.Update(ctx, node)
	if err != nil {
		s.logger.Error("更新节点状态失败: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update node state: %v", err))
	}

	s.logger.Info("节点状态更新成功: %s, 新状态=%s", req.NodeId, targetState)

	return &pb.StateChangeResponse{
		Success:      true,
		ErrorMessage: "",
		ConfirmTime:  timestamppb.Now(),
	}, nil
}

// Stop 停止服务器
func (s *GameNodeServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.done)
	s.server.GracefulStop()

	return nil
}

// Start 启动服务器
func (s *GameNodeServer) Start(ctx context.Context, opts *GameNodeServerOptions) error {
	// 创建 gRPC 服务器
	pb.RegisterGameNodeGRPCServiceServer(s.server, s)

	// 创建监听器
	listener, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		s.logger.Error("监听地址失败: %v", err)
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.logger.Info("服务器开始监听地址: %s", opts.ListenAddr)

	// 启动心跳检查
	go s.startHeartbeatCheck(ctx)

	// 启动服务器
	go func() {
		if err := s.server.Serve(listener); err != nil {
			s.logger.Error("gRPC服务器运行失败: %v", err)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()
	s.logger.Info("收到停止信号，正在关闭服务器")

	// 优雅关闭
	s.server.GracefulStop()
	return nil
}

// startHeartbeatCheck 启动心跳检查
func (s *GameNodeServer) startHeartbeatCheck(ctx context.Context) {
	ticker := time.NewTicker(s.config.HeartbeatPeriod)
	defer ticker.Stop()

	s.logger.Info("心跳检查服务启动，周期: %v", s.config.HeartbeatPeriod)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("心跳检查服务停止")
			return
		case <-ticker.C:
			s.checkHeartbeats(ctx)
		}
	}
}

// checkHeartbeats 检查心跳
func (s *GameNodeServer) checkHeartbeats(ctx context.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	for nodeID := range s.connections {
		// 获取节点状态
		node, err := s.nodeService.Get(ctx, nodeID)
		if err != nil {
			s.logger.Error("获取节点状态失败: %v", err)
			continue
		}

		// 检查最后在线时间
		if now.Sub(node.Status.LastOnline) > s.config.HeartbeatPeriod*2 {
			// 更新节点状态为离线
			if err := s.nodeService.UpdateStatusOnlineStatus(ctx, nodeID, false); err != nil {
				s.logger.Error("更新节点状态失败: %v", err)
				continue
			}

			s.logger.Warn("节点心跳超时，标记为离线: %s", nodeID)

			// 从连接池中移除
			delete(s.connections, nodeID)
		}
	}
}
