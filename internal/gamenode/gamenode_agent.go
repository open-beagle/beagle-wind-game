package gamenode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// GameNodeAgent 代表一个游戏节点代理
type GameNodeAgent struct {
	sync.RWMutex

	// 基本信息
	nodeID   string
	hostname string
	info     *pb.NodeInfo

	// 连接相关
	serverAddr string
	conn       *grpc.ClientConn
	client     pb.GameNodeServiceClient
	sessionID  string

	// Docker 客户端
	dockerClient *client.Client

	// 运行状态
	running   bool
	stopCh    chan struct{}
	pipelines map[string]*GameNodePipeline

	// 监控数据
	metrics *pb.NodeMetrics

	// 事件管理
	eventManager *GameNodeEventManager

	// 日志收集
	logCollector *GameNodeLogCollector

	// 其他状态
	status        string
	lastHeartbeat time.Time
	lastSeen      time.Time
}

// NewGameNodeAgent 创建新的游戏节点代理实例
func NewGameNodeAgent(nodeID string, dockerClient *client.Client) *GameNodeAgent {
	agent := &GameNodeAgent{
		nodeID:        nodeID,
		dockerClient:  dockerClient,
		eventManager:  NewGameNodeEventManager(),
		logCollector:  NewGameNodeLogCollector(dockerClient),
		status:        "disconnected",
		lastHeartbeat: time.Now(),
		pipelines:     make(map[string]*GameNodePipeline),
		stopCh:        make(chan struct{}),
	}

	return agent
}

// Start 启动游戏节点代理
func (a *GameNodeAgent) Start() error {
	a.Lock()
	defer a.Unlock()

	if a.running {
		return fmt.Errorf("agent is already running")
	}

	// 初始化系统信息
	if err := a.initSystemInfo(); err != nil {
		return fmt.Errorf("failed to init system info: %v", err)
	}

	// 建立gRPC连接
	conn, err := grpc.Dial(a.serverAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	a.conn = conn
	a.client = pb.NewGameNodeServiceClient(conn)

	// 注册节点
	resp, err := a.client.Register(context.Background(), &pb.RegisterRequest{
		NodeId:   a.nodeID,
		Hostname: a.hostname,
		NodeInfo: a.info,
	})
	if err != nil {
		return fmt.Errorf("failed to register node: %v", err)
	}

	a.sessionID = resp.SessionId
	a.status = "connected"
	a.running = true

	// 启动后台任务
	go a.heartbeatLoop()
	go a.metricsCollector()
	go a.eventLoop()

	return nil
}

// Stop 停止游戏节点代理
func (a *GameNodeAgent) Stop() error {
	a.Lock()
	defer a.Unlock()

	if !a.running {
		return fmt.Errorf("agent is not running")
	}

	// 停止所有pipeline
	for _, pipeline := range a.pipelines {
		pipeline.UpdateStatus(PipelineStateCanceled)
		a.eventManager.Publish(NewGameNodePipelineEvent(a.nodeID, pipeline.GetName(), "canceled", "Pipeline canceled"))
	}

	// 关闭所有后台任务
	close(a.stopCh)

	// 关闭gRPC连接
	if a.conn != nil {
		a.conn.Close()
	}

	a.running = false
	a.status = "stopped"
	return nil
}

// initSystemInfo 初始化系统信息
func (a *GameNodeAgent) initSystemInfo() error {
	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		return fmt.Errorf("failed to get host info: %v", err)
	}

	// 获取CPU信息
	cpuInfo, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("failed to get cpu info: %v", err)
	}

	// 获取内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory info: %v", err)
	}

	// 获取磁盘信息
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return fmt.Errorf("failed to get disk info: %v", err)
	}

	// 获取网络信息
	netInfo, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get network info: %v", err)
	}

	// 设置节点信息
	a.hostname = hostInfo.Hostname
	a.nodeID = fmt.Sprintf("%s-%d", hostInfo.Hostname, time.Now().UnixNano())
	a.info = &pb.NodeInfo{
		Hostname: hostInfo.Hostname,
		Os:       hostInfo.OS,
		Arch:     hostInfo.KernelArch,
		Kernel:   hostInfo.KernelVersion,
		Hardware: &pb.HardwareInfo{
			Cpu: &pb.CpuInfo{
				Cores:      int32(len(cpuInfo)),
				Model:      cpuInfo[0].ModelName,
				ClockSpeed: float32(cpuInfo[0].Mhz),
			},
			Memory: &pb.MemoryInfo{
				Total: int64(memInfo.Total),
				Type:  "DDR4", // 这里需要根据实际情况获取
			},
			Disk: &pb.DiskInfo{
				Total: int64(diskInfo.Total),
				Type:  "SSD", // 这里需要根据实际情况获取
			},
			Network: &pb.NetworkInfo{
				PrimaryInterface: netInfo[0].Name,
				Bandwidth:        1000, // 这里需要根据实际情况获取
			},
		},
	}

	return nil
}

// heartbeatLoop 心跳循环
func (a *GameNodeAgent) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				a.status = "disconnected"
				a.eventManager.Publish(NewGameNodeNodeEvent(a.nodeID, "disconnected", err.Error()))
			}
		}
	}
}

// sendHeartbeat 发送心跳
func (a *GameNodeAgent) sendHeartbeat() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := a.client.Heartbeat(ctx, &pb.HeartbeatRequest{
		NodeId:    a.nodeID,
		SessionId: a.sessionID,
		Metrics:   a.metrics,
	})
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %v", err)
	}

	a.lastHeartbeat = time.Now()
	a.status = "connected"
	return nil
}

// metricsCollector 指标收集器
func (a *GameNodeAgent) metricsCollector() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.collectMetrics(); err != nil {
				a.eventManager.Publish(NewGameNodeNodeEvent(a.nodeID, "error", err.Error()))
			}
		}
	}
}

// collectMetrics 收集系统指标
func (a *GameNodeAgent) collectMetrics() error {
	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Errorf("failed to get cpu usage: %v", err)
	}

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory usage: %v", err)
	}

	// 获取磁盘使用率
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return fmt.Errorf("failed to get disk usage: %v", err)
	}

	// 获取网络指标
	netStats, err := net.IOCounters(true)
	if err != nil {
		return fmt.Errorf("failed to get network stats: %v", err)
	}

	// 更新指标
	a.metrics = &pb.NodeMetrics{
		CpuUsage:    float32(cpuPercent[0]),
		MemoryUsage: float32(memInfo.UsedPercent),
		DiskUsage:   float32(diskInfo.UsedPercent),
		NetworkMetrics: &pb.NetworkMetrics{
			RxBytesPerSec: float32(netStats[0].BytesRecv),
			TxBytesPerSec: float32(netStats[0].BytesSent),
		},
		ContainerCount: int32(len(a.pipelines)),
		CollectedAt:    timestamppb.Now(),
	}

	return nil
}

// eventLoop 事件循环
func (a *GameNodeAgent) eventLoop() {
	subscriber := a.eventManager.Subscribe([]string{GameNodeEventTypeContainer, GameNodeEventTypePipeline, GameNodeEventTypeNode})
	defer a.eventManager.Unsubscribe(subscriber)

	for {
		select {
		case <-a.stopCh:
			return
		case event := <-subscriber.ch:
			// 处理事件
			a.handleEvent(event)
		}
	}
}

// handleEvent 处理事件
func (a *GameNodeAgent) handleEvent(event *pb.Event) {
	// 根据事件类型进行处理
	switch event.Type {
	case GameNodeEventTypePipeline:
		// 处理pipeline事件
	case GameNodeEventTypeContainer:
		// 处理容器事件
	case GameNodeEventTypeNode:
		// 处理节点事件
	}
}

// ExecutePipeline 执行Pipeline
func (a *GameNodeAgent) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) (*pb.ExecutePipelineResponse, error) {
	a.RLock()
	defer a.RUnlock()

	if !a.running {
		return nil, fmt.Errorf("agent is not running")
	}

	// 创建新的pipeline
	pipeline, err := NewGameNodePipelineFromYAML(req.PipelineData)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %v", err)
	}

	// 保存pipeline
	a.pipelines[req.PipelineId] = pipeline

	// 异步执行pipeline
	go func() {
		if err := a.runPipeline(ctx, pipeline); err != nil {
			a.eventManager.Publish(NewGameNodePipelineEvent(a.nodeID, req.PipelineId, "failed", err.Error()))
		}
	}()

	return &pb.ExecutePipelineResponse{
		ExecutionId: req.PipelineId,
		Accepted:    true,
		Message:     "Pipeline execution started",
	}, nil
}

// runPipeline 执行流水线
func (a *GameNodeAgent) runPipeline(ctx context.Context, pipeline *GameNodePipeline) error {
	if a.dockerClient == nil {
		return fmt.Errorf("docker client is not initialized")
	}

	// 设置开始时间
	pipeline.SetStartTime(time.Now().Unix())
	pipeline.UpdateStatus(PipelineStateRunning)

	// 执行每个步骤
	for i, step := range pipeline.GetSteps() {
		select {
		case <-ctx.Done():
			pipeline.UpdateStatus(PipelineStateCanceled)
			return fmt.Errorf("pipeline canceled")
		default:
			// 更新当前步骤
			pipeline.UpdateProgress(int32(i))
			pipeline.UpdateStepStatus(int32(i), StepStateRunning)

			// 创建容器配置
			config := &container.Config{
				Image:      step.Container.Image,
				Hostname:   step.Container.Hostname,
				Env:        convertMapToSlice(step.Container.Environment),
				Cmd:        step.Container.Command,
				WorkingDir: "/app",
			}

			// 创建主机配置
			hostConfig := &container.HostConfig{
				Privileged:    step.Container.Privileged,
				SecurityOpt:   step.Container.SecurityOpt,
				CapAdd:        step.Container.CapAdd,
				Tmpfs:         convertSliceToMap(step.Container.Tmpfs),
				Binds:         step.Container.Volumes,
				PortBindings:  nat.PortMap{},
				RestartPolicy: container.RestartPolicy{Name: "no"},
			}

			// 创建容器
			containerName := fmt.Sprintf("%s-%s-%d", pipeline.GetName(), step.Name, i)
			resp, err := a.dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
			if err != nil {
				pipeline.SetStepError(int32(i), err)
				pipeline.UpdateStepStatus(int32(i), StepStateFailed)
				pipeline.UpdateStatus(PipelineStateFailed)
				return fmt.Errorf("failed to create container: %v", err)
			}

			// 启动容器
			if err := a.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
				pipeline.SetStepError(int32(i), err)
				pipeline.UpdateStepStatus(int32(i), StepStateFailed)
				pipeline.UpdateStatus(PipelineStateFailed)
				return fmt.Errorf("failed to start container: %v", err)
			}

			// 等待容器完成
			statusCh, errCh := a.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
			select {
			case err := <-errCh:
				pipeline.SetStepError(int32(i), err)
				pipeline.UpdateStepStatus(int32(i), StepStateFailed)
				pipeline.UpdateStatus(PipelineStateFailed)
				return fmt.Errorf("error waiting for container: %v", err)
			case status := <-statusCh:
				if status.StatusCode != 0 {
					pipeline.SetStepError(int32(i), fmt.Errorf("container exited with status code %d", status.StatusCode))
					pipeline.UpdateStepStatus(int32(i), StepStateFailed)
					pipeline.UpdateStatus(PipelineStateFailed)
					return fmt.Errorf("container exited with status code %d", status.StatusCode)
				}
			}

			// 获取容器日志
			logs, err := a.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
			})
			if err == nil {
				defer logs.Close()
				// 读取日志
				scanner := bufio.NewScanner(logs)
				var output string
				for scanner.Scan() {
					output += scanner.Text() + "\n"
				}
				pipeline.SetStepOutput(int32(i), output)
			}

			// 更新步骤状态
			pipeline.UpdateStepStatus(int32(i), StepStateCompleted)
		}
	}

	// 设置结束时间
	pipeline.SetEndTime(time.Now().Unix())
	pipeline.UpdateStatus(PipelineStateCompleted)
	return nil
}

// convertMapToSlice 将 map[string]string 转换为 []string
func convertMapToSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// convertSliceToMap 将 []string 转换为 map[string]string
func convertSliceToMap(s []string) map[string]string {
	result := make(map[string]string)
	for _, v := range s {
		parts := strings.Split(v, ":")
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// GetPipelineStatus 获取Pipeline状态
func (a *GameNodeAgent) GetPipelineStatus(ctx context.Context, req *pb.PipelineStatusRequest) (*pb.PipelineStatusResponse, error) {
	a.RLock()
	defer a.RUnlock()

	if !a.running {
		return nil, fmt.Errorf("agent is not running")
	}

	pipeline, ok := a.pipelines[req.ExecutionId]
	if !ok {
		return nil, fmt.Errorf("pipeline not found")
	}

	status := pipeline.GetStatus()
	stepStatuses := make([]*pb.ContainerStatus, len(status.StepStatuses))
	for i, stepStatus := range status.StepStatuses {
		stepStatuses[i] = &pb.ContainerStatus{
			Name:         fmt.Sprintf("step-%d", i),
			Status:       string(stepStatus.State),
			ErrorMessage: stepStatus.Error,
			StartTime:    timestamppb.New(time.Unix(stepStatus.StartTime, 0)),
			EndTime:      timestamppb.New(time.Unix(stepStatus.EndTime, 0)),
		}
	}

	return &pb.PipelineStatusResponse{
		ExecutionId:            req.ExecutionId,
		Status:                 string(status.State),
		CurrentStep:            status.CurrentStep,
		TotalSteps:             status.TotalSteps,
		CurrentStepDescription: fmt.Sprintf("Step %d of %d", status.CurrentStep+1, status.TotalSteps),
		Progress:               status.Progress,
		ErrorMessage:           status.ErrorMessage,
		StartTime:              timestamppb.New(time.Unix(status.StartTime, 0)),
		EndTime:                timestamppb.New(time.Unix(status.EndTime, 0)),
		ContainerStatuses:      stepStatuses,
	}, nil
}

// CancelPipeline 取消Pipeline
func (a *GameNodeAgent) CancelPipeline(ctx context.Context, req *pb.PipelineCancelRequest) (*pb.PipelineCancelResponse, error) {
	a.RLock()
	pipeline, exists := a.pipelines[req.ExecutionId]
	a.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pipeline not found")
	}

	pipeline.UpdateStatus(PipelineStateCanceled)
	return &pb.PipelineCancelResponse{
		Success: true,
		Message: "Pipeline cancelled successfully",
	}, nil
}

// SubscribeEvents 订阅事件
func (a *GameNodeAgent) SubscribeEvents(req *pb.EventSubscriptionRequest) *GameNodeEventStream {
	return NewGameNodeEventStream(a.eventManager, req.EventTypes)
}

// StreamNodeLogs 流式获取节点日志
func (a *GameNodeAgent) StreamNodeLogs(ctx context.Context, req *pb.NodeLogsRequest) (<-chan *pb.LogEntry, error) {
	return a.logCollector.CollectNodeLogs(ctx, int64(req.TailLines), req.Follow)
}

// StreamContainerLogs 流式获取容器日志
func (a *GameNodeAgent) StreamContainerLogs(ctx context.Context, req *pb.ContainerLogsRequest) (<-chan *pb.LogEntry, error) {
	return a.logCollector.CollectContainerLogs(ctx, req.ContainerId, int64(req.TailLines), req.Follow)
}

// StartContainer 启动容器
func (a *GameNodeAgent) StartContainer(ctx context.Context, req *pb.StartContainerRequest) (*pb.StartContainerResponse, error) {
	a.RLock()
	defer a.RUnlock()

	if !a.running {
		return nil, fmt.Errorf("agent is not running")
	}

	// 将环境变量转换为字符串切片
	env := make([]string, 0, len(req.Config.Environment))
	for k, v := range req.Config.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// 创建容器配置
	config := &container.Config{
		Image:      req.Config.Image,
		Hostname:   req.Config.Hostname,
		Env:        env,
		Cmd:        req.Config.Command,
		WorkingDir: "/app",
	}

	// 将卷映射转换为字符串切片
	volumes := make([]string, 0, len(req.Config.Volumes))
	for _, v := range req.Config.Volumes {
		volumes = append(volumes, fmt.Sprintf("%s:%s", v.HostPath, v.ContainerPath))
	}

	// 创建主机配置
	hostConfig := &container.HostConfig{
		Privileged:    req.Config.Privileged,
		SecurityOpt:   req.Config.SecurityOpt,
		CapAdd:        req.Config.CapAdd,
		Tmpfs:         convertSliceToMap(req.Config.Tmpfs),
		Binds:         volumes,
		PortBindings:  nat.PortMap{},
		RestartPolicy: container.RestartPolicy{Name: "no"},
	}

	// 创建容器
	resp, err := a.dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, req.Config.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	// 启动容器
	if err := a.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	return &pb.StartContainerResponse{
		Success:     true,
		ContainerId: resp.ID,
		Message:     "Container started successfully",
	}, nil
}

// StopContainer 停止容器
func (a *GameNodeAgent) StopContainer(ctx context.Context, req *pb.StopContainerRequest) (*pb.StopContainerResponse, error) {
	a.RLock()
	defer a.RUnlock()

	if !a.running {
		return nil, fmt.Errorf("agent is not running")
	}

	timeout := int(req.Timeout)
	if err := a.dockerClient.ContainerStop(ctx, req.ContainerId, container.StopOptions{Timeout: &timeout}); err != nil {
		return nil, fmt.Errorf("failed to stop container: %v", err)
	}

	return &pb.StopContainerResponse{
		Success: true,
		Message: "Container stopped successfully",
	}, nil
}

// RestartContainer 重启容器
func (a *GameNodeAgent) RestartContainer(ctx context.Context, req *pb.RestartContainerRequest) (*pb.RestartContainerResponse, error) {
	a.RLock()
	defer a.RUnlock()

	if !a.running {
		return nil, fmt.Errorf("agent is not running")
	}

	timeout := int(req.Timeout)
	if err := a.dockerClient.ContainerRestart(ctx, req.ContainerId, container.StopOptions{Timeout: &timeout}); err != nil {
		return nil, fmt.Errorf("failed to restart container: %v", err)
	}

	return &pb.RestartContainerResponse{
		Success: true,
		Message: "Container restarted successfully",
	}, nil
}

// GetNodeMetrics 获取节点指标
func (a *GameNodeAgent) GetNodeMetrics(ctx context.Context, req *pb.NodeMetricsRequest) (*pb.NodeMetricsResponse, error) {
	a.RLock()
	defer a.RUnlock()

	if !a.running {
		return nil, fmt.Errorf("agent is not running")
	}

	// 获取容器指标
	containers, err := a.dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	containerMetrics := make([]*pb.ContainerMetrics, 0, len(containers))
	for _, c := range containers {
		stats, err := a.dockerClient.ContainerStats(ctx, c.ID, false)
		if err != nil {
			continue
		}
		defer stats.Body.Close()

		var statsJSON struct {
			CPUStats struct {
				CPUUsage struct {
					TotalUsage uint64 `json:"total_usage"`
				} `json:"cpu_usage"`
			} `json:"cpu_stats"`
			MemoryStats struct {
				Usage uint64 `json:"usage"`
			} `json:"memory_stats"`
		}
		if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
			continue
		}

		containerMetrics = append(containerMetrics, &pb.ContainerMetrics{
			ContainerId: c.ID,
			CpuUsage:    float32(statsJSON.CPUStats.CPUUsage.TotalUsage),
			MemoryUsage: float32(statsJSON.MemoryStats.Usage),
		})
	}

	return &pb.NodeMetricsResponse{
		NodeId:           a.nodeID,
		Metrics:          a.metrics,
		ContainerMetrics: containerMetrics,
	}, nil
}

// UpdateMetrics 更新节点指标
func (a *GameNodeAgent) UpdateMetrics(metrics *pb.NodeMetrics) {
	a.Lock()
	defer a.Unlock()
	a.metrics = metrics
	a.lastSeen = time.Now()
}
