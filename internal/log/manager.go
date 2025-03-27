package log

import (
	"context"
	"sync"
	"time"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// LogManager 日志管理器接口
type LogManager interface {
	// AddLog 添加日志
	AddLog(nodeID string, entry *pb.LogEntry)
	// GetLogs 获取日志
	GetLogs(nodeID string, since time.Time) []*pb.LogEntry
	// ClearLogs 清除日志
	ClearLogs(nodeID string)
	// StreamLogs 流式获取日志
	StreamLogs(ctx context.Context, nodeID string, since time.Time) <-chan *pb.LogEntry
}

// DefaultLogManager 默认日志管理器实现
type DefaultLogManager struct {
	mu     sync.RWMutex
	logs   map[string][]*pb.LogEntry
	maxLog int
}

// NewDefaultLogManager 创建新的日志管理器
func NewDefaultLogManager() *DefaultLogManager {
	return &DefaultLogManager{
		logs:   make(map[string][]*pb.LogEntry),
		maxLog: 1000, // 每个节点最多保存1000条日志
	}
}

// AddLog 添加日志
func (m *DefaultLogManager) AddLog(nodeID string, entry *pb.LogEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logs := m.logs[nodeID]
	logs = append(logs, entry)

	// 如果日志数量超过限制，删除最旧的日志
	if len(logs) > m.maxLog {
		logs = logs[1:]
	}

	m.logs[nodeID] = logs
}

// GetLogs 获取日志
func (m *DefaultLogManager) GetLogs(nodeID string, since time.Time) []*pb.LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := m.logs[nodeID]
	if len(logs) == 0 {
		return nil
	}

	// 过滤指定时间之后的日志
	var filtered []*pb.LogEntry
	for _, log := range logs {
		if log.Timestamp.AsTime().After(since) {
			filtered = append(filtered, log)
		}
	}

	return filtered
}

// ClearLogs 清除日志
func (m *DefaultLogManager) ClearLogs(nodeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.logs, nodeID)
}

// StreamLogs 流式获取日志
func (m *DefaultLogManager) StreamLogs(ctx context.Context, nodeID string, since time.Time) <-chan *pb.LogEntry {
	ch := make(chan *pb.LogEntry, 100)

	go func() {
		defer close(ch)

		logs := m.GetLogs(nodeID, since)
		for _, log := range logs {
			select {
			case <-ctx.Done():
				return
			case ch <- log:
			}
		}
	}()

	return ch
}
