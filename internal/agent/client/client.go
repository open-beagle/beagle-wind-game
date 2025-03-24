package client

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-beagle/beagle-wind-game/internal/agent/proto"
)

// AgentClient 是节点Agent客户端
type AgentClient struct {
	opts          ClientOptions
	conn          *grpc.ClientConn
	client        proto.AgentServiceClient
	nodeID        string
	sessionID     string
	registered    bool
	lastHeartbeat time.Time
	metrics       *proto.NodeMetrics
	metricsMutex  sync.RWMutex
	reconnecting  bool
	reconnectChan chan struct{}
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// ClientOptions 客户端配置选项
type ClientOptions struct {
	ServerAddr            string
	NodeID                string
	TLSEnabled            bool
	TLSCertFile           string
	HeartbeatInterval     time.Duration
	ReconnectInterval     time.Duration
	MaxReconnectAttempts  int
	ConnectionTimeout     time.Duration
	InsecureSkipTLSVerify bool
}

// DefaultClientOptions 默认客户端配置
var DefaultClientOptions = ClientOptions{
	ServerAddr:            "localhost:50051",
	NodeID:                "",
	TLSEnabled:            false,
	TLSCertFile:           "",
	HeartbeatInterval:     10 * time.Second,
	ReconnectInterval:     5 * time.Second,
	MaxReconnectAttempts:  5,
	ConnectionTimeout:     30 * time.Second,
	InsecureSkipTLSVerify: false,
}

// NewAgentClient 创建新的Agent客户端
func NewAgentClient(opts ClientOptions) *AgentClient {
	if opts.NodeID == "" {
		opts.NodeID = fmt.Sprintf("node-%d", time.Now().UnixNano())
	}

	return &AgentClient{
		opts:          opts,
		nodeID:        opts.NodeID,
		reconnectChan: make(chan struct{}, 1),
		stopChan:      make(chan struct{}),
	}
}

// Start 启动Agent客户端
func (c *AgentClient) Start() error {
	// 连接到服务端
	err := c.connect()
	if err != nil {
		return fmt.Errorf("连接服务端失败: %w", err)
	}

	// 注册节点
	err = c.register()
	if err != nil {
		c.closeConnection()
		return fmt.Errorf("注册节点失败: %w", err)
	}

	// 启动心跳
	c.wg.Add(1)
	go c.heartbeatLoop()

	// 启动连接监控
	c.wg.Add(1)
	go c.connectionMonitor()

	return nil
}

// Stop 停止Agent客户端
func (c *AgentClient) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	c.closeConnection()
}

// connect 连接到服务端
func (c *AgentClient) connect() error {
	// 准备连接选项
	var dialOpts []grpc.DialOption

	// TLS配置
	if c.opts.TLSEnabled {
		if c.opts.TLSCertFile != "" {
			creds, err := credentials.NewClientTLSFromFile(c.opts.TLSCertFile, "")
			if err != nil {
				return fmt.Errorf("加载TLS证书失败: %w", err)
			}
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
		} else if c.opts.InsecureSkipTLSVerify {
			// 不安全模式，跳过证书验证
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
	} else {
		// 非加密连接
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 设置保活参数
	kaParams := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(kaParams))

	// 使用新的API创建客户端连接
	conn, err := grpc.NewClient(c.opts.ServerAddr, dialOpts...)
	if err != nil {
		return fmt.Errorf("连接gRPC服务端失败: %w", err)
	}

	c.conn = conn
	c.client = proto.NewAgentServiceClient(conn)
	return nil
}

// closeConnection 关闭连接
func (c *AgentClient) closeConnection() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}
}

// register 注册节点
func (c *AgentClient) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.opts.ConnectionTimeout)
	defer cancel()

	// 准备节点信息
	nodeInfo := &proto.NodeInfo{
		Hostname: "example-host", // TODO: 获取真实主机名
		Ip:       "127.0.0.1",    // TODO: 获取真实IP
		Os:       "linux",        // TODO: 获取真实操作系统
		Arch:     "amd64",        // TODO: 获取真实架构
		Labels:   map[string]string{"environment": "dev"},
		Hardware: &proto.HardwareInfo{
			Cpu: &proto.CpuInfo{
				Cores:      4,
				Model:      "Intel Core i7",
				ClockSpeed: 2.6,
			},
			Memory: &proto.MemoryInfo{
				Total: 8 * 1024 * 1024 * 1024, // 8GB
				Type:  "DDR4",
			},
			Disk: &proto.DiskInfo{
				Total: 256 * 1024 * 1024 * 1024, // 256GB
				Type:  "SSD",
			},
			Gpus: []*proto.GpuInfo{
				{
					Model:  "NVIDIA GeForce RTX 3080",
					Memory: 10 * 1024 * 1024 * 1024, // 10GB
					Driver: "460.32.03",
				},
			},
			Network: &proto.NetworkInfo{
				PrimaryInterface: "eth0",
				Bandwidth:        1000, // 1Gbps
			},
		},
	}

	// 发送注册请求
	req := &proto.RegisterRequest{
		NodeId:   c.nodeID,
		Hostname: nodeInfo.Hostname,
		NodeInfo: nodeInfo,
	}

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("注册请求失败: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("注册失败: %s", resp.Message)
	}

	c.sessionID = resp.SessionId
	c.registered = true
	fmt.Printf("节点 %s 注册成功，会话ID: %s\n", c.nodeID, c.sessionID)
	return nil
}

// sendHeartbeat 发送心跳
func (c *AgentClient) sendHeartbeat() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 准备指标数据
	c.metricsMutex.RLock()
	metrics := c.metrics
	if metrics == nil {
		// 创建示例指标
		metrics = &proto.NodeMetrics{
			CpuUsage:       float32(30 + time.Now().Unix()%20), // 30-50%随机波动
			MemoryUsage:    40.0,
			DiskUsage:      60.0,
			ContainerCount: 3,
			CollectedAt:    timestamppb.Now(),
			GpuMetrics: []*proto.GpuMetrics{
				{
					Index:       0,
					Usage:       25.0,
					MemoryUsage: 30.0,
					Temperature: 65.0,
				},
			},
			NetworkMetrics: &proto.NetworkMetrics{
				RxBytesPerSec: 1024 * 1024, // 1MB/s
				TxBytesPerSec: 512 * 1024,  // 512KB/s
			},
		}
	}
	c.metricsMutex.RUnlock()

	// 发送心跳请求
	req := &proto.HeartbeatRequest{
		NodeId:    c.nodeID,
		SessionId: c.sessionID,
		Metrics:   metrics,
	}

	resp, err := c.client.Heartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("心跳请求失败: %w", err)
	}

	if !resp.Success {
		c.registered = false
		return fmt.Errorf("心跳响应失败")
	}

	c.lastHeartbeat = time.Now()
	return nil
}

// heartbeatLoop 心跳循环
func (c *AgentClient) heartbeatLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.opts.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			if c.registered && c.client != nil {
				err := c.sendHeartbeat()
				if err != nil {
					fmt.Printf("发送心跳失败: %v\n", err)
					// 触发重连
					c.triggerReconnect()
				}
			}
		}
	}
}

// triggerReconnect 触发重连
func (c *AgentClient) triggerReconnect() {
	if !c.reconnecting {
		c.reconnecting = true
		select {
		case c.reconnectChan <- struct{}{}:
		default:
			// 已经有重连请求在队列中
		}
	}
}

// connectionMonitor 连接监控
func (c *AgentClient) connectionMonitor() {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			return
		case <-c.reconnectChan:
			c.handleReconnect()
		}
	}
}

// handleReconnect 处理重连
func (c *AgentClient) handleReconnect() {
	fmt.Println("开始重连...")

	// 关闭当前连接
	c.closeConnection()

	// 尝试重连
	attempt := 0
	for attempt < c.opts.MaxReconnectAttempts {
		attempt++
		fmt.Printf("重连尝试 %d/%d...\n", attempt, c.opts.MaxReconnectAttempts)

		// 连接服务端
		err := c.connect()
		if err != nil {
			fmt.Printf("重连失败: %v\n", err)
			time.Sleep(c.opts.ReconnectInterval)
			continue
		}

		// 重新注册
		err = c.register()
		if err != nil {
			fmt.Printf("重新注册失败: %v\n", err)
			c.closeConnection()
			time.Sleep(c.opts.ReconnectInterval)
			continue
		}

		// 重连成功
		fmt.Println("重连成功")
		c.reconnecting = false
		return
	}

	fmt.Printf("重连失败，已达到最大尝试次数: %d\n", c.opts.MaxReconnectAttempts)
	c.reconnecting = false
}

// UpdateMetrics 更新节点指标
func (c *AgentClient) UpdateMetrics(metrics *proto.NodeMetrics) {
	c.metricsMutex.Lock()
	defer c.metricsMutex.Unlock()

	c.metrics = metrics
}

// ExecuteCommand 执行命令
func (c *AgentClient) ExecuteCommand(command string, args []string) (string, error) {
	// TODO: 实现命令执行逻辑
	return fmt.Sprintf("执行命令: %s %v", command, args), nil
}

// SubscribeEvents 订阅事件
func (c *AgentClient) SubscribeEvents(ctx context.Context, eventTypes []string) (<-chan *proto.Event, error) {
	if !c.registered || c.client == nil {
		return nil, fmt.Errorf("客户端未连接或未注册")
	}

	req := &proto.EventSubscriptionRequest{
		NodeId:     c.nodeID,
		EventTypes: eventTypes,
	}

	stream, err := c.client.SubscribeEvents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("订阅事件失败: %w", err)
	}

	eventChan := make(chan *proto.Event, 10)

	// 启动接收协程
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(eventChan)

		for {
			event, err := stream.Recv()
			if err == io.EOF {
				// 流结束
				return
			}
			if err != nil {
				fmt.Printf("接收事件错误: %v\n", err)
				// 触发重连
				c.triggerReconnect()
				return
			}

			// 发送事件到通道
			select {
			case eventChan <- event:
			case <-ctx.Done():
				return
			case <-c.stopChan:
				return
			}
		}
	}()

	return eventChan, nil
}
