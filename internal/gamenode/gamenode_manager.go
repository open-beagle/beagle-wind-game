package gamenode

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// GameNodeManager 定义游戏节点管理器接口
type GameNodeManager interface {
	// 节点管理
	RegisterNode(ctx context.Context, nodeID string, info *pb.NodeInfo) error
	UpdateNodeStatus(ctx context.Context, nodeID string, status string) error
	UpdateNodeMetrics(ctx context.Context, nodeID string, metrics *pb.NodeMetrics) error
	GetNode(nodeID string) (*GameNode, error)
	ListNodes() []*GameNode
	DeleteNode(nodeID string) error

	// 流水线管理
	CreatePipeline(ctx context.Context, nodeID string, pipeline *pb.Pipeline) error
	GetPipeline(nodeID string, pipelineID string) (*GameNodePipeline, error)
	ListPipelines(nodeID string) []*GameNodePipeline
	DeletePipeline(nodeID string, pipelineID string) error
}

// GameNode 表示一个游戏节点
type GameNode struct {
	ID        string
	Info      *pb.NodeInfo
	Status    string
	Metrics   *pb.NodeMetrics
	Pipelines map[string]*GameNodePipeline
	LastSeen  time.Time
	mu        sync.RWMutex
}

// NewGameNode 创建新的游戏节点
func NewGameNode(id string, info *pb.NodeInfo) *GameNode {
	return &GameNode{
		ID:        id,
		Info:      info,
		Status:    "INITIALIZING",
		Pipelines: make(map[string]*GameNodePipeline),
		LastSeen:  time.Now(),
	}
}

// DefaultGameNodeManager 默认的游戏节点管理器实现
type DefaultGameNodeManager struct {
	nodes map[string]*GameNode
	mu    sync.RWMutex
}

// NewDefaultGameNodeManager 创建新的默认游戏节点管理器
func NewDefaultGameNodeManager() *DefaultGameNodeManager {
	return &DefaultGameNodeManager{
		nodes: make(map[string]*GameNode),
	}
}

// RegisterNode 注册新节点
func (m *DefaultGameNodeManager) RegisterNode(ctx context.Context, nodeID string, info *pb.NodeInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[nodeID]; exists {
		return fmt.Errorf("node already registered: %s", nodeID)
	}

	m.nodes[nodeID] = NewGameNode(nodeID, info)
	return nil
}

// UpdateNodeStatus 更新节点状态
func (m *DefaultGameNodeManager) UpdateNodeStatus(ctx context.Context, nodeID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	node.Status = status
	node.LastSeen = time.Now()
	return nil
}

// UpdateNodeMetrics 更新节点指标
func (m *DefaultGameNodeManager) UpdateNodeMetrics(ctx context.Context, nodeID string, metrics *pb.NodeMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	node.Metrics = metrics
	node.LastSeen = time.Now()
	return nil
}

// GetNode 获取节点信息
func (m *DefaultGameNodeManager) GetNode(nodeID string) (*GameNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	return node, nil
}

// ListNodes 列出所有节点
func (m *DefaultGameNodeManager) ListNodes() []*GameNode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make([]*GameNode, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// DeleteNode 删除节点
func (m *DefaultGameNodeManager) DeleteNode(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	delete(m.nodes, nodeID)
	return nil
}

// CreatePipeline 创建流水线
func (m *DefaultGameNodeManager) CreatePipeline(ctx context.Context, nodeID string, pipeline *pb.Pipeline) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	if _, exists := node.Pipelines[pipeline.Id]; exists {
		return fmt.Errorf("pipeline already exists: %s", pipeline.Id)
	}

	// TODO: 创建流水线实例
	return nil
}

// GetPipeline 获取流水线
func (m *DefaultGameNodeManager) GetPipeline(nodeID string, pipelineID string) (*GameNodePipeline, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	pipeline, exists := node.Pipelines[pipelineID]
	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	return pipeline, nil
}

// ListPipelines 列出节点的所有流水线
func (m *DefaultGameNodeManager) ListPipelines(nodeID string) []*GameNodePipeline {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return nil
	}

	pipelines := make([]*GameNodePipeline, 0, len(node.Pipelines))
	for _, pipeline := range node.Pipelines {
		pipelines = append(pipelines, pipeline)
	}

	return pipelines
}

// DeletePipeline 删除流水线
func (m *DefaultGameNodeManager) DeletePipeline(nodeID string, pipelineID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	if _, exists := node.Pipelines[pipelineID]; !exists {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	delete(node.Pipelines, pipelineID)
	return nil
}

// GameNodeManager 实现节点管理接口
type GameNodeManager struct {
	nodeService *service.GameNodeService
}

// Get 获取节点信息
func (m *GameNodeManager) Get(nodeID string) (*pb.NodeInfo, error) {
	node, err := m.nodeService.Get(nodeID)
	if err != nil {
		return nil, fmt.Errorf("获取节点信息失败: %w", err)
	}
	if node == nil {
		return nil, fmt.Errorf("节点不存在: %s", nodeID)
	}

	// 转换为 NodeInfo
	return &pb.NodeInfo{
		Hostname: node.Name,
		Ip:       node.Network["ip"],
		Os:       "linux", // TODO: 从系统获取
		Arch:     "amd64", // TODO: 从系统获取
		Kernel:   "",      // TODO: 从系统获取
		Hardware: &pb.HardwareInfo{
			Cpu: &pb.CpuInfo{
				Model: node.Hardware["cpu"],
			},
			Memory: &pb.MemoryInfo{
				Total: 0, // TODO: 从硬件配置中获取
				Type:  "Unknown",
			},
			Disk: &pb.DiskInfo{
				Total: 0, // TODO: 从硬件配置中获取
				Type:  "Unknown",
			},
			Network: &pb.NetworkInfo{
				PrimaryInterface: "", // TODO: 从网络配置中获取
				Bandwidth:        0,  // TODO: 从网络配置中获取
			},
		},
		Labels: node.Labels,
	}, nil
}

// UpdateStatusState 更新节点状态
func (m *GameNodeManager) UpdateStatusState(nodeID string, state string) error {
	return m.nodeService.UpdateStatusState(nodeID, state)
}

// UpdateStatusMetrics 更新节点指标
func (m *GameNodeManager) UpdateStatusMetrics(nodeID string, metrics *pb.NodeMetrics) error {
	// 转换指标数据
	nodeMetrics := make(map[string]interface{})

	// CPU指标
	nodeMetrics["cpu"] = map[string]interface{}{
		"usage": metrics.CpuUsage,
	}

	// 内存指标
	nodeMetrics["memory"] = map[string]interface{}{
		"usage": metrics.MemoryUsage,
	}

	// 磁盘指标
	nodeMetrics["disk"] = map[string]interface{}{
		"usage": metrics.DiskUsage,
	}

	// GPU指标
	if len(metrics.GpuMetrics) > 0 {
		gpuMetrics := make([]map[string]interface{}, len(metrics.GpuMetrics))
		for i, gpu := range metrics.GpuMetrics {
			gpuMetrics[i] = map[string]interface{}{
				"index":        gpu.Index,
				"usage":        gpu.Usage,
				"memory_usage": gpu.MemoryUsage,
				"temperature":  gpu.Temperature,
			}
		}
		nodeMetrics["gpu"] = gpuMetrics
	}

	// 网络指标
	if metrics.NetworkMetrics != nil {
		nodeMetrics["network"] = map[string]interface{}{
			"rx_bytes_per_sec": metrics.NetworkMetrics.RxBytesPerSec,
			"tx_bytes_per_sec": metrics.NetworkMetrics.TxBytesPerSec,
		}
	}

	return m.nodeService.UpdateStatusMetrics(nodeID, nodeMetrics)
}

// UpdateStatusResources 更新节点资源
func (m *GameNodeManager) UpdateStatusResources(nodeID string, resources map[string]interface{}) error {
	return m.nodeService.UpdateStatusResources(nodeID, resources)
}

// UpdateStatusOnlineStatus 更新节点在线状态
func (m *GameNodeManager) UpdateStatusOnlineStatus(nodeID string, online bool) error {
	return m.nodeService.UpdateStatusOnlineStatus(nodeID, online)
}
