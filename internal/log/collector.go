package log

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// LogSource 日志来源
type LogSource string

const (
	LogSourceContainer LogSource = "container" // 容器日志
	LogSourceNode      LogSource = "node"      // 节点日志
	LogSourceSystem    LogSource = "system"    // 系统日志
)

// LogCollector 日志收集器接口
type LogCollector interface {
	// CollectContainerLogs 收集容器日志
	CollectContainerLogs(ctx context.Context, containerID string, tailLines int64, follow bool) (<-chan *pb.LogEntry, error)
	// CollectNodeLogs 收集节点日志
	CollectNodeLogs(ctx context.Context, tailLines int64, follow bool) (<-chan *pb.LogEntry, error)
	// CollectSystemLogs 收集系统日志
	CollectSystemLogs(ctx context.Context, tailLines int64, follow bool) (<-chan *pb.LogEntry, error)
}

// DefaultLogCollector 默认日志收集器实现
type DefaultLogCollector struct {
	dockerClient *client.Client
	logDir       string
}

// NewDefaultLogCollector 创建新的日志收集器
func NewDefaultLogCollector(dockerClient *client.Client, logDir string) *DefaultLogCollector {
	return &DefaultLogCollector{
		dockerClient: dockerClient,
		logDir:       logDir,
	}
}

// CollectContainerLogs 收集容器日志
func (lc *DefaultLogCollector) CollectContainerLogs(ctx context.Context, containerID string, tailLines int64, follow bool) (<-chan *pb.LogEntry, error) {
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
		return nil, fmt.Errorf("failed to get container logs: %w", err)
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
					Message:   line,
					Timestamp: timestamppb.Now(),
					Level:     "info",
				}

				// 根据日志内容判断来源
				if len(line) > 8 {
					switch line[8] {
					case '1':
						logEntry.Level = "stdout"
					case '2':
						logEntry.Level = "stderr"
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
func (lc *DefaultLogCollector) CollectNodeLogs(ctx context.Context, tailLines int64, follow bool) (<-chan *pb.LogEntry, error) {
	ch := make(chan *pb.LogEntry, 100)

	// 获取节点日志文件路径
	logFile := filepath.Join(lc.logDir, "node.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		close(ch)
		return nil, fmt.Errorf("node log file not found: %s", logFile)
	}

	file, err := os.Open(logFile)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to open node log file: %w", err)
	}

	go func() {
		defer file.Close()
		defer close(ch)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				logEntry := &pb.LogEntry{
					Message:   line,
					Timestamp: timestamppb.Now(),
					Level:     "info",
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

// CollectSystemLogs 收集系统日志
func (lc *DefaultLogCollector) CollectSystemLogs(ctx context.Context, tailLines int64, follow bool) (<-chan *pb.LogEntry, error) {
	ch := make(chan *pb.LogEntry, 100)

	// 获取系统日志文件路径
	logFile := filepath.Join(lc.logDir, "system.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		close(ch)
		return nil, fmt.Errorf("system log file not found: %s", logFile)
	}

	file, err := os.Open(logFile)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to open system log file: %w", err)
	}

	go func() {
		defer file.Close()
		defer close(ch)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				logEntry := &pb.LogEntry{
					Message:   line,
					Timestamp: timestamppb.Now(),
					Level:     "info",
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
