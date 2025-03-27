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
    Name      string            `json:"name" yaml:"name"`             // 节点名称
    Model     string            `json:"model" yaml:"model"`           // 节点型号
    Type      GameNodeType      `json:"type" yaml:"type"`             // 节点类型
    Location  string            `json:"location" yaml:"location"`     // 节点位置
    Hardware  map[string]string `json:"hardware" yaml:"hardware"`     // 硬件配置
    Network   map[string]string `json:"network" yaml:"network"`       // 网络配置
    Labels    map[string]string `json:"labels" yaml:"labels"`         // 标签
    Status    GameNodeStatus    `json:"status" yaml:"status"`         // 节点状态信息
    CreatedAt time.Time         `json:"created_at" yaml:"created_at"` // 创建时间
    UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"` // 更新时间
}
```

#### 2.1.1 静态属性

- ID：节点唯一标识
- Name：节点名称
- Model：节点型号
- Type：节点类型（物理机/虚拟机）
- Location：节点地理位置
- Hardware：硬件配置（CPU、内存、GPU 等）
- Network：网络信息（IP、带宽等）
- Labels：节点标签
- CreatedAt：创建时间
- UpdatedAt：更新时间

#### 2.1.2 状态信息（GameNodeStatus）

- State：节点状态（offline/online/ready/busy/error）
- Online：在线状态
- LastOnline：最后在线时间
- UpdatedAt：状态更新时间
- Resources：资源使用情况
- Metrics：性能指标

#### 2.1.3 资源信息（NodeResources）

- CPU：CPU 资源使用情况
- Memory：内存资源使用情况
- Disk：磁盘资源使用情况

#### 2.1.4 性能指标（NodeMetrics）

- CPU：CPU 使用率
- Memory：内存使用率
- Disk：磁盘使用率
- Network：网络指标（接收/发送速率）

#### 2.1.5 状态管理

- 状态转换
- 资源监控
- 心跳检测

### 2.2 GameNodeHandler

HTTP API 服务实体，负责处理前端请求：

#### 2.2.1 API 接口

```go
// 节点管理
GET    /api/v1/nodes           // 获取节点列表
GET    /api/v1/nodes/{id}      // 获取节点详情
POST   /api/v1/nodes           // 创建新节点
PUT    /api/v1/nodes/{id}      // 更新节点信息
DELETE /api/v1/nodes/{id}      // 删除节点

// 节点状态
PUT    /api/v1/nodes/{id}/status  // 更新节点状态
GET    /api/v1/nodes/{id}/metrics // 获取节点指标
GET    /api/v1/nodes/{id}/logs    // 获取节点日志

// 流水线管理
POST   /api/v1/nodes/{id}/pipelines    // 创建流水线
GET    /api/v1/nodes/{id}/pipelines    // 获取流水线列表
GET    /api/v1/nodes/{id}/pipelines/{pid}  // 获取流水线详情
DELETE /api/v1/nodes/{id}/pipelines/{pid}  // 删除流水线
```

#### 2.2.2 请求处理

- 参数验证
- 权限检查
- 错误处理
- 响应格式化

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

数据存储管理器，负责数据的持久化：

#### 2.4.1 存储接口

```go
type GameNodeStore interface {
    // 节点管理
    List() ([]*GameNode, error)
    Get(id string) (*GameNode, error)
    Create(node *GameNode) error
    Update(node *GameNode) error
    Delete(id string) error

    // 状态管理
    UpdateStatus(id string, status string) error
    UpdateMetrics(id string, metrics map[string]interface{}) error

    // 流水线管理
    ListPipelines(nodeID string) ([]*GameNodePipeline, error)
    GetPipeline(nodeID, pipelineID string) (*GameNodePipeline, error)
    SavePipeline(nodeID string, pipeline *GameNodePipeline) error
    DeletePipeline(nodeID, pipelineID string) error
}
```

#### 2.4.2 数据模型

- 节点信息
- 状态数据
- 指标数据
- 流水线数据

### 2.5 GameNode Communication

节点通信设计，实现节点间的通信。详细设计请参考[通信设计文档](gamenode_communication.md)：

#### 2.5.1 通信组件

- GameNodeServer：Pipeline 任务管理器
- GameNodeAgent：Pipeline 任务执行器
- GameNodeProto：Pipeline 任务接口设计
- GameNodePipeline：Pipeline 任务模板

#### 2.5.2 通信协议

```protobuf
service GameNodeService {
    // 节点管理
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);

    // 流水线管理
    rpc ExecutePipeline(ExecutePipelineRequest) returns (ExecutePipelineResponse);
    rpc GetPipelineStatus(PipelineStatusRequest) returns (PipelineStatusResponse);
    rpc CancelPipeline(PipelineCancelRequest) returns (PipelineCancelResponse);

    // 监控和日志
    rpc GetNodeMetrics(NodeMetricsRequest) returns (NodeMetricsResponse);
    rpc StreamNodeLogs(NodeLogsRequest) returns (stream LogEntry);
    rpc StreamContainerLogs(ContainerLogsRequest) returns (stream LogEntry);

    // 事件订阅
    rpc SubscribeEvents(EventSubscriptionRequest) returns (stream Event);
}
```

#### 2.5.3 通信特性

- 双向流式通信
- 心跳机制
- 会话管理
- 错误处理

## 3. 错误处理

### 3.1 重试机制

```go
type RetryConfig struct {
    MaxRetries    int
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    BackoffFactor float64
    JitterFactor  float64
}
```

### 3.2 错误类型

- RetryableError：可重试的错误
- RetryNowError：需要立即重试的错误

## 4. 安全机制

### 4.1 认证授权

- 节点认证
- 会话管理
- 权限控制

### 4.2 数据安全

- 通信加密
- 数据验证
- 访问控制

## 5. 部署和运维

### 5.1 部署要求

- 系统要求
- 网络要求
- 资源要求

### 5.2 监控和维护

- 系统监控
- 日志管理
- 故障处理
