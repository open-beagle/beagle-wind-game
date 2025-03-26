package agent

import (
	"fmt"

	pb "github.com/open-beagle/beagle-wind-game/internal/agent/proto"
	"github.com/open-beagle/beagle-wind-game/internal/models"
)

// NodeManager 节点管理器接口
type AgentServerManager interface {
	Get(id string) (*models.GameNode, error)
	UpdateStatusState(id string, status string) error
	UpdateStatusMetrics(id string, metrics map[string]interface{}) error
	UpdateStatusResources(id string, resources map[string]interface{}) error
	UpdateStatusOnlineStatus(id string, online bool) error
}

// UpdateStatusState 更新节点状态
func (s *AgentServer) updateStatusState(nodeID string, status string) error {
	if s.manager == nil {
		return fmt.Errorf("节点管理器未初始化")
	}

	return s.manager.UpdateStatusState(nodeID, status)
}

// updateMetrics 更新节点指标
func (s *AgentServer) updateStatusMetrics(nodeID string, metrics *pb.NodeMetrics) error {
	if s.manager == nil {
		return fmt.Errorf("节点管理器未初始化")
	}

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

	return s.manager.UpdateStatusMetrics(nodeID, nodeMetrics)
}

// updateResources 更新节点资源使用情况
func (s *AgentServer) updateStatusResources(nodeID string, metrics *pb.NodeMetrics) error {
	if s.manager == nil {
		return fmt.Errorf("节点管理器未初始化")
	}

	// 转换资源使用情况
	resources := make(map[string]interface{})

	resources["CPU_Usage"] = fmt.Sprintf("%.1f%%", metrics.CpuUsage)
	resources["RAM_Usage"] = fmt.Sprintf("%.1f%%", metrics.MemoryUsage)
	resources["Disk_Usage"] = fmt.Sprintf("%.1f%%", metrics.DiskUsage)

	if len(metrics.GpuMetrics) > 0 {
		resources["GPU_Usage"] = fmt.Sprintf("%.1f%%", metrics.GpuMetrics[0].Usage)
	}

	return s.manager.UpdateStatusResources(nodeID, resources)
}

// handleHeartbeat 处理节点心跳
func (s *AgentServer) handleHeartbeat(nodeID string, metrics *pb.NodeMetrics) error {
	if s.manager == nil {
		return fmt.Errorf("节点管理器未初始化")
	}

	// 更新节点在线状态
	if err := s.manager.UpdateStatusOnlineStatus(nodeID, true); err != nil {
		return fmt.Errorf("更新节点在线状态失败: %v", err)
	}

	// 更新节点指标
	if err := s.updateStatusMetrics(nodeID, metrics); err != nil {
		return fmt.Errorf("更新节点指标失败: %v", err)
	}

	// 更新节点资源使用情况
	if err := s.updateStatusResources(nodeID, metrics); err != nil {
		return fmt.Errorf("更新节点资源使用情况失败: %v", err)
	}

	return nil
}

// handleRegistration 处理节点注册
func (s *AgentServer) handleRegistration(nodeID string, info *pb.NodeInfo) error {
	if s.manager == nil {
		return fmt.Errorf("节点管理器未初始化")
	}

	// 检查节点是否存在
	_, err := s.manager.Get(nodeID)
	if err != nil {
		return fmt.Errorf("获取节点信息失败: %v", err)
	}

	// 更新节点状态
	if err := s.manager.UpdateStatusState(nodeID, string(models.GameNodeStateReady)); err != nil {
		return fmt.Errorf("更新节点状态失败: %v", err)
	}

	return nil
}

// handleDisconnection 处理节点断开连接
func (s *AgentServer) handleDisconnection(nodeID string) error {
	if s.manager == nil {
		return fmt.Errorf("节点管理器未初始化")
	}

	// 更新节点在线状态
	if err := s.manager.UpdateStatusOnlineStatus(nodeID, false); err != nil {
		return fmt.Errorf("更新节点在线状态失败: %v", err)
	}

	return nil
}
