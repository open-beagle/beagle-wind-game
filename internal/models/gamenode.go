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
	CPU struct {
		Devices []struct {
			Model        string  `json:"model" yaml:"model"`               // CPU型号
			Cores        int32   `json:"cores" yaml:"cores"`               // 物理核心数
			Threads      int32   `json:"threads" yaml:"threads"`           // 线程数
			Frequency    float64 `json:"frequency" yaml:"frequency"`       // 基准频率
			Cache        int64   `json:"cache" yaml:"cache"`               // 缓存大小
			Socket       string  `json:"socket" yaml:"socket"`             // CPU插槽
			Manufacturer string  `json:"manufacturer" yaml:"manufacturer"` // 制造商
			Architecture string  `json:"architecture" yaml:"architecture"` // 架构
		} `json:"devices" yaml:"devices"`
	} `json:"cpu" yaml:"cpu"`

	Memory struct {
		Devices []struct {
			Size         int64   `json:"size" yaml:"size"`                 // 内存条容量
			Type         string  `json:"type" yaml:"type"`                 // 内存类型
			Frequency    float64 `json:"frequency" yaml:"frequency"`       // 频率
			Manufacturer string  `json:"manufacturer" yaml:"manufacturer"` // 制造商
			Serial       string  `json:"serial" yaml:"serial"`             // 序列号
			Slot         string  `json:"slot" yaml:"slot"`                 // 插槽位置
			PartNumber   string  `json:"part_number" yaml:"part_number"`   // 部件号
			FormFactor   string  `json:"form_factor" yaml:"form_factor"`   // 内存形式
		} `json:"devices" yaml:"devices"`
	} `json:"memory" yaml:"memory"`

	GPU struct {
		Devices []struct {
			Model        string `json:"model" yaml:"model"`               // 显卡型号
			MemoryTotal  int64  `json:"memory_total" yaml:"memory_total"` // 显存总量
			CudaCores    int32  `json:"cuda_cores" yaml:"cuda_cores"`     // CUDA核心数
			Manufacturer string `json:"manufacturer" yaml:"manufacturer"` // 制造商
			Bus          string `json:"bus" yaml:"bus"`                   // 总线类型
			PciSlot      string `json:"pci_slot" yaml:"pci_slot"`         // PCI插槽
			Serial       string `json:"serial" yaml:"serial"`             // 序列号
			Architecture string `json:"architecture" yaml:"architecture"` // GPU架构
			TDP          int32  `json:"tdp" yaml:"tdp"`                   // 功耗指标(W)
		} `json:"devices" yaml:"devices"`
	} `json:"gpu" yaml:"gpu"`

	Storage struct {
		Devices []struct {
			Type         string `json:"type" yaml:"type"`                 // 存储类型(SSD/HDD/NVMe)
			Model        string `json:"model" yaml:"model"`               // 设备型号
			Capacity     int64  `json:"capacity" yaml:"capacity"`         // 总容量
			Path         string `json:"path" yaml:"path"`                 // 挂载路径
			Serial       string `json:"serial" yaml:"serial"`             // 序列号
			Interface    string `json:"interface" yaml:"interface"`       // 接口类型(SATA/NVMe/SAS)
			Manufacturer string `json:"manufacturer" yaml:"manufacturer"` // 制造商
			FormFactor   string `json:"form_factor" yaml:"form_factor"`   // 尺寸规格(2.5"/3.5"/M.2)
			Firmware     string `json:"firmware" yaml:"firmware"`         // 固件版本
		} `json:"devices" yaml:"devices"`
	} `json:"storage" yaml:"storage"`

	Network struct {
		Devices []struct {
			Name         string `json:"name" yaml:"name"`                 // 网卡名称
			Model        string `json:"model" yaml:"model"`               // 网卡型号
			MacAddress   string `json:"mac_address" yaml:"mac_address"`   // MAC地址
			IpAddress    string `json:"ip_address" yaml:"ip_address"`     // IP地址
			Speed        int64  `json:"speed" yaml:"speed"`               // 网卡速率(Mbps)
			Duplex       string `json:"duplex" yaml:"duplex"`             // 双工模式
			Manufacturer string `json:"manufacturer" yaml:"manufacturer"` // 制造商
			Interface    string `json:"interface" yaml:"interface"`       // 接口类型
			PciSlot      string `json:"pci_slot" yaml:"pci_slot"`         // PCI插槽
		} `json:"devices" yaml:"devices"`
	} `json:"network" yaml:"network"`
}

// SystemInfo 系统信息
type SystemInfo struct {
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

// MetricsInfo 监控指标信息
type MetricsInfo struct {
	CPU struct {
		Devices []struct {
			Model       string  `json:"model" yaml:"model"`             // CPU型号
			Cores       int32   `json:"cores" yaml:"cores"`             // 物理核心数
			Threads     int32   `json:"threads" yaml:"threads"`         // 线程数
			Usage       float64 `json:"usage" yaml:"usage"`             // CPU使用率
			Temperature float64 `json:"temperature" yaml:"temperature"` // CPU温度
		} `json:"devices" yaml:"devices"`
	} `json:"cpu" yaml:"cpu"`

	Memory struct {
		Total     int64   `json:"total" yaml:"total"`         // 总容量
		Available int64   `json:"available" yaml:"available"` // 可用内存
		Used      int64   `json:"used" yaml:"used"`           // 已用内存
		Usage     float64 `json:"usage" yaml:"usage"`         // 使用率
	} `json:"memory" yaml:"memory"`

	GPU struct {
		Devices []struct {
			Model       string  `json:"model" yaml:"model"`               // GPU型号
			MemoryTotal int64   `json:"memory_total" yaml:"memory_total"` // 显存总量
			Usage       float64 `json:"usage" yaml:"usage"`               // GPU使用率
			MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`   // 已用显存
			MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`   // 可用显存
			MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"` // 显存使用率
			Temperature float64 `json:"temperature" yaml:"temperature"`   // GPU温度
			Power       float64 `json:"power" yaml:"power"`               // 功耗
		} `json:"devices" yaml:"devices"`
	} `json:"gpu" yaml:"gpu"`

	Storage struct {
		Devices []struct {
			Path       string  `json:"path" yaml:"path"`               // 挂载路径
			Type       string  `json:"type" yaml:"type"`               // 存储类型
			Model      string  `json:"model" yaml:"model"`             // 设备型号
			Capacity   int64   `json:"capacity" yaml:"capacity"`       // 总容量
			Used       int64   `json:"used" yaml:"used"`               // 已用空间
			Free       int64   `json:"free" yaml:"free"`               // 可用空间
			Usage      float64 `json:"usage" yaml:"usage"`             // 使用率
			ReadSpeed  float64 `json:"read_speed" yaml:"read_speed"`   // 读取速度
			WriteSpeed float64 `json:"write_speed" yaml:"write_speed"` // 写入速度
		} `json:"devices" yaml:"devices"`
	} `json:"storage" yaml:"storage"`

	Network struct {
		InboundTraffic  float64 `json:"inbound_traffic" yaml:"inbound_traffic"`   // 流入流量(Mbps)
		OutboundTraffic float64 `json:"outbound_traffic" yaml:"outbound_traffic"` // 流出流量(Mbps)
		Connections     int32   `json:"connections" yaml:"connections"`           // 连接数
	} `json:"network" yaml:"network"`
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

// ResourceInfo 资源信息
type ResourceInfo struct {
	ID        string       `json:"id" yaml:"id"`               // 节点ID
	Timestamp int64        `json:"timestamp" yaml:"timestamp"` // 时间戳
	Hardware  HardwareInfo `json:"hardware" yaml:"hardware"`   // 硬件信息
	Metrics   MetricsInfo  `json:"metrics" yaml:"metrics"`     // 监控指标
}

// GameNode 游戏节点
type GameNode struct {
	ID        string            `json:"id" yaml:"id"`                 // 节点ID
	Alias     string            `json:"alias" yaml:"alias"`           // 节点别名
	Model     string            `json:"model" yaml:"model"`           // 节点型号
	Type      GameNodeType      `json:"type" yaml:"type"`             // 节点类型
	Location  string            `json:"location" yaml:"location"`     // 节点位置
	Labels    map[string]string `json:"labels" yaml:"labels"`         // 标签
	Hardware  map[string]string `json:"hardware" yaml:"hardware"`     // 硬件配置(简化版)
	System    map[string]string `json:"system" yaml:"system"`         // 系统配置(简化版)
	Status    GameNodeStatus    `json:"status" yaml:"status"`         // 节点状态信息
	CreatedAt time.Time         `json:"created_at" yaml:"created_at"` // 创建时间
	UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"` // 更新时间
}
