package gamenode

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	dockerclient "github.com/docker/docker/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-beagle/beagle-wind-game/internal/event"
	"github.com/open-beagle/beagle-wind-game/internal/log"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/sysinfo"
)

// GameNodeAgent 游戏节点代理
type GameNodeAgent struct {
	// 基本信息
	id string

	// 连接信息
	serverAddr string
	conn       *grpc.ClientConn
	client     pb.GameNodeGRPCServiceClient

	// 状态管理
	mu        sync.RWMutex
	status    *models.GameNodeStatus
	pipelines map[string]*models.GameNodePipeline

	// 事件管理
	eventManager event.EventManager

	// 日志管理
	logManager log.LogManager

	// 配置
	opts *Options

	// Docker 客户端
	dockerClient *dockerclient.Client

	// 系统信息采集器
	hardwareCollector *sysinfo.HardwareCollector
	systemCollector   *sysinfo.SystemCollector
	metricsCollector  *sysinfo.MetricsCollector

	// 节点基本信息（原来存储在status中，但根据模型定义需要分开存储）
	nodeInfo struct {
		name      string
		namespace string
		nodeType  models.GameNodeType
		ip        string
		labels    map[string]string
	}
}

// NewGameNodeAgent 创建新的游戏节点代理
func NewGameNodeAgent(eventManager event.EventManager, logManager log.LogManager, dockerClient *dockerclient.Client, opts ...Option) *GameNodeAgent {
	options := applyOptions(opts...)

	// 创建系统信息采集器配置
	sysInfoOpts := make(map[string]string)

	// 检查是否提供了自定义GPU工具路径
	if path, exists := os.LookupEnv("NVIDIA_SMI_PATH"); exists {
		sysInfoOpts["nvidia_smi_path"] = path
	}
	if path, exists := os.LookupEnv("ROCM_SMI_PATH"); exists {
		sysInfoOpts["rocm_smi_path"] = path
	}
	if path, exists := os.LookupEnv("INTEL_GPU_TOP_PATH"); exists {
		sysInfoOpts["intel_gpu_top_path"] = path
	}

	agent := &GameNodeAgent{
		id:           options.ID,
		serverAddr:   options.ServerAddr,
		eventManager: eventManager,
		logManager:   logManager,
		opts:         options,
		dockerClient: dockerClient,
		status: &models.GameNodeStatus{
			State:      models.GameNodeStateOffline,
			Online:     false,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
			Hardware:   models.HardwareInfo{},
			System:     models.SystemInfo{},
			Metrics:    models.MetricsInfo{},
		},
		pipelines: make(map[string]*models.GameNodePipeline),

		// 初始化系统信息采集器
		hardwareCollector: sysinfo.NewHardwareCollector(sysInfoOpts),
		systemCollector:   sysinfo.NewSystemCollector(sysInfoOpts),
		metricsCollector:  sysinfo.NewMetricsCollector(sysInfoOpts),
	}

	// 设置节点基本信息
	agent.nodeInfo.name = options.Name
	agent.nodeInfo.namespace = options.Namespace
	agent.nodeInfo.nodeType = options.NodeType
	agent.nodeInfo.ip = options.IP
	agent.nodeInfo.labels = options.Labels

	// 初始化 gRPC 连接
	err := agent.connectClient()
	if err != nil {
		// 记录错误但不中断初始化流程，将在Start时重试
		eventManager.Publish(event.NewNodeEvent(agent.id, "warn", fmt.Sprintf("初始化连接失败，将在启动时重试: %v", err)))
	}

	return agent
}

// connectClient 连接gRPC服务
func (a *GameNodeAgent) connectClient() error {
	// 实现指数退避重试逻辑
	var err error
	maxRetries := 5
	initialBackoff := 500 * time.Millisecond
	maxBackoff := 10 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// 计算退避时间
			backoff := initialBackoff * time.Duration(1<<uint(i-1))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			a.eventManager.Publish(event.NewNodeEvent(a.id, "info", fmt.Sprintf("重试连接服务器 (尝试 %d/%d)，等待 %v", i+1, maxRetries, backoff)))
			time.Sleep(backoff)
		}

		// 尝试建立连接
		if err = a.tryConnect(); err == nil {
			// 连接成功
			return nil
		}
	}

	// 所有重试都失败
	return fmt.Errorf("连接服务器失败，已重试 %d 次: %v", maxRetries, err)
}

// tryConnect 尝试建立单次连接
func (a *GameNodeAgent) tryConnect() error {
	// 设置连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 关闭旧连接
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
		a.client = nil
	}

	// 使用NewClient建立连接
	clientConn, err := grpc.NewClient(a.serverAddr, opts...)
	if err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "error", fmt.Sprintf("连接服务器失败: %v", err)))
		return err
	}

	// 检查连接状态
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for {
			state := clientConn.GetState()
			if state == connectivity.Ready {
				a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "连接已就绪"))
				return
			} else if state == connectivity.TransientFailure || state == connectivity.Shutdown {
				a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("连接状态异常: %v", state)))
				return
			}

			if !clientConn.WaitForStateChange(ctx, state) {
				a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", "等待连接状态变化超时"))
				return
			}
		}
	}()

	// 设置连接和客户端
	a.conn = clientConn
	a.client = pb.NewGameNodeGRPCServiceClient(clientConn)
	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "已成功创建连接到服务器"))

	return nil
}

// Start 启动代理
func (a *GameNodeAgent) Start(ctx context.Context) error {
	// 检测节点类型
	nodeType := a.DetectNodeType()
	a.nodeInfo.nodeType = nodeType
	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", fmt.Sprintf("检测到节点类型: %s", nodeType)))

	// 建立连接
	if a.client == nil {
		if err := a.connectClient(); err != nil {
			return fmt.Errorf("建立连接失败: %v", err)
		}
	}

	// 注册节点
	if err := a.Register(ctx); err != nil {
		return fmt.Errorf("注册失败: %v", err)
	}

	// 启动心跳
	go a.startHeartbeat(ctx)

	// 启动指标采集
	go a.startMetricsCollection(ctx)

	return nil
}

// DetectNodeType 自动检测节点类型
func (a *GameNodeAgent) DetectNodeType() models.GameNodeType {
	// 1. 检查是否在容器内运行
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return models.GameNodeTypeContainer
	}

	// 2. 检查 cgroup 路径
	cgroupPath := "/proc/1/cgroup"
	if content, err := os.ReadFile(cgroupPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "docker") ||
			strings.Contains(contentStr, "kubepods") ||
			strings.Contains(contentStr, "containerd") {
			return models.GameNodeTypeContainer
		}
	}

	// 3. 检查虚拟化特征
	if _, err := os.Stat("/sys/hypervisor"); err == nil {
		return models.GameNodeTypeVirtual
	}

	if content, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "hypervisor") ||
			strings.Contains(contentStr, "VMware") ||
			strings.Contains(contentStr, "KVM") ||
			strings.Contains(contentStr, "Microsoft Hv") {
			return models.GameNodeTypeVirtual
		}
	}

	// 默认为物理节点
	return models.GameNodeTypePhysical
}

// startHeartbeat 启动心跳
func (a *GameNodeAgent) startHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(a.opts.HeartbeatPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.Heartbeat(ctx); err != nil {
				a.eventManager.Publish(event.NewNodeEvent(a.id, "error", fmt.Sprintf("Heartbeat failed: %v", err)))
			}
		}
	}
}

// startMetricsCollection 启动指标采集
func (a *GameNodeAgent) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(a.opts.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.ReportMetrics(ctx); err != nil {
				a.eventManager.Publish(event.NewNodeEvent(a.id, "error", fmt.Sprintf("Metrics collection failed: %v", err)))
			}
		}
	}
}

// Register 注册节点
func (a *GameNodeAgent) Register(ctx context.Context) error {
	// 收集节点信息
	if err := a.collectNodeInfo(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "error", fmt.Sprintf("收集节点信息失败: %v", err)))
		return err
	}

	// 收集硬件和系统信息
	if err := a.collectHardwareAndSystemInfo(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("收集硬件和系统信息失败: %v", err)))
	}

	// 收集指标信息
	if err := a.collectMetricsInfo(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("收集指标信息失败: %v", err)))
	}

	// 准备硬件和系统信息的简化版表示（兼容现有API）
	hardwareConfig, err := a.hardwareCollector.GetSimplifiedHardwareInfo()
	if err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("获取简化版硬件信息失败: %v", err)))
	}

	systemConfig, err := a.systemCollector.GetSimplifiedSystemInfo()
	if err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("获取简化版系统信息失败: %v", err)))
	}

	// 发送注册请求
	req := &pb.RegisterRequest{
		Id:       a.id,
		Type:     string(a.nodeInfo.nodeType),
		Hardware: hardwareConfig,
		System:   systemConfig,
		Labels:   a.nodeInfo.labels,
	}

	resp, err := a.client.Register(ctx, req)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to register: %v", err))
	}

	// 如果注册失败，返回错误
	if !resp.Success {
		return status.Error(codes.Internal, fmt.Sprintf("registration failed: %s", resp.Message))
	}

	// 更新状态
	a.mu.Lock()
	a.status.State = models.GameNodeStateOnline
	a.status.Online = true
	a.status.LastOnline = time.Now()
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	// 发布注册成功事件
	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "节点注册成功"))

	return nil
}

// Heartbeat 发送心跳
func (a *GameNodeAgent) Heartbeat(ctx context.Context) error {
	// 首先更新硬件和系统信息
	if err := a.collectHardwareAndSystemInfo(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("收集硬件和系统信息失败: %v", err)))
	}

	// 收集最新的指标信息
	if err := a.collectMetricsInfo(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("收集指标信息失败: %v", err)))
	}

	// 更新在线状态和时间
	a.mu.Lock()
	a.status.Online = true
	a.status.State = models.GameNodeStateOnline
	a.status.LastOnline = time.Now()
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	// 发送心跳请求
	req := &pb.HeartbeatRequest{
		Id:        a.id,
		SessionId: a.id,
		Timestamp: time.Now().Unix(),
		Status:    a.collectProtobufNodeStatus(),
	}

	_, err := a.client.Heartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("发送心跳失败: %v", err)
	}

	return nil
}

// ReportMetrics 报告指标信息
func (a *GameNodeAgent) ReportMetrics(ctx context.Context) error {
	// 首先更新硬件和系统信息
	if err := a.collectHardwareAndSystemInfo(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("收集硬件和系统信息失败: %v", err)))
	}

	// 收集最新的指标信息
	if err := a.collectMetricsInfo(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("收集指标信息失败: %v", err)))
		return err
	}

	// 更新状态
	a.mu.Lock()
	a.status.UpdatedAt = time.Now()
	// 确保节点状态为在线
	if a.status.State != models.GameNodeStateOffline {
		a.status.State = models.GameNodeStateOnline
		a.status.Online = true
	}
	a.mu.Unlock()

	// 准备指标报告
	report := a.prepareMetricsReport()

	// 发送指标请求
	_, err := a.client.ReportMetrics(ctx, report)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to report metrics: %v", err))
	}

	// 同步更新状态 - 发送完整的资源信息
	resourceInfo := a.collectProtobufResourceInfo()
	_, updateErr := a.client.UpdateResourceInfo(ctx, resourceInfo)
	if updateErr != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("更新资源信息失败: %v", updateErr)))
	}

	return nil
}

// prepareMetricsReport 准备指标报告
func (a *GameNodeAgent) prepareMetricsReport() *pb.MetricsReport {
	a.mu.RLock()
	defer a.mu.RUnlock()

	report := &pb.MetricsReport{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Metrics:   []*pb.Metric{},
	}

	// 添加CPU指标
	if len(a.status.Metrics.CPUs) > 0 {
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "cpu.usage",
			Type:  "gauge",
			Value: a.status.Metrics.CPUs[0].Usage,
		})
	}

	// 添加内存指标
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "memory.usage",
		Type:  "gauge",
		Value: a.status.Metrics.Memory.Usage,
	})

	// 添加GPU指标
	if len(a.status.Metrics.GPUs) > 0 {
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "gpu.usage",
			Type:  "gauge",
			Value: a.status.Metrics.GPUs[0].Usage,
		})
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "gpu.memory_usage",
			Type:  "gauge",
			Value: a.status.Metrics.GPUs[0].MemoryUsage,
		})
	}

	// 添加存储指标
	if len(a.status.Metrics.Storages) > 0 {
		// 只报告重要的存储设备（最多3个）
		maxDevicesToReport := 3
		devicesToReport := len(a.status.Metrics.Storages)
		if devicesToReport > maxDevicesToReport {
			devicesToReport = maxDevicesToReport
		}

		for i := 0; i < devicesToReport; i++ {
			device := a.status.Metrics.Storages[i]

			// 确保设备有路径和容量
			if device.Path == "" || device.Capacity == 0 {
				continue
			}

			prefix := fmt.Sprintf("storage.%d", i)
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".usage",
				Type:  "gauge",
				Value: device.Usage,
			})
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".used",
				Type:  "gauge",
				Value: float64(device.Used),
			})
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".free",
				Type:  "gauge",
				Value: float64(device.Free),
			})
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".capacity",
				Type:  "gauge",
				Value: float64(device.Capacity),
			})

			// 添加可选的路径标识
			if device.Path != "" {
				report.Metrics = append(report.Metrics, &pb.Metric{
					Name:   prefix + ".path",
					Type:   "label",
					Value:  0,
					Labels: map[string]string{"path": device.Path},
				})
			}

			// 添加可选的类型标识
			if device.Type != "" {
				report.Metrics = append(report.Metrics, &pb.Metric{
					Name:   prefix + ".type",
					Type:   "label",
					Value:  0,
					Labels: map[string]string{"type": device.Type},
				})
			}
		}
	}

	// 添加网络指标
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "network.inbound",
		Type:  "gauge",
		Value: a.status.Metrics.Network.InboundTraffic,
	})
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "network.outbound",
		Type:  "gauge",
		Value: a.status.Metrics.Network.OutboundTraffic,
	})

	return report
}

// collectProtobufResourceInfo 收集protobuf格式资源信息
func (a *GameNodeAgent) collectProtobufResourceInfo() *pb.ResourceInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 创建HardwareInfo - 使用新的扁平结构
	pbHardware := &pb.HardwareInfo{
		Cpus:     []*pb.CPUHardware{},
		Memories: []*pb.MemoryHardware{},
		Gpus:     []*pb.GPUHardware{},
		Storages: []*pb.StorageDevice{},
		Networks: []*pb.NetworkDevice{},
	}

	// CPU信息
	for _, device := range a.status.Hardware.CPUs {
		pbHardware.Cpus = append(pbHardware.Cpus, &pb.CPUHardware{
			Model:        device.Model,
			Cores:        device.Cores,
			Threads:      device.Threads,
			Frequency:    device.Frequency,
			Cache:        device.Cache,
			Architecture: device.Architecture,
		})
	}

	// 内存信息
	for _, device := range a.status.Hardware.Memories {
		pbHardware.Memories = append(pbHardware.Memories, &pb.MemoryHardware{
			Size:      device.Size,
			Type:      device.Type,
			Frequency: device.Frequency,
		})
	}

	// GPU信息
	for _, device := range a.status.Hardware.GPUs {
		pbHardware.Gpus = append(pbHardware.Gpus, &pb.GPUHardware{
			Model:             device.Model,
			MemoryTotal:       device.MemoryTotal,
			Architecture:      device.Architecture,
			DriverVersion:     device.DriverVersion,
			ComputeCapability: device.ComputeCapability,
			Tdp:               device.TDP,
		})
	}

	// 存储信息
	for _, device := range a.status.Hardware.Storages {
		pbHardware.Storages = append(pbHardware.Storages, &pb.StorageDevice{
			Type:     device.Type,
			Model:    device.Model,
			Capacity: device.Capacity,
			Path:     device.Path,
		})
	}

	// 网络信息
	for _, device := range a.status.Hardware.Networks {
		pbHardware.Networks = append(pbHardware.Networks, &pb.NetworkDevice{
			Name:       device.Name,
			MacAddress: device.MacAddress,
			IpAddress:  device.IpAddress,
			Speed:      device.Speed,
		})
	}

	// 创建SystemInfo
	pbSystem := &pb.SystemInfo{
		OsDistribution:       a.status.System.OSDistribution,
		OsVersion:            a.status.System.OSVersion,
		OsArchitecture:       a.status.System.OSArchitecture,
		KernelVersion:        a.status.System.KernelVersion,
		GpuDriverVersion:     a.status.System.GPUDriverVersion,
		GpuComputeApiVersion: a.status.System.GPUComputeAPIVersion,
		DockerVersion:        a.status.System.DockerVersion,
		ContainerdVersion:    a.status.System.ContainerdVersion,
		RuncVersion:          a.status.System.RuncVersion,
	}

	// 创建MetricsInfo
	pbMetrics := &pb.MetricsInfo{
		Cpus:     []*pb.CPUMetrics{},
		Memory:   &pb.MemoryMetrics{},
		Gpus:     []*pb.GPUMetrics{},
		Storages: []*pb.StorageMetrics{},
		Network:  &pb.NetworkMetrics{},
	}

	// CPU指标
	for _, metric := range a.status.Metrics.CPUs {
		pbMetrics.Cpus = append(pbMetrics.Cpus, &pb.CPUMetrics{
			Model:   metric.Model,
			Cores:   metric.Cores,
			Threads: metric.Threads,
			Usage:   metric.Usage,
		})
	}

	// 内存指标
	pbMetrics.Memory = &pb.MemoryMetrics{
		Total:     a.status.Metrics.Memory.Total,
		Available: a.status.Metrics.Memory.Available,
		Used:      a.status.Metrics.Memory.Used,
		Usage:     a.status.Metrics.Memory.Usage,
	}

	// GPU指标
	for _, metric := range a.status.Metrics.GPUs {
		pbMetrics.Gpus = append(pbMetrics.Gpus, &pb.GPUMetrics{
			Model:       metric.Model,
			MemoryTotal: metric.MemoryTotal,
			Usage:       metric.Usage,
			MemoryUsed:  metric.MemoryUsed,
			MemoryFree:  metric.MemoryFree,
			MemoryUsage: metric.MemoryUsage,
		})
	}

	// 存储指标
	for _, metric := range a.status.Metrics.Storages {
		pbMetrics.Storages = append(pbMetrics.Storages, &pb.StorageMetrics{
			Path:     metric.Path,
			Type:     metric.Type,
			Model:    metric.Model,
			Capacity: metric.Capacity,
			Used:     metric.Used,
			Free:     metric.Free,
			Usage:    metric.Usage,
		})
	}

	// 网络指标
	pbMetrics.Network = &pb.NetworkMetrics{
		InboundTraffic:  a.status.Metrics.Network.InboundTraffic,
		OutboundTraffic: a.status.Metrics.Network.OutboundTraffic,
		Connections:     a.status.Metrics.Network.Connections,
	}

	// 节点类型转换
	var nodeType pb.NodeType
	switch a.nodeInfo.nodeType {
	case models.GameNodeTypePhysical:
		nodeType = pb.NodeType_NODE_TYPE_PHYSICAL
	case models.GameNodeTypeVirtual:
		nodeType = pb.NodeType_NODE_TYPE_VIRTUAL
	case models.GameNodeTypeContainer:
		nodeType = pb.NodeType_NODE_TYPE_CONTAINER
	default:
		nodeType = pb.NodeType_NODE_TYPE_UNKNOWN
	}

	// 创建ResourceInfo
	resourceInfo := &pb.ResourceInfo{
		NodeId:    a.id,
		NodeName:  a.id,
		NodeState: pb.GameNodeState_NODE_STATE_OFFLINE, // 默认值，后面会根据实际状态更新
		NodeType:  nodeType,
		Hardware:  pbHardware,
		System:    pbSystem,
		Metrics:   pbMetrics,
	}

	// 更新节点状态
	switch a.status.State {
	case models.GameNodeStateOffline:
		resourceInfo.NodeState = pb.GameNodeState_NODE_STATE_OFFLINE
	case models.GameNodeStateOnline:
		resourceInfo.NodeState = pb.GameNodeState_NODE_STATE_ONLINE
	case models.GameNodeStateMaintenance:
		resourceInfo.NodeState = pb.GameNodeState_NODE_STATE_MAINTENANCE
	case models.GameNodeStateReady:
		resourceInfo.NodeState = pb.GameNodeState_NODE_STATE_READY
	case models.GameNodeStateBusy:
		resourceInfo.NodeState = pb.GameNodeState_NODE_STATE_BUSY
	case models.GameNodeStateError:
		resourceInfo.NodeState = pb.GameNodeState_NODE_STATE_ERROR
	default:
		resourceInfo.NodeState = pb.GameNodeState_NODE_STATE_OFFLINE
	}

	return resourceInfo
}

// collectProtobufNodeStatus 收集protobuf格式的节点状态
func (a *GameNodeAgent) collectProtobufNodeStatus() *pb.GameNodeStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 处理状态转换
	var pbState pb.GameNodeState
	switch a.status.State {
	case models.GameNodeStateOffline:
		pbState = pb.GameNodeState_NODE_STATE_OFFLINE
	case models.GameNodeStateOnline:
		pbState = pb.GameNodeState_NODE_STATE_ONLINE
	case models.GameNodeStateMaintenance:
		pbState = pb.GameNodeState_NODE_STATE_MAINTENANCE
	case models.GameNodeStateReady:
		pbState = pb.GameNodeState_NODE_STATE_READY
	case models.GameNodeStateBusy:
		pbState = pb.GameNodeState_NODE_STATE_BUSY
	case models.GameNodeStateError:
		pbState = pb.GameNodeState_NODE_STATE_ERROR
	default:
		pbState = pb.GameNodeState_NODE_STATE_OFFLINE
	}

	// 创建状态对象
	pbStatus := &pb.GameNodeStatus{
		State:    pbState,
		Online:   a.status.Online,
		Hardware: a.collectProtobufResourceInfo().Hardware,
		System:   a.collectProtobufResourceInfo().System,
		Metrics:  a.collectProtobufResourceInfo().Metrics,
	}

	// 转换时间戳
	if !a.status.LastOnline.IsZero() {
		pbStatus.LastOnline = timestamppb.New(a.status.LastOnline)
	}
	if !a.status.UpdatedAt.IsZero() {
		pbStatus.UpdatedAt = timestamppb.New(a.status.UpdatedAt)
	}

	return pbStatus
}

// collectNodeInfo 收集节点基本信息
func (a *GameNodeAgent) collectNodeInfo() error {
	// 基本信息已在初始化时设置
	return nil
}

// collectHardwareAndSystemInfo 收集硬件和系统信息
func (a *GameNodeAgent) collectHardwareAndSystemInfo() error {
	// 收集硬件信息
	hardwareInfo, err := a.hardwareCollector.GetHardwareInfo()
	if err != nil {
		return fmt.Errorf("获取硬件信息失败: %v", err)
	}

	// 收集系统信息
	systemInfo, err := a.systemCollector.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("获取系统信息失败: %v", err)
	}

	// 更新状态
	a.mu.Lock()
	a.status.Hardware = hardwareInfo
	a.status.System = systemInfo
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	// 如果客户端连接有效，异步发送资源信息更新
	go func() {
		if a.client == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 使用已有的ResourceInfo接口发送硬件和系统信息
		resourceInfo := a.collectProtobufResourceInfo()
		_, err := a.client.UpdateResourceInfo(ctx, resourceInfo)
		if err != nil {
			a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("更新资源信息失败: %v", err)))
		}
	}()

	return nil
}

// collectMetricsInfo 收集指标信息
func (a *GameNodeAgent) collectMetricsInfo() error {
	// 收集指标信息
	newMetricsInfo, err := a.metricsCollector.GetMetricsInfo()
	if err != nil {
		return fmt.Errorf("获取指标信息失败: %v", err)
	}

	// 更新状态，保留历史数据（仅更新有采集到数据的指标）
	a.mu.Lock()
	defer a.mu.Unlock()

	// 获取当前状态的副本，用于保留历史数据
	currentMetrics := a.status.Metrics

	// 1. 合并CPU指标
	if len(newMetricsInfo.CPUs) > 0 {
		// 新指标中有CPU数据，更新CPU数据
		if len(currentMetrics.CPUs) == 0 {
			// 当前无CPU数据，直接使用新数据
			currentMetrics.CPUs = newMetricsInfo.CPUs
		} else {
			// 合并CPU指标：保留基本信息，更新最新的用量指标
			for i, newCPU := range newMetricsInfo.CPUs {
				if i < len(currentMetrics.CPUs) {
					// 仅更新有效的使用率
					if newCPU.Usage > 0 {
						currentMetrics.CPUs[i].Usage = newCPU.Usage
					}

					// 只有当当前值为空且新值不为空时才更新这些基础信息
					if currentMetrics.CPUs[i].Model == "" && newCPU.Model != "" {
						currentMetrics.CPUs[i].Model = newCPU.Model
					}
					if currentMetrics.CPUs[i].Cores == 0 && newCPU.Cores > 0 {
						currentMetrics.CPUs[i].Cores = newCPU.Cores
					}
					if currentMetrics.CPUs[i].Threads == 0 && newCPU.Threads > 0 {
						currentMetrics.CPUs[i].Threads = newCPU.Threads
					}
				} else {
					// 添加新的CPU
					currentMetrics.CPUs = append(currentMetrics.CPUs, newCPU)
				}
			}
		}
	}

	// 2. 合并内存指标
	if newMetricsInfo.Memory.Total > 0 {
		// 内存总量不为0，表示采集到了有效数据
		currentMetrics.Memory = newMetricsInfo.Memory
	}

	// 3. 合并GPU指标
	if len(newMetricsInfo.GPUs) > 0 {
		// 新指标中有GPU数据，更新GPU数据
		if len(currentMetrics.GPUs) == 0 {
			// 当前无GPU数据，直接使用新数据
			currentMetrics.GPUs = newMetricsInfo.GPUs
		} else {
			// 合并GPU指标：保留基本信息，更新最新的用量指标
			for i, newGPU := range newMetricsInfo.GPUs {
				if i < len(currentMetrics.GPUs) {
					// 只更新有效的使用率数据
					if newGPU.Usage > 0 {
						currentMetrics.GPUs[i].Usage = newGPU.Usage
					}
					if newGPU.MemoryUsage > 0 {
						currentMetrics.GPUs[i].MemoryUsage = newGPU.MemoryUsage
					}
					if newGPU.MemoryUsed > 0 {
						currentMetrics.GPUs[i].MemoryUsed = newGPU.MemoryUsed
					}
					if newGPU.MemoryFree > 0 {
						currentMetrics.GPUs[i].MemoryFree = newGPU.MemoryFree
					}

					// 只有当当前值为空且新值不为空时才更新这些基础信息
					if currentMetrics.GPUs[i].Model == "" && newGPU.Model != "" {
						currentMetrics.GPUs[i].Model = newGPU.Model
					}
					if currentMetrics.GPUs[i].MemoryTotal == 0 && newGPU.MemoryTotal > 0 {
						currentMetrics.GPUs[i].MemoryTotal = newGPU.MemoryTotal
					}
				} else {
					// 添加新的GPU
					currentMetrics.GPUs = append(currentMetrics.GPUs, newGPU)
				}
			}
		}
	}

	// 4. 合并存储指标
	if len(newMetricsInfo.Storages) > 0 {
		// 定义最小有效存储大小（100MB），过滤掉太小的设备
		const minValidStorageSize int64 = 100 * 1024 * 1024

		// 首先使用路径和容量范围作为键构建映射
		storageMap := make(map[string]models.StorageMetrics)

		// 工具函数：生成存储设备的唯一键
		generateStorageKey := func(storage models.StorageMetrics) string {
			if storage.Path != "" {
				return storage.Path // 直接使用路径作为唯一键
			}

			// 如果是WSL驱动器，使用特殊标识
			if storage.Type == "Virtual" && storage.Path != "" && len(storage.Path) >= 6 && strings.HasPrefix(storage.Path, "/mnt/") {
				driveLetter := storage.Path[5:6]
				return fmt.Sprintf("wsl-drive-%s", strings.ToUpper(driveLetter))
			}

			// 其他设备类型，使用容量范围作为键
			sizeGB := float64(storage.Capacity) / (1024 * 1024 * 1024)
			// 舍入到最接近的10GB
			sizeRounded := math.Round(sizeGB/10) * 10
			return fmt.Sprintf("%.0fGB-%s", sizeRounded, storage.Type)
		}

		// 首先添加当前存储设备（保留有效的设备）
		for _, storage := range currentMetrics.Storages {
			// 跳过无效或容量太小的设备
			if storage.Path == "" || storage.Capacity < minValidStorageSize {
				continue
			}

			key := generateStorageKey(storage)
			storageMap[key] = storage
		}

		// 然后处理新的存储指标
		for _, newStorage := range newMetricsInfo.Storages {
			// 跳过无效或太小的存储设备
			if newStorage.Path == "" || newStorage.Capacity < minValidStorageSize {
				continue
			}

			// 检查是否有相同路径的设备
			key := generateStorageKey(newStorage)
			if existing, exists := storageMap[key]; exists {
				// 只更新有效的指标数据
				if newStorage.Usage > 0 {
					existing.Usage = newStorage.Usage
				}
				if newStorage.Used > 0 {
					existing.Used = newStorage.Used
				}
				if newStorage.Free > 0 {
					existing.Free = newStorage.Free
				}

				// 只在当前值为空且新值不为空时才更新静态信息
				if existing.Model == "" && newStorage.Model != "" {
					existing.Model = newStorage.Model
				}
				if existing.Type == "" && newStorage.Type != "" {
					existing.Type = newStorage.Type
				}

				storageMap[key] = existing
			} else {
				// 添加新设备
				storageMap[key] = newStorage
			}
		}

		// 转换回数组并按容量排序
		var storageSlice []models.StorageMetrics
		for _, storage := range storageMap {
			storageSlice = append(storageSlice, storage)
		}

		// 按容量降序排序
		sort.Slice(storageSlice, func(i, j int) bool {
			return storageSlice[i].Capacity > storageSlice[j].Capacity
		})

		// 更新存储设备列表
		currentMetrics.Storages = storageSlice
	}

	// 5. 更新网络指标 - 只有当采集到的数据有效时才更新
	if newMetricsInfo.Network.InboundTraffic > 0 || newMetricsInfo.Network.OutboundTraffic > 0 || newMetricsInfo.Network.Connections > 0 {
		currentMetrics.Network = newMetricsInfo.Network
	}

	// 更新状态
	a.status.Metrics = currentMetrics
	a.status.UpdatedAt = time.Now()

	// 解锁以便调用collectMetricsFromHardware
	a.mu.Unlock()

	// 从硬件信息中补充完善指标信息（CPU/GPU型号、存储路径等）
	if err := a.collectMetricsFromHardware(); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("从硬件信息补充指标失败: %v", err)))
	}

	// 重新上锁，以便后续操作
	a.mu.Lock()

	// 异步发送指标报告
	go a.asyncReportMetrics()

	return nil
}

// asyncReportMetrics 异步发送指标报告
func (a *GameNodeAgent) asyncReportMetrics() {
	if a.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取当前指标
	a.mu.RLock()
	metricsInfo := a.status.Metrics
	a.mu.RUnlock()

	// 将指标信息转换为MetricsReport格式
	report := &pb.MetricsReport{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Metrics:   []*pb.Metric{},
	}

	// 添加CPU指标
	if len(metricsInfo.CPUs) > 0 {
		device := metricsInfo.CPUs[0]
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "cpu.usage",
			Type:  "gauge",
			Value: device.Usage,
		})
	}

	// 添加内存指标
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "memory.usage",
		Type:  "gauge",
		Value: metricsInfo.Memory.Usage,
	})
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "memory.used",
		Type:  "gauge",
		Value: float64(metricsInfo.Memory.Used),
	})
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "memory.available",
		Type:  "gauge",
		Value: float64(metricsInfo.Memory.Available),
	})
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "memory.total",
		Type:  "gauge",
		Value: float64(metricsInfo.Memory.Total),
	})

	// 添加GPU指标
	if len(metricsInfo.GPUs) > 0 {
		device := metricsInfo.GPUs[0]
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "gpu.usage",
			Type:  "gauge",
			Value: device.Usage,
		})
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "gpu.memory_usage",
			Type:  "gauge",
			Value: device.MemoryUsage,
		})
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "gpu.memory_used",
			Type:  "gauge",
			Value: float64(device.MemoryUsed),
		})
		report.Metrics = append(report.Metrics, &pb.Metric{
			Name:  "gpu.memory_free",
			Type:  "gauge",
			Value: float64(device.MemoryFree),
		})
	}

	// 添加存储指标
	if len(metricsInfo.Storages) > 0 {
		// 只报告重要的存储设备（最多3个）
		maxDevicesToReport := 3
		devicesToReport := len(metricsInfo.Storages)
		if devicesToReport > maxDevicesToReport {
			devicesToReport = maxDevicesToReport
		}

		for i := 0; i < devicesToReport; i++ {
			device := metricsInfo.Storages[i]

			// 确保设备有路径和容量
			if device.Path == "" || device.Capacity == 0 {
				continue
			}

			prefix := fmt.Sprintf("storage.%d", i)
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".usage",
				Type:  "gauge",
				Value: device.Usage,
			})
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".used",
				Type:  "gauge",
				Value: float64(device.Used),
			})
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".free",
				Type:  "gauge",
				Value: float64(device.Free),
			})
			report.Metrics = append(report.Metrics, &pb.Metric{
				Name:  prefix + ".capacity",
				Type:  "gauge",
				Value: float64(device.Capacity),
			})

			// 添加可选的路径标识
			if device.Path != "" {
				report.Metrics = append(report.Metrics, &pb.Metric{
					Name:   prefix + ".path",
					Type:   "label",
					Value:  0,
					Labels: map[string]string{"path": device.Path},
				})
			}

			// 添加可选的类型标识
			if device.Type != "" {
				report.Metrics = append(report.Metrics, &pb.Metric{
					Name:   prefix + ".type",
					Type:   "label",
					Value:  0,
					Labels: map[string]string{"type": device.Type},
				})
			}
		}
	}

	// 添加网络指标
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "network.inbound",
		Type:  "gauge",
		Value: metricsInfo.Network.InboundTraffic,
	})
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "network.outbound",
		Type:  "gauge",
		Value: metricsInfo.Network.OutboundTraffic,
	})
	report.Metrics = append(report.Metrics, &pb.Metric{
		Name:  "network.connections",
		Type:  "gauge",
		Value: float64(metricsInfo.Network.Connections),
	})

	// 发送指标报告
	_, err := a.client.ReportMetrics(ctx, report)
	if err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("发送指标报告失败: %v", err)))
	}
}

// NewDefaultAgentConfig 创建默认代理配置 - 兼容旧代码，保留以备将来使用
func NewDefaultAgentConfig() *Options {
	return &Options{
		HeartbeatPeriod: 30 * time.Second,
		MetricsInterval: 60 * time.Second,
		Labels:          make(map[string]string),
	}
}

// Stop 停止代理，关闭连接和释放资源
func (a *GameNodeAgent) Stop() {
	// 关闭gRPC连接
	if a.conn != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "关闭与服务器的连接"))
		a.conn.Close()
		a.conn = nil
		a.client = nil
	}

	// 更新状态为离线
	a.mu.Lock()
	a.status.Online = false
	a.status.State = models.GameNodeStateOffline
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "代理已停止"))
}

// SendHeartbeat 发送心跳信息
func (a *GameNodeAgent) SendHeartbeat(ctx context.Context) error {
	// 获取和更新状态
	nodeStatus := a.collectProtobufNodeStatus()

	// 发送心跳请求
	req := &pb.HeartbeatRequest{
		Id:        a.id,
		SessionId: a.id, // 使用节点ID作为会话ID
		Timestamp: time.Now().Unix(),
		Status:    nodeStatus,
	}

	_, err := a.client.Heartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("发送心跳失败: %v", err)
	}

	return nil
}

// collectMetricsFromHardware 从硬件信息中收集监控数据
// 只收集status.hardware中有的设备的指标
func (a *GameNodeAgent) collectMetricsFromHardware() error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 检查硬件信息是否已初始化
	if a.status == nil {
		return fmt.Errorf("硬件信息尚未初始化")
	}

	// 获取实时采集的指标
	metricsInfo, err := a.metricsCollector.GetMetricsInfo()
	if err != nil {
		return fmt.Errorf("获取实时指标失败: %v", err)
	}

	// 创建一个新的指标信息
	var metrics models.MetricsInfo

	// 1. 只收集hardware.cpus中的CPU设备指标
	if len(a.status.Hardware.CPUs) > 0 {
		// 如果有硬件CPU信息，就只收集这些设备的指标
		for _, device := range a.status.Hardware.CPUs {
			cpuMetric := models.CPUMetrics{
				Model:   device.Model,
				Cores:   device.Cores,
				Threads: device.Threads,
			}

			// 从采集到的指标中找到使用率
			if len(metricsInfo.CPUs) > 0 {
				cpuMetric.Usage = metricsInfo.CPUs[0].Usage
			}

			metrics.CPUs = append(metrics.CPUs, cpuMetric)
		}
	} else if len(metricsInfo.CPUs) > 0 {
		// 如果没有硬件CPU信息，但有采集到的CPU指标，就使用采集到的
		metrics.CPUs = metricsInfo.CPUs
	}

	// 2. 只收集hardware.gpus中的GPU设备指标
	if len(a.status.Hardware.GPUs) > 0 {
		// 如果有硬件GPU信息，就只收集这些设备的指标
		for _, device := range a.status.Hardware.GPUs {
			gpuMetric := models.GPUMetrics{
				Model:       device.Model,
				MemoryTotal: device.MemoryTotal,
			}

			// 从采集到的指标中找到匹配的GPU
			for _, gpu := range metricsInfo.GPUs {
				// 尝试匹配GPU型号（不区分大小写）
				if strings.Contains(strings.ToLower(gpu.Model), strings.ToLower(device.Model)) ||
					strings.Contains(strings.ToLower(device.Model), strings.ToLower(gpu.Model)) {
					gpuMetric.Usage = gpu.Usage
					gpuMetric.MemoryUsed = gpu.MemoryUsed
					gpuMetric.MemoryFree = gpu.MemoryFree
					gpuMetric.MemoryUsage = gpu.MemoryUsage
					break
				}
			}

			metrics.GPUs = append(metrics.GPUs, gpuMetric)
		}
	} else if len(metricsInfo.GPUs) > 0 {
		// 如果没有硬件GPU信息，但有采集到的GPU指标，就使用采集到的
		metrics.GPUs = metricsInfo.GPUs
	}

	// 3. 只收集hardware.storages中的存储设备指标
	if len(a.status.Hardware.Storages) > 0 {
		// 如果有硬件Storage信息，就只收集这些设备的指标
		for _, device := range a.status.Hardware.Storages {
			storageMetric := models.StorageMetrics{
				Path:     device.Path,
				Type:     device.Type,
				Model:    device.Model,
				Capacity: device.Capacity,
			}

			// 从采集到的指标中找到匹配的存储设备
			for _, storage := range metricsInfo.Storages {
				// 尝试匹配路径或设备名称
				if storage.Path == device.Path ||
					(storage.Path != "" && device.Path != "" &&
						(strings.HasSuffix(storage.Path, device.Path) ||
							strings.HasSuffix(device.Path, storage.Path))) {
					storageMetric.Used = storage.Used
					storageMetric.Free = storage.Free
					storageMetric.Usage = storage.Usage
					break
				}
			}

			metrics.Storages = append(metrics.Storages, storageMetric)
		}
	} else if len(metricsInfo.Storages) > 0 {
		// 如果没有硬件Storage信息，但有采集到的存储指标，就使用采集到的
		metrics.Storages = metricsInfo.Storages
	}

	// 4. 内存指标（假设一台机器只有一组内存）
	if len(a.status.Hardware.Memories) > 0 {
		// 计算总内存容量
		totalMemory := int64(0)
		for _, mem := range a.status.Hardware.Memories {
			totalMemory += mem.Size
		}

		// 使用采集到的内存使用情况
		if metricsInfo.Memory.Total > 0 {
			metrics.Memory = metricsInfo.Memory
			// 确保Total与硬件信息一致
			if totalMemory > 0 {
				metrics.Memory.Total = totalMemory
			}
		} else {
			// 如果没有采集到内存使用情况，至少设置总量
			metrics.Memory.Total = totalMemory
		}
	} else if metricsInfo.Memory.Total > 0 {
		// 如果没有硬件内存信息，但有采集到的内存指标，就使用采集到的
		metrics.Memory = metricsInfo.Memory
	}

	// 5. 网络指标直接使用采集到的
	metrics.Network = metricsInfo.Network

	// 更新状态
	a.mu.Lock()
	a.status.Metrics = metrics
	a.mu.Unlock()

	return nil
}
