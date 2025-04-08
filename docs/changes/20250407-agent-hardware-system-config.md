# Agent 硬件和系统配置优化

## 设计约束分析

### 1. GameNodeAgent 职责边界

1. 静态参数限制

   - GameNodeAgent 只关注 id 和 serveraddr
   - 不应在 Agent 中存储其他静态参数(alias, model, location 等)
   - 这些静态参数由管理员通过前端页面管理

2. 注册流程

   - 在节点注册(Register)时采集必要信息
   - 只采集可自动获取的信息(节点类型、硬件配置、系统配置、标签)
   - 不处理需要人工配置的信息(alias, model, location)

3. 参数维护

   - 注册完成后不再维护静态参数
   - 不向 Server 传递静态参数的变化
   - 专注于动态状态的采集和更新

4. 状态管理重点
   - 重点关注 GameNodeStatus 对象
   - 定时采集动态属性(CPU、内存、GPU 等使用率)
   - 通过心跳等机制向 Server 传递状态更新

### 2. 职责分离原则

1. 入口程序(cmd/agent/main.go)职责

   - 解析命令行参数
   - 创建和配置核心组件
   - 启动和停止服务
   - 处理信号和生命周期
   - 不包含具体业务逻辑实现

2. 业务模块(internal/gamenode/gamenode_agent.go)职责
   - 实现所有业务逻辑
   - 处理节点类型检测
   - 处理硬件和系统配置采集
   - 管理节点状态和资源监控
   - 处理注册、心跳和指标上报

## 变更目标调整

1. 优化代码结构和职责分配

   - 将业务逻辑从入口程序迁移至业务模块
   - 重构节点类型检测、硬件配置和系统配置采集等功能
   - 明确各组件职责边界

2. 优化注册阶段的信息采集

   - 规范化可自动获取的硬件配置
   - 规范化可自动获取的系统配置
   - 移除对静态配置的依赖

3. 完善状态监控
   - 优化资源使用率的采集
   - 规范化状态更新机制
   - 完善心跳数据结构

## 实现方案

### 1. 入口程序优化

重构 cmd/agent/main.go，移除业务逻辑:

```txt
main() {
    // 解析命令行参数
   // 初始化组件
   // 创建并启动Agent
   // 等待信号处理退出
}
```

### 2. 业务模块优化

修改 internal/gamenode/gamenode_agent.go，添加业务逻辑:

```txt
// 节点类型检测
DetectNodeType() {
   // 检查容器特征
   // 检查虚拟化特征
   // 返回节点类型
}

// 硬件配置获取
GetHardwareConfig() {
   // 获取CPU信息
   // 获取GPU信息
   // 获取内存信息
   // 获取存储信息
}

// 系统配置获取
GetSystemConfig() {
   // 获取OS信息
   // 获取GPU驱动信息
   // 获取CUDA版本
   // 获取网络信息
}
```

## 硬件信息采集改进

针对节点硬件信息采集，需要进行以下优化：

1. RAM（内存）显示格式优化

   - 修改前：`"RAM": "16382832 kB"`
   - 修改后：`"RAM": "16 GB"`
   - 修改位置：`internal/gamenode/gamenode_agent.go` 中的 `GetHardwareConfig()` 函数
   - 改进原因：
     - 硬件配置应反映标准化信息，而不是实际运行时的精确值
     - 实际硬件内存应该是标准容量，实际可用内存会在状态信息中准确反映
     - 避免在硬件规格中显示非标准值，保持专业性
   - 修改思路：
     1. 读取 `/proc/meminfo` 中的 `MemTotal` 值
     2. 将 kB 转换为 GB（除以 1024\*1024）
     3. 向上取整到最接近的整数 GB 值
     4. 格式化为"XX GB"的字符串

2. Storage（存储）信息完善

   - 当前问题：仅显示存储设备型号，缺少容量信息
   - 优化方案：结合设备型号和容量，格式为 "{设备型号} {容量}"（如 "Virtual Disk 500G"）
   - 具体改进：添加获取存储容量的逻辑，并与设备型号组合
   - 改进效果：
     - 修改前：`"Storage": "Virtual Disk"`
     - 修改后：`"Storage": "Virtual Disk 500G"` 或 `"Storage": "Virtual Disk 2TB"`
   - 修改位置：`internal/gamenode/gamenode_agent.go` 中的 `GetHardwareConfig()` 函数

3. GPU 驱动版本格式规范化

   - 当前问题：仅显示版本号（如 "572.83"）
   - 优化方案：添加厂商信息，格式为 "NVIDIA Driver {版本号}"（如 "NVIDIA Driver 572.83"）
   - 具体改进：在 GetSystemConfig 方法中修改 gpu_driver 的值格式
   - 改进效果：
     - 修改前：`"gpu_driver": "572.83"`
     - 修改后：`"gpu_driver": "NVIDIA Driver 572.83"`
   - 修改位置：`internal/gamenode/gamenode_agent.go` 中的 `GetSystemConfig()` 函数

4. CPU 和 GPU 插槽信息完善

   - 当前问题：CPU 和 GPU 信息中缺少插槽编号，多 CPU/GPU 系统无法准确识别每个组件位置
   - 优化方案：在静态硬件配置中添加插槽编号，格式为 "{插槽编号},{硬件详情}"
   - 具体改进：
     - CPU 信息：添加插槽编号，格式化为"0,Intel i7-9700K 2Core 3.6GHz 95W; 1,Intel i7-9700K 2Core 3.6GHz 95W"
     - GPU 信息：添加插槽编号，格式化为"0,NVIDIA GeForce RTX 4090 24GB 450W; 1,NVIDIA GeForce RTX 4090 24GB 450W"
   - 改进效果：
     - 修改前（CPU）：`"CPU": "Intel i7-9700K 2Core 3.6GHz 95W"`
     - 修改后（CPU）：`"CPU": "0,Intel i7-9700K 2Core 3.6GHz 95W; 1,Intel i7-9700K 2Core 3.6GHz 95W"`
     - 修改前（GPU）：`"GPU": "NVIDIA GeForce RTX 4090 24GB"`
     - 修改后（GPU）：`"GPU": "0,NVIDIA GeForce RTX 4090 24GB 450W; 1,NVIDIA GeForce RTX 4090 24GB 450W"`
   - 修改位置：`internal/gamenode/gamenode_agent.go` 中的 `GetHardwareConfig()` 函数
   - 前端调整：
     - 前端显示需要解析新格式，提取插槽编号和硬件详情进行分组显示
     - 前端编辑界面需添加插槽编号输入/选择功能
     - 列表视图应合并显示所有硬件，详情视图应分插槽展示

5. GPU 功耗(TDP)信息添加

   - 当前问题：GPU 信息中缺少功耗数据，无法准确评估电源需求和散热需求
   - 优化方案：在 GPU 信息中添加 TDP 数据，格式为"{型号} {显存}GB {TDP}W"
   - 具体改进：
     - 添加根据 GPU 型号估算 TDP 值的函数
     - 在 GPU 信息字符串和硬件信息结构中都添加 TDP 值
     - GPU 信息格式："{插槽号},{型号} {显存大小} {TDP 值}W"
   - 改进效果：
     - 修改前：`"GPU": "0,NVIDIA GeForce RTX 4090 24GB"`
     - 修改后：`"GPU": "0,NVIDIA GeForce RTX 4090 24GB 450W"`
   - 修改位置：
     - `internal/gamenode/gamenode_agent.go` 中的 `GetHardwareConfig()` 函数
     - `internal/models/gamenode.go` 中的 `HardwareInfo` 结构
   - 前端调整：
     - 前端显示需解析新格式，提取并显示功耗信息
     - 可增加功耗总计统计，帮助评估电源需求

6. Storage 信息修复

   - 当前问题：Storage 信息有时无法正确显示，导致节点没有存储设备信息
   - 优化方案：确保 Storage 信息始终包含至少一个默认项，修复格式错误
   - 具体改进：
     - 添加验证逻辑确保 Storage 信息正确存在
     - 对于缺失的情况，提供默认的根目录存储信息
     - 修复可能导致 Storage 信息未正确设置的逻辑
   - 改进效果：
     - 修改前：可能出现 `"Storage": ""` 或完全缺失
     - 修改后：至少包含根目录存储信息，如 `"Storage": "/,HDD 500GB"`
   - 修改位置：`internal/gamenode/gamenode_agent.go` 中的 `GetHardwareConfig()` 函数

7. 状态信息结构优化

   4.1 静态硬件配置信息（GameNode.Hardware）

   - 职责定位：反映节点的硬件配置信息，以简化、标准化、人类可读的形式呈现
   - 采集时机：Agent 启动时自动采集，注册时传递给服务器，后续由管理员维护修改
   - 存储位置：作为静态信息直接存储在 GameNode 结构中，而非状态中
   - 存储形式：使用键值对（map[string]string）简化表示，便于管理员查看和修改
   - 信息类型：

     - CPU：字符串形式，例如"0,Intel i9 13900k 4.2Ghz 16 核心 125W; 1,Intel i9 13900k 4.2Ghz 16 核心 125W"
     - 内存：字符串形式，例如"64 GB DDR5-4800"
     - GPU：字符串形式，例如"0,NVIDIA GeForce RTX 4090 24GB; 1,NVIDIA GeForce RTX 4090 24GB"
     - 存储：字符串形式，例如"Samsung 980 Pro 2TB NVMe SSD;WD Black 4TB HDD"

     4.2 静态系统配置信息（GameNode.System）

   - 职责定位：反映节点的操作系统和驱动等信息，以简化、标准化形式呈现
   - 采集时机：Agent 启动时自动采集，注册时传递给服务器，后续由管理员维护修改
   - 存储位置：作为静态信息直接存储在 GameNode 结构中，而非状态中
   - 存储形式：使用键值对（map[string]string）简化表示，便于管理员查看和修改
   - 信息类型：

     - os_distribution: 操作系统发行版，例如"Ubuntu"
     - os_version: 操作系统版本，例如"22.04 LTS"
     - os_architecture: 操作系统架构，例如"x86_64"
     - kernel_version: 内核版本，例如"5.15.0-58-generic"
     - gpu_driver_version: GPU 驱动版本，例如"NVIDIA Driver 535.129.03"
     - cuda_version: CUDA 版本，例如"12.2"

     4.3 动态硬件配置信息（GameNodeStatus.Hardware）

   - 职责定位：提供详细的硬件设备信息，以专业、结构化的形式呈现
   - 采集时机：Agent 启动时采集，与静态硬件配置同步采集，但存储路径不同
   - 存储位置：作为状态的一部分存储在 GameNodeStatus 结构中
   - 存储形式：使用结构化的 HardwareInfo 类型，包含详细的硬件参数
   - 信息类型：详细的 CPU、内存、GPU、存储设备列表，每个设备包含完整的技术参数

     4.4 动态系统配置信息（GameNodeStatus.System）

   - 职责定位：提供详细的系统软件信息，以专业、结构化的形式呈现
   - 采集时机：Agent 启动时采集，与静态系统配置同步采集，但存储路径不同
   - 存储位置：作为状态的一部分存储在 GameNodeStatus 结构中
   - 存储形式：使用结构化的 SystemInfo 类型，包含详细的系统参数
   - 信息类型：详细的操作系统信息、驱动版本、容器运行时版本等

     4.5 监控指标信息（GameNodeStatus.Metrics）

   - 职责定位：反映节点的实时运行状态及相关硬件参数
   - 采集时机：定期采集更新（如每 30 秒）
   - 存储位置：作为动态信息存储在 GameNodeStatus 结构中
   - 存储形式：使用结构化的 MetricsInfo 类型，包含实时监控数据
   - 信息类型：
     - CPU 使用率、温度等
     - 内存使用情况
     - GPU 使用率、温度、功耗等
     - 存储设备使用情况、读写速度
     - 网络流量、连接数

这种结构优化将使节点信息组织更加合理：

1. 静态硬件配置(GameNode.Hardware)和系统配置(GameNode.System)：

   - 简化为键值对形式，便于管理员查看和编辑
   - 强调人类可读性，使用标准化、简洁的表述
   - 注册后由管理员维护，不再自动更新

2. 动态硬件信息(GameNodeStatus.Hardware)和系统信息(GameNodeStatus.System)：

   - 使用结构化的详细数据，体现专业性
   - 包含完整的技术参数，支持深入分析和监控
   - 仍保持自动采集，但不作为主要展示内容

3. 监控指标(GameNodeStatus.Metrics)：
   - 专注于实时变化的运行状态
   - 定期自动更新，反映最新状态
   - 作为性能监控和报警的数据来源

## 具体实现步骤

### 1. GameNode 模型改进

将在 `internal/models/gamenode.go` 中重构数据模型：

```go
// HardwareInfo 硬件信息
type HardwareInfo struct {
	CPU struct {
		Devices []struct {
			Model     string  `json:"model" yaml:"model"`         // CPU型号
			Cores     int32   `json:"cores" yaml:"cores"`         // 物理核心数
			Threads   int32   `json:"threads" yaml:"threads"`     // 线程数
			Frequency float64 `json:"frequency" yaml:"frequency"` // 基准频率
			Cache     int64   `json:"cache" yaml:"cache"`         // 缓存大小
		} `json:"devices" yaml:"devices"`
	} `json:"cpu" yaml:"cpu"`

	Memory struct {
		Devices []struct {
			Size        int64   `json:"size" yaml:"size"`               // 内存条容量
			Type        string  `json:"type" yaml:"type"`               // 内存类型
			Frequency   float64 `json:"frequency" yaml:"frequency"`     // 频率
			Manufacturer string  `json:"manufacturer" yaml:"manufacturer"` // 制造商
			Serial      string  `json:"serial" yaml:"serial"`           // 序列号
			Slot        string  `json:"slot" yaml:"slot"`               // 插槽位置
		} `json:"devices" yaml:"devices"`
	} `json:"memory" yaml:"memory"`

	GPU struct {
		Devices []struct {
			Model       string `json:"model" yaml:"model"`             // 显卡型号
			MemoryTotal int64  `json:"memory_total" yaml:"memory_total"` // 显存总量
			CudaCores   int32  `json:"cuda_cores" yaml:"cuda_cores"`   // CUDA核心数
		} `json:"devices" yaml:"devices"`
	} `json:"gpu" yaml:"gpu"`

	Storage struct {
		Devices []struct {
			Type     string `json:"type" yaml:"type"`         // 存储类型
			Capacity int64  `json:"capacity" yaml:"capacity"` // 总容量
		} `json:"devices" yaml:"devices"`
	} `json:"storage" yaml:"storage"`
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
		Available int64   `json:"available" yaml:"available"` // 可用内存
		Used      int64   `json:"used" yaml:"used"`           // 已用内存
		Usage     float64 `json:"usage" yaml:"usage"`         // 使用率
	} `json:"memory" yaml:"memory"`

	GPU struct {
		Devices []struct {
			Model       string  `json:"model" yaml:"model"`             // 显卡型号
			MemoryTotal int64   `json:"memory_total" yaml:"memory_total"` // 显存总量
			Usage       float64 `json:"usage" yaml:"usage"`             // GPU使用率
			MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`   // 已用显存
			MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`   // 可用显存
			MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"` // 显存使用率
			Temperature float64 `json:"temperature" yaml:"temperature"`   // GPU温度
			Power       float64 `json:"power" yaml:"power"`             // 功耗
		} `json:"devices" yaml:"devices"`
	} `json:"gpu" yaml:"gpu"`

	Storage struct {
		Devices []struct {
			Path       string  `json:"path" yaml:"path"`             // 挂载路径
			Type       string  `json:"type" yaml:"type"`             // 存储类型
			Model      string  `json:"model" yaml:"model"`           // 设备型号
			Capacity   int64   `json:"capacity" yaml:"capacity"`     // 总容量
			Used       int64   `json:"used" yaml:"used"`             // 已用空间
			Free       int64   `json:"free" yaml:"free"`             // 可用空间
			Usage      float64 `json:"usage" yaml:"usage"`           // 使用率
			ReadSpeed  float64 `json:"read_speed" yaml:"read_speed"` // 读取速度
			WriteSpeed float64 `json:"write_speed" yaml:"write_speed"` // 写入速度
		} `json:"devices" yaml:"devices"`
	} `json:"storage" yaml:"storage"`

	Network struct {
		Bandwidth    float64 `json:"bandwidth" yaml:"bandwidth"`       // 带宽使用
		Latency      float64 `json:"latency" yaml:"latency"`         // 延迟
		Connections  int32   `json:"connections" yaml:"connections"` // 连接数
		PacketLoss   float64 `json:"packet_loss" yaml:"packet_loss"` // 丢包率
	} `json:"network" yaml:"network"`
}

// GameNodeStatus 节点状态信息
type GameNodeStatus struct {
	State      GameNodeState `json:"state" yaml:"state"`             // 节点状态
	Online     bool          `json:"online" yaml:"online"`           // 是否在线
	LastOnline time.Time     `json:"last_online" yaml:"last_online"` // 最后在线时间
	UpdatedAt  time.Time     `json:"updated_at" yaml:"updated_at"`   // 状态更新时间
	Hardware   HardwareInfo  `json:"hardware" yaml:"hardware"`       // 硬件配置(详细版)
	System     SystemInfo    `json:"system" yaml:"system"`           // 系统配置(详细版)
	Metrics    MetricsInfo   `json:"metrics" yaml:"metrics"`         // 监控指标
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
```

### 2. GameNodeAgent 硬件信息采集改进

在 `internal/gamenode/gamenode_agent.go` 中修改 `GetHardwareConfig` 方法：

```go
// GetHardwareConfig 获取硬件配置信息
func (agent *GameNodeAgent) GetHardwareConfig() (map[string]string, HardwareInfo, error) {
	// 简化版硬件配置（用于兼容现有接口）
    config := make(map[string]string)

	// 完整硬件信息结构
	var hardwareInfo HardwareInfo

	// 1. 获取CPU信息
	cpuModel, cpuCores, cpuThreads, err := agent.getCPUInfo()
	if err != nil {
		log.Printf("获取CPU信息失败: %v", err)
	} else {
		// 设置简化版配置
		config["CPU"] = cpuModel

		// 设置详细硬件信息
		hardwareInfo.CPU.Devices = append(hardwareInfo.CPU.Devices, struct {
			Model     string  `json:"model" yaml:"model"`
			Cores     int32   `json:"cores" yaml:"cores"`
			Threads   int32   `json:"threads" yaml:"threads"`
			Frequency float64 `json:"frequency" yaml:"frequency"`
			Cache     int64   `json:"cache" yaml:"cache"`
		}{
			Model:     cpuModel,
			Cores:     int32(cpuCores),
			Threads:   int32(cpuThreads),
			Frequency: float64(cpuThreads) * cpuCores,
			Cache:     cpuCores * cpuThreads * cpuCores,
		})
	}

	// 2. 获取内存信息
	memDevices, err := agent.getMemoryDevices()
	if err != nil {
		log.Printf("获取内存信息失败: %v", err)
	} else {
		// 计算总内存容量
		var totalMemGB int64
		for _, device := range memDevices {
			totalMemGB += device.Size
		}

		// 设置简化版配置
		config["RAM"] = fmt.Sprintf("%d GB", totalMemGB)

		// 设置详细硬件信息
		for _, device := range memDevices {
			hardwareInfo.Memory.Devices = append(hardwareInfo.Memory.Devices, struct {
				Size        int64   `json:"size" yaml:"size"`
				Type        string  `json:"type" yaml:"type"`
				Frequency   float64 `json:"frequency" yaml:"frequency"`
				Manufacturer string  `json:"manufacturer" yaml:"manufacturer"`
				Serial      string  `json:"serial" yaml:"serial"`
				Slot        string  `json:"slot" yaml:"slot"`
			}{
				Size:        device.Size,
				Type:        device.Type,
				Frequency:   device.Frequency,
				Manufacturer: device.Manufacturer,
				Serial:      device.Serial,
				Slot:        device.Slot,
			})
		}
	}

	// 3. 获取GPU信息
	gpuModels, gpuMemory, err := agent.getGPUInfo()
	if err != nil {
		log.Printf("获取GPU信息失败: %v", err)
	} else {
		// 设置简化版配置
		config["GPU"] = strings.Join(gpuModels, ",")

		// 设置详细硬件信息
		for _, gpuModel := range gpuModels {
			hardwareInfo.GPU.Devices = append(hardwareInfo.GPU.Devices, struct {
				Model       string `json:"model" yaml:"model"`
				MemoryTotal int64  `json:"memory_total" yaml:"memory_total"`
				CudaCores   int32  `json:"cuda_cores" yaml:"cuda_cores"`
			}{
				Model:       gpuModel,
				MemoryTotal: gpuMemory,
			})
		}
	}

	// 4. 获取存储信息
	storageDevices, err := agent.getStorageInfo()
	if err != nil {
		log.Printf("获取存储信息失败: %v", err)
	} else {
		// 设置详细硬件信息
		for _, device := range storageDevices {
			hardwareInfo.Storage.Devices = append(hardwareInfo.Storage.Devices, struct {
				Type     string `json:"type" yaml:"type"`
				Capacity int64  `json:"capacity" yaml:"capacity"`
			}{
				Type:     device.Type,
				Capacity: device.Capacity,
			})
		}

		// 对于简化版配置，合并所有存储设备信息
		var storageInfos []string
		for _, device := range storageDevices {
			// 根据容量格式化大小字符串
			capacityStr := formatStorageSize(device.Capacity)
			storageInfos = append(storageInfos, fmt.Sprintf("%s %s", device.Model, capacityStr))
		}
		config["Storage"] = strings.Join(storageInfos, ", ")
	}

	return config, hardwareInfo, nil
}
```

### 3. 系统配置信息采集改进

在 `internal/gamenode/gamenode_agent.go` 中添加 `GetSystemConfig` 方法：

```go
// GetSystemConfig 获取系统配置信息
func (agent *GameNodeAgent) GetSystemConfig() (map[string]string, SystemInfo, error) {
	// 简化版系统配置（用于兼容现有接口）
    config := make(map[string]string)

	// 完整系统信息结构
	var systemInfo SystemInfo

	// 1. 获取操作系统信息
	osDistribution, osVersion, osArch, kernelVersion, err := agent.getOSInfo()
	if err != nil {
		log.Printf("获取操作系统信息失败: %v", err)
	} else {
		// 设置简化版配置
		config["os_type"] = osDistribution
		config["os_version"] = osVersion

		// 设置详细系统信息
		systemInfo.OSDistribution = osDistribution
		systemInfo.OSVersion = osVersion
		systemInfo.OSArchitecture = osArch
		systemInfo.KernelVersion = kernelVersion
	}

	// 2. 获取GPU驱动信息
	gpuDriverVersion, err := agent.getGPUDriverInfo()
	if err != nil {
		log.Printf("获取GPU驱动信息失败: %v", err)
	} else {
		// 格式化GPU驱动版本
		formattedDriverVersion := fmt.Sprintf("NVIDIA Driver %s", gpuDriverVersion)

		// 设置简化版配置
		config["gpu_driver"] = formattedDriverVersion

		// 设置详细系统信息
		systemInfo.GPUDriverVersion = formattedDriverVersion
	}

	// 3. 获取CUDA版本
	cudaVersion, err := agent.getCUDAInfo()
	if err != nil {
		log.Printf("获取CUDA信息失败: %v", err)
	} else {
		// 设置简化版配置
		config["cuda_version"] = cudaVersion

		// 设置详细系统信息
		systemInfo.CUDAVersion = cudaVersion
	}

	// 4. 获取容器运行时信息
	dockerVersion, containerdVersion, runcVersion, err := agent.getContainerRuntimeInfo()
	if err != nil {
		log.Printf("获取容器运行时信息失败: %v", err)
	} else {
		// 设置详细系统信息
		systemInfo.DockerVersion = dockerVersion
		systemInfo.ContainerdVersion = containerdVersion
		systemInfo.RuncVersion = runcVersion
	}

	return config, systemInfo, nil
}
```

### 4. 监控指标采集改进

在 `internal/gamenode/gamenode_agent.go` 中添加 `CollectMetrics` 方法：

```go
// CollectMetrics 采集监控指标
func (agent *GameNodeAgent) CollectMetrics() (MetricsInfo, error) {
	var metricsInfo MetricsInfo

	// 1. 采集CPU使用率和温度
	cpuUsages, cpuTemps, err := agent.collectCPUMetrics()
	if err != nil {
		log.Printf("采集CPU指标失败: %v", err)
	} else {
		// 获取硬件信息中的CPU配置
		for i, cpuDevice := range agent.hardwareInfo.CPU.Devices {
			if i < len(cpuUsages) && i < len(cpuTemps) {
				metricsInfo.CPU.Devices = append(metricsInfo.CPU.Devices, struct {
					Model       string  `json:"model" yaml:"model"`
					Cores       int32   `json:"cores" yaml:"cores"`
					Threads     int32   `json:"threads" yaml:"threads"`
					Usage       float64 `json:"usage" yaml:"usage"`
					Temperature float64 `json:"temperature" yaml:"temperature"`
				}{
					Model:       cpuDevice.Model,
					Cores:       cpuDevice.Cores,
					Threads:     cpuDevice.Threads,
					Usage:       cpuUsages[i],
					Temperature: cpuTemps[i],
				})
			}
		}
	}

	// 2. 采集内存使用情况
	memAvailable, memUsed, memUsage, err := agent.collectMemoryMetrics()
	if err != nil {
		log.Printf("采集内存指标失败: %v", err)
	} else {
		metricsInfo.Memory.Available = memAvailable
		metricsInfo.Memory.Used = memUsed
		metricsInfo.Memory.Usage = memUsage
	}

	// 3. 采集GPU使用情况
	gpuDevicesInfo, err := agent.collectGPUMetrics()
	if err != nil {
		log.Printf("采集GPU指标失败: %v", err)
	} else {
		// 遍历所有GPU设备进行指标采集
		for _, deviceInfo := range gpuDevicesInfo {
			metricsInfo.GPU.Devices = append(metricsInfo.GPU.Devices, struct {
				Model       string  `json:"model" yaml:"model"`
				MemoryTotal int64   `json:"memory_total" yaml:"memory_total"`
				Usage       float64 `json:"usage" yaml:"usage"`
				MemoryUsed  int64   `json:"memory_used" yaml:"memory_used"`
				MemoryFree  int64   `json:"memory_free" yaml:"memory_free"`
				MemoryUsage float64 `json:"memory_usage" yaml:"memory_usage"`
				Temperature float64 `json:"temperature" yaml:"temperature"`
				Power       float64 `json:"power" yaml:"power"`
			}{
				Model:       deviceInfo.Model,
				MemoryTotal: deviceInfo.MemoryTotal,
				Usage:       deviceInfo.Usage,
				MemoryUsed:  deviceInfo.MemoryUsed,
				MemoryFree:  deviceInfo.MemoryFree,
				MemoryUsage: deviceInfo.MemoryUsage,
				Temperature: deviceInfo.Temperature,
				Power:       deviceInfo.Power,
			})
		}
	}

	// 4. 采集存储使用情况
	storageDevicesInfo, err := agent.collectStorageMetrics()
	if err != nil {
		log.Printf("采集存储指标失败: %v", err)
	} else {
		// 遍历所有存储设备进行指标采集
		for _, deviceInfo := range storageDevicesInfo {
			metricsInfo.Storage.Devices = append(metricsInfo.Storage.Devices, struct {
				Path       string  `json:"path" yaml:"path"`
				Type       string  `json:"type" yaml:"type"`
				Model      string  `json:"model" yaml:"model"`
				Capacity   int64   `json:"capacity" yaml:"capacity"`
				Used       int64   `json:"used" yaml:"used"`
				Free       int64   `json:"free" yaml:"free"`
				Usage      float64 `json:"usage" yaml:"usage"`
				ReadSpeed  float64 `json:"read_speed" yaml:"read_speed"`
				WriteSpeed float64 `json:"write_speed" yaml:"write_speed"`
			}{
				Path:       deviceInfo.Path,
				Type:       deviceInfo.Type,
				Model:      deviceInfo.Model,
				Capacity:   deviceInfo.Capacity,
				Used:       deviceInfo.Used,
				Free:       deviceInfo.Free,
				Usage:      deviceInfo.Usage,
				ReadSpeed:  deviceInfo.ReadSpeed,
				WriteSpeed: deviceInfo.WriteSpeed,
			})
		}
	}

	// 5. 采集网络状况
	netBandwidth, netLatency, netConnections, netPacketLoss, err := agent.collectNetworkMetrics()
	if err != nil {
		log.Printf("采集网络指标失败: %v", err)
	} else {
		metricsInfo.Network.Bandwidth = netBandwidth
		metricsInfo.Network.Latency = netLatency
		metricsInfo.Network.Connections = netConnections
		metricsInfo.Network.PacketLoss = netPacketLoss
	}

	return metricsInfo, nil
}
```

## 影响分析

1. 代码结构影响

   - 更清晰的职责分离
   - 入口程序专注于启动和配置
   - 业务逻辑集中在业务模块

2. 功能影响

   - 业务逻辑更加内聚
   - 更便于测试和维护
   - 提高代码可读性

3. 性能影响

   - 无明显性能影响
   - 代码结构优化不影响运行时行为

4. 数据模型影响
   - 更清晰地区分了静态硬件信息和动态监控指标
   - 提高了数据结构的合理性和可维护性
   - 有利于未来的功能扩展和性能优化

## 测试计划

1. 单元测试

   - 测试节点类型检测逻辑
   - 测试硬件和系统配置采集
   - 测试资源监控功能

2. 集成测试

   - 测试节点注册流程
   - 测试心跳和状态更新
   - 测试与服务器的交互

3. 端到端测试

   - 测试完整的节点生命周期
   - 测试异常情况处理
   - 测试资源监控准确性

## 后续优化

1. 完善资源监控

   - 增加更多资源指标
   - 优化采集性能
   - 添加异常检测

2. 改进错误处理

   - 添加更详细的错误日志
   - 实现错误重试机制
   - 优化异常情况下的行为

3. 指标聚合与分析

   - 实现历史指标存储和趋势分析
   - 添加阈值告警机制
   - 提供资源使用预测功能

4. 可视化改进

   - 优化前端展示逻辑
   - 提供更丰富的图表分析
   - 支持自定义监控面板

## 总结

本次 Agent 硬件和系统配置优化是对游戏节点管理的一次重要升级，通过重构数据模型和优化采集逻辑，使系统具备了更清晰的职责划分和更合理的数据结构。这些改进将直接提升系统的可维护性和可扩展性，为未来的功能迭代奠定坚实基础。

同时，此次优化与 GameNode 模型重构紧密结合，共同构成了对游戏节点核心功能的全面升级。这两部分变更相辅相成，共同推动系统向更加专业和高效的方向发展。
