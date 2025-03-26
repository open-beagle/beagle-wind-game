package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/open-beagle/beagle-wind-game/internal/agent/proto"
)

// TestNewAgent 测试创建新的Agent实例
func TestNewAgent(t *testing.T) {
	tests := []struct {
		name       string
		serverAddr string
		wantErr    bool
	}{
		{
			name:       "创建有效的Agent",
			serverAddr: "localhost:50051",
			wantErr:    false,
		},
		{
			name:       "空服务器地址",
			serverAddr: "",
			wantErr:    false, // 空地址是允许的，但不推荐
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewAgent(tt.serverAddr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, agent)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, agent)
			assert.Equal(t, tt.serverAddr, agent.serverAddr)
			assert.NotEmpty(t, agent.nodeID)
			assert.NotEmpty(t, agent.hostname)
			assert.NotNil(t, agent.dockerClient)
			assert.NotNil(t, agent.eventManager)
			assert.NotNil(t, agent.logCollector)
			assert.NotNil(t, agent.stopCh)
			assert.NotNil(t, agent.pipelines)
		})
	}
}

// TestAgent_StartStop 测试Agent的启动和停止
func TestAgent_StartStop(t *testing.T) {
	agent, err := NewAgent("localhost:50051")
	require.NoError(t, err)
	require.NotNil(t, agent)

	// 测试启动
	ctx := context.Background()
	err = agent.Start(ctx)
	// 由于无法连接到实际服务器，预期会返回错误
	assert.Error(t, err)

	// 测试重复启动
	err = agent.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent 已经在运行")

	// 测试停止
	agent.Stop()
	assert.False(t, agent.running)

	// 测试重复停止
	agent.Stop() // 应该不会panic
}

// TestAgent_ExecutePipeline 测试Pipeline执行
func TestAgent_ExecutePipeline(t *testing.T) {
	agent, err := NewAgent("localhost:50051")
	require.NoError(t, err)
	require.NotNil(t, agent)

	ctx := context.Background()
	req := &pb.ExecutePipelineRequest{
		NodeId:     agent.nodeID,
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

	resp, err := agent.ExecutePipeline(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Accepted)
	assert.Equal(t, req.PipelineId, resp.ExecutionId)

	// 验证pipeline是否被正确保存
	agent.RLock()
	pipeline, exists := agent.pipelines[req.PipelineId]
	agent.RUnlock()
	assert.True(t, exists)
	assert.NotNil(t, pipeline)

	// 等待一段时间让pipeline开始执行
	time.Sleep(100 * time.Millisecond)

	// 测试获取pipeline状态
	status, err := agent.GetPipelineStatus(ctx, &pb.PipelineStatusRequest{
		NodeId:      agent.nodeID,
		ExecutionId: req.PipelineId,
	})
	require.NoError(t, err)
	assert.NotNil(t, status)

	// 测试取消pipeline
	cancelResp, err := agent.CancelPipeline(ctx, &pb.PipelineCancelRequest{
		NodeId:      agent.nodeID,
		ExecutionId: req.PipelineId,
	})
	require.NoError(t, err)
	assert.NotNil(t, cancelResp)
}

// TestAgent_EventSubscription 测试事件订阅
func TestAgent_EventSubscription(t *testing.T) {
	agent, err := NewAgent("localhost:50051")
	require.NoError(t, err)
	require.NotNil(t, agent)

	// 订阅事件
	req := &pb.EventSubscriptionRequest{
		NodeId:     agent.nodeID,
		EventTypes: []string{"node", "pipeline"},
	}
	stream := agent.SubscribeEvents(req)
	require.NotNil(t, stream)

	// 发布一个测试事件
	testEvent := NewNodeEvent(agent.nodeID, "test", "测试事件")
	agent.eventManager.Publish(testEvent)

	// 等待事件接收
	event, err := stream.Recv()
	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, agent.nodeID, event.NodeId)
	assert.Equal(t, "test", event.Status)
}

// TestAgent_MetricsCollection 测试指标收集
func TestAgent_MetricsCollection(t *testing.T) {
	agent, err := NewAgent("localhost:50051")
	require.NoError(t, err)
	require.NotNil(t, agent)

	// 测试节点信息收集
	info, err := agent.collectNodeInfo()
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.Hostname)
	assert.NotEmpty(t, info.Ip)

	// 测试节点指标收集
	metrics, err := agent.collectNodeMetrics()
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.True(t, metrics.CpuUsage >= 0 && metrics.CpuUsage <= 100)
	assert.True(t, metrics.MemoryUsage >= 0)
	assert.True(t, metrics.DiskUsage >= 0)
}
