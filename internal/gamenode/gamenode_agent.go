package gamenode

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	dockerclient "github.com/docker/docker/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/sysinfo"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
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

	// 日志框架
	logger utils.Logger

	// 配置
	opts *GameNodeOptions

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

// NewGameNodeAgent 创建新的gRPC客户端
func NewGameNodeAgent(
	ctx context.Context,
	id string,
	serverAddr string,
	dockerClient *dockerclient.Client,
	opts *GameNodeOptions,
) (*GameNodeAgent, error) {
	agent := &GameNodeAgent{
		id:           id,
		serverAddr:   serverAddr,
		dockerClient: dockerClient,
		opts:         opts,
		logger:       utils.New("GameNodeAgent"),
	}

	// 初始化系统信息采集器
	agent.hardwareCollector = sysinfo.NewHardwareCollector(nil)
	agent.systemCollector = sysinfo.NewSystemCollector(nil)
	agent.metricsCollector = sysinfo.NewMetricsCollector(nil)

	// 检测节点类型
	agent.nodeInfo.nodeType = agent.DetectNodeType()
	agent.logger.Info("检测到节点类型: %s", agent.nodeInfo.nodeType)

	// 建立连接
	if err := agent.connectClient(ctx); err != nil {
		return nil, fmt.Errorf("connect failed: %v", err)
	}

	// 注册节点
	if err := agent.Register(ctx); err != nil {
		agent.Close() // 确保关闭连接
		return nil, fmt.Errorf("register failed: %v", err)
	}

	return agent, nil
}

// connectClient 建立gRPC连接
func (a *GameNodeAgent) connectClient(ctx context.Context) error {
	a.logger.Debug("开始连接到服务器: %s", a.serverAddr)

	// 使用 NewClient 建立连接
	conn, err := grpc.NewClient(a.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		a.logger.Error("连接服务器失败: %v", err)
		return fmt.Errorf("连接服务器失败: %v", err)
	}

	// 等待连接就绪
	state := conn.GetState()
	if state != connectivity.Ready {
		conn.Connect()
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}

	a.conn = conn
	a.client = pb.NewGameNodeGRPCServiceClient(conn)
	a.logger.Info("成功连接到服务器")

	return nil
}

// Close 关闭连接
func (a *GameNodeAgent) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// Register 注册节点
func (a *GameNodeAgent) Register(ctx context.Context) error {
	a.logger.Debug("开始注册节点: %s", a.id)

	// 1. 参数验证
	if a.id == "" {
		return fmt.Errorf("节点ID不能为空")
	}
	if a.nodeInfo.nodeType == "" {
		return fmt.Errorf("节点类型不能为空")
	}

	// 2. 收集硬件和系统信息（带超时控制）
	collectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := a.collectHardwareAndSystemInfo(collectCtx); err != nil {
		a.logger.Error("收集硬件和系统信息失败: %v", err)
		return fmt.Errorf("收集硬件和系统信息失败: %w", err)
	}

	// 3. 获取标准化的硬件信息
	hardware, err := a.hardwareCollector.GetSimplifiedHardwareInfo()
	if err != nil {
		a.logger.Error("获取标准化硬件信息失败: %v", err)
		return fmt.Errorf("获取标准化硬件信息失败: %w", err)
	}

	// 4. 准备注册请求
	req := &pb.RegisterRequest{
		Id:       a.id,
		Alias:    a.id,
		Model:    "Beagle-Wind-2024",
		Type:     string(a.nodeInfo.nodeType),
		Location: "default",
		Hardware: hardware, // 使用标准化的硬件信息
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
		a.logger.Warn("操作系统发行版信息为空")
	}
	if req.System["os_version"] == "" {
		a.logger.Warn("操作系统版本信息为空")
	}
	if req.System["kernel_version"] == "" {
		a.logger.Warn("内核版本信息为空")
	}

	// 6. 发送注册请求（带超时控制）
	rpcCtx, cancel := context.WithTimeout(ctx, defaultRPCTimeout)
	defer cancel()

	resp, err := a.client.Register(rpcCtx, req)
	if err != nil {
		a.logger.Error("注册请求失败: %v", err)
		return fmt.Errorf("注册请求失败: %w", err)
	}

	if !resp.Success {
		a.logger.Error("注册失败: %s", resp.Message)
		return fmt.Errorf("注册失败: %s", resp.Message)
	}

	a.logger.Info("节点注册成功")
	return nil
}

// SendHeartbeat 发送心跳
func (a *GameNodeAgent) SendHeartbeat(ctx context.Context) error {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return fmt.Errorf("上下文已取消: %v", ctx.Err())
	}

	// 检查客户端连接
	if a.client == nil {
		return fmt.Errorf("客户端连接不可用")
	}

	// 准备心跳请求
	req := &pb.HeartbeatRequest{
		Id:        a.id,
		SessionId: a.id, // 使用节点ID作为会话ID
		Timestamp: time.Now().Unix(),
	}

	// 发送心跳请求，最多重试3次
	var lastErr error
	for i := 0; i < 3; i++ {
		_, err := a.client.Heartbeat(ctx, req)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
	}

	a.logger.Error("心跳发送失败: %v", lastErr)
	return fmt.Errorf("心跳发送失败: %v", lastErr)
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
			Usage:       gpu.Usage,
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

// UpdateMetrics 更新指标
func (a *GameNodeAgent) UpdateMetrics(ctx context.Context) error {
	// 使用 MetricsCollector 收集指标
	metricsInfo, err := a.metricsCollector.GetMetricsInfo()
	if err != nil {
		return fmt.Errorf("收集指标信息失败: %v", err)
	}

	// 更新本地状态
	a.mu.Lock()
	defer a.mu.Unlock()
	a.status.Metrics = metricsInfo
	a.status.UpdatedAt = time.Now()

	return nil
}

// Start 启动代理
func (a *GameNodeAgent) Start(ctx context.Context) error {
	// 1. 建立gRPC连接
	if err := a.connectClient(ctx); err != nil {
		return fmt.Errorf("connect failed: %v", err)
	}

	// 2. 注册节点
	if err := a.Register(ctx); err != nil {
		a.Close() // 确保关闭连接
		return fmt.Errorf("register failed: %v", err)
	}

	a.logger.Info("GameNodeAgent 已启动并完成注册")
	return nil
}

// Stop 停止代理
func (a *GameNodeAgent) Stop() error {
	if err := a.Close(); err != nil {
		a.logger.Error("关闭连接失败: %v", err)
		return err
	}
	a.logger.Info("GameNodeAgent 已停止")
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

// Heartbeat 发送心跳
func (a *GameNodeAgent) Heartbeat(ctx context.Context) error {
	a.logger.Debug("发送心跳: %s", a.id)

	req := &pb.HeartbeatRequest{
		Id:        a.id,
		SessionId: a.id, // 使用节点ID作为会话ID
		Timestamp: time.Now().Unix(),
	}

	_, err := a.client.Heartbeat(ctx, req)
	if err != nil {
		a.logger.Error("心跳失败: %v", err)
		return fmt.Errorf("心跳失败: %w", err)
	}

	a.logger.Info("成功发送心跳: %s", a.id)
	return nil
}

// ReportMetrics 发送指标报告
func (a *GameNodeAgent) ReportMetrics(ctx context.Context) error {
	// 使用 MetricsCollector 收集指标
	metricsInfo, err := a.metricsCollector.GetMetricsInfo()
	if err != nil {
		return fmt.Errorf("收集指标信息失败: %v", err)
	}

	// 准备请求
	req := &pb.MetricsRequest{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Metrics:   convertToProtoMetricsInfo(metricsInfo),
	}

	// 发送请求
	_, err = a.client.ReportMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("上报指标失败: %v", err)
	}

	return nil
}

// collectHardwareAndSystemInfo 收集硬件和系统信息
func (a *GameNodeAgent) collectHardwareAndSystemInfo(ctx context.Context) error {
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

	a.mu.Lock()
	defer a.mu.Unlock()

	// 更新状态
	a.status.Hardware = hardwareInfo
	a.status.System = systemInfo
	a.status.UpdatedAt = time.Now()

	return nil
}
