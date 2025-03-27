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
