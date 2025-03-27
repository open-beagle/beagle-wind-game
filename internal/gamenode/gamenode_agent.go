package gamenode

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
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
}

// NewGameNodeAgent 创建新的游戏节点代理实例
func NewGameNodeAgent(serverAddr string, dockerClient *client.Client) *GameNodeAgent {
	if dockerClient == nil {
		return nil
	}

	agent := &GameNodeAgent{
		serverAddr:    serverAddr,
		dockerClient:  dockerClient,
		eventManager:  NewGameNodeEventManager(),
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
		a.conn.Close()
		return fmt.Errorf("failed to register node: %v", err)
	}

	if !resp.Success {
		a.conn.Close()
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	a.sessionID = resp.SessionId
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
		return nil
	}

	close(a.stopCh)
	if a.conn != nil {
		a.conn.Close()
	}

	a.running = false
	return nil
}

// initSystemInfo 初始化系统信息
func (a *GameNodeAgent) initSystemInfo() error {
	hostInfo, err := host.Info()
	if err != nil {
		return err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return err
	}

	a.hostname = hostInfo.Hostname
	a.info = &pb.NodeInfo{
		Hostname: hostInfo.Hostname,
		Ip:       hostInfo.HostID,
		Os:       hostInfo.OS,
		Arch:     hostInfo.Platform,
		Kernel:   hostInfo.KernelVersion,
		Hardware: &pb.HardwareInfo{
			Cpu: &pb.CpuInfo{
				Cores:      int32(cpuInfo[0].Cores),
				Model:      cpuInfo[0].ModelName,
				ClockSpeed: float32(cpuInfo[0].Mhz) / 1000,
			},
			Memory: &pb.MemoryInfo{
				Total: int64(memInfo.Total),
				Type:  "DDR4", // TODO: 获取实际内存类型
			},
			Disk: &pb.DiskInfo{
				Total: int64(diskInfo.Total),
				Type:  "SSD", // TODO: 获取实际磁盘类型
			},
			Network: &pb.NetworkInfo{
				PrimaryInterface: "eth0", // TODO: 获取实际网络接口
				Bandwidth:        1000,   // TODO: 获取实际带宽
			},
		},
	}

	return nil
}

// heartbeatLoop 维护与服务器的心跳
func (a *GameNodeAgent) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				// TODO: 处理心跳失败
				continue
			}
		}
	}
}

// sendHeartbeat 发送单次心跳
func (a *GameNodeAgent) sendHeartbeat() error {
	a.RLock()
	if !a.running {
		a.RUnlock()
		return fmt.Errorf("agent is not running")
	}
	client := a.client
	a.RUnlock()

	_, err := client.Heartbeat(context.Background(), &pb.HeartbeatRequest{
		NodeId:    a.nodeID,
		SessionId: a.sessionID,
		Metrics:   a.metrics,
	})
	return err
}

// metricsCollector 收集系统指标
func (a *GameNodeAgent) metricsCollector() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.collectMetrics(); err != nil {
				// TODO: 处理指标收集失败
				continue
			}
		}
	}
}

// collectMetrics 收集当前系统指标
func (a *GameNodeAgent) collectMetrics() error {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return err
	}

	a.Lock()
	a.metrics = &pb.NodeMetrics{
		CpuUsage:    float32(cpuPercent[0]),
		MemoryUsage: float32(memInfo.UsedPercent),
		DiskUsage:   float32(diskInfo.UsedPercent),
		CollectedAt: timestamppb.Now(),
	}
	a.Unlock()

	return nil
}

// eventLoop 处理事件循环
func (a *GameNodeAgent) eventLoop() {
	for {
		select {
		case <-a.stopCh:
			return
		default:
			// TODO: 实现事件处理逻辑
		}
	}
}

// ExecutePipeline 执行流水线
func (a *GameNodeAgent) ExecutePipeline(ctx context.Context, req *pb.ExecutePipelineRequest) (*pb.ExecutePipelineResponse, error) {
	// 创建流水线
	pipeline := NewGameNodePipeline(req.Pipeline, a.dockerClient)

	// 发布流水线开始事件
	a.eventManager.Publish(NewGameNodePipelineEvent(a.nodeID, req.PipelineId, "started", "流水线开始执行"))

	// 执行流水线
	go func() {
		if err := pipeline.Execute(ctx); err != nil {
			a.eventManager.Publish(NewGameNodePipelineEvent(a.nodeID, req.PipelineId, "failed", fmt.Sprintf("流水线执行失败: %v", err)))
			return
		}
		a.eventManager.Publish(NewGameNodePipelineEvent(a.nodeID, req.PipelineId, "completed", "流水线执行完成"))
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
func (a *GameNodeAgent) GetPipelineStatus(ctx context.Context, req *pb.PipelineStatusRequest) (*pb.PipelineStatusResponse, error) {
	a.RLock()
	pipeline, exists := a.pipelines[req.ExecutionId]
	a.RUnlock()

	if !exists {
		return nil, fmt.Errorf("流水线不存在: %s", req.ExecutionId)
	}

	return pipeline.GetStatus(), nil
}

// CancelPipeline 取消流水线执行
func (a *GameNodeAgent) CancelPipeline(ctx context.Context, req *pb.PipelineCancelRequest) (*pb.PipelineCancelResponse, error) {
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
	a.eventManager.Publish(NewGameNodePipelineEvent(a.nodeID, req.ExecutionId, "canceled", "流水线已取消"))

	return &pb.PipelineCancelResponse{
		Success: true,
		Message: "流水线已取消",
	}, nil
}

// SubscribeEvents 订阅事件
func (a *GameNodeAgent) SubscribeEvents(req *pb.EventSubscriptionRequest) *GameNodeEventStream {
	return NewGameNodeEventStream(a.eventManager, req.EventTypes)
}

// StreamNodeLogs 流式获取节点日志
func (a *GameNodeAgent) StreamNodeLogs(ctx context.Context, req *pb.NodeLogsRequest) (<-chan *pb.LogEntry, error) {
	// TODO: 实现节点日志收集
	return nil, fmt.Errorf("节点日志收集尚未实现")
}

// StreamContainerLogs 流式获取容器日志
func (a *GameNodeAgent) StreamContainerLogs(ctx context.Context, req *pb.ContainerLogsRequest) (<-chan *pb.LogEntry, error) {
	return a.logCollector.CollectContainerLogs(ctx, req.ContainerId, int64(req.TailLines), req.Follow)
}
