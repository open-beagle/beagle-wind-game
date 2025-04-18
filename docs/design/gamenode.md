# GameNode 系统设计文档

## 相关文档

- [GameNode 通信设计](grpc_communication.md)：详细描述了 GameNode 的通信机制和协议，包括 Pipeline 任务管理、执行和接口设计
- [GameNode Pipeline 设计](gamepipeline.md)：详细描述了 GameNode 的流水线执行机制

## 1. 系统概述

GameNode 是 Beagle Wind Game 平台的核心组件之一，负责管理和执行游戏节点上的各种任务。它由以下五个核心部分组成：

1. GameNode 模型：系统的核心基础设施，定义了游戏节点的基本结构和行为
2. GameNodeHandler：HTTP API 服务实体，处理前端交互
3. GameNodeService：核心业务实现，管理游戏节点状态和生命周期
4. GameNodeStore：数据存储管理器，负责数据持久化
5. GameNodeCommunication：节点通信设计，实现节点间通信，包括 Pipeline 任务管理、执行和接口设计，详见[通信设计文档](grpc_communication.md)

## 2. 核心组件设计

### 2.1 GameNode 模型

GameNode 模型是整个系统的基础，定义了游戏节点的核心属性和行为。模型分为静态属性和动态状态两部分：

```go
// GameNode 游戏节点
type GameNode struct {
    ID        string            `json:"id" yaml:"id"`                 // 节点ID
    Alias     string            `json:"alias" yaml:"alias"`           // 节点别名
    Model     string            `json:"model" yaml:"model"`           // 节点型号
    Type      GameNodeType      `json:"type" yaml:"type"`             // 节点类型
    Location  string            `json:"location" yaml:"location"`     // 节点位置
    Labels    map[string]string `json:"labels" yaml:"labels"`         // 标签
    State     GameNodeStaticState `json:"state" yaml:"state"`         // 节点维护状态
    Hardware  map[string]string `json:"hardware" yaml:"hardware"`     // 硬件配置(简化版)
    System    map[string]string `json:"system" yaml:"system"`         // 系统配置(简化版)
    Status    GameNodeStatus    `json:"status" yaml:"status"`         // 节点状态信息
    CreatedAt time.Time         `json:"created_at" yaml:"created_at"` // 创建时间
    UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"` // 更新时间
}

// GameNodeStaticState 节点维护状态
type GameNodeStaticState string

const (
    GameNodeStaticStateNormal    GameNodeStaticState = "normal"     // 正常状态
    GameNodeStaticStateMaintenance GameNodeStaticState = "maintenance" // 维护状态
    GameNodeStaticStateDisabled   GameNodeStaticState = "disabled"   // 禁用状态
)
```

#### 2.1.1 静态属性

- ID：节点唯一标识
- Alias：节点别名（用于显示）
- Model：节点型号（如 Beagle-Wind-2024）
- Type：节点类型（物理机/虚拟机/容器）
- Location：节点地理位置
- Labels：节点标签（用于分类和筛选）
- State：节点维护状态
  - normal：正常状态，节点可以正常处理所有业务
  - maintenance：维护状态，节点不处理业务，但保持心跳和状态报告
  - disabled：禁用状态，节点只响应心跳，不处理其他请求
- Hardware：硬件配置（简化版，键值对形式）
- System：系统配置（简化版，键值对形式）
- CreatedAt：创建时间
- UpdatedAt：更新时间

#### 2.1.2 维护状态说明

1. 正常状态（normal）：

   - 节点可以正常处理所有业务请求
   - 保持正常的心跳和状态报告
   - 可以接收和执行 Pipeline 任务
   - 可以响应资源查询和指标采集请求

2. 维护状态（maintenance）：

   - 节点不处理任何业务请求
   - 保持正常的心跳和状态报告
   - 不接收新的 Pipeline 任务
   - 可以响应资源查询和指标采集请求
   - 已运行的任务可以继续执行
   - 管理员可以执行维护操作

3. 禁用状态（disabled）：

   - 节点只响应心跳请求
   - 不处理任何业务请求
   - 不接收新的 Pipeline 任务
   - 不响应资源查询和指标采集请求
   - 已运行的任务会被终止
   - 管理员可以执行恢复操作

#### 2.1.3 状态转换规则

1. 正常状态转换：

   - normal -> maintenance：管理员手动设置
   - normal -> disabled：管理员手动设置
   - maintenance -> normal：管理员手动设置
   - maintenance -> disabled：管理员手动设置
   - disabled -> normal：管理员手动设置
   - disabled -> maintenance：不允许转换

2. 状态转换影响：

   - 转换为 maintenance：
     - 停止接收新的业务请求
     - 保持现有任务运行
     - 继续报告状态和指标
   - 转换为 disabled：
     - 停止所有业务处理
     - 终止正在运行的任务
     - 只保持心跳连接
   - 转换为 normal：
     - 恢复所有业务处理
     - 可以接收新的任务
     - 正常报告状态和指标

#### 2.1.4 动态状态（GameNodeStatus）

动态状态包含节点的实时运行状态和资源使用情况：

```go
type GameNodeStatus struct {
    State      GameNodeState `json:"state" yaml:"state"`             // 节点状态
    Online     bool          `json:"online" yaml:"online"`           // 是否在线
    LastOnline time.Time     `json:"last_online" yaml:"last_online"` // 最后在线时间
    UpdatedAt  time.Time     `json:"updated_at" yaml:"updated_at"`   // 状态更新时间
    Hardware   HardwareInfo  `json:"hardware" yaml:"hardware"`       // 硬件配置
    System     SystemInfo    `json:"system" yaml:"system"`           // 系统配置
    Metrics    MetricsInfo   `json:"metrics" yaml:"metrics"`         // 监控指标
}
```

1. 基本状态信息：

   - State：节点状态（ready/offline/online/maintenance/busy/error）
   - Online：是否在线
   - LastOnline：最后在线时间
   - UpdatedAt：状态更新时间

2. 详细硬件信息（HardwareInfo）：

   - CPUs：CPU 设备列表
     - Model：CPU 型号
     - Cores：物理核心数
     - Threads：线程数
     - Frequency：基准频率
     - Cache：缓存大小
     - Architecture：架构
   - Memories：内存设备列表
     - Size：内存条容量
     - Type：内存类型
     - Frequency：频率
   - GPUs：GPU 设备列表
     - Model：显卡型号
     - MemoryTotal：显存总量
     - Architecture：GPU 架构
     - DriverVersion：驱动版本
     - ComputeCapability：计算能力
     - TDP：功耗指标
   - Storages：存储设备列表
     - Type：存储类型(SSD/HDD/NVMe)
     - Model：设备型号
     - Capacity：总容量
     - Path：挂载路径
   - Networks：网络设备列表
     - Name：网卡名称
     - MacAddress：MAC 地址
     - IpAddress：IP 地址
     - Speed：网卡速率

3. 详细系统信息（SystemInfo）：

   - OSDistribution：操作系统发行版
   - OSVersion：操作系统版本
   - OSArchitecture：操作系统架构
   - KernelVersion：内核版本
   - GPUDriverVersion：GPU 驱动版本
   - GPUComputeAPIVersion：GPU 计算框架版本
   - DockerVersion：Docker 版本
   - ContainerdVersion：Containerd 版本
   - RuncVersion：Runc 版本

4. 监控指标（MetricsInfo）：
   - CPUs：CPU 指标列表
     - Model：CPU 型号
     - Cores：物理核心数
     - Threads：线程数
     - Usage：CPU 使用率
   - Memory：内存指标
     - Total：总容量
     - Available：可用内存
     - Used：已用内存
     - Usage：使用率
   - GPUs：GPU 指标列表
     - Model：GPU 型号
     - MemoryTotal：显存总量
     - Usage：GPU 使用率
     - MemoryUsed：已用显存
     - MemoryFree：可用显存
     - MemoryUsage：显存使用率
   - Storages：存储指标列表
     - Path：挂载路径
     - Type：存储类型
     - Model：设备型号
     - Capacity：总容量
     - Used：已用空间
     - Free：可用空间
     - Usage：使用率
   - Network：网络指标
     - InboundTraffic：流入流量
     - OutboundTraffic：流出流量
     - Connections：连接数

#### 2.1.5 数据采集

1. 静态属性：

   - 硬件和系统信息：由 Agent 自动采集和上报
   - 管理属性：由管理员通过管理界面维护
   - 变更需要记录审计日志

2. 动态状态：

   - 由系统自动采集
   - 定期更新（默认间隔 5 分钟）
   - 支持实时推送更新
   - 历史数据保留 30 天

3. 指标数据：
   - 支持多种采集方式（Prometheus、自定义采集器）
   - 支持自定义指标定义
   - 支持告警阈值设置
   - 支持数据聚合和统计

### 2.2 GameNodeHandler

GameNodeHandler 是系统的 HTTP API 服务实体，负责处理前端交互。详细设计请参考 [GameNodeHandler 设计文档](gamenode_handler.md)。

#### 2.2.1 核心职责

1. 提供节点基础信息的查询和管理
2. 提供 Pipeline 状态查询和取消功能

#### 2.2.2 非职责范围

1. 节点注册：由 GameNodeAgent 通过 gRPC 实现
2. 节点心跳：由 GameNodeAgent 通过 gRPC 实现
3. Pipeline 创建和执行：由其他组件负责
4. 容器管理：暂缓实现
5. 日志管理：暂缓实现
6. 事件流管理：由 Event Handler 负责，详见 [Event Handler 设计文档](event_handler.md)

### 2.3 GameNodeService

核心业务实现，负责游戏节点的生命周期管理：

#### 2.3.1 核心功能

- 节点生命周期管理
  - 创建和初始化
  - 状态维护
  - 资源调度
  - 任务执行
- 流水线管理
  - 任务编排
  - 状态跟踪
  - 错误处理
- 资源管理
  - CPU 分配
  - 内存管理
  - GPU 调度
  - 网络控制

#### 2.3.2 业务逻辑

- 节点注册和认证
- 状态同步和更新
- 资源分配和回收
- 任务调度和执行

### 2.4 GameNodeStore

数据存储管理器，负责数据的持久化。存储接口的设计和实现细节请参考：

- [GameNodeHandler 设计文档](gamenode_handler.md) 中的 GameNodeStore 设计
- [GamePipelineHandler 设计文档](gamepipeline_handler.md) 中的 GamePipelineStore 设计

### 2.5 GameNode Communication

节点通信设计，实现节点间的通信。详细设计请参考[通信设计文档](grpc_communication.md)：

#### 2.5.1 通信组件

- GameNodeServer：Pipeline 任务管理器
- GameNodeAgent：Pipeline 任务执行器
- GameNodeProto：Pipeline 任务接口设计
- GamePipeline：Pipeline 任务模板

#### 2.5.2 通信协议

详细设计请参考 [GameNodeServer 设计文档](gamenode_server.md) 中的 gRPC 服务接口部分。

#### 2.5.3 通信特性

- 双向流式通信
- 心跳机制
- 会话管理
- 错误处理

## 3. 生命周期管理

### 3.1 初始化创建

#### 3.1.1 Agent 注册创建

- GameNodeAgent 启动时通过 gRPC 向 GameNodeServer 注册
- 提供完整的节点信息：
  - 硬件信息（Hardware）：
    - CPU：型号、核心数、线程数、频率、缓存、架构
    - 内存：总量、类型、频率
    - GPU：型号、显存大小、架构、驱动版本、计算能力、功耗
    - 硬盘：型号、容量、类型、路径
    - 网络：网卡名称、MAC 地址、IP 地址、速率
  - 系统信息（System）：
    - 操作系统：发行版、版本号、架构
    - 内核版本：主版本、次版本、修订号
    - 显卡驱动：驱动版本、计算框架版本
    - 运行时环境：Docker 版本、Containerd 版本、Runc 版本

#### 3.1.2 数据存储

- 系统自动创建节点记录
- 使用 YAML 文件存储节点信息
- 存储在 `data/gamenodes.yaml` 中

### 3.2 管理员维护

#### 3.2.1 静态信息维护

- 通过 GameNodeHandler 的 HTTP API 接口更新
- 可更新的信息包括：
  - Alias：节点别名（用于显示）
  - Location：节点位置
  - Labels：节点标签
  - 其他管理属性

#### 3.2.2 状态管理

- 查看节点状态
- 监控资源使用
- 处理异常情况

### 3.3 运行维护

#### 3.3.1 状态维护

- 心跳检测：定期发送心跳包，支持重试机制
- 资源监控：
  - CPU 使用率
  - 内存使用情况
  - GPU 状态和使用率
  - 磁盘使用情况
  - 网络状态和流量

#### 3.3.2 任务执行

- Pipeline 任务管理
- 容器生命周期管理
- 资源调度和分配

#### 3.3.3 数据同步

- 状态更新
- 指标上报
- 日志收集

### 3.4 终止处理

#### 3.4.1 正常终止

- 通过 API 接口删除节点
- 清理相关资源
- 保存历史数据

#### 3.4.2 异常终止

- 节点离线处理
- 资源回收
- 状态清理
