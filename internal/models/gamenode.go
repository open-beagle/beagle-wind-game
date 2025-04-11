package models

import (
	"time"
)

// GameNodeType 游戏节点类型
type GameNodeType string

const (
	GameNodeTypePhysical  GameNodeType = "physical"  // 物理节点
	GameNodeTypeVirtual   GameNodeType = "virtual"   // 虚拟节点
	GameNodeTypeContainer GameNodeType = "container" // 容器节点
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

// GameNodeStaticState 节点维护状态
type GameNodeStaticState string

const (
	GameNodeStaticStateNormal      GameNodeStaticState = "normal"      // 正常状态
	GameNodeStaticStateMaintenance GameNodeStaticState = "maintenance" // 维护状态
	GameNodeStaticStateDisabled    GameNodeStaticState = "disabled"    // 禁用状态
)

// CPUDevice CPU设备信息
type CPUDevice struct {
	Model        string  `json:"model" yaml:"model"`               // CPU型号
	Cores        int32   `json:"cores" yaml:"cores"`               // 物理核心数
	Threads      int32   `json:"threads" yaml:"threads"`           // 线程数
	Frequency    float64 `json:"frequency" yaml:"frequency"`       // 基准频率
	Cache        int64   `json:"cache" yaml:"cache"`               // 缓存大小
	Architecture string  `json:"architecture" yaml:"architecture"` // 架构
}

// MemoryDevice 内存设备信息
type MemoryDevice struct {
	Size      int64   `json:"size" yaml:"size"`           // 内存条容量
	Type      string  `json:"type" yaml:"type"`           // 内存类型
	Frequency float64 `json:"frequency" yaml:"frequency"` // 频率
}

// GPUDevice GPU设备信息
type GPUDevice struct {
	Model             string `json:"model" yaml:"model"`                           // 显卡型号
	MemoryTotal       int64  `json:"memory_total" yaml:"memory_total"`             // 显存总量
	Architecture      string `json:"architecture" yaml:"architecture"`             // GPU架构
	DriverVersion     string `json:"driver_version" yaml:"driver_version"`         // 驱动版本
	ComputeCapability string `json:"compute_capability" yaml:"compute_capability"` // 计算能力
	TDP               int32  `json:"tdp" yaml:"tdp"`                               // 功耗指标(W)
}

// StorageDevice 存储设备信息
type StorageDevice struct {
	Type     string `json:"type" yaml:"type"`         // 存储类型(SSD/HDD/NVMe)
	Model    string `json:"model" yaml:"model"`       // 设备型号
	Capacity int64  `json:"capacity" yaml:"capacity"` // 总容量
	Path     string `json:"path" yaml:"path"`         // 挂载路径
}

// NetworkDevice 网络设备信息
type NetworkDevice struct {
	Name       string `json:"name" yaml:"name"`               // 网卡名称
	MacAddress string `json:"mac_address" yaml:"mac_address"` // MAC地址
	IpAddress  string `json:"ip_address" yaml:"ip_address"`   // IP地址
	Speed      int64  `json:"speed" yaml:"speed"`             // 网卡速率(Mbps)
}

// CPUMetrics CPU监控指标
type CPUMetrics struct {
	Model   string  `json:"model" yaml:"model"`     // CPU型号
	Cores   int32   `json:"cores" yaml:"cores"`     // 物理核心数
	Threads int32   `json:"threads" yaml:"threads"` // 线程数
	Usage   float64 `json:"usage" yaml:"usage"`     // CPU使用率
}

// MemoryMetrics 内存监控指标
type MemoryMetrics struct {
	Total     int64   `json:"total" yaml:"total"`         // 总容量
	Available int64   `json:"available" yaml:"available"` // 可用内存
	Used      int64   `json:"used" yaml:"used"`           // 已用内存
	Usage     float64 `json:"usage" yaml:"usage"`         // 使用率
}

// GPUMetrics GPU监控指标
type GPUMetrics struct {
	Model       string  `json:"model" yaml:"model"`               // GPU型号
	GPUUsage    float64 `json:"gpu_usage" yaml:"gpu_usage"`       // GPU使用率
	MemoryTotal int64   `json:"memory_total" yaml:"memory_total"` // 显存总量
	MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`   // 已用显存
	MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`   // 可用显存
	MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"` // 显存使用率
}

// StorageMetrics 存储设备监控指标
type StorageMetrics struct {
	Path     string  `json:"path" yaml:"path"`         // 挂载路径
	Type     string  `json:"type" yaml:"type"`         // 存储类型
	Model    string  `json:"model" yaml:"model"`       // 设备型号
	Capacity int64   `json:"capacity" yaml:"capacity"` // 总容量
	Used     int64   `json:"used" yaml:"used"`         // 已用空间
	Free     int64   `json:"free" yaml:"free"`         // 可用空间
	Usage    float64 `json:"usage" yaml:"usage"`       // 使用率
}

// NetworkMetrics 网络监控指标
type NetworkMetrics struct {
	InboundTraffic  float64 `json:"inbound_traffic" yaml:"inbound_traffic"`   // 流入流量(Mbps)
	OutboundTraffic float64 `json:"outbound_traffic" yaml:"outbound_traffic"` // 流出流量(Mbps)
	Connections     int32   `json:"connections" yaml:"connections"`           // 连接数
}

// HardwareInfo 硬件信息 - 重构为扁平结构
type HardwareInfo struct {
	CPUs     []CPUDevice     `json:"cpus" yaml:"cpus"`         // CPU设备列表
	Memories []MemoryDevice  `json:"memories" yaml:"memories"` // 内存设备列表
	GPUs     []GPUDevice     `json:"gpus" yaml:"gpus"`         // GPU设备列表
	Storages []StorageDevice `json:"storages" yaml:"storages"` // 存储设备列表
	Networks []NetworkDevice `json:"networks" yaml:"networks"` // 网络设备列表
}

// SystemInfo 系统信息
type SystemInfo struct {
	OSDistribution       string `json:"os_distribution" yaml:"os_distribution"`                 // 操作系统发行版
	OSVersion            string `json:"os_version" yaml:"os_version"`                           // 操作系统版本
	OSArchitecture       string `json:"os_architecture" yaml:"os_architecture"`                 // 操作系统架构
	KernelVersion        string `json:"kernel_version" yaml:"kernel_version"`                   // 内核版本
	GPUDriverVersion     string `json:"gpu_driver_version" yaml:"gpu_driver_version"`           // GPU驱动版本
	GPUComputeAPIVersion string `json:"gpu_compute_api_version" yaml:"gpu_compute_api_version"` // GPU计算框架版本(CUDA/ROCm/oneAPI/OpenCL)
	DockerVersion        string `json:"docker_version" yaml:"docker_version"`                   // Docker版本
	ContainerdVersion    string `json:"containerd_version" yaml:"containerd_version"`           // Containerd版本
	RuncVersion          string `json:"runc_version" yaml:"runc_version"`                       // Runc版本
}

// MetricsInfo 监控指标信息 - 重构为扁平结构
type MetricsInfo struct {
	CPUs     []CPUMetrics     `json:"cpus" yaml:"cpus"`         // CPU指标列表
	Memory   MemoryMetrics    `json:"memory" yaml:"memory"`     // 内存指标
	GPUs     []GPUMetrics     `json:"gpus" yaml:"gpus"`         // GPU指标列表
	Storages []StorageMetrics `json:"storages" yaml:"storages"` // 存储指标列表
	Network  NetworkMetrics   `json:"network" yaml:"network"`   // 网络指标
}

// GameNodeStatus 节点状态信息
type GameNodeStatus struct {
	State      GameNodeState `json:"state" yaml:"state"`             // 节点状态
	Online     bool          `json:"online" yaml:"online"`           // 是否在线
	LastOnline time.Time     `json:"last_online" yaml:"last_online"` // 最后在线时间
	UpdatedAt  time.Time     `json:"updated_at" yaml:"updated_at"`   // 状态更新时间
	Hardware   HardwareInfo  `json:"hardware" yaml:"hardware"`       // 硬件配置
	System     SystemInfo    `json:"system" yaml:"system"`           // 系统配置
	Metrics    MetricsInfo   `json:"metrics" yaml:"metrics"`         // 监控指标
}

// GameNode 游戏节点
type GameNode struct {
	ID        string              `json:"id" yaml:"id"`                 // 节点ID
	Alias     string              `json:"alias" yaml:"alias"`           // 节点别名
	Model     string              `json:"model" yaml:"model"`           // 节点型号
	Type      GameNodeType        `json:"type" yaml:"type"`             // 节点类型
	Location  string              `json:"location" yaml:"location"`     // 节点位置
	Labels    map[string]string   `json:"labels" yaml:"labels"`         // 标签
	State     GameNodeStaticState `json:"state" yaml:"state"`           // 节点维护状态
	Hardware  map[string]string   `json:"hardware" yaml:"hardware"`     // 硬件配置(简化版)
	System    map[string]string   `json:"system" yaml:"system"`         // 系统配置(简化版)
	Status    GameNodeStatus      `json:"status" yaml:"status"`         // 节点状态信息
	CreatedAt time.Time           `json:"created_at" yaml:"created_at"` // 创建时间
	UpdatedAt time.Time           `json:"updated_at" yaml:"updated_at"` // 更新时间
}
