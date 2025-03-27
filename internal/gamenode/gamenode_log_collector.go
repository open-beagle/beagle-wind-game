package gamenode

import (
	"bufio"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GameNodeLogCollector 日志收集器
type GameNodeLogCollector struct {
	dockerClient *client.Client
}

// NewGameNodeLogCollector 创建新的日志收集器
func NewGameNodeLogCollector(dockerClient *client.Client) *GameNodeLogCollector {
	return &GameNodeLogCollector{
		dockerClient: dockerClient,
	}
}

// CollectContainerLogs 收集容器日志
func (lc *GameNodeLogCollector) CollectContainerLogs(ctx context.Context, containerID string, tailLines int64, follow bool) (<-chan *pb.LogEntry, error) {
	ch := make(chan *pb.LogEntry, 100)

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tailLines),
		Follow:     follow,
		Timestamps: true,
	}

	reader, err := lc.dockerClient.ContainerLogs(ctx, containerID, options)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to get container logs: %v", err)
	}

	go func() {
		defer reader.Close()
		defer close(ch)

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				// 解析日志行
				logEntry := &pb.LogEntry{
					Content:   line,
					Timestamp: timestamppb.Now(),
				}

				// 根据日志内容判断来源
				if len(line) > 8 {
					switch line[8] {
					case '1':
						logEntry.Source = "stdout"
					case '2':
						logEntry.Source = "stderr"
					}
				}

				select {
				case ch <- logEntry:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// CollectNodeLogs 收集节点日志
func (lc *GameNodeLogCollector) CollectNodeLogs(ctx context.Context, tailLines int64, follow bool) (<-chan *pb.LogEntry, error) {
	// TODO: 实现节点日志收集
	// 这里可以收集系统日志、应用日志等
	return nil, fmt.Errorf("not implemented")
}
