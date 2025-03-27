package models

import "time"

// GameNodeType 游戏节点类型
type GameNodeType string

const (
	GameNodeTypePhysical GameNodeType = "physical" // 物理节点
	GameNodeTypeVirtual  GameNodeType = "virtual"  // 虚拟节点
)

// GameNodeState 游戏节点状态
type GameNodeState string

const (
	GameNodeStateOffline     GameNodeState = "offline"     // 离线
	GameNodeStateOnline      GameNodeState = "online"      // 在线
	GameNodeStateMaintenance GameNodeState = "maintenance" // 维护
	GameNodeStateReady       GameNodeState = "ready"       // 就绪
	GameNodeStateBusy        GameNodeState = "busy"        // 忙碌
	GameNodeStateError       GameNodeState = "error"       // 错误
)

// CPUInfo CPU信息
type CPUInfo struct {
	Model       string  `json:"model" yaml:"model"`             // CPU型号
	Cores       int32   `json:"cores" yaml:"cores"`             // 核心数
	Threads     int32   `json:"threads" yaml:"threads"`         // 线程数
	Frequency   float64 `json:"frequency" yaml:"frequency"`     // 频率
	Temperature float64 `json:"temperature" yaml:"temperature"` // 温度
	Usage       float64 `json:"usage" yaml:"usage"`             // 使用率
	Cache       int64   `json:"cache" yaml:"cache"`             // 缓存大小
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total     int64   `json:"total" yaml:"total"`         // 总量
	Available int64   `json:"available" yaml:"available"` // 可用量
	Used      int64   `json:"used" yaml:"used"`           // 已用量
	Usage     float64 `json:"usage" yaml:"usage"`         // 使用率
	Type      string  `json:"type" yaml:"type"`           // 内存类型
	Frequency float64 `json:"frequency" yaml:"frequency"` // 频率
	Channels  int32   `json:"channels" yaml:"channels"`   // 通道数
}

// GPUInfo GPU信息
type GPUInfo struct {
	Model       string  `json:"model" yaml:"model"`               // GPU型号
	MemoryTotal int64   `json:"memory_total" yaml:"memory_total"` // 显存总量
	MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`   // 显存已用
	MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`   // 显存剩余
	MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"` // 显存使用率
	Usage       float64 `json:"usage" yaml:"usage"`               // GPU使用率
	Temperature float64 `json:"temperature" yaml:"temperature"`   // 温度
	Power       float64 `json:"power" yaml:"power"`               // 功耗
	CUDACores   int32   `json:"cuda_cores" yaml:"cuda_cores"`     // CUDA核心数
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Model      string  `json:"model" yaml:"model"`             // 磁盘型号
	Capacity   int64   `json:"capacity" yaml:"capacity"`       // 容量
	Used       int64   `json:"used" yaml:"used"`               // 已用空间
	Free       int64   `json:"free" yaml:"free"`               // 剩余空间
	Usage      float64 `json:"usage" yaml:"usage"`             // 使用率
	Type       string  `json:"type" yaml:"type"`               // 磁盘类型
	Interface  string  `json:"interface" yaml:"interface"`     // 接口类型
	ReadSpeed  float64 `json:"read_speed" yaml:"read_speed"`   // 读取速度
	WriteSpeed float64 `json:"write_speed" yaml:"write_speed"` // 写入速度
	IOPS       int64   `json:"iops" yaml:"iops"`               // IOPS
}

// SoftwareInfo 软件信息
type SoftwareInfo struct {
	OSDistribution    string `json:"os_distribution" yaml:"os_distribution"`       // 操作系统发行版
	OSVersion         string `json:"os_version" yaml:"os_version"`                 // 操作系统版本
	OSArchitecture    string `json:"os_architecture" yaml:"os_architecture"`       // 操作系统架构
	KernelVersion     string `json:"kernel_version" yaml:"kernel_version"`         // 内核版本
	GPUDriverVersion  string `json:"gpu_driver_version" yaml:"gpu_driver_version"` // GPU驱动版本
	CUDAVersion       string `json:"cuda_version" yaml:"cuda_version"`             // CUDA版本
	DockerVersion     string `json:"docker_version" yaml:"docker_version"`         // Docker版本
	ContainerdVersion string `json:"containerd_version" yaml:"containerd_version"` // Containerd版本
	RuncVersion       string `json:"runc_version" yaml:"runc_version"`             // Runc版本
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Bandwidth   float64 `json:"bandwidth" yaml:"bandwidth"`     // 带宽
	Latency     float64 `json:"latency" yaml:"latency"`         // 延迟
	Connections int32   `json:"connections" yaml:"connections"` // 连接数
	PacketLoss  float64 `json:"packet_loss" yaml:"packet_loss"` // 丢包率
}

// HardwareInfo 硬件信息
type HardwareInfo struct {
	CPU    CPUInfo    `json:"cpu" yaml:"cpu"`       // CPU信息
	Memory MemoryInfo `json:"memory" yaml:"memory"` // 内存信息
	GPU    GPUInfo    `json:"gpu" yaml:"gpu"`       // GPU信息
	Disk   DiskInfo   `json:"disk" yaml:"disk"`     // 磁盘信息
}

// ResourceInfo 资源信息
type ResourceInfo struct {
	ID        string       `json:"id" yaml:"id"`               // 节点ID
	Timestamp int64        `json:"timestamp" yaml:"timestamp"` // 时间戳
	Hardware  HardwareInfo `json:"hardware" yaml:"hardware"`   // 硬件信息
	Software  SoftwareInfo `json:"software" yaml:"software"`   // 软件信息
	Network   NetworkInfo  `json:"network" yaml:"network"`     // 网络信息
}

// Metric 指标数据
type Metric struct {
	Name   string            `json:"name" yaml:"name"`     // 指标名称
	Type   string            `json:"type" yaml:"type"`     // 指标类型
	Value  float64           `json:"value" yaml:"value"`   // 指标值
	Labels map[string]string `json:"labels" yaml:"labels"` // 标签
}

// MetricsReport 指标报告
type MetricsReport struct {
	ID        string   `json:"id" yaml:"id"`               // 节点ID
	Timestamp int64    `json:"timestamp" yaml:"timestamp"` // 时间戳
	Metrics   []Metric `json:"metrics" yaml:"metrics"`     // 指标列表
}

// GameNodeStatus 节点状态信息
type GameNodeStatus struct {
	State      GameNodeState `json:"state" yaml:"state"`             // 节点状态
	Online     bool          `json:"online" yaml:"online"`           // 是否在线
	LastOnline time.Time     `json:"last_online" yaml:"last_online"` // 最后在线时间
	UpdatedAt  time.Time     `json:"updated_at" yaml:"updated_at"`   // 状态更新时间
	Resource   ResourceInfo  `json:"resource" yaml:"resource"`       // 资源信息
	Metrics    MetricsReport `json:"metrics" yaml:"metrics"`         // 指标报告
}

// GameNode 游戏节点
type GameNode struct {
	ID        string            `json:"id" yaml:"id"`                 // 节点ID
	Alias     string            `json:"alias" yaml:"alias"`           // 节点别名
	Model     string            `json:"model" yaml:"model"`           // 节点型号
	Type      GameNodeType      `json:"type" yaml:"type"`             // 节点类型
	Location  string            `json:"location" yaml:"location"`     // 节点位置
	Labels    map[string]string `json:"labels" yaml:"labels"`         // 标签
	Status    GameNodeStatus    `json:"status" yaml:"status"`         // 节点状态信息
	CreatedAt time.Time         `json:"created_at" yaml:"created_at"` // 创建时间
	UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"` // 更新时间
}

// TableName 返回表名
func (GameNode) TableName() string {
	return "game_nodes"
}

// Resource 资源信息
type Resource struct {
	Total int64 `json:"total" yaml:"total"` // 总量
	Used  int64 `json:"used" yaml:"used"`   // 已使用量
}

// NodeResources 节点资源信息
type NodeResources struct {
	CPU    Resource `json:"cpu" yaml:"cpu"`       // CPU资源
	Memory Resource `json:"memory" yaml:"memory"` // 内存资源
	Disk   Resource `json:"disk" yaml:"disk"`     // 磁盘资源
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	RxBytesPerSec float32 `json:"rx_bytes_per_sec" yaml:"rx_bytes_per_sec"` // 接收速率
	TxBytesPerSec float32 `json:"tx_bytes_per_sec" yaml:"tx_bytes_per_sec"` // 发送速率
}

// NodeMetrics 节点指标
type NodeMetrics struct {
	CPU     float32        `json:"cpu" yaml:"cpu"`         // CPU使用率
	Memory  float32        `json:"memory" yaml:"memory"`   // 内存使用率
	Disk    float32        `json:"disk" yaml:"disk"`       // 磁盘使用率
	Network NetworkMetrics `json:"network" yaml:"network"` // 网络指标
}
