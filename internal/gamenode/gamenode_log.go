package gamenode

import (
	"bufio"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// LogCollector 日志收集器
type LogCollector struct {
	sync.RWMutex

	dockerClient *client.Client
	logs         map[string]*LogBuffer
	stopCh       chan struct{}
}

// LogBuffer 日志缓冲区
type LogBuffer struct {
	sync.RWMutex

	containerID string
	lines       []*pb.LogEntry
	maxLines    int
}

// NewLogCollector 创建新的日志收集器
func NewLogCollector(dockerClient *client.Client) *LogCollector {
	return &LogCollector{
		dockerClient: dockerClient,
		logs:         make(map[string]*LogBuffer),
		stopCh:       make(chan struct{}),
	}
}

// Start 启动日志收集器
func (lc *LogCollector) Start(ctx context.Context) {
	go lc.run(ctx)
}

// Stop 停止日志收集器
func (lc *LogCollector) Stop() {
	close(lc.stopCh)
}

// run 运行日志收集器
func (lc *LogCollector) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-lc.stopCh:
			return
		}
	}
}

// CollectContainerLogs 收集容器日志
func (lc *LogCollector) CollectContainerLogs(ctx context.Context, containerID string, tailLines int32, follow bool) (<-chan *pb.LogEntry, error) {
	// 创建日志缓冲区
	buffer := &LogBuffer{
		containerID: containerID,
		maxLines:    1000, // 默认保留1000行
	}
	lc.Lock()
	lc.logs[containerID] = buffer
	lc.Unlock()

	// 获取容器日志
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tailLines),
		Follow:     follow,
		Timestamps: true,
	}

	reader, err := lc.dockerClient.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("获取容器日志失败: %v", err)
	}

	// 创建日志通道
	logCh := make(chan *pb.LogEntry, 100)

	// 启动日志处理协程
	go func() {
		defer reader.Close()
		defer close(logCh)

		scanner := bufio.NewScanner(reader)
		// 处理日志流
		for {
			select {
			case <-ctx.Done():
				return
			case <-lc.stopCh:
				return
			default:
				// 读取日志
				if !scanner.Scan() {
					if err := scanner.Err(); err != nil {
						fmt.Printf("读取日志失败: %v\n", err)
					}
					return
				}
				line := scanner.Text()

				// 解析日志
				entry, err := parseLogEntry(line)
				if err != nil {
					continue
				}

				// 添加到缓冲区
				buffer.Lock()
				buffer.lines = append(buffer.lines, entry)
				if len(buffer.lines) > buffer.maxLines {
					buffer.lines = buffer.lines[1:]
				}
				buffer.Unlock()

				// 发送到通道
				select {
				case logCh <- entry:
				default:
					// 如果通道已满，跳过该日志
				}
			}
		}
	}()

	return logCh, nil
}

// GetContainerLogs 获取容器日志
func (lc *LogCollector) GetContainerLogs(containerID string, tailLines int32) ([]*pb.LogEntry, error) {
	lc.RLock()
	buffer, exists := lc.logs[containerID]
	lc.RUnlock()

	if !exists {
		return nil, fmt.Errorf("容器日志不存在: %s", containerID)
	}

	buffer.RLock()
	defer buffer.RUnlock()

	// 如果请求的行数大于缓冲区大小，返回所有日志
	if tailLines >= int32(len(buffer.lines)) {
		return buffer.lines, nil
	}

	// 返回最后 tailLines 行日志
	return buffer.lines[len(buffer.lines)-int(tailLines):], nil
}

// parseLogEntry 解析日志条目
func parseLogEntry(line string) (*pb.LogEntry, error) {
	// 解析时间戳
	timestamp, err := time.Parse(time.RFC3339Nano, line[:30])
	if err != nil {
		return nil, err
	}

	// 确定日志来源
	var source string
	if line[31] == 's' {
		source = "stdout"
	} else if line[31] == 'e' {
		source = "stderr"
	} else {
		return nil, fmt.Errorf("无效的日志来源: %c", line[31])
	}

	// 提取日志内容
	content := line[32:]

	return &pb.LogEntry{
		Source:    source,
		Content:   content,
		Timestamp: timestamppb.New(timestamp),
	}, nil
}
