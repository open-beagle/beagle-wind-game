package gamenode

import (
	"fmt"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// AgentServerManager 定义节点管理器接口，负责与平台服务交互
type AgentServerManager interface {
	// Get 获取节点信息
	Get(nodeID string) (*pb.NodeInfo, error)
	// UpdateStatusState 更新节点状态
	UpdateStatusState(nodeID string, state string) error
	// UpdateStatusMetrics 更新节点指标
	UpdateStatusMetrics(nodeID string, metrics *pb.NodeMetrics) error
	// UpdateStatusResources 更新节点资源
	UpdateStatusResources(nodeID string, resources map[string]interface{}) error
	// UpdateStatusOnlineStatus 更新节点在线状态
	UpdateStatusOnlineStatus(nodeID string, online bool) error
}

// NewAgentServerManager 创建新的节点管理器
func NewAgentServerManager(nodeService *service.GameNodeService) AgentServerManager {
	return &agentServerManager{
		nodeService: nodeService,
	}
}

// agentServerManager 实现节点管理接口
type agentServerManager struct {
	nodeService *service.GameNodeService
}

// Get 获取节点信息
func (m *agentServerManager) Get(nodeID string) (*pb.NodeInfo, error) {
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
func (m *agentServerManager) UpdateStatusState(nodeID string, state string) error {
	return m.nodeService.UpdateStatusState(nodeID, state)
}

// UpdateStatusMetrics 更新节点指标
func (m *agentServerManager) UpdateStatusMetrics(nodeID string, metrics *pb.NodeMetrics) error {
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
func (m *agentServerManager) UpdateStatusResources(nodeID string, resources map[string]interface{}) error {
	return m.nodeService.UpdateStatusResources(nodeID, resources)
}

// UpdateStatusOnlineStatus 更新节点在线状态
func (m *agentServerManager) UpdateStatusOnlineStatus(nodeID string, online bool) error {
	return m.nodeService.UpdateStatusOnlineStatus(nodeID, online)
}
