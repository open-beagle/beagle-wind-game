package agent

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/agent/proto"
)

// Agent 代表一个节点代理
type Agent struct {
	sync.RWMutex

	// 基本信息
	nodeID   string
	hostname string
	info     *pb.NodeInfo

	// 连接相关
	serverAddr string
	conn       *grpc.ClientConn
	client     pb.AgentServiceClient
	sessionID  string
	retryCount int
	maxRetries int
	retryDelay time.Duration
	maxDelay   time.Duration

	// Docker 客户端
	dockerClient *client.Client

	// 运行状态
	running   bool
	stopCh    chan struct{}
	pipelines map[string]*Pipeline

	// 监控数据
	metrics *pb.NodeMetrics

	// 事件管理
	eventManager *EventManager

	// 日志收集
	logCollector *LogCollector
}

// NewAgent 创建一个新的 Agent 实例
func NewAgent(serverAddr string) (*Agent, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("获取主机名失败: %v", err)
	}

	// 创建 Docker 客户端
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("创建 Docker 客户端失败: %v", err)
	}

	return &Agent{
		nodeID:       hostname, // 使用主机名作为节点ID
		hostname:     hostname,
		serverAddr:   serverAddr,
		maxRetries:   5,
		retryDelay:   100 * time.Millisecond,
		maxDelay:     30 * time.Second,
		stopCh:       make(chan struct{}),
		pipelines:    make(map[string]*Pipeline),
		dockerClient: dockerClient,
		eventManager: NewEventManager(),
		logCollector: NewLogCollector(dockerClient),
	}, nil
}

// Start 启动 Agent
func (a *Agent) Start(ctx context.Context) error {
	a.Lock()
	if a.running {
		a.Unlock()
		return fmt.Errorf("agent 已经在运行")
	}
	a.running = true
	a.Unlock()

	// 启动事件管理器
	a.eventManager.Start(ctx)

	// 启动日志收集器
	a.logCollector.Start(ctx)

	// 连接服务器
	if err := a.connect(); err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}

	// 注册节点
	if err := a.register(ctx); err != nil {
		return fmt.Errorf("注册节点失败: %v", err)
	}

	// 发布节点启动事件
	a.eventManager.Publish(NewNodeEvent(a.nodeID, "started", "节点已启动"))

	// 启动心跳
	go a.heartbeat(ctx)

	// 启动指标收集
	go a.collectMetrics(ctx)

	return nil
}

// Stop 停止 Agent
func (a *Agent) Stop() {
	a.Lock()
	defer a.Unlock()

	if !a.running {
		return
	}

	// 发布节点停止事件
	a.eventManager.Publish(NewNodeEvent(a.nodeID, "stopped", "节点正在停止"))

	close(a.stopCh)
	a.running = false

	// 停止事件管理器
	a.eventManager.Stop()

	// 停止日志收集器
	a.logCollector.Stop()

	if a.conn != nil {
		a.conn.Close()
	}
}

// ExecutePipeline 执行流水线
func (a *Agent) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) (*pb.ExecutePipelineResponse, error) {
	// 创建流水线
	pipeline := NewPipeline(req, a.dockerClient)

	// 发布流水线开始事件
	a.eventManager.Publish(NewPipelineEvent(a.nodeID, req.PipelineId, "started", "流水线开始执行"))

	// 执行流水线
	go func() {
		if err := pipeline.Execute(ctx); err != nil {
			a.eventManager.Publish(NewPipelineEvent(a.nodeID, req.PipelineId, "failed", fmt.Sprintf("流水线执行失败: %v", err)))
			return
		}
		a.eventManager.Publish(NewPipelineEvent(a.nodeID, req.PipelineId, "completed", "流水线执行完成"))
	}()

	// 保存流水线实例
	a.Lock()
	a.pipelines[req.PipelineId] = pipeline
	a.Unlock()

	return &pb.ExecutePipelineResponse{
		ExecutionId: req.PipelineId,
		Accepted:    true,
		Message:     "流水线已开始执行",
	}, nil
}

// GetPipelineStatus 获取流水线状态
func (a *Agent) GetPipelineStatus(ctx context.Context, req *pb.PipelineStatusRequest) (*pb.PipelineStatusResponse, error) {
	a.RLock()
	pipeline, exists := a.pipelines[req.ExecutionId]
	a.RUnlock()

	if !exists {
		return nil, fmt.Errorf("流水线不存在: %s", req.ExecutionId)
	}

	return pipeline.GetStatus(), nil
}

// CancelPipeline 取消流水线执行
func (a *Agent) CancelPipeline(ctx context.Context, req *pb.PipelineCancelRequest) (*pb.PipelineCancelResponse, error) {
	a.RLock()
	pipeline, exists := a.pipelines[req.ExecutionId]
	a.RUnlock()

	if !exists {
		return nil, fmt.Errorf("流水线不存在: %s", req.ExecutionId)
	}

	if err := pipeline.Cancel(); err != nil {
		return &pb.PipelineCancelResponse{
			Success: false,
			Message: fmt.Sprintf("取消流水线失败: %v", err),
		}, nil
	}

	// 发布流水线取消事件
	a.eventManager.Publish(NewPipelineEvent(a.nodeID, req.ExecutionId, "canceled", "流水线已取消"))

	return &pb.PipelineCancelResponse{
		Success: true,
		Message: "流水线已取消",
	}, nil
}

// SubscribeEvents 订阅事件
func (a *Agent) SubscribeEvents(req *pb.EventSubscriptionRequest) *EventStream {
	return NewEventStream(a.eventManager, req.EventTypes)
}

// connect 连接到服务器
func (a *Agent) connect() error {
	conn, err := grpc.Dial(a.serverAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	a.conn = conn
	a.client = pb.NewAgentServiceClient(conn)
	return nil
}

// register 注册节点
func (a *Agent) register(ctx context.Context) error {
	info, err := a.collectNodeInfo()
	if err != nil {
		return fmt.Errorf("收集节点信息失败: %v", err)
	}

	req := &pb.RegisterRequest{
		NodeId:   a.nodeID,
		Hostname: a.hostname,
		NodeInfo: info,
	}

	resp, err := a.client.Register(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("注册失败: %s", resp.Message)
	}

	a.sessionID = resp.SessionId
	a.info = info
	return nil
}

// heartbeat 发送心跳
func (a *Agent) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			req := &pb.HeartbeatRequest{
				NodeId:    a.nodeID,
				SessionId: a.sessionID,
				Metrics:   a.metrics,
			}

			_, err := a.client.Heartbeat(ctx, req)
			if err != nil {
				// 重试逻辑
				a.retryConnection(ctx)
			}
		}
	}
}

// collectMetrics 收集节点指标
func (a *Agent) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			metrics, err := a.collectNodeMetrics()
			if err != nil {
				continue
			}
			a.Lock()
			a.metrics = metrics
			a.Unlock()
		}
	}
}

// collectNodeInfo 收集节点信息
func (a *Agent) collectNodeInfo() (*pb.NodeInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &pb.NodeInfo{
		Hostname: a.hostname,
		Ip:       "", // TODO: 获取IP地址
		Os:       hostInfo.OS,
		Arch:     hostInfo.KernelArch,
		Kernel:   hostInfo.KernelVersion,
		Hardware: &pb.HardwareInfo{
			Cpu: &pb.CpuInfo{
				Cores:      int32(len(cpuInfo)),
				Model:      cpuInfo[0].ModelName,
				ClockSpeed: float32(cpuInfo[0].Mhz) / 1000.0,
			},
			Memory: &pb.MemoryInfo{
				Total: int64(memInfo.Total),
				Type:  "Unknown",
			},
			Disk: &pb.DiskInfo{
				Total: int64(diskInfo.Total),
				Type:  "Unknown",
			},
			// TODO: 添加GPU信息
			Network: &pb.NetworkInfo{
				PrimaryInterface: "", // TODO: 获取主网卡信息
				Bandwidth:        0,  // TODO: 获取带宽信息
			},
		},
		Labels: make(map[string]string),
	}, nil
}

// collectNodeMetrics 收集节点指标
func (a *Agent) collectNodeMetrics() (*pb.NodeMetrics, error) {
	// 获取 CPU 使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// 获取磁盘使用率
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	// 获取容器数量
	containers, err := a.dockerClient.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.NodeMetrics{
		CpuUsage:       float32(cpuPercent[0]),
		MemoryUsage:    float32(memInfo.UsedPercent),
		DiskUsage:      float32(diskInfo.UsedPercent),
		ContainerCount: int32(len(containers)),
		CollectedAt:    timestamppb.Now(),
	}, nil
}

// retryConnection 重试连接
func (a *Agent) retryConnection(ctx context.Context) {
	delay := a.retryDelay

	for a.retryCount < a.maxRetries {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			if err := a.connect(); err == nil {
				if err := a.register(ctx); err == nil {
					a.retryCount = 0
					return
				}
			}

			a.retryCount++
			delay = time.Duration(float64(delay) * 1.5)
			if delay > a.maxDelay {
				delay = a.maxDelay
			}
		}
	}
}

// StreamNodeLogs 流式获取节点日志
func (a *Agent) StreamNodeLogs(ctx context.Context, req *pb.NodeLogsRequest) (<-chan *pb.LogEntry, error) {
	// TODO: 实现节点日志收集
	return nil, fmt.Errorf("节点日志收集尚未实现")
}

// StreamContainerLogs 流式获取容器日志
func (a *Agent) StreamContainerLogs(ctx context.Context, req *pb.ContainerLogsRequest) (<-chan *pb.LogEntry, error) {
	return a.logCollector.CollectContainerLogs(ctx, req.ContainerId, req.TailLines, req.Follow)
}
