package gamenode

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-beagle/beagle-wind-game/internal/event"
	"github.com/open-beagle/beagle-wind-game/internal/log"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
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
	config *AgentConfig

	// 资源采集
	resourceCollector *ResourceCollector

	// Docker 客户端
	dockerClient *dockerclient.Client

	// 硬件和系统信息
	hardwareConfig map[string]string   // 兼容现有简化版硬件配置
	hardware       models.HardwareInfo // 详细硬件信息
	system         models.SystemInfo   // 详细系统信息
	cpuCores       int
	cpuThreads     int
}

// AgentConfig 代理配置
type AgentConfig struct {
	// 基本配置
	Alias    string            // 节点别名
	Model    string            // 节点型号
	Type     string            // 节点类型
	Location string            // 节点位置
	Labels   map[string]string // 节点标签

	// 运行配置
	HeartbeatPeriod time.Duration // 心跳周期
	RetryCount      int           // 重试次数
	RetryDelay      time.Duration // 重试延迟
	MetricsInterval time.Duration // 指标采集间隔
}

// ResourceCollector 资源采集器
type ResourceCollector struct {
	mu sync.RWMutex
	// 系统指标
	cpuUsage    float64
	memoryUsage float64
	diskUsage   float64
	networkIO   struct {
		read  int64
		write int64
	}
	// GPU指标
	gpuUsage  float64
	gpuMemory float64
	gpuTemp   float64
	gpuPower  float64
}

// NewGameNodeAgent 创建新的游戏节点代理
func NewGameNodeAgent(
	id string,
	serverAddr string,
	eventManager event.EventManager,
	logManager log.LogManager,
	dockerClient *dockerclient.Client,
	config *AgentConfig,
) *GameNodeAgent {
	if config == nil {
		config = NewDefaultAgentConfig()
	}

	agent := &GameNodeAgent{
		id:                id,
		serverAddr:        serverAddr,
		eventManager:      eventManager,
		logManager:        logManager,
		config:            config,
		resourceCollector: &ResourceCollector{},
		dockerClient:      dockerClient,
		hardwareConfig:    make(map[string]string),
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
	}

	// 初始化 gRPC 连接
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(serverAddr, opts...)
	if err == nil {
		agent.conn = conn
		agent.client = pb.NewGameNodeGRPCServiceClient(conn)
	}

	return agent
}

// Start 启动代理
func (a *GameNodeAgent) Start(ctx context.Context) error {
	// 检测节点类型
	nodeType := a.DetectNodeType()
	a.config.Type = string(nodeType)
	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", fmt.Sprintf("检测到节点类型: %s", nodeType)))

	// 建立连接
	if err := a.connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	// 注册节点
	if err := a.Register(ctx); err != nil {
		return fmt.Errorf("failed to register: %v", err)
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
	if content, err := ioutil.ReadFile(cgroupPath); err == nil {
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

	if content, err := ioutil.ReadFile("/proc/cpuinfo"); err == nil {
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

// connect 建立连接
func (a *GameNodeAgent) connect(_ context.Context) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(a.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	a.conn = conn
	a.client = pb.NewGameNodeGRPCServiceClient(conn)
	return nil
}

// startHeartbeat 启动心跳
func (a *GameNodeAgent) startHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(a.config.HeartbeatPeriod)
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
	ticker := time.NewTicker(a.config.MetricsInterval)
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
	// 获取节点信息
	nodeInfo, err := a.collectNodeInfo()
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to collect node info: %v", err))
	}

	// 获取硬件配置和系统配置
	hardwareConfig, hardwareInfo, err := a.GetHardwareConfig()
	if err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("获取硬件配置失败: %v", err)))
	}
	a.hardwareConfig = hardwareConfig
	a.hardware = hardwareInfo

	systemConfig, systemInfo, err := a.GetSystemConfig()
	if err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "warn", fmt.Sprintf("获取系统配置失败: %v", err)))
	}
	a.system = systemInfo

	// 使用获取到的配置完善注册请求
	nodeInfo.Hardware = hardwareConfig
	nodeInfo.System = systemConfig

	// 发送注册请求
	resp, err := a.client.Register(ctx, nodeInfo)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to register: %v", err))
	}

	// 如果注册失败，返回错误
	if !resp.Success {
		return status.Error(codes.Internal, fmt.Sprintf("registration failed: %s", resp.Message))
	}

	// 初始化指标信息
	metricsInfo := models.MetricsInfo{
		CPU: struct {
			Devices []struct {
				Model       string  `json:"model" yaml:"model"`
				Cores       int32   `json:"cores" yaml:"cores"`
				Threads     int32   `json:"threads" yaml:"threads"`
				Usage       float64 `json:"usage" yaml:"usage"`
				Temperature float64 `json:"temperature" yaml:"temperature"`
			} `json:"devices" yaml:"devices"`
		}{
			Devices: []struct {
				Model       string  `json:"model" yaml:"model"`
				Cores       int32   `json:"cores" yaml:"cores"`
				Threads     int32   `json:"threads" yaml:"threads"`
				Usage       float64 `json:"usage" yaml:"usage"`
				Temperature float64 `json:"temperature" yaml:"temperature"`
			}{},
		},
		Memory: struct {
			Total     int64   `json:"total" yaml:"total"`
			Available int64   `json:"available" yaml:"available"`
			Used      int64   `json:"used" yaml:"used"`
			Usage     float64 `json:"usage" yaml:"usage"`
		}{},
		GPU: struct {
			Devices []struct {
				Model       string  `json:"model" yaml:"model"`
				MemoryTotal int64   `json:"memory_total" yaml:"memory_total"`
				Usage       float64 `json:"usage" yaml:"usage"`
				MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`
				MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`
				MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"`
				Temperature float64 `json:"temperature" yaml:"temperature"`
				Power       float64 `json:"power" yaml:"power"`
			} `json:"devices" yaml:"devices"`
		}{
			Devices: []struct {
				Model       string  `json:"model" yaml:"model"`
				MemoryTotal int64   `json:"memory_total" yaml:"memory_total"`
				Usage       float64 `json:"usage" yaml:"usage"`
				MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`
				MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`
				MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"`
				Temperature float64 `json:"temperature" yaml:"temperature"`
				Power       float64 `json:"power" yaml:"power"`
			}{},
		},
		Storage: struct {
			Devices []struct {
				Path       string  `json:"path" yaml:"path"`
				Type       string  `json:"type" yaml:"type"`
				Model      string  `json:"model" yaml:"model"`
				Capacity   int64   `json:"capacity" yaml:"capacity"`
				Used       int64   `json:"used" yaml:"used"`
				Free       int64   `json:"free" yaml:"free"`
				Usage      float64 `json:"usage" yaml:"usage"`
				ReadSpeed  float64 `json:"read_speed" yaml:"read_speed"`
				WriteSpeed float64 `json:"write_speed" yaml:"write_speed"`
			} `json:"devices" yaml:"devices"`
		}{
			Devices: []struct {
				Path       string  `json:"path" yaml:"path"`
				Type       string  `json:"type" yaml:"type"`
				Model      string  `json:"model" yaml:"model"`
				Capacity   int64   `json:"capacity" yaml:"capacity"`
				Used       int64   `json:"used" yaml:"used"`
				Free       int64   `json:"free" yaml:"free"`
				Usage      float64 `json:"usage" yaml:"usage"`
				ReadSpeed  float64 `json:"read_speed" yaml:"read_speed"`
				WriteSpeed float64 `json:"write_speed" yaml:"write_speed"`
			}{},
		},
		Network: struct {
			InboundTraffic  float64 `json:"inbound_traffic" yaml:"inbound_traffic"`
			OutboundTraffic float64 `json:"outbound_traffic" yaml:"outbound_traffic"`
			Connections     int32   `json:"connections" yaml:"connections"`
		}{},
	}

	// 如果存在CPU信息，添加到指标中
	if len(hardwareInfo.CPU.Devices) > 0 {
		for _, cpu := range hardwareInfo.CPU.Devices {
			metricsInfo.CPU.Devices = append(metricsInfo.CPU.Devices, struct {
				Model       string  `json:"model" yaml:"model"`
				Cores       int32   `json:"cores" yaml:"cores"`
				Threads     int32   `json:"threads" yaml:"threads"`
				Usage       float64 `json:"usage" yaml:"usage"`
				Temperature float64 `json:"temperature" yaml:"temperature"`
			}{
				Model:       cpu.Model,
				Cores:       cpu.Cores,
				Threads:     cpu.Threads,
				Usage:       0,
				Temperature: 45, // 默认温度
			})
		}
	}

	// 如果存在GPU信息，添加到指标中
	if len(hardwareInfo.GPU.Devices) > 0 {
		for _, gpu := range hardwareInfo.GPU.Devices {
			metricsInfo.GPU.Devices = append(metricsInfo.GPU.Devices, struct {
				Model       string  `json:"model" yaml:"model"`
				MemoryTotal int64   `json:"memory_total" yaml:"memory_total"`
				Usage       float64 `json:"usage" yaml:"usage"`
				MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`
				MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`
				MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"`
				Temperature float64 `json:"temperature" yaml:"temperature"`
				Power       float64 `json:"power" yaml:"power"`
			}{
				Model:       gpu.Model,
				MemoryTotal: gpu.MemoryTotal,
				MemoryFree:  gpu.MemoryTotal, // 初始时全部可用
				Temperature: 38,              // 默认温度
				Power:       30,              // 默认功耗
			})
		}
	}

	// 如果存在存储设备信息，添加到指标中
	if len(hardwareInfo.Storage.Devices) > 0 {
		for _, storage := range hardwareInfo.Storage.Devices {
			used := storage.Capacity / 4 // 假设使用了25%
			free := storage.Capacity - used
			metricsInfo.Storage.Devices = append(metricsInfo.Storage.Devices, struct {
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
				Path:     "/",
				Type:     storage.Type,
				Model:    storage.Model,
				Capacity: storage.Capacity,
				Used:     used,
				Free:     free,
				Usage:    25, // 25%
			})
		}
	}

	// 如果存在内存设备信息，添加到指标中
	if len(hardwareInfo.Memory.Devices) > 0 {
		totalMem := hardwareInfo.Memory.Devices[0].Size
		usedMem := totalMem / 8 // 假设使用了12.5%
		availMem := totalMem - usedMem
		metricsInfo.Memory.Total = totalMem
		metricsInfo.Memory.Used = usedMem
		metricsInfo.Memory.Available = availMem
		metricsInfo.Memory.Usage = 12.5
	}

	// 设置默认网络信息
	metricsInfo.Network.InboundTraffic = 10.5
	metricsInfo.Network.OutboundTraffic = 2.3
	metricsInfo.Network.Connections = 120

	// 更新状态
	a.mu.Lock()
	a.status.State = models.GameNodeStateOnline
	a.status.Online = true
	a.status.LastOnline = time.Now()
	a.status.UpdatedAt = time.Now()
	a.status.Hardware = hardwareInfo
	a.status.System = systemInfo
	a.status.Metrics = metricsInfo
	a.mu.Unlock()

	// 发布注册成功事件
	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "节点注册成功"))

	return nil
}

// collectResourceInfo 采集资源信息
func (a *GameNodeAgent) collectResourceInfo() (*pb.ResourceInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 获取硬件信息
	hardwareInfo := &pb.HardwareInfo{
		Cpu:     &pb.CPUHardware{},
		Memory:  &pb.MemoryHardware{},
		Gpu:     &pb.GPUHardware{},
		Storage: &pb.StorageHardware{},
	}

	// 添加CPU信息
	if len(a.hardware.CPU.Devices) > 0 {
		cpu := a.hardware.CPU.Devices[0]
		hardwareInfo.Cpu = &pb.CPUHardware{
			Model:     cpu.Model,
			Cores:     cpu.Cores,
			Threads:   cpu.Threads,
			Frequency: cpu.Frequency,
			Cache:     cpu.Cache,
		}
	}

	// 添加内存信息
	if len(a.hardware.Memory.Devices) > 0 {
		mem := a.hardware.Memory.Devices[0]
		hardwareInfo.Memory = &pb.MemoryHardware{
			Total:     mem.Size,
			Type:      mem.Type,
			Frequency: mem.Frequency,
		}
	}

	// 添加GPU信息
	if len(a.hardware.GPU.Devices) > 0 {
		gpu := a.hardware.GPU.Devices[0]
		hardwareInfo.Gpu = &pb.GPUHardware{
			Model:       gpu.Model,
			MemoryTotal: gpu.MemoryTotal,
			CudaCores:   gpu.CudaCores,
		}
	}

	// 添加存储设备信息
	hardwareInfo.Storage.Devices = make([]*pb.StorageDevice, len(a.hardware.Storage.Devices))
	for i, device := range a.hardware.Storage.Devices {
		hardwareInfo.Storage.Devices[i] = &pb.StorageDevice{
			Type:     device.Type,
			Capacity: device.Capacity,
		}
	}

	return &pb.ResourceInfo{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Hardware:  hardwareInfo,
	}, nil
}

// parseMemorySize 从内存大小字符串解析为数值（GB）
func parseMemorySize(memStr string) int64 {
	// 示例解析，实际应根据输入格式调整
	// 预期格式："{size} GB"，如 "16 GB"
	parts := strings.Split(memStr, " ")
	if len(parts) < 2 {
		return 0
	}

	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}

	return size
}

// Heartbeat 发送心跳
func (a *GameNodeAgent) Heartbeat(ctx context.Context) error {
	// 获取资源信息
	resourceInfo, err := a.collectResourceInfo()
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to collect resource info: %v", err))
	}

	// 发送心跳请求
	req := &pb.HeartbeatRequest{
		Id:           a.id,
		Timestamp:    time.Now().Unix(),
		ResourceInfo: resourceInfo,
	}

	_, err = a.client.Heartbeat(ctx, req)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to send heartbeat: %v", err))
	}

	// 更新状态
	a.mu.Lock()
	a.status.LastOnline = time.Now()
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	return nil
}

// ReportMetrics 上报指标
func (a *GameNodeAgent) ReportMetrics(ctx context.Context) error {
	// 获取指标数据
	metrics := a.collectMetrics()

	// 发送指标报告
	req := &pb.MetricsReport{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
	}

	_, err := a.client.ReportMetrics(ctx, req)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to report metrics: %v", err))
	}

	// 更新状态中的指标信息
	a.mu.Lock()
	// 初始化指标信息
	var metricsInfo models.MetricsInfo

	// 初始化CPU指标
	cpuDevice := struct {
		Model       string  `json:"model" yaml:"model"`
		Cores       int32   `json:"cores" yaml:"cores"`
		Threads     int32   `json:"threads" yaml:"threads"`
		Usage       float64 `json:"usage" yaml:"usage"`
		Temperature float64 `json:"temperature" yaml:"temperature"`
	}{
		Cores:   int32(a.cpuCores),
		Threads: int32(a.cpuThreads),
	}

	// 初始化GPU指标
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

	// 初始化存储指标
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

	// 转换指标值
	for _, m := range metrics {
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

	a.status.Metrics = metricsInfo
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	return nil
}

// collectMetrics 采集指标数据
func (a *GameNodeAgent) collectMetrics() []*pb.Metric {
	a.resourceCollector.mu.RLock()
	defer a.resourceCollector.mu.RUnlock()

	return []*pb.Metric{
		{
			Name:  "cpu_usage",
			Type:  "gauge",
			Value: a.resourceCollector.cpuUsage,
		},
		{
			Name:  "cpu_temperature",
			Type:  "gauge",
			Value: 45.0, // 示例值，实际应从硬件获取
		},
		{
			Name:  "memory_usage",
			Type:  "gauge",
			Value: a.resourceCollector.memoryUsage,
		},
		{
			Name:  "memory_used",
			Type:  "gauge",
			Value: float64(8 * 1024 * 1024 * 1024), // 示例值，实际应计算
		},
		{
			Name:  "memory_available",
			Type:  "gauge",
			Value: float64(8 * 1024 * 1024 * 1024), // 示例值，实际应计算
		},
		{
			Name:  "disk_usage",
			Type:  "gauge",
			Value: a.resourceCollector.diskUsage,
		},
		{
			Name:  "disk_used",
			Type:  "gauge",
			Value: float64(100 * 1024 * 1024 * 1024), // 示例值，实际应计算
		},
		{
			Name:  "disk_free",
			Type:  "gauge",
			Value: float64(900 * 1024 * 1024 * 1024), // 示例值，实际应计算
		},
		{
			Name:  "gpu_usage",
			Type:  "gauge",
			Value: a.resourceCollector.gpuUsage,
		},
		{
			Name:  "gpu_memory_used",
			Type:  "gauge",
			Value: float64(2 * 1024), // 示例值：2GB，实际应从硬件获取
		},
		{
			Name:  "gpu_memory_free",
			Type:  "gauge",
			Value: float64(6 * 1024), // 示例值：6GB，实际应从硬件获取
		},
		{
			Name:  "gpu_memory_usage",
			Type:  "gauge",
			Value: a.resourceCollector.gpuMemory,
		},
		{
			Name:  "gpu_temperature",
			Type:  "gauge",
			Value: a.resourceCollector.gpuTemp,
		},
		{
			Name:  "gpu_power",
			Type:  "gauge",
			Value: a.resourceCollector.gpuPower,
		},
		{
			Name:  "network_inbound",
			Type:  "counter",
			Value: float64(a.resourceCollector.networkIO.read),
		},
		{
			Name:  "network_outbound",
			Type:  "counter",
			Value: float64(a.resourceCollector.networkIO.write),
		},
		{
			Name:  "network_connections",
			Type:  "gauge",
			Value: 100, // 示例值，实际应计算网络连接数
		},
	}
}

// GetHardwareConfig 获取硬件配置信息
func (a *GameNodeAgent) GetHardwareConfig() (map[string]string, models.HardwareInfo, error) {
	// 简化版硬件配置（用于兼容现有接口）
	config := make(map[string]string)

	// 完整硬件信息结构
	var hardwareInfo models.HardwareInfo

	// 1. 获取CPU信息
	cpuModel := ""
	cpuCores := 0
	cpuThreads := 0
	var cpuFrequency float64 = 0

	if content, err := ioutil.ReadFile("/proc/cpuinfo"); err == nil {
		lines := strings.Split(string(content), "\n")
		var cores []string
		var processors []string

		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					cpuModel = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "core id") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					coreID := strings.TrimSpace(parts[1])
					if !contains(cores, coreID) {
						cores = append(cores, coreID)
					}
				}
			} else if strings.HasPrefix(line, "processor") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					processorID := strings.TrimSpace(parts[1])
					if !contains(processors, processorID) {
						processors = append(processors, processorID)
					}
				}
			} else if strings.HasPrefix(line, "cpu MHz") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					if freq, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						// 记录每个核心的频率，这里简化处理取第一个有效值
						if cpuFrequency == 0 {
							cpuFrequency = freq
						}
					}
				}
			}
		}

		cpuCores = len(cores)
		if cpuCores == 0 {
			// 备选方法
			if out, err := exec.Command("nproc").Output(); err == nil {
				if val, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil {
					cpuCores = val
				}
			}
		}

		cpuThreads = len(processors)
		if cpuThreads == 0 {
			cpuThreads = cpuCores * 2 // 默认假设超线程
		}
	}

	// 设置简化版配置
	config["CPU"] = cpuModel

	// 存储CPU核心数和线程数
	a.cpuCores = cpuCores
	a.cpuThreads = cpuThreads

	// 提取CPU型号中的主频信息
	freqInModelName := 0.0
	cpuModelLower := strings.ToLower(cpuModel)
	// 尝试从型号中提取频率信息，例如: "Intel Core i7-9700K @ 3.60GHz"
	if strings.Contains(cpuModelLower, "ghz") {
		// 查找数字 + GHz 模式
		freqRegex := regexp.MustCompile(`(\d+\.\d+)ghz`)
		matches := freqRegex.FindStringSubmatch(cpuModelLower)
		if len(matches) > 1 {
			if freq, err := strconv.ParseFloat(matches[1], 64); err == nil {
				freqInModelName = freq
			}
		}
	}

	// 如果从型号中提取到了频率，优先使用，否则使用从cpuinfo中读取的
	if freqInModelName > 0 {
		cpuFrequency = freqInModelName * 1000 // 转换为MHz
	} else if cpuFrequency == 0 {
		// 如果还是没有频率信息，尝试从lscpu获取
		if out, err := exec.Command("lscpu").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "CPU MHz") || strings.Contains(line, "CPU max MHz") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						if freq, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
							cpuFrequency = freq
							break
						}
					}
				}
			}
		}
	}

	// 获取CPU插槽信息
	var cpuSockets []string
	if out, err := exec.Command("lscpu").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Socket(s)") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					socketCount, err := strconv.Atoi(strings.TrimSpace(parts[1]))
					if err == nil {
						// 根据socket数量创建socket标识
						for i := 0; i < socketCount; i++ {
							cpuSockets = append(cpuSockets, fmt.Sprintf("%d", i))
						}
					}
				}
				break
			}
		}
	}

	// 如果没有找到socket信息，假设只有一个socket
	if len(cpuSockets) == 0 {
		cpuSockets = append(cpuSockets, "0")
	}

	// 更新CPU静态信息，添加核心数、主频和TDP
	var cpuInfos []string
	if cpuFrequency > 0 {
		// 将MHz转换为GHz
		cpuFreqGHz := cpuFrequency / 1000.0

		// 基于CPU型号和核心数估算TDP
		var cpuTDP int = 0
		if strings.Contains(cpuModelLower, "i9") {
			if cpuCores >= 12 {
				cpuTDP = 125
			} else {
				cpuTDP = 95
			}
		} else if strings.Contains(cpuModelLower, "i7") {
			if cpuCores >= 8 {
				cpuTDP = 95
			} else {
				cpuTDP = 65
			}
		} else if strings.Contains(cpuModelLower, "i5") {
			cpuTDP = 65
		} else if strings.Contains(cpuModelLower, "i3") {
			cpuTDP = 58
		} else if strings.Contains(cpuModelLower, "xeon") {
			if cpuCores >= 24 {
				cpuTDP = 225
			} else if cpuCores >= 16 {
				cpuTDP = 165
			} else if cpuCores >= 8 {
				cpuTDP = 125
			} else {
				cpuTDP = 95
			}
		} else if strings.Contains(cpuModelLower, "ryzen") {
			if strings.Contains(cpuModelLower, "threadripper") {
				cpuTDP = 280
			} else if strings.Contains(cpuModelLower, "9") {
				cpuTDP = 105
			} else if strings.Contains(cpuModelLower, "7") {
				cpuTDP = 95
			} else if strings.Contains(cpuModelLower, "5") {
				cpuTDP = 65
			} else {
				cpuTDP = 65
			}
		} else {
			// 基于核心数的通用估算
			if cpuCores <= 4 {
				cpuTDP = 65
			} else if cpuCores <= 8 {
				cpuTDP = 95
			} else if cpuCores <= 16 {
				cpuTDP = 125
			} else {
				cpuTDP = 165
			}
		}

		// 为每个CPU插槽创建信息字符串
		for _, socketID := range cpuSockets {
			// 简化CPU型号
			simplifiedModel := simpleCPUModel(cpuModel)
			cpuInfos = append(cpuInfos, fmt.Sprintf("%s,%s %dCore %.1fGHz %dW", socketID, simplifiedModel, cpuCores/len(cpuSockets), cpuFreqGHz, cpuTDP))
		}
	} else {
		// 没有频率信息时
		for _, socketID := range cpuSockets {
			// 简化CPU型号
			simplifiedModel := simpleCPUModel(cpuModel)
			cpuInfos = append(cpuInfos, fmt.Sprintf("%s,%s %dCore", socketID, simplifiedModel, cpuCores/len(cpuSockets)))
		}
	}

	// 设置格式化后的CPU信息
	config["CPU"] = strings.Join(cpuInfos, ";")

	// 设置详细硬件信息 - CPU
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
		Model:     cpuModel,
		Cores:     int32(cpuCores),
		Threads:   int32(cpuThreads),
		Frequency: cpuFrequency,
		Socket:    "0", // 默认插槽
	}
	hardwareInfo.CPU.Devices = append(hardwareInfo.CPU.Devices, cpuDevice)

	// 2. 获取内存信息
	var memorySize int64 = 0
	if content, err := ioutil.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					memoryInfo := strings.TrimSpace(parts[1])
					// 将 KB 转换为 GB 表示
					if strings.Contains(memoryInfo, "kB") {
						memoryValue := strings.TrimSpace(strings.Replace(memoryInfo, "kB", "", -1))
						if memVal, err := strconv.ParseInt(memoryValue, 10, 64); err == nil {
							// 获取 KB 转为 GB，并向上取整
							actualGB := float64(memVal) / 1024.0 / 1024.0
							roundedGB := int(math.Ceil(actualGB))
							config["RAM"] = fmt.Sprintf("%d GB", roundedGB)

							// 记录内存大小（字节）
							memorySize = int64(roundedGB) * 1024 * 1024 * 1024
						} else {
							config["RAM"] = memoryInfo // 如果转换失败，保留原始信息
						}
					} else {
						config["RAM"] = memoryInfo
					}
					break
				}
			}
		}
	}

	// 设置详细硬件信息 - 内存
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
		Size: memorySize,
		Type: "Unknown",
	}
	hardwareInfo.Memory.Devices = append(hardwareInfo.Memory.Devices, memoryDevice)

	// 3. 获取GPU信息
	gpuModels := []string{}
	gpuMemories := []int64{}
	gpuBuses := []string{}

	// 使用nvidia-smi列出所有GPU
	if out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,pci.bus_id", "--format=csv,noheader").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if len(line) > 0 {
				fields := strings.Split(line, ", ")
				if len(fields) >= 3 {
					gpuModels = append(gpuModels, strings.TrimSpace(fields[0]))

					// 解析显存
					memStr := strings.TrimSpace(fields[1])
					var memory int64 = 0
					if strings.Contains(memStr, "MiB") {
						memVal := strings.TrimSuffix(memStr, " MiB")
						if val, err := strconv.ParseInt(memVal, 10, 64); err == nil {
							memory = val
						}
					}
					gpuMemories = append(gpuMemories, memory)

					// 解析PCI总线ID
					busID := strings.TrimSpace(fields[2])
					gpuBuses = append(gpuBuses, busID)
				}
			}
		}
	}

	// 如果没有找到GPU，尝试使用简单命令获取
	if len(gpuModels) == 0 {
		if out, err := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader").Output(); err == nil {
			gpuModel := strings.TrimSpace(string(out))
			if len(gpuModel) > 0 {
				gpuModels = append(gpuModels, gpuModel)
				gpuBuses = append(gpuBuses, "0") // 默认总线ID

				// 尝试获取显存信息
				if out, err := exec.Command("nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader").Output(); err == nil {
					memStr := strings.TrimSpace(string(out))
					var memory int64 = 0
					if strings.Contains(memStr, "MiB") {
						memVal := strings.TrimSuffix(memStr, " MiB")
						if val, err := strconv.ParseInt(memVal, 10, 64); err == nil {
							memory = val
						}
					}
					gpuMemories = append(gpuMemories, memory)
				} else {
					gpuMemories = append(gpuMemories, 0)
				}
			}
		}
	}

	// 格式化GPU信息
	var gpuInfos []string
	for i, model := range gpuModels {
		// 获取GPU序号
		gpuIndex := i

		// 从PCI总线ID提取序号
		if i < len(gpuBuses) && len(gpuBuses[i]) > 0 {
			parts := strings.Split(gpuBuses[i], ":")
			if len(parts) > 1 {
				// 使用最后一部分作为简化的序号
				lastPart := parts[len(parts)-1]
				if val, err := strconv.Atoi(lastPart); err == nil {
					gpuIndex = val
				}
			}
		}

		// 添加显存信息
		var memory int64 = 0
		if i < len(gpuMemories) {
			memory = gpuMemories[i]
		}

		// 估算GPU TDP值
		tdp := estimateGPUTDP(model)

		// 添加到指标设备
		gpuDevice := struct {
			Model        string `json:"model" yaml:"model"`
			MemoryTotal  int64  `json:"memory_total" yaml:"memory_total"`
			CudaCores    int32  `json:"cuda_cores" yaml:"cuda_cores"`
			Manufacturer string `json:"manufacturer" yaml:"manufacturer"`
			Bus          string `json:"bus" yaml:"bus"`
			PciSlot      string `json:"pci_slot" yaml:"pci_slot"`
			Serial       string `json:"serial" yaml:"serial"`
			Architecture string `json:"architecture" yaml:"architecture"`
			TDP          int32  `json:"tdp" yaml:"tdp"`
		}{
			Model:        model,
			MemoryTotal:  memory,
			CudaCores:    0,
			Manufacturer: "NVIDIA",
			Bus:          gpuBuses[i],
			PciSlot:      gpuBuses[i],
			TDP:          int32(tdp),
		}
		hardwareInfo.GPU.Devices = append(hardwareInfo.GPU.Devices, gpuDevice)

		if memory > 0 {
			// 将MiB转换为GB
			gpuMemGB := float64(memory) / 1024
			if gpuMemGB >= 1 {
				// 四舍五入到整数
				gpuInfos = append(gpuInfos, fmt.Sprintf("%d,%s %dGB %dW", gpuIndex, model, int(math.Round(gpuMemGB)), tdp))
			} else {
				// 如果小于1GB，显示MB
				gpuInfos = append(gpuInfos, fmt.Sprintf("%d,%s %dMB %dW", gpuIndex, model, memory, tdp))
			}
		} else {
			gpuInfos = append(gpuInfos, fmt.Sprintf("%d,%s %dW", gpuIndex, model, tdp))
		}
	}

	// 设置格式化后的GPU信息
	if len(gpuInfos) > 0 {
		config["GPU"] = strings.Join(gpuInfos, ";")
	} else {
		// 尝试使用替代方法检测GPU
		if out, err := exec.Command("lspci", "-v").Output(); err == nil {
			output := string(out)
			if strings.Contains(strings.ToLower(output), "vga") ||
				strings.Contains(strings.ToLower(output), "nvidia") ||
				strings.Contains(strings.ToLower(output), "amd") ||
				strings.Contains(strings.ToLower(output), "intel") ||
				strings.Contains(strings.ToLower(output), "display") {

				// 提取显卡信息
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.Contains(strings.ToLower(line), "vga") ||
						strings.Contains(strings.ToLower(line), "display") ||
						strings.Contains(strings.ToLower(line), "3d") {

						parts := strings.SplitN(line, ":", 3)
						if len(parts) >= 2 {
							gpuModel := strings.TrimSpace(parts[1])
							if gpuModel != "" {
								// 估算TDP
								tdp := 75 // 默认集成显卡功耗

								// 添加到硬件信息结构
								gpuDevice := struct {
									Model        string `json:"model" yaml:"model"`
									MemoryTotal  int64  `json:"memory_total" yaml:"memory_total"`
									CudaCores    int32  `json:"cuda_cores" yaml:"cuda_cores"`
									Manufacturer string `json:"manufacturer" yaml:"manufacturer"`
									Bus          string `json:"bus" yaml:"bus"`
									PciSlot      string `json:"pci_slot" yaml:"pci_slot"`
									Serial       string `json:"serial" yaml:"serial"`
									Architecture string `json:"architecture" yaml:"architecture"`
									TDP          int32  `json:"tdp" yaml:"tdp"`
								}{
									Model: gpuModel,
									TDP:   int32(tdp),
								}

								// 确定制造商
								if strings.Contains(strings.ToLower(gpuModel), "nvidia") {
									gpuDevice.Manufacturer = "NVIDIA"
								} else if strings.Contains(strings.ToLower(gpuModel), "amd") ||
									strings.Contains(strings.ToLower(gpuModel), "radeon") {
									gpuDevice.Manufacturer = "AMD"
								} else if strings.Contains(strings.ToLower(gpuModel), "intel") {
									gpuDevice.Manufacturer = "Intel"
								} else {
									gpuDevice.Manufacturer = "Unknown"
								}

								hardwareInfo.GPU.Devices = append(hardwareInfo.GPU.Devices, gpuDevice)
								config["GPU"] = fmt.Sprintf("0,%s %dW", gpuModel, tdp)
								break
							}
						}
					}
				}
			}
		}

		// 如果还是没有找到GPU信息，检查是否在WSL环境中
		if config["GPU"] == "" {
			isWSL := false

			// 检查是否在WSL环境
			if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
				isWSL = true
			} else if out, err := exec.Command("uname", "-r").Output(); err == nil {
				if strings.Contains(strings.ToLower(string(out)), "microsoft") ||
					strings.Contains(strings.ToLower(string(out)), "wsl") {
					isWSL = true
				}
			}

			// 在WSL2环境中，强制尝试使用nvidia-smi直接获取GPU信息
			if isWSL {
				// 输出调试信息
				fmt.Println("检测到WSL环境，尝试强制获取GPU信息")

				// 尝试在WSL2环境下使用nvidia-smi获取GPU名称
				if out, err := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader").Output(); err == nil {
					gpuModel := strings.TrimSpace(string(out))
					if len(gpuModel) > 0 {
						// 获取显存
						var memory int64 = 8 * 1024 // 默认8GB，单位MiB
						if memOut, err := exec.Command("nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader").Output(); err == nil {
							memStr := strings.TrimSpace(string(memOut))
							if strings.Contains(memStr, "MiB") {
								memVal := strings.TrimSuffix(memStr, " MiB")
								if val, err := strconv.ParseInt(memVal, 10, 64); err == nil {
									memory = val
								}
							}
						}

						// 估算TDP
						tdp := 100
						if strings.Contains(gpuModel, "4070") {
							tdp = 100
						} else if strings.Contains(gpuModel, "4080") {
							tdp = 320
						} else if strings.Contains(gpuModel, "4090") {
							tdp = 450
						}

						// 配置GPU信息
						gpuMemGB := float64(memory) / 1024
						config["GPU"] = fmt.Sprintf("0,%s %dGB %dW", gpuModel, int(math.Round(gpuMemGB)), tdp)

						// 添加到硬件信息结构
						gpuDevice := struct {
							Model        string `json:"model" yaml:"model"`
							MemoryTotal  int64  `json:"memory_total" yaml:"memory_total"`
							CudaCores    int32  `json:"cuda_cores" yaml:"cuda_cores"`
							Manufacturer string `json:"manufacturer" yaml:"manufacturer"`
							Bus          string `json:"bus" yaml:"bus"`
							PciSlot      string `json:"pci_slot" yaml:"pci_slot"`
							Serial       string `json:"serial" yaml:"serial"`
							Architecture string `json:"architecture" yaml:"architecture"`
							TDP          int32  `json:"tdp" yaml:"tdp"`
						}{
							Model:        gpuModel,
							MemoryTotal:  memory,
							Manufacturer: "NVIDIA",
							TDP:          int32(tdp),
						}

						hardwareInfo.GPU.Devices = append(hardwareInfo.GPU.Devices, gpuDevice)
						fmt.Println("成功配置GPU: ", config["GPU"])
					}
				}
			}
		}
	}

	// 4. 获取存储信息
	var storageDevices []struct {
		Name     string
		Model    string
		Type     string
		Capacity int64
	}

	// 获取存储设备列表
	if out, err := exec.Command("lsblk", "-d", "-o", "NAME,MODEL,SIZE,TYPE", "-n").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				var device struct {
					Name     string
					Model    string
					Type     string
					Capacity int64
				}

				// 设备名称
				if len(fields) >= 1 {
					device.Name = fields[0]
				}

				// 设备模型，可能为空，需要处理
				if len(fields) >= 2 {
					device.Model = fields[1]
				} else {
					device.Model = "Unknown"
				}

				// 设备容量
				sizeStr := fields[2]
				device.Capacity = parseStorageSize(sizeStr)

				// 设备类型
				if len(fields) >= 4 {
					device.Type = fields[3]
				} else {
					device.Type = "disk"
				}

				// 细化存储类型
				if device.Type == "disk" {
					device.Type = getStorageDeviceType(device.Name, device.Model)
				}

				// 跳过loop设备和空设备
				if device.Type != "loop" && device.Capacity > 0 {
					storageDevices = append(storageDevices, device)
				}
			}
		}
	}

	// 如果找到了存储设备，设置配置
	if len(storageDevices) > 0 {
		// 设置详细硬件信息 - 存储设备
		for _, device := range storageDevices {
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
				Model:    device.Model,
				Capacity: device.Capacity,
			}
			hardwareInfo.Storage.Devices = append(hardwareInfo.Storage.Devices, storageDevice)
		}

		// 尝试获取挂载点信息
		var mountInfos []struct {
			Device string
			Mount  string
			Type   string
		}

		// 获取挂载信息
		if out, err := exec.Command("df", "-T").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			// 跳过标题行
			for i := 1; i < len(lines); i++ {
				if i < len(lines) && len(lines[i]) > 0 {
					fields := strings.Fields(lines[i])
					if len(fields) >= 7 {
						mountInfos = append(mountInfos, struct {
							Device string
							Mount  string
							Type   string
						}{
							Device: fields[0],
							Type:   fields[1],
							Mount:  fields[6],
						})
					}
				}
			}
		}

		// 对于简化版配置，按照挂载点展示存储设备信息
		var storageInfos []string

		// 创建设备名到设备信息的映射，方便查找
		deviceMap := make(map[string]struct {
			Name     string
			Model    string
			Type     string
			Capacity int64
		})

		for _, device := range storageDevices {
			deviceMap[device.Name] = device
			// 有些情况下需要添加/dev/前缀进行匹配
			deviceMap["dev/"+device.Name] = device
			deviceMap["/dev/"+device.Name] = device
		}

		// 首先尝试匹配根目录(/)
		foundRoot := false
		for _, mount := range mountInfos {
			if mount.Mount == "/" {
				foundRoot = true
				deviceName := ""

				// 提取设备名
				if strings.HasPrefix(mount.Device, "/dev/") {
					deviceName = strings.TrimPrefix(mount.Device, "/dev/")
				} else {
					deviceName = mount.Device
				}

				// 查找匹配的设备
				if device, ok := deviceMap[deviceName]; ok {
					capacityStr := formatStorageSize(device.Capacity)
					storageInfos = append(storageInfos, fmt.Sprintf("/,%s %s", device.Type, capacityStr))
				} else {
					// 如果没有直接匹配，尝试部分匹配
					for _, device := range storageDevices {
						if strings.Contains(deviceName, device.Name) || strings.Contains(device.Name, deviceName) {
							capacityStr := formatStorageSize(device.Capacity)
							storageInfos = append(storageInfos, fmt.Sprintf("/,%s %s", device.Type, capacityStr))
							break
						}
					}
				}
				break
			}
		}

		// 如果没有找到根目录挂载点，使用第一个存储设备作为根目录
		if !foundRoot && len(storageDevices) > 0 {
			device := storageDevices[0]
			capacityStr := formatStorageSize(device.Capacity)
			storageInfos = append(storageInfos, fmt.Sprintf("/,%s %s", device.Type, capacityStr))
		} else if !foundRoot {
			// 如果根本没有找到任何存储设备，添加一个带"-"类型的根目录
			storageInfos = append(storageInfos, "/,-")
		}

		// 添加其他主要挂载点（非系统挂载点）
		for _, mount := range mountInfos {
			// 排除系统挂载点和已处理的根目录
			if mount.Mount != "/" &&
				!strings.HasPrefix(mount.Mount, "/boot") &&
				!strings.HasPrefix(mount.Mount, "/sys") &&
				!strings.HasPrefix(mount.Mount, "/proc") &&
				!strings.HasPrefix(mount.Mount, "/dev") &&
				!strings.HasPrefix(mount.Mount, "/run") &&
				!strings.HasPrefix(mount.Mount, "/tmp") {

				deviceName := ""

				// 提取设备名
				if strings.HasPrefix(mount.Device, "/dev/") {
					deviceName = strings.TrimPrefix(mount.Device, "/dev/")
				} else {
					deviceName = mount.Device
				}

				// 查找匹配的设备
				if device, ok := deviceMap[deviceName]; ok {
					capacityStr := formatStorageSize(device.Capacity)
					storageInfos = append(storageInfos, fmt.Sprintf("%s,%s %s", mount.Mount, device.Type, capacityStr))
				} else {
					// 如果没有直接匹配，尝试部分匹配
					found := false
					for _, device := range storageDevices {
						if strings.Contains(deviceName, device.Name) || strings.Contains(device.Name, deviceName) {
							capacityStr := formatStorageSize(device.Capacity)
							storageInfos = append(storageInfos, fmt.Sprintf("%s,%s %s", mount.Mount, device.Type, capacityStr))
							found = true
							break
						}
					}

					// 如果还是找不到匹配，但挂载点看起来是重要的数据目录，则添加
					if !found && (strings.HasPrefix(mount.Mount, "/data") ||
						strings.HasPrefix(mount.Mount, "/mnt") ||
						strings.HasPrefix(mount.Mount, "/media") ||
						strings.HasPrefix(mount.Mount, "/home")) {
						// 使用挂载点信息，类型设为"-"
						storageInfos = append(storageInfos, fmt.Sprintf("%s,-", mount.Mount))
					}
				}
			}
		}

		// 如果没有找到任何有效的挂载点信息，回退到简单列出所有存储设备
		if len(storageInfos) == 0 {
			for _, device := range storageDevices {
				capacityStr := formatStorageSize(device.Capacity)
				storageInfos = append(storageInfos, fmt.Sprintf("-,%s %s", device.Type, capacityStr))
			}
		}

		config["Storage"] = strings.Join(storageInfos, ";")
	}

	// 确保Storage信息正确赋值，修复Storage信息可能缺失的问题
	if config["Storage"] == "" || !strings.Contains(config["Storage"], ",") {
		// 如果没有Storage信息或者格式不正确，尝试从根目录获取一个基本信息
		rootInfo := "/"
		rootType := "HDD"
		rootSize := "Unknown"

		// 尝试使用df命令直接获取根目录信息
		if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 2 {
					rootSize = fields[1]
				}
			}
		}

		// 如果有存储设备信息，使用第一个设备信息
		if len(storageDevices) > 0 {
			device := storageDevices[0]
			rootType = device.Type
			rootSize = formatStorageSize(device.Capacity)
		}

		config["Storage"] = fmt.Sprintf("%s,%s %s", rootInfo, rootType, rootSize)

		// 确保硬件信息结构中也有对应的存储设备
		if len(hardwareInfo.Storage.Devices) == 0 {
			var capacityBytes int64 = 0
			// 尝试解析容量字符串
			if strings.HasSuffix(rootSize, "GB") || strings.HasSuffix(rootSize, "G") {
				sizeStr := strings.TrimSuffix(strings.TrimSuffix(rootSize, "GB"), "G")
				if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
					capacityBytes = int64(size * 1024 * 1024 * 1024)
				}
			} else if strings.HasSuffix(rootSize, "TB") || strings.HasSuffix(rootSize, "T") {
				sizeStr := strings.TrimSuffix(strings.TrimSuffix(rootSize, "TB"), "T")
				if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
					capacityBytes = int64(size * 1024 * 1024 * 1024 * 1024)
				}
			}

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
				Type:     rootType,
				Model:    "System Disk",
				Capacity: capacityBytes,
				Path:     "/",
			}
			hardwareInfo.Storage.Devices = append(hardwareInfo.Storage.Devices, storageDevice)
		}
	}

	return config, hardwareInfo, nil
}

// GetSystemConfig 获取系统配置信息
func (a *GameNodeAgent) GetSystemConfig() (map[string]string, models.SystemInfo, error) {
	// 简化版系统配置（用于兼容现有接口）
	config := make(map[string]string)

	// 完整系统信息结构
	var systemInfo models.SystemInfo

	// 1. 获取操作系统信息
	if content, err := ioutil.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NAME=") {
				osType := strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
				config["os_type"] = osType
				systemInfo.OSDistribution = osType
			} else if strings.HasPrefix(line, "VERSION_ID=") {
				osVersion := strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
				config["os_version"] = osVersion
				systemInfo.OSVersion = osVersion
			}
		}
	}

	// 2. 获取操作系统架构
	if out, err := exec.Command("uname", "-m").Output(); err == nil {
		systemInfo.OSArchitecture = strings.TrimSpace(string(out))
	}

	// 3. 获取内核版本
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		systemInfo.KernelVersion = strings.TrimSpace(string(out))
	}

	// 4. 获取GPU驱动信息
	if out, err := exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader").Output(); err == nil {
		driverVersion := strings.TrimSpace(string(out))
		formattedDriverVersion := fmt.Sprintf("NVIDIA Driver %s", driverVersion)
		config["gpu_driver"] = formattedDriverVersion
		systemInfo.GPUDriverVersion = formattedDriverVersion
	}

	// 5. 获取CUDA版本
	if out, err := exec.Command("nvidia-smi", "--query-gpu=cuda_version", "--format=csv,noheader").Output(); err == nil {
		cudaVersion := strings.TrimSpace(string(out))
		config["cuda_version"] = cudaVersion
		systemInfo.CUDAVersion = cudaVersion
	}

	// 6. 获取容器运行时信息
	if out, err := exec.Command("docker", "--version").Output(); err == nil {
		systemInfo.DockerVersion = strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("containerd", "--version").Output(); err == nil {
		systemInfo.ContainerdVersion = strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("runc", "--version").Output(); err == nil {
		systemInfo.RuncVersion = strings.TrimSpace(string(out))
	}

	// 7. 获取网络信息
	if out, err := exec.Command("ip", "route", "get", "1").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "src") {
				fields := strings.Fields(line)
				for i, field := range fields {
					if field == "src" && i+1 < len(fields) {
						config["ip_address"] = fields[i+1]
						break
					}
				}
			}
		}
	}

	return config, systemInfo, nil
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// parseStorageSize 解析存储大小字符串为字节数
func parseStorageSize(sizeStr string) int64 {
	// 移除空格
	sizeStr = strings.TrimSpace(sizeStr)

	// 检查单位
	var multiplier int64 = 1
	unit := ""

	if len(sizeStr) > 0 {
		lastChar := sizeStr[len(sizeStr)-1]
		if !('0' <= lastChar && lastChar <= '9') {
			// 最后一个字符是单位
			unit = sizeStr[len(sizeStr)-1:]
			sizeStr = sizeStr[:len(sizeStr)-1]

			// 可能有两个字符的单位，如GB
			if len(sizeStr) > 0 && !('0' <= sizeStr[len(sizeStr)-1] && sizeStr[len(sizeStr)-1] <= '9') {
				unit = sizeStr[len(sizeStr)-1:] + unit
				sizeStr = sizeStr[:len(sizeStr)-1]
			}
		}
	}

	// 解析数值部分
	value, err := strconv.ParseFloat(strings.TrimSpace(sizeStr), 64)
	if err != nil {
		return 0
	}

	// 根据单位确定乘数
	unit = strings.ToUpper(unit)
	switch unit {
	case "K", "KB":
		multiplier = 1 << 10 // 1024
	case "M", "MB":
		multiplier = 1 << 20 // 1024^2
	case "G", "GB":
		multiplier = 1 << 30 // 1024^3
	case "T", "TB":
		multiplier = 1 << 40 // 1024^4
	case "P", "PB":
		multiplier = 1 << 50 // 1024^5
	}

	return int64(value * float64(multiplier))
}

// formatStorageSize 格式化存储大小
func formatStorageSize(bytes int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
		PB
	)

	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= int64(PB):
		unit = "PB"
		value = value / PB
	case bytes >= int64(TB):
		unit = "TB"
		value = value / TB
	case bytes >= int64(GB):
		unit = "GB"
		value = value / GB
	case bytes >= int64(MB):
		unit = "MB"
		value = value / MB
	case bytes >= int64(KB):
		unit = "KB"
		value = value / KB
	default:
		unit = "B"
	}

	if value >= 10 {
		return fmt.Sprintf("%.0f %s", value, unit)
	}
	return fmt.Sprintf("%.1f %s", value, unit)
}

// collectNodeInfo 采集节点信息
func (a *GameNodeAgent) collectNodeInfo() (*pb.RegisterRequest, error) {
	// 获取硬件和系统配置 - 静态信息，使用简化版格式
	hardware, hardwareInfo, err := a.GetHardwareConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get hardware config: %v", err)
	}

	system, systemInfo, err := a.GetSystemConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get system config: %v", err)
	}

	// 初始化状态中的硬件和系统信息 - 动态信息，使用详细版格式
	a.mu.Lock()
	if a.status == nil {
		a.status = &models.GameNodeStatus{
			State:      models.GameNodeStateOffline,
			Online:     false,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
		}
	}
	a.status.Hardware = hardwareInfo
	a.status.System = systemInfo
	a.mu.Unlock()

	return &pb.RegisterRequest{
		Id:       a.id,
		Type:     a.config.Type,
		Hardware: hardware, // 静态信息，键值对格式
		System:   system,   // 静态信息，键值对格式
	}, nil
}

// ExecutePipeline 执行流水线
func (a *GameNodeAgent) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) error {
	// 创建流水线
	pipeline, err := models.NewGameNodePipelineFromYAML(req.PipelineData)
	if err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "error", fmt.Sprintf("failed to create pipeline: %v", err)))
		return status.Error(codes.Internal, fmt.Sprintf("failed to create pipeline: %v", err))
	}

	// 保存流水线
	a.mu.Lock()
	a.pipelines[req.PipelineId] = pipeline
	a.mu.Unlock()

	// 执行流水线
	if err := a.executePipelineSteps(ctx, pipeline); err != nil {
		a.eventManager.Publish(event.NewNodeEvent(a.id, "error", fmt.Sprintf("failed to execute pipeline: %v", err)))
		return status.Error(codes.Internal, fmt.Sprintf("failed to execute pipeline: %v", err))
	}

	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "Pipeline completed successfully"))
	return nil
}

// executePipelineSteps 执行流水线步骤
func (a *GameNodeAgent) executePipelineSteps(ctx context.Context, pipeline *models.GameNodePipeline) error {
	for _, step := range pipeline.Steps {
		// 更新步骤状态
		statusUpdate := &pb.StepStatusUpdate{
			PipelineId: pipeline.Name,
			StepId:     step.Name,
			Status:     pb.StepStatus_RUNNING,
			StartTime:  time.Now().Unix(),
		}

		_, err := a.client.UpdateStepStatus(ctx, statusUpdate)
		if err != nil {
			return fmt.Errorf("failed to update step status: %v", err)
		}

		// 执行步骤
		if err := a.executeStep(ctx, &step); err != nil {
			// 更新失败状态
			statusUpdate.Status = pb.StepStatus_FAILED
			statusUpdate.EndTime = time.Now().Unix()
			statusUpdate.ErrorMessage = err.Error()
			_, _ = a.client.UpdateStepStatus(ctx, statusUpdate)
			return err
		}

		// 更新完成状态
		statusUpdate.Status = pb.StepStatus_COMPLETED
		statusUpdate.EndTime = time.Now().Unix()
		_, err = a.client.UpdateStepStatus(ctx, statusUpdate)
		if err != nil {
			return fmt.Errorf("failed to update step status: %v", err)
		}
	}

	return nil
}

// executeStep 执行单个步骤
func (a *GameNodeAgent) executeStep(ctx context.Context, step *models.PipelineStep) error {
	if a.dockerClient == nil {
		return fmt.Errorf("Docker client not initialized")
	}

	// 创建容器配置
	config := &container.Config{
		Image:      step.Container.Image,
		Cmd:        step.Container.Command,
		Env:        convertEnvMapToSlice(step.Container.Environment),
		WorkingDir: "/",
	}

	// 创建主机配置
	hostConfig := &container.HostConfig{
		Binds:       step.Container.Volumes,
		NetworkMode: container.NetworkMode("host"),
		Resources: container.Resources{
			Memory:     1024 * 1024 * 1024, // 1GB
			MemorySwap: 1024 * 1024 * 1024, // 1GB
			CPUShares:  1024,
			CPUPeriod:  100000,
			CPUQuota:   100000,
		},
	}

	// 创建容器
	resp, err := a.dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, fmt.Sprintf("step-%s-%s", step.Name, uuid.New().String()))
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	// 启动容器
	err = a.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// 等待容器完成
	statusCh, errCh := a.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to wait container: %v", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with status %d", status.StatusCode)
		}
	}

	// 获取容器日志
	logs, err := a.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to get container logs: %v", err)
	}

	// 读取日志内容
	logContent, err := io.ReadAll(logs)
	if err != nil {
		return fmt.Errorf("failed to read container logs: %v", err)
	}

	// 记录日志
	logEntry := &pb.LogEntry{
		PipelineId: step.Name,
		StepId:     step.Name,
		Level:      "info",
		Message:    string(logContent),
		Timestamp:  timestamppb.Now(),
	}
	a.logManager.AddLog(step.Name, logEntry)

	// 删除容器
	err = a.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	return nil
}

// convertEnvMapToSlice 将环境变量 map 转换为 slice
func convertEnvMapToSlice(envMap map[string]string) []string {
	envSlice := make([]string, 0, len(envMap))
	for k, v := range envMap {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}

// GetPipelineStatus 获取流水线状态
func (a *GameNodeAgent) GetPipelineStatus(ctx context.Context, pipelineID string) (*models.PipelineStatus, error) {
	// 获取流水线状态
	status := &models.PipelineStatus{
		ID:          pipelineID,
		NodeID:      a.id,
		State:       "unknown",
		CurrentStep: 0,
		TotalSteps:  0,
		Progress:    0,
		StartTime:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return status, nil
}

// CancelPipeline 取消流水线
func (a *GameNodeAgent) CancelPipeline(ctx context.Context, pipelineID string) error {
	// 发布取消事件
	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", fmt.Sprintf("Pipeline %s cancelled", pipelineID)))
	return nil
}

// SubscribeEvents 订阅事件
func (a *GameNodeAgent) SubscribeEvents(ctx context.Context, types []string) (<-chan *pb.Event, error) {
	subscriber := a.eventManager.Subscribe(types)
	eventCh := make(chan *pb.Event, 100)

	go func() {
		defer close(eventCh)
		for {
			select {
			case <-ctx.Done():
				a.eventManager.Unsubscribe(subscriber)
				return
			default:
				// 事件处理逻辑
			}
		}
	}()

	return eventCh, nil
}

// handleEvent 处理事件
func (a *GameNodeAgent) handleEvent(event *pb.Event) {
	switch event.Type {
	case "container":
		// 处理容器事件
		a.handleContainerEvent(event)
	case "pipeline":
		// 处理流水线事件
		a.handlePipelineEvent(event)
	case "node":
		// 处理节点事件
		a.handleNodeEvent(event)
	}
}

// handleContainerEvent 处理容器事件
func (a *GameNodeAgent) handleContainerEvent(event *pb.Event) {
	// TODO: 实现容器事件处理逻辑
}

// handlePipelineEvent 处理流水线事件
func (a *GameNodeAgent) handlePipelineEvent(event *pb.Event) {
	// TODO: 实现流水线事件处理逻辑
}

// handleNodeEvent 处理节点事件
func (a *GameNodeAgent) handleNodeEvent(event *pb.Event) {
	// TODO: 实现节点事件处理逻辑
}

// StreamLogs 流式获取日志
func (a *GameNodeAgent) StreamLogs(ctx context.Context, pipelineID string) (<-chan *pb.LogEntry, error) {
	return a.logManager.StreamLogs(ctx, pipelineID, time.Now().Add(-24*time.Hour)), nil
}

// StartContainer 启动容器
func (a *GameNodeAgent) StartContainer(ctx context.Context, containerID string) error {
	// 发布容器启动事件
	a.eventManager.Publish(event.NewContainerEvent(a.id, containerID, "started", "Container started"))
	return nil
}

// StopContainer 停止容器
func (a *GameNodeAgent) StopContainer(ctx context.Context, containerID string) error {
	// 发布容器停止事件
	a.eventManager.Publish(event.NewContainerEvent(a.id, containerID, "stopped", "Container stopped"))
	return nil
}

// StreamNodeLogs 流式获取节点日志
func (a *GameNodeAgent) StreamNodeLogs(ctx context.Context) (<-chan *pb.LogEntry, error) {
	return a.logManager.StreamLogs(ctx, a.id, time.Now().Add(-24*time.Hour)), nil
}

// StreamContainerLogs 流式获取容器日志
func (a *GameNodeAgent) StreamContainerLogs(ctx context.Context, containerID string) (<-chan *pb.LogEntry, error) {
	// 从过去24小时开始获取日志
	return a.logManager.StreamLogs(ctx, containerID, time.Now().Add(-24*time.Hour)), nil
}

// NewDefaultAgentConfig 创建默认配置
func NewDefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		HeartbeatPeriod: 30 * time.Second,
		RetryCount:      3,
		RetryDelay:      5 * time.Second,
		MetricsInterval: 60 * time.Second,
		Labels:          make(map[string]string),
	}
}

// Stop 停止代理并清理资源
func (a *GameNodeAgent) Stop() error {
	// 关闭gRPC连接
	if a.conn != nil {
		if err := a.conn.Close(); err != nil {
			return fmt.Errorf("failed to close gRPC connection: %v", err)
		}
	}

	// 发布节点关闭事件
	a.eventManager.Publish(event.NewNodeEvent(a.id, "info", "节点代理正在关闭"))

	// 更新状态
	a.mu.Lock()
	a.status.State = models.GameNodeStateOffline
	a.status.Online = false
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	return nil
}

// getStorageDeviceType 识别存储设备类型
func getStorageDeviceType(deviceName, deviceModel string) string {
	deviceName = strings.ToLower(deviceName)
	deviceModel = strings.ToLower(deviceModel)

	// 检查是否为NVMe设备
	if strings.Contains(deviceName, "nvme") || strings.Contains(deviceModel, "nvme") {
		return "NVMe"
	}

	// 检查是否为SSD设备
	if strings.Contains(deviceName, "ssd") || strings.Contains(deviceModel, "ssd") {
		return "SSD"
	}

	// 尝试使用lsblk获取设备旋转属性来判断是SSD还是HDD
	if out, err := exec.Command("lsblk", "-d", "-o", "NAME,ROTA", "-n", "/dev/"+deviceName).Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == "0" {
				return "SSD"
			} else if len(fields) >= 2 && fields[1] == "1" {
				return "HDD"
			}
		}
	}

	// 默认为普通硬盘
	return "HDD"
}

// estimateGPUTDP 估算GPU TDP值（瓦特）
func estimateGPUTDP(model string) int32 {
	modelLower := strings.ToLower(model)

	// RTX 40系列
	if strings.Contains(modelLower, "rtx 4090") {
		return 450
	} else if strings.Contains(modelLower, "rtx 4080") {
		return 320
	} else if strings.Contains(modelLower, "rtx 4070 ti") {
		return 285
	} else if strings.Contains(modelLower, "rtx 4070") {
		return 200
	} else if strings.Contains(modelLower, "rtx 4060 ti") {
		return 160
	} else if strings.Contains(modelLower, "rtx 4060") {
		return 115
	}

	// RTX 30系列
	if strings.Contains(modelLower, "rtx 3090 ti") {
		return 450
	} else if strings.Contains(modelLower, "rtx 3090") {
		return 350
	} else if strings.Contains(modelLower, "rtx 3080 ti") {
		return 350
	} else if strings.Contains(modelLower, "rtx 3080") {
		return 320
	} else if strings.Contains(modelLower, "rtx 3070 ti") {
		return 290
	} else if strings.Contains(modelLower, "rtx 3070") {
		return 220
	} else if strings.Contains(modelLower, "rtx 3060 ti") {
		return 200
	} else if strings.Contains(modelLower, "rtx 3060") {
		return 170
	} else if strings.Contains(modelLower, "rtx 3050") {
		return 130
	}

	// RTX 20系列
	if strings.Contains(modelLower, "rtx 2080 ti") {
		return 250
	} else if strings.Contains(modelLower, "rtx 2080 super") {
		return 250
	} else if strings.Contains(modelLower, "rtx 2080") {
		return 215
	} else if strings.Contains(modelLower, "rtx 2070 super") {
		return 215
	} else if strings.Contains(modelLower, "rtx 2070") {
		return 175
	} else if strings.Contains(modelLower, "rtx 2060 super") {
		return 175
	} else if strings.Contains(modelLower, "rtx 2060") {
		return 160
	}

	// GTX 16系列
	if strings.Contains(modelLower, "gtx 1660 ti") {
		return 120
	} else if strings.Contains(modelLower, "gtx 1660 super") {
		return 125
	} else if strings.Contains(modelLower, "gtx 1660") {
		return 120
	} else if strings.Contains(modelLower, "gtx 1650 super") {
		return 100
	} else if strings.Contains(modelLower, "gtx 1650") {
		return 75
	}

	// GTX 10系列
	if strings.Contains(modelLower, "gtx 1080 ti") {
		return 250
	} else if strings.Contains(modelLower, "gtx 1080") {
		return 180
	} else if strings.Contains(modelLower, "gtx 1070 ti") {
		return 180
	} else if strings.Contains(modelLower, "gtx 1070") {
		return 150
	} else if strings.Contains(modelLower, "gtx 1060") {
		return 120
	} else if strings.Contains(modelLower, "gtx 1050 ti") {
		return 75
	} else if strings.Contains(modelLower, "gtx 1050") {
		return 75
	}

	// 专业卡
	if strings.Contains(modelLower, "tesla") {
		if strings.Contains(modelLower, "v100") {
			return 300
		} else if strings.Contains(modelLower, "p100") {
			return 250
		} else if strings.Contains(modelLower, "t4") {
			return 70
		} else if strings.Contains(modelLower, "k80") {
			return 300
		} else if strings.Contains(modelLower, "a100") {
			return 400
		}
	} else if strings.Contains(modelLower, "quadro") {
		if strings.Contains(modelLower, "rtx") {
			return 295
		} else {
			return 185
		}
	} else if strings.Contains(modelLower, "titan") {
		return 280
	}

	// 默认值，如果无法识别
	return 150
}

// 简化CPU型号显示格式
func simpleCPUModel(fullModel string) string {
	// 移除常见的前缀和多余信息
	simplifiedModel := fullModel

	// 确定制造商
	var manufacturer string
	if strings.Contains(strings.ToLower(simplifiedModel), "intel") {
		manufacturer = "Intel "
	} else if strings.Contains(strings.ToLower(simplifiedModel), "amd") {
		manufacturer = "AMD "
	} else {
		manufacturer = ""
	}

	// 移除型号前的多余信息
	prefixes := []string{
		"Intel(R) Core(TM) ",
		"Intel(R) Core(TM)2 ",
		"Intel(R) ",
		"AMD ",
		"Genuine ",
		"Authentic ",
	}

	for _, prefix := range prefixes {
		if strings.Contains(simplifiedModel, prefix) {
			simplifiedModel = strings.Replace(simplifiedModel, prefix, "", 1)
		}
	}

	// 提取核心型号信息
	re := regexp.MustCompile(`(i[3579]-\d{4,5}[A-Z]*|Ryzen \d+ \d{4}[A-Z]*|Xeon [A-Z]?-?\d{4,5}[A-Z]*|A\d+-\d{4,5}[A-Z]*)`)
	matches := re.FindStringSubmatch(simplifiedModel)
	if len(matches) > 0 {
		return manufacturer + matches[0]
	}

	return manufacturer + simplifiedModel
}
