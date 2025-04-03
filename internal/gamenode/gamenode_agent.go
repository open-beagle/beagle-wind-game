package gamenode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
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
		status: &models.GameNodeStatus{
			State:      models.GameNodeStateOffline,
			Online:     false,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
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

// connect 建立连接
func (a *GameNodeAgent) connect(_ context.Context) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(a.serverAddr, opts...)
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
	// 采集节点信息
	req, err := a.collectNodeInfo()
	if err != nil {
		return fmt.Errorf("采集节点信息失败: %v", err)
	}

	// 发送注册请求
	resp, err := a.client.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("发送注册请求失败: %v", err)
	}

	// 检查响应
	if !resp.Success {
		return fmt.Errorf("注册失败: %s", resp.Message)
	}

	// 更新状态
	a.mu.Lock()
	a.status.State = models.GameNodeStateOnline
	a.status.Online = true
	a.status.LastOnline = time.Now()
	a.status.UpdatedAt = time.Now()
	a.mu.Unlock()

	// 发布注册事件
	a.eventManager.Publish(event.NewNodeEvent(a.id, "registered", "Node registered successfully"))

	return nil
}

// collectResourceInfo 采集资源信息
func (a *GameNodeAgent) collectResourceInfo() (*pb.ResourceInfo, error) {
	a.resourceCollector.mu.RLock()
	defer a.resourceCollector.mu.RUnlock()

	return &pb.ResourceInfo{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Hardware: &pb.HardwareInfo{
			Cpu: &pb.CPUInfo{
				Usage: a.resourceCollector.cpuUsage,
			},
			Memory: &pb.MemoryInfo{
				Usage: a.resourceCollector.memoryUsage,
			},
			Gpu: &pb.GPUInfo{
				Usage:       a.resourceCollector.gpuUsage,
				MemoryUsage: a.resourceCollector.gpuMemory,
				Temperature: a.resourceCollector.gpuTemp,
				Power:       a.resourceCollector.gpuPower,
			},
			Disk: &pb.DiskInfo{
				Usage: a.resourceCollector.diskUsage,
			},
		},
		Network: &pb.NetworkInfo{
			Bandwidth: float64(a.resourceCollector.networkIO.read + a.resourceCollector.networkIO.write),
		},
	}, nil
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
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		ResourceInfo: &pb.ResourceInfo{
			Id:        a.id,
			Timestamp: time.Now().Unix(),
			Hardware:  resourceInfo.Hardware,
			Software:  resourceInfo.Software,
			Network:   resourceInfo.Network,
		},
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

	// 更新状态
	a.mu.Lock()
	a.status.Metrics = models.MetricsReport{
		ID:        a.id,
		Timestamp: time.Now().Unix(),
		Metrics:   make([]models.Metric, len(metrics)),
	}
	for i, m := range metrics {
		a.status.Metrics.Metrics[i] = models.Metric{
			Name:   m.Name,
			Type:   m.Type,
			Value:  m.Value,
			Labels: m.Labels,
		}
	}
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
			Name:  "memory_usage",
			Type:  "gauge",
			Value: a.resourceCollector.memoryUsage,
		},
		{
			Name:  "disk_usage",
			Type:  "gauge",
			Value: a.resourceCollector.diskUsage,
		},
		{
			Name:  "gpu_usage",
			Type:  "gauge",
			Value: a.resourceCollector.gpuUsage,
		},
		{
			Name:  "gpu_memory",
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
			Name:  "network_read",
			Type:  "counter",
			Value: float64(a.resourceCollector.networkIO.read),
		},
		{
			Name:  "network_write",
			Type:  "counter",
			Value: float64(a.resourceCollector.networkIO.write),
		},
	}
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
	return a.logManager.StreamLogs(ctx, containerID, time.Now().Add(-24*time.Hour)), nil
}

// GetNodeMetrics 获取节点指标
func (a *GameNodeAgent) GetNodeMetrics(ctx context.Context) (*pb.MetricsReport, error) {
	a.resourceCollector.mu.RLock()
	defer a.resourceCollector.mu.RUnlock()

	return &pb.MetricsReport{
		Id:        a.id,
		Timestamp: time.Now().Unix(),
		Metrics: []*pb.Metric{
			{
				Name:  "cpu_usage",
				Type:  "gauge",
				Value: a.resourceCollector.cpuUsage,
			},
			{
				Name:  "memory_usage",
				Type:  "gauge",
				Value: a.resourceCollector.memoryUsage,
			},
			{
				Name:  "gpu_usage",
				Type:  "gauge",
				Value: a.resourceCollector.gpuUsage,
			},
			{
				Name:  "gpu_memory",
				Type:  "gauge",
				Value: a.resourceCollector.gpuMemory,
			},
		},
	}, nil
}

// UpdateMetrics 更新指标
func (a *GameNodeAgent) UpdateMetrics(metrics *pb.MetricsReport) {
	a.resourceCollector.mu.Lock()
	defer a.resourceCollector.mu.Unlock()

	for _, metric := range metrics.Metrics {
		switch metric.Name {
		case "cpu_usage":
			a.resourceCollector.cpuUsage = metric.Value
		case "memory_usage":
			a.resourceCollector.memoryUsage = metric.Value
		case "gpu_usage":
			a.resourceCollector.gpuUsage = metric.Value
		case "gpu_memory":
			a.resourceCollector.gpuMemory = metric.Value
		}
	}
}

// Stop 停止代理
func (a *GameNodeAgent) Stop() {
	if a.conn != nil {
		a.conn.Close()
	}
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

// collectNodeInfo 采集节点信息
func (a *GameNodeAgent) collectNodeInfo() (*pb.RegisterRequest, error) {
	// 1. 采集硬件信息
	hardware, err := a.collectHardwareInfo()
	if err != nil {
		return nil, fmt.Errorf("采集硬件信息失败: %v", err)
	}

	// 2. 采集系统信息
	system, err := a.collectSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("采集系统信息失败: %v", err)
	}

	// 3. 获取节点标签
	labels, err := a.getNodeLabels()
	if err != nil {
		return nil, fmt.Errorf("获取节点标签失败: %v", err)
	}

	return &pb.RegisterRequest{
		Id:       a.id,
		Alias:    a.config.Alias,
		Model:    a.config.Model,
		Type:     a.config.Type,
		Location: a.config.Location,
		Hardware: hardware,
		System:   system,
		Labels:   labels,
	}, nil
}

// collectSystemInfo 采集系统信息
func (a *GameNodeAgent) collectSystemInfo() (map[string]string, error) {
	info := make(map[string]string)

	// 采集操作系统信息
	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH

	// 采集网络信息
	network, err := a.collectNetworkInfo()
	if err != nil {
		return nil, err
	}

	// 将网络信息序列化为 JSON
	networkJSON, err := json.Marshal(network)
	if err != nil {
		return nil, fmt.Errorf("序列化网络信息失败: %v", err)
	}
	info["network"] = string(networkJSON)

	// 采集其他系统信息
	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}

	// 采集 Docker 信息
	if dockerVersion, err := a.getDockerVersion(); err == nil {
		info["docker_version"] = dockerVersion
	}

	return info, nil
}

// collectNetworkInfo 采集网络信息
func (a *GameNodeAgent) collectNetworkInfo() (map[string]string, error) {
	info := make(map[string]string)

	// 采集网络接口信息
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				info[iface.Name] = ipnet.IP.String()
			}
		}
	}

	return info, nil
}

// getDockerVersion 获取 Docker 版本
func (a *GameNodeAgent) getDockerVersion() (string, error) {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// collectHardwareInfo 采集硬件信息
func (a *GameNodeAgent) collectHardwareInfo() (map[string]string, error) {
	info := make(map[string]string)

	// 采集 CPU 信息
	cpuInfo, err := a.collectCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("采集 CPU 信息失败: %v", err)
	}
	info["cpu_model"] = cpuInfo.Model
	info["cpu_cores"] = strconv.FormatInt(int64(cpuInfo.Cores), 10)
	info["cpu_threads"] = strconv.FormatInt(int64(cpuInfo.Threads), 10)

	// 采集内存信息
	memInfo, err := a.collectMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("采集内存信息失败: %v", err)
	}
	info["memory_total"] = strconv.FormatInt(memInfo.Total, 10)
	info["memory_available"] = strconv.FormatInt(memInfo.Available, 10)

	// 采集 GPU 信息
	gpuInfo, err := a.collectGPUInfo()
	if err != nil {
		return nil, fmt.Errorf("采集 GPU 信息失败: %v", err)
	}
	info["gpu_model"] = gpuInfo.Model
	info["gpu_memory"] = strconv.FormatInt(gpuInfo.MemoryTotal, 10)

	return info, nil
}

// collectCPUInfo 采集 CPU 信息
func (a *GameNodeAgent) collectCPUInfo() (*models.CPUInfo, error) {
	// 读取 /proc/cpuinfo 文件
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("读取 CPU 信息失败: %v", err)
	}

	info := &models.CPUInfo{
		Model:   "Unknown",
		Cores:   0,
		Threads: 0,
	}

	// 解析 CPU 信息
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			info.Model = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.HasPrefix(line, "processor") {
			info.Threads++
		} else if strings.HasPrefix(line, "cpu cores") {
			cores, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1]))
			if err == nil {
				info.Cores = int32(cores)
			}
		}
	}

	return info, nil
}

// collectMemoryInfo 采集内存信息
func (a *GameNodeAgent) collectMemoryInfo() (*models.MemoryInfo, error) {
	// 读取 /proc/meminfo 文件
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("读取内存信息失败: %v", err)
	}

	info := &models.MemoryInfo{
		Total:     0,
		Available: 0,
	}

	// 解析内存信息
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				total, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					info.Total = total * 1024 // 转换为字节
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				available, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					info.Available = available * 1024 // 转换为字节
				}
			}
		}
	}

	return info, nil
}

// collectGPUInfo 采集 GPU 信息
func (a *GameNodeAgent) collectGPUInfo() (*models.GPUInfo, error) {
	// 尝试使用 nvidia-smi 命令获取 GPU 信息
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取 GPU 信息失败: %v", err)
	}

	info := &models.GPUInfo{
		Model:       "Unknown",
		MemoryTotal: 0,
	}

	// 解析 GPU 信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		fields := strings.Split(lines[0], ",")
		if len(fields) >= 2 {
			info.Model = strings.TrimSpace(fields[0])
			memory, err := strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 64)
			if err == nil {
				info.MemoryTotal = memory * 1024 * 1024 // 转换为字节
			}
		}
	}

	return info, nil
}

// getNodeLabels 获取节点标签
func (a *GameNodeAgent) getNodeLabels() (map[string]string, error) {
	labels := make(map[string]string)

	// 从配置文件加载标签
	for k, v := range a.config.Labels {
		labels[k] = v
	}

	// 添加系统标签
	labels["os"] = runtime.GOOS
	labels["arch"] = runtime.GOARCH

	return labels, nil
}
