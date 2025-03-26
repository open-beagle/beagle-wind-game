package agent

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLogCollector 测试创建新的日志收集器
func TestNewLogCollector(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	collector := NewLogCollector(dockerClient)
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.dockerClient)
	assert.NotNil(t, collector.logs)
	assert.NotNil(t, collector.stopCh)
}

// TestLogCollector_StartStop 测试日志收集器的启动和停止
func TestLogCollector_StartStop(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	collector := NewLogCollector(dockerClient)
	assert.NotNil(t, collector)

	// 测试启动
	ctx := context.Background()
	collector.Start(ctx)

	// 测试停止
	collector.Stop()
}

// TestLogCollector_CollectContainerLogs 测试容器日志收集
func TestLogCollector_CollectContainerLogs(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	collector := NewLogCollector(dockerClient)
	assert.NotNil(t, collector)

	// 启动一个测试容器
	ctx := context.Background()
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "busybox:latest",
		Cmd:   []string{"sh", "-c", "echo 'test log'; sleep 1"},
	}, nil, nil, nil, "test-container")
	require.NoError(t, err)
	containerID := resp.ID

	err = dockerClient.ContainerStart(ctx, containerID, container.StartOptions{})
	require.NoError(t, err)

	// 收集容器日志
	logCh, err := collector.CollectContainerLogs(ctx, containerID, 10, true)
	require.NoError(t, err)

	// 等待日志
	select {
	case log := <-logCh:
		assert.NotNil(t, log)
		assert.Contains(t, log.Content, "test log")
	case <-time.After(5 * time.Second):
		t.Fatal("未能收到容器日志")
	}

	// 获取容器日志
	logs, err := collector.GetContainerLogs(containerID, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Contains(t, logs[0].Content, "test log")

	// 清理
	err = dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	require.NoError(t, err)
}

// TestLogCollector_GetContainerLogs 测试获取容器日志
func TestLogCollector_GetContainerLogs(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	collector := NewLogCollector(dockerClient)
	assert.NotNil(t, collector)

	// 测试获取不存在的容器日志
	logs, err := collector.GetContainerLogs("nonexistent", 10)
	assert.Error(t, err)
	assert.Nil(t, logs)

	// 启动一个测试容器
	ctx := context.Background()
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "busybox:latest",
		Cmd:   []string{"sh", "-c", "for i in $(seq 1 5); do echo \"log $i\"; done"},
	}, nil, nil, nil, "test-container")
	require.NoError(t, err)
	containerID := resp.ID

	err = dockerClient.ContainerStart(ctx, containerID, container.StartOptions{})
	require.NoError(t, err)

	// 收集容器日志
	_, err = collector.CollectContainerLogs(ctx, containerID, 10, true)
	require.NoError(t, err)

	// 等待一段时间让日志收集完成
	time.Sleep(1 * time.Second)

	// 测试获取全部日志
	logs, err = collector.GetContainerLogs(containerID, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.LessOrEqual(t, len(logs), 10)

	// 测试获取部分日志
	logs, err = collector.GetContainerLogs(containerID, 3)
	require.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Equal(t, 3, len(logs))

	// 清理
	err = dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	require.NoError(t, err)
}

// TestLogCollector_ParseLogEntry 测试日志解析
func TestLogCollector_ParseLogEntry(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{
			name:    "解析stdout日志",
			line:    "2024-03-26T12:34:56.789012345Z stdout test log",
			wantErr: false,
		},
		{
			name:    "解析stderr日志",
			line:    "2024-03-26T12:34:56.789012345Z stderr error log",
			wantErr: false,
		},
		{
			name:    "无效的时间戳",
			line:    "invalid timestamp stdout test log",
			wantErr: true,
		},
		{
			name:    "无效的日志来源",
			line:    "2024-03-26T12:34:56.789012345Z invalid test log",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parseLogEntry(tt.line)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, entry)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, entry)
			assert.NotNil(t, entry.Timestamp)
			assert.NotEmpty(t, entry.Source)
			assert.NotEmpty(t, entry.Content)
		})
	}
}
