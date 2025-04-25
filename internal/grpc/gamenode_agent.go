package grpc

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/sysinfo"
)

const (
	defaultRPCTimeout = 10 * time.Second
)

// GameNodeOptions 配置选项
type GameNodeOptions struct {
	HeartbeatPeriod time.Duration
	MetricsInterval time.Duration
}

// GameNodeAgent gRPC客户端实现
type GameNodeAgent struct {
	*Agent // 嵌入基础 Agent

	// 节点状态
	state models.GameNodeStaticState

	// 状态管理
	mu     sync.RWMutex
	status *models.GameNodeStatus

	pipelines map[string]*models.GamePipeline

	// 系统信息采集器
	hardwareCollector *sysinfo.HardwareCollector
	systemCollector   *sysinfo.SystemCollector
	metricsCollector  *sysinfo.MetricsCollector

	// 节点基本信息
	nodeInfo struct {
		name      string
		namespace string
		nodeType  models.GameNodeType
		ip        string
		labels    map[string]string
	}
}

// NewGameNodeAgent 创建新的 GameNode Agent
func NewGameNodeAgent(base *Agent) (*GameNodeAgent, error) {
	agent := &GameNodeAgent{
		Agent:     base,
		pipelines: make(map[string]*models.GamePipeline),
	}

	// 初始化系统信息采集器
	agent.hardwareCollector = sysinfo.NewHardwareCollector(nil)
	agent.systemCollector = sysinfo.NewSystemCollector(nil)
	agent.metricsCollector = sysinfo.NewMetricsCollector(nil)

	// 检测节点类型
	agent.nodeInfo.nodeType = agent.DetectNodeType()
	agent.GetLogger().Info("检测到节点类型: %s", agent.nodeInfo.nodeType)

	return agent, nil
}

// Register 注册节点
func (a *GameNodeAgent) Register(ctx context.Context) error {
	a.GetLogger().Debug("开始注册节点: %s", a.id)

	// 1. 参数验证
	if a.id == "" {
		a.GetLogger().Fatal("节点ID不能为空")
	}
	if a.nodeInfo.nodeType == "" {
		a.GetLogger().Fatal("节点类型不能为空")
	}

	// 2. 收集硬件和系统信息（带超时控制）
	collectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := a.collectHardwareAndSystemInfo(collectCtx); err != nil {
		a.GetLogger().Fatal("收集硬件和系统信息失败: %v", err)
	}

	// 3. 获取标准化的硬件信息
	hardware, err := a.hardwareCollector.GetSimplifiedHardwareInfo()
	if err != nil {
		a.GetLogger().Fatal("获取标准化硬件信息失败: %v", err)
	}

	// 4. 准备注册请求
	req := &pb.RegisterRequest{
		Id:       a.id,
		Alias:    a.id,
		Model:    "Beagle-Wind-2024",
		Type:     string(a.nodeInfo.nodeType),
		Location: "default",
		Hardware: hardware,
		System: map[string]string{
			"os_distribution":         a.status.System.OSDistribution,
			"os_version":              a.status.System.OSVersion,
			"os_architecture":         a.status.System.OSArchitecture,
			"kernel_version":          a.status.System.KernelVersion,
			"gpu_driver_version":      a.status.System.GPUDriverVersion,
			"gpu_compute_api_version": a.status.System.GPUComputeAPIVersion,
		},
		Labels: a.nodeInfo.labels,
	}

	// 5. 验证必要的系统信息
	if req.System["os_distribution"] == "" {
		a.GetLogger().Warn("操作系统发行版信息为空")
	}
	if req.System["os_version"] == "" {
		a.GetLogger().Warn("操作系统版本信息为空")
	}
	if req.System["kernel_version"] == "" {
		a.GetLogger().Warn("内核版本信息为空")
	}

	// 6. 发送注册请求（带超时控制）
	rpcCtx, cancel := context.WithTimeout(ctx, defaultRPCTimeout)
	defer cancel()

	client := a.Agent.GetGameNodeClient()
	resp, err := client.Register(rpcCtx, req)
	if err != nil {
		a.GetLogger().Fatal("注册请求失败: %v", err)
	}

	if !resp.Success {
		a.GetLogger().Fatal("注册失败: %s", resp.Message)
	}

	// 7. 存储节点维护状态
	a.state = convertFromProtoStaticState(resp.State)

	return nil
}

// GetState 获取节点维护状态
func (a *GameNodeAgent) GetState() models.GameNodeStaticState {
	return a.state
}

// SendHeartbeat 发送心跳
func (a *GameNodeAgent) SendHeartbeat(ctx context.Context) error {
	// 准备心跳请求
	req := &pb.HeartbeatRequest{
		Id:        a.id,
		SessionId: a.id, // 使用节点ID作为会话ID
		Timestamp: time.Now().Unix(),
	}

	// 发送心跳请求，最多重试3次
	var lastErr error
	for i := 0; i < 3; i++ {
		client := a.Agent.GetGameNodeClient()
		_, err := client.Heartbeat(ctx, req)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
	}

	a.GetLogger().Error("心跳发送失败: %v", lastErr)
	return fmt.Errorf("心跳发送失败: %v", lastErr)
}

// ReportMetrics 发送指标报告
func (a *GameNodeAgent) ReportMetrics(ctx context.Context) error {
	// 获取硬件信息
	hardwareInfo := a.status.Hardware

	// 使用 MetricsCollector 收集指标
	metricsInfo, err := a.metricsCollector.GetMetricsInfo(&hardwareInfo)
	if err != nil {
		return fmt.Errorf("收集指标信息失败: %v", err)
	}

	// 更新本地状态
	a.mu.Lock()
	a.status.Metrics = metricsInfo
	a.mu.Unlock()

	// 准备请求
	req := &pb.MetricsRequest{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Metrics:   convertToProtoMetricsInfo(metricsInfo),
	}

	// 发送请求
	client := a.Agent.GetGameNodeClient()
	_, err = client.ReportMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("上报指标失败: %v", err)
	}

	return nil
}

// Start 启动代理
func (a *GameNodeAgent) Start(ctx context.Context) error {
	// 使用基础 Agent 的 Start 方法
	if err := a.Agent.Start(ctx); err != nil {
		return err
	}

	// 注册节点
	if err := a.Register(ctx); err != nil {
		return fmt.Errorf("register failed: %v", err)
	}

	a.GetLogger().Info("GameNodeAgent 已启动并完成注册")
	return nil
}

// Stop 停止代理
func (a *GameNodeAgent) Stop(ctx context.Context) {
	a.Agent.Stop(ctx)
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

// collectHardwareAndSystemInfo 收集硬件和系统信息
func (a *GameNodeAgent) collectHardwareAndSystemInfo(ctx context.Context) error {
	// 检查采集器是否初始化
	if a.hardwareCollector == nil || a.systemCollector == nil {
		return fmt.Errorf("系统信息采集器未初始化")
	}

	// 使用硬件采集器收集信息
	hardwareInfo, err := a.hardwareCollector.GetHardwareInfo()
	if err != nil {
		return fmt.Errorf("收集硬件信息失败: %v", err)
	}

	// 使用系统采集器收集信息
	systemInfo, err := a.systemCollector.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("收集系统信息失败: %v", err)
	}

	// 确保状态对象已初始化
	if a.status == nil {
		a.status = &models.GameNodeStatus{
			Hardware: hardwareInfo,
			System:   systemInfo,
		}
	}

	// 更新状态
	a.mu.Lock()
	a.status.Hardware = hardwareInfo
	a.status.System = systemInfo
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	return nil
}

// convertFromProtoStaticState 将 proto.GameNodeStaticState 转换为 models.GameNodeStaticState
func convertFromProtoStaticState(state pb.GameNodeStaticState) models.GameNodeStaticState {
	switch state {
	case pb.GameNodeStaticState_NODE_STATE_NORMAL:
		return models.GameNodeStaticStateNormal
	case pb.GameNodeStaticState_NODE_STATE_MAINTENANCE:
		return models.GameNodeStaticStateMaintenance
	case pb.GameNodeStaticState_NODE_STATE_DISABLED:
		return models.GameNodeStaticStateDisabled
	default:
		return models.GameNodeStaticStateNormal
	}
}

// convertToProtoMetricsInfo 将 models.MetricsInfo 转换为 proto.MetricsInfo
func convertToProtoMetricsInfo(info models.MetricsInfo) *pb.MetricsInfo {
	protoInfo := &pb.MetricsInfo{
		Memory: &pb.MemoryMetrics{
			Total:     info.Memory.Total,
			Available: info.Memory.Available,
			Used:      info.Memory.Used,
			Usage:     info.Memory.Usage,
		},
		Network: &pb.NetworkMetrics{
			InboundTraffic:  info.Network.InboundTraffic,
			OutboundTraffic: info.Network.OutboundTraffic,
			Connections:     info.Network.Connections,
		},
	}

	// 转换 CPU 指标
	for _, cpu := range info.CPUs {
		protoInfo.Cpus = append(protoInfo.Cpus, &pb.CPUMetrics{
			Model:   cpu.Model,
			Cores:   cpu.Cores,
			Threads: cpu.Threads,
			Usage:   cpu.Usage,
		})
	}

	// 转换 GPU 指标
	for _, gpu := range info.GPUs {
		protoInfo.Gpus = append(protoInfo.Gpus, &pb.GPUMetrics{
			Model:       gpu.Model,
			MemoryTotal: gpu.MemoryTotal,
			GpuUsage:    gpu.GPUUsage,
			MemoryUsed:  gpu.MemoryUsed,
			MemoryFree:  gpu.MemoryFree,
			MemoryUsage: gpu.MemoryUsage,
		})
	}

	// 转换存储指标
	for _, storage := range info.Storages {
		protoInfo.Storages = append(protoInfo.Storages, &pb.StorageMetrics{
			Path:     storage.Path,
			Type:     storage.Type,
			Model:    storage.Model,
			Capacity: storage.Capacity,
			Used:     storage.Used,
			Free:     storage.Free,
			Usage:    storage.Usage,
		})
	}

	return protoInfo
}

// ReportResource 上报节点资源信息
func (a *GameNodeAgent) ReportResource(ctx context.Context) error {
	// 收集硬件和系统信息
	hardwareInfo, err := a.hardwareCollector.GetHardwareInfo()
	if err != nil {
		return fmt.Errorf("收集硬件信息失败: %v", err)
	}

	systemInfo, err := a.systemCollector.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("收集系统信息失败: %v", err)
	}

	// 准备请求
	req := &pb.ResourceRequest{
		NodeId: a.id,
		Hardware: &pb.HardwareInfo{
			Cpus:     convertToProtoCPUDevices(hardwareInfo.CPUs),
			Memories: convertToProtoMemoryDevices(hardwareInfo.Memories),
			Gpus:     convertToProtoGPUDevices(hardwareInfo.GPUs),
			Storages: convertToProtoStorageDevices(hardwareInfo.Storages),
			Networks: convertToProtoNetworkDevices(hardwareInfo.Networks),
		},
		System: &pb.SystemInfo{
			OsDistribution:       systemInfo.OSDistribution,
			OsVersion:            systemInfo.OSVersion,
			OsArchitecture:       systemInfo.OSArchitecture,
			KernelVersion:        systemInfo.KernelVersion,
			GpuDriverVersion:     systemInfo.GPUDriverVersion,
			GpuComputeApiVersion: systemInfo.GPUComputeAPIVersion,
			DockerVersion:        systemInfo.DockerVersion,
			ContainerdVersion:    systemInfo.ContainerdVersion,
			RuncVersion:          systemInfo.RuncVersion,
		},
	}

	// 发送请求
	client := a.Agent.GetGameNodeClient()
	_, err = client.ReportResource(ctx, req)
	if err != nil {
		return fmt.Errorf("上报资源信息失败: %v", err)
	}

	return nil
}

// convertToProtoCPUDevices 将 models.CPUDevice 转换为 proto.CPUHardware
func convertToProtoCPUDevices(devices []models.CPUDevice) []*pb.CPUHardware {
	result := make([]*pb.CPUHardware, len(devices))
	for i, device := range devices {
		result[i] = &pb.CPUHardware{
			Model:        device.Model,
			Cores:        device.Cores,
			Threads:      device.Threads,
			Frequency:    device.Frequency,
			Cache:        device.Cache,
			Architecture: device.Architecture,
		}
	}
	return result
}

// convertToProtoMemoryDevices 将 models.MemoryDevice 转换为 proto.MemoryHardware
func convertToProtoMemoryDevices(devices []models.MemoryDevice) []*pb.MemoryHardware {
	result := make([]*pb.MemoryHardware, len(devices))
	for i, device := range devices {
		result[i] = &pb.MemoryHardware{
			Size:      device.Size,
			Type:      device.Type,
			Frequency: device.Frequency,
		}
	}
	return result
}

// convertToProtoGPUDevices 将 models.GPUDevice 转换为 proto.GPUHardware
func convertToProtoGPUDevices(devices []models.GPUDevice) []*pb.GPUHardware {
	result := make([]*pb.GPUHardware, len(devices))
	for i, device := range devices {
		result[i] = &pb.GPUHardware{
			Model:             device.Model,
			MemoryTotal:       device.MemoryTotal,
			Architecture:      device.Architecture,
			DriverVersion:     device.DriverVersion,
			ComputeCapability: device.ComputeCapability,
			Tdp:               device.TDP,
		}
	}
	return result
}

// convertToProtoStorageDevices 将 models.StorageDevice 转换为 proto.StorageDevice
func convertToProtoStorageDevices(devices []models.StorageDevice) []*pb.StorageDevice {
	result := make([]*pb.StorageDevice, len(devices))
	for i, device := range devices {
		result[i] = &pb.StorageDevice{
			Type:     device.Type,
			Model:    device.Model,
			Capacity: device.Capacity,
			Path:     device.Path,
		}
	}
	return result
}

// convertToProtoNetworkDevices 将 models.NetworkDevice 转换为 proto.NetworkDevice
func convertToProtoNetworkDevices(devices []models.NetworkDevice) []*pb.NetworkDevice {
	result := make([]*pb.NetworkDevice, len(devices))
	for i, device := range devices {
		result[i] = &pb.NetworkDevice{
			Name:       device.Name,
			MacAddress: device.MacAddress,
			IpAddress:  device.IpAddress,
			Speed:      device.Speed,
		}
	}
	return result
}
