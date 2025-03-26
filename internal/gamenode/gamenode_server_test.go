package gamenode

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// MockDockerClient 模拟 Docker 客户端
type MockDockerClient struct {
	*client.Client
}

func NewMockDockerClient() *client.Client {
	return &client.Client{}
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	return container.CreateResponse{
		ID: "mock-container-id",
	}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return nil
}

func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)

	go func() {
		statusCh <- container.WaitResponse{
			StatusCode: 0,
		}
	}()

	return statusCh, errCh
}

func (m *MockDockerClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

// MockAgentServerManager 模拟AgentServerManager
type MockAgentServerManager struct {
	nodes map[string]*models.GameNode
}

func NewMockAgentServerManager() *MockAgentServerManager {
	return &MockAgentServerManager{
		nodes: make(map[string]*models.GameNode),
	}
}

func (m *MockAgentServerManager) Get(id string) (*models.GameNode, error) {
	if node, exists := m.nodes[id]; exists {
		return node, nil
	}
	return nil, nil
}

func (m *MockAgentServerManager) UpdateStatusState(id string, status string) error {
	if node, exists := m.nodes[id]; exists {
		node.Status.State = models.GameNodeState(status)
		return nil
	}
	return nil
}

func (m *MockAgentServerManager) UpdateStatusMetrics(id string, metrics map[string]interface{}) error {
	if node, exists := m.nodes[id]; exists {
		node.Status.Metrics = metrics
		return nil
	}
	return nil
}

func (m *MockAgentServerManager) UpdateStatusResources(id string, resources map[string]interface{}) error {
	if node, exists := m.nodes[id]; exists {
		node.Status.Resources = make(map[string]string)
		for k, v := range resources {
			node.Status.Resources[k] = v.(string)
		}
		return nil
	}
	return nil
}

func (m *MockAgentServerManager) UpdateStatusOnlineStatus(id string, online bool) error {
	if node, exists := m.nodes[id]; exists {
		node.Status.Online = online
		if online {
			node.Status.LastOnline = time.Now()
		}
		return nil
	}
	return nil
}

// TestNewAgentServer 测试创建新的AgentServer实例
func TestNewAgentServer(t *testing.T) {
	opts := ServerOptions{
		ListenAddr:   ":50051",
		MaxHeartbeat: 30 * time.Second,
	}
	manager := NewMockAgentServerManager()
	server := NewAgentServer(opts, manager)
	assert.NotNil(t, server)
	assert.Equal(t, opts, server.opts)
	assert.Equal(t, manager, server.manager)
	assert.NotNil(t, server.nodeConnections)
	assert.NotNil(t, server.dockerClient)
}

// TestAgentServer_Register 测试节点注册
func TestAgentServer_Register(t *testing.T) {
	opts := ServerOptions{
		ListenAddr:   ":50051",
		MaxHeartbeat: 30 * time.Second,
	}
	manager := NewMockAgentServerManager()
	server := NewAgentServer(opts, manager)

	// 测试注册请求
	req := &pb.RegisterRequest{
		NodeId: "test-node-1",
		NodeInfo: &pb.NodeInfo{
			Hostname: "test-host",
			Ip:       "192.168.1.1",
		},
	}

	resp, err := server.Register(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.SessionId)

	// 验证节点连接信息
	server.connectionsMutex.RLock()
	conn, exists := server.nodeConnections[req.NodeId]
	server.connectionsMutex.RUnlock()
	assert.True(t, exists)
	assert.NotNil(t, conn)
	assert.Equal(t, req.NodeId, conn.nodeID)
	assert.Equal(t, resp.SessionId, conn.sessionID)
	assert.Equal(t, req.NodeInfo, conn.info)
}

// TestAgentServer_Heartbeat 测试节点心跳
func TestAgentServer_Heartbeat(t *testing.T) {
	opts := ServerOptions{
		ListenAddr:   ":50051",
		MaxHeartbeat: 30 * time.Second,
	}
	manager := NewMockAgentServerManager()
	server := NewAgentServer(opts, manager)

	// 先注册节点
	registerReq := &pb.RegisterRequest{
		NodeId: "test-node-1",
		NodeInfo: &pb.NodeInfo{
			Hostname: "test-host",
			Ip:       "192.168.1.1",
		},
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)
	require.True(t, registerResp.Success)

	// 发送心跳
	heartbeatReq := &pb.HeartbeatRequest{
		NodeId:    "test-node-1",
		SessionId: registerResp.SessionId,
		Metrics: &pb.NodeMetrics{
			CpuUsage:    50.0,
			MemoryUsage: 60.0,
			DiskUsage:   70.0,
		},
	}

	resp, err := server.Heartbeat(context.Background(), heartbeatReq)
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// 验证节点状态更新
	server.connectionsMutex.RLock()
	conn, exists := server.nodeConnections[heartbeatReq.NodeId]
	server.connectionsMutex.RUnlock()
	assert.True(t, exists)
	assert.NotNil(t, conn)
	assert.True(t, conn.lastSeen.After(time.Now().Add(-time.Second)))
}

// TestAgentServer_SubscribeEvents 测试事件订阅
func TestAgentServer_SubscribeEvents(t *testing.T) {
	opts := ServerOptions{
		ListenAddr:   ":50051",
		MaxHeartbeat: 30 * time.Second,
	}
	manager := NewMockAgentServerManager()
	server := NewAgentServer(opts, manager)

	// 订阅事件
	req := &pb.EventSubscriptionRequest{
		NodeId:     "test-node-1",
		EventTypes: []string{"node", "pipeline"},
	}

	// 创建一个模拟的事件流
	mockStream := &mockEventStream{
		ch: make(chan *pb.Event, 1),
	}

	stream := server.SubscribeEvents(req, mockStream)
	assert.NotNil(t, stream)

	// 发布测试事件
	testEvent := &pb.Event{
		NodeId:  "test-node-1",
		Type:    "node",
		Status:  "started",
		Message: "节点已启动",
	}

	// 验证事件接收
	select {
	case event := <-mockStream.ch:
		assert.Equal(t, testEvent, event)
	case <-time.After(time.Second):
		t.Fatal("未能接收到事件")
	}
}

// mockEventStream 模拟事件流
type mockEventStream struct {
	pb.AgentService_SubscribeEventsServer
	ch chan *pb.Event
}

// TestAgentServer_ExecutePipeline 测试Pipeline执行
func TestAgentServer_ExecutePipeline(t *testing.T) {
	opts := ServerOptions{
		ListenAddr:   ":50051",
		MaxHeartbeat: 30 * time.Second,
	}
	manager := NewMockAgentServerManager()
	server := NewAgentServer(opts, manager)

	// 确保 Docker 客户端已初始化
	require.NotNil(t, server.dockerClient)

	// 先注册节点
	registerReq := &pb.RegisterRequest{
		NodeId: "test-node-1",
		NodeInfo: &pb.NodeInfo{
			Hostname: "test-host",
			Ip:       "192.168.1.1",
		},
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)
	require.True(t, registerResp.Success)

	// 执行Pipeline
	req := &pb.ExecutePipelineRequest{
		NodeId:     "test-node-1",
		PipelineId: "test-pipeline-1",
		Pipeline: &pb.Pipeline{
			Steps: []*pb.PipelineStep{
				{
					Name: "test-step",
					Type: "docker",
					Container: &pb.ContainerConfig{
						Image: "nginx:latest",
					},
				},
			},
		},
	}

	// 创建 Pipeline 实例
	pipeline := NewPipeline(req, server.dockerClient)

	// 保存 Pipeline 实例到服务器
	server.pipelinesMutex.Lock()
	server.pipelines[req.PipelineId] = pipeline
	server.pipelinesMutex.Unlock()

	resp, err := server.ExecutePipeline(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.Accepted)
	assert.Equal(t, req.PipelineId, resp.ExecutionId)
}

// TestAgentServer_GetPipelineStatus 测试获取Pipeline状态
func TestAgentServer_GetPipelineStatus(t *testing.T) {
	opts := ServerOptions{
		ListenAddr:   ":50051",
		MaxHeartbeat: 30 * time.Second,
	}
	manager := NewMockAgentServerManager()
	server := NewAgentServer(opts, manager)

	// 先注册节点
	registerReq := &pb.RegisterRequest{
		NodeId: "test-node-1",
		NodeInfo: &pb.NodeInfo{
			Hostname: "test-host",
			Ip:       "192.168.1.1",
		},
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)
	require.True(t, registerResp.Success)

	// 执行Pipeline
	executeReq := &pb.ExecutePipelineRequest{
		NodeId:     "test-node-1",
		PipelineId: "test-pipeline-1",
		Pipeline: &pb.Pipeline{
			Steps: []*pb.PipelineStep{
				{
					Name: "test-step",
					Type: "docker",
					Container: &pb.ContainerConfig{
						Image: "nginx:latest",
					},
				},
			},
		},
	}
	executeResp, err := server.ExecutePipeline(context.Background(), executeReq)
	require.NoError(t, err)
	require.True(t, executeResp.Accepted)

	// 获取Pipeline状态
	statusReq := &pb.PipelineStatusRequest{
		NodeId:      "test-node-1",
		ExecutionId: executeReq.PipelineId,
	}

	status, err := server.GetPipelineStatus(context.Background(), statusReq)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, executeReq.PipelineId, status.ExecutionId)
}

// TestAgentServer_CancelPipeline 测试取消Pipeline
func TestAgentServer_CancelPipeline(t *testing.T) {
	opts := ServerOptions{
		ListenAddr:   ":50051",
		MaxHeartbeat: 30 * time.Second,
	}
	manager := NewMockAgentServerManager()
	server := NewAgentServer(opts, manager)

	// 先注册节点
	registerReq := &pb.RegisterRequest{
		NodeId: "test-node-1",
		NodeInfo: &pb.NodeInfo{
			Hostname: "test-host",
			Ip:       "192.168.1.1",
		},
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)
	require.True(t, registerResp.Success)

	// 执行Pipeline
	executeReq := &pb.ExecutePipelineRequest{
		NodeId:     "test-node-1",
		PipelineId: "test-pipeline-1",
		Pipeline: &pb.Pipeline{
			Steps: []*pb.PipelineStep{
				{
					Name: "test-step",
					Type: "docker",
					Container: &pb.ContainerConfig{
						Image: "nginx:latest",
					},
				},
			},
		},
	}
	executeResp, err := server.ExecutePipeline(context.Background(), executeReq)
	require.NoError(t, err)
	require.True(t, executeResp.Accepted)

	// 取消Pipeline
	cancelReq := &pb.PipelineCancelRequest{
		NodeId:      "test-node-1",
		ExecutionId: executeReq.PipelineId,
		Force:       false,
	}

	resp, err := server.CancelPipeline(context.Background(), cancelReq)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}
