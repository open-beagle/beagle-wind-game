package gamenode

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// TestNewPipeline 测试创建新的Pipeline实例
func TestNewPipeline(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	req := &pb.ExecutePipelineRequest{
		NodeId:     "test-node",
		PipelineId: "test-pipeline",
		Pipeline: &pb.Pipeline{
			Name:        "测试流水线",
			Description: "用于测试的流水线",
			Steps: []*pb.PipelineStep{
				{
					Name: "step1",
					Type: "docker",
					Container: &pb.ContainerConfig{
						Image: "nginx:latest",
					},
				},
			},
		},
		Envs: map[string]string{
			"ENV1": "value1",
		},
		Args: map[string]string{
			"ARG1": "value1",
		},
	}

	pipeline := NewPipeline(req, dockerClient)
	assert.NotNil(t, pipeline)
	assert.Equal(t, req.PipelineId, pipeline.id)
	assert.Equal(t, req.Pipeline.Name, pipeline.name)
	assert.Equal(t, req.Pipeline.Description, pipeline.description)
	assert.Equal(t, req.Pipeline.Steps, pipeline.steps)
	assert.Equal(t, req.Envs, pipeline.envs)
	assert.Equal(t, req.Args, pipeline.args)
	assert.Equal(t, "pending", pipeline.status)
	assert.NotNil(t, pipeline.startTime)
	assert.NotNil(t, pipeline.containerStatuses)
	assert.NotNil(t, pipeline.dockerClient)
	assert.NotNil(t, pipeline.eventManager)
	assert.Equal(t, req.NodeId, pipeline.nodeID)
}

// TestPipeline_Execute 测试Pipeline执行
func TestPipeline_Execute(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	req := &pb.ExecutePipelineRequest{
		NodeId:     "test-node",
		PipelineId: "test-pipeline",
		Pipeline: &pb.Pipeline{
			Steps: []*pb.PipelineStep{
				{
					Name: "step1",
					Type: "docker",
					Container: &pb.ContainerConfig{
						Image:   "busybox:latest",
						Command: []string{"echo", "hello"},
					},
				},
			},
		},
	}

	pipeline := NewPipeline(req, dockerClient)
	assert.NotNil(t, pipeline)

	// 测试执行
	ctx := context.Background()
	err = pipeline.Execute(ctx)
	require.NoError(t, err)

	// 验证状态
	status := pipeline.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "completed", status.Status)
	assert.Equal(t, float32(1.0), status.Progress)
	assert.NotNil(t, status.StartTime)
	assert.NotNil(t, status.EndTime)
}

// TestPipeline_ExecuteError 测试Pipeline执行错误
func TestPipeline_ExecuteError(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	req := &pb.ExecutePipelineRequest{
		NodeId:     "test-node",
		PipelineId: "test-pipeline",
		Pipeline: &pb.Pipeline{
			Steps: []*pb.PipelineStep{
				{
					Name: "step1",
					Type: "docker",
					Container: &pb.ContainerConfig{
						Image:   "busybox:latest",
						Command: []string{"nonexistent-command"},
					},
				},
			},
		},
	}

	pipeline := NewPipeline(req, dockerClient)
	assert.NotNil(t, pipeline)

	// 测试执行
	ctx := context.Background()
	err = pipeline.Execute(ctx)
	assert.Error(t, err)

	// 验证状态
	status := pipeline.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "failed", status.Status)
	assert.NotEmpty(t, status.ErrorMessage)
}

// TestPipeline_Cancel 测试Pipeline取消
func TestPipeline_Cancel(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	req := &pb.ExecutePipelineRequest{
		NodeId:     "test-node",
		PipelineId: "test-pipeline",
		Pipeline: &pb.Pipeline{
			Steps: []*pb.PipelineStep{
				{
					Name: "step1",
					Type: "docker",
					Container: &pb.ContainerConfig{
						Image:   "busybox:latest",
						Command: []string{"sleep", "30"},
					},
				},
			},
		},
	}

	pipeline := NewPipeline(req, dockerClient)
	assert.NotNil(t, pipeline)

	// 在后台执行pipeline
	ctx := context.Background()
	errCh := make(chan error)
	go func() {
		errCh <- pipeline.Execute(ctx)
	}()

	// 等待pipeline开始执行
	time.Sleep(1 * time.Second)

	// 取消pipeline
	err = pipeline.Cancel()
	require.NoError(t, err)

	// 等待pipeline结束
	select {
	case err := <-errCh:
		assert.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("pipeline未能及时取消")
	}

	// 验证状态
	status := pipeline.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "canceled", status.Status)
}
