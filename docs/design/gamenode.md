# GameNode 系统设计文档

## 相关文档

- [GameNode 通信设计](gamenode_communication.md)：详细描述了 GameNode 的通信机制和协议，包括 Pipeline 任务管理、执行和接口设计
- [GameNode Pipeline 设计](gamenode_pipeline.md)：详细描述了 GameNode 的流水线执行机制

## 1. 系统概述

GameNode 是 Beagle Wind Game 平台的核心组件之一，负责管理和执行游戏节点上的各种任务。它由以下五个核心部分组成：

1. GameNode 模型：系统的核心基础设施，定义了游戏节点的基本结构和行为
2. GameNodeHandler：HTTP API 服务实体，处理前端交互
3. GameNodeService：核心业务实现，管理游戏节点状态和生命周期
4. GameNodeStore：数据存储管理器，负责数据持久化
5. GameNodeCommunication：节点通信设计，实现节点间通信，包括 Pipeline 任务管理、执行和接口设计，详见[通信设计文档](gamenode_communication.md)

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
    Hardware  map[string]string `json:"hardware" yaml:"hardware"`     // 硬件配置
    System    map[string]string `json:"system" yaml:"system"`         // 系统配置
    Status    GameNodeStatus    `json:"status" yaml:"status"`         // 节点状态信息
    CreatedAt time.Time         `json:"created_at" yaml:"created_at"` // 创建时间
    UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"` // 更新时间
}
```

#### 2.1.1 静态属性

- ID：节点唯一标识
- Alias：节点别名（用于显示）
- Model：节点型号（如 Beagle-Wind-2024）
- Type：节点类型（物理机/虚拟机）
- Location：节点地理位置
- Labels：节点标签（用于分类和筛选）
- Hardware：硬件配置（CPU、内存、GPU、硬盘）
- System：系统配置（操作系统、内核版本、IP）
- CreatedAt：创建时间
- UpdatedAt：更新时间

#### 2.1.2 动态状态（GameNodeStatus）

动态状态包含节点的实时运行状态和资源使用情况：

1. 基本状态信息：

   - State：节点状态（ready/offline/online/maintenance/busy/error）
   - Online：是否在线
   - LastOnline：最后在线时间
   - UpdatedAt：状态更新时间

2. 资源信息（ResourceInfo）：

   - 硬件信息（CPU、内存、GPU、磁盘）
   - 软件信息（操作系统、驱动等）
   - 网络信息（带宽、延迟等）

3. 指标报告（MetricsReport）：
   - 系统指标（CPU、内存、GPU、磁盘使用率）
   - 网络指标（带宽使用、连接数等）
   - 自定义指标

#### 2.1.3 数据采集

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

#### 2.2.3 接口设计原则

1. RESTful 规范

   - 使用 HTTP 方法表示操作
   - 使用 URL 表示资源
   - 使用状态码表示结果

2. 安全性

   - 所有接口必须认证
   - 敏感数据加密传输
   - 防止 CSRF 攻击

3. 可用性

   - 接口幂等性
   - 合理的超时设置
   - 优雅的错误处理

4. 可维护性

   - 清晰的接口命名
   - 统一的响应格式
   - 完整的接口文档

5. 性能
   - 合理的缓存策略
   - 分页查询支持
   - 流式响应支持

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
- [GameNodePipelineHandler 设计文档](gamenode_pipeline_handler.md) 中的 GameNodePipelineStore 设计

### 2.5 GameNode Communication

节点通信设计，实现节点间的通信。详细设计请参考[通信设计文档](gamenode_communication.md)：

#### 2.5.1 通信组件

- GameNodeServer：Pipeline 任务管理器
- GameNodeAgent：Pipeline 任务执行器
- GameNodeProto：Pipeline 任务接口设计
- GameNodePipeline：Pipeline 任务模板

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
    - CPU：型号、核心数、线程数、频率、缓存
    - 内存：总量、类型、频率、通道数
    - GPU：型号、显存大小、CUDA 核心数
    - 硬盘：型号、容量、类型、接口
  - 系统信息（System）：
    - 操作系统：发行版、版本号、架构
    - 内核版本：主版本、次版本、修订号
    - 显卡驱动：驱动版本、CUDA 版本
    - 运行时环境：Docker 版本、Containerd 版本、Runc 版本

#### 3.1.2 数据存储

- 系统自动创建节点记录
- 使用 YAML 文件存储节点信息
- 存储在 `data/game-nodes.yaml` 中

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

- 心跳检测：定期发送心跳包
- 资源监控：
  - CPU 使用率、温度
  - 内存使用情况
  - GPU 状态
  - 磁盘使用情况
  - 网络状态

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
