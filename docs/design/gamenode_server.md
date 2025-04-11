# GameNodeServer 设计文档

## 1. 系统概述

GameNodeServer 是 GameNode 系统的核心服务组件之一，负责管理游戏节点的 Pipeline 任务。它是 GameNodeCommunication 的重要组成部分，通过 gRPC 与 GameNodeAgent 进行通信。

### 1.1 设计原则

1.1.1 **状态管理原则**

- 状态数据分为两类：
  - GameNode 状态数据：由 GameNodeService 统一管理
  - GamePipeline 状态数据：由 GamePipelineGRPCService 统一管理
- GameNodeServer 不存储任何状态数据
- 所有状态更新通过对应的 Service 进行
- 保持状态管理的一致性和集中性

  1.1.2 **职责定位**

- GameNodeServer：纯通信服务组件

  - 负责与 Agent 的 gRPC 通信
  - 处理通信协议转换
  - 通过 GameNodeService 处理节点状态
  - 通过 GamePipelineGRPCService 处理 Pipeline 状态
  - 不存储任何状态数据

- GameNodeService：节点状态管理组件

  - 负责 GameNode 状态数据的存储
  - 处理节点状态更新逻辑
  - 维护节点数据一致性
  - 提供节点状态查询接口

- GamePipelineGRPCService：Pipeline 状态管理组件

  - 负责 Pipeline 状态数据的存储
  - 处理 Pipeline 状态更新逻辑
  - 维护 Pipeline 数据一致性
  - 提供 Pipeline 状态查询接口

### 1.2 核心职责

1.2.1 **节点服务**

节点服务是 GameNodeServer 的核心功能之一，负责处理节点的生命周期管理和状态维护。

#### 1.2.1.1 节点注册服务

- 功能描述：
  - 处理节点的首次注册请求
  - 验证节点信息的完整性和有效性
  - 创建节点记录并初始化状态
  - 建立与节点的 gRPC 连接

- 注册请求（RegisterRequest）：
  - 节点 ID：节点的唯一标识符
  - 节点类型：物理机/虚拟机/容器
  - 节点硬件配置：
    - CPU：型号、核心数、线程数、频率、缓存、架构
    - 内存：总量、类型、频率
    - GPU：型号、显存大小、架构、驱动版本、计算能力、功耗
    - 硬盘：型号、容量、类型、路径
    - 网络：网卡名称、MAC 地址、IP 地址、速率
  - 节点系统配置：
    - 操作系统：发行版、版本号、架构
    - 内核版本：主版本、次版本、修订号
    - 显卡驱动：驱动版本、计算框架版本
    - 运行时环境：Docker 版本、Containerd 版本、Runc 版本

- 注册响应（RegisterResponse）：
  - 注册结果：
    - 成功：返回成功状态和欢迎消息
    - 失败：返回失败状态和错误原因
  - 节点维护状态：
    - normal：正常状态，可以处理所有业务
    - maintenance：维护状态，不处理业务但保持心跳
    - disabled：禁用状态，只响应心跳

- 处理流程：
  1. 接收注册请求
  2. 验证节点信息完整性
  3. 检查节点是否已存在
     - 新节点：创建新记录
     - 已有节点：更新信息
  4. 获取节点维护状态
  5. 根据维护状态决定后续行为：
     - normal：允许注册，建立连接
     - maintenance：允许注册，建立连接
     - disabled：拒绝注册，返回错误
  6. 返回注册响应

- 状态管理：
  - 正常状态：允许注册，建立连接
  - 维护状态：允许注册，建立连接
  - 禁用状态：拒绝注册，返回错误

- 错误处理：
  - 信息不完整：返回错误，要求补充信息
  - 信息无效：返回错误，说明原因
  - 注册失败：返回错误，提供重试建议
  - 状态异常：返回错误，说明状态限制

#### 1.2.1.2 节点资源服务

- 功能描述：
  - 收集节点硬件和系统配置信息
  - 上报节点资源配置信息
  - 更新节点状态中的资源配置
  - 不处理资源调度等业务逻辑

- 资源上报请求（ResourceInfoRequest）：
  - 节点 ID：目标节点标识符
  - 时间戳：信息采集时间
  - 硬件配置：
    - CPU：型号、核心数、线程数、频率、缓存、架构
    - 内存：总量、类型、频率
    - GPU：型号、显存大小、架构、驱动版本、计算能力、功耗
    - 硬盘：型号、容量、类型、路径
    - 网络：网卡名称、MAC 地址、IP 地址、速率
  - 系统配置：
    - 操作系统：发行版、版本号、架构
    - 内核版本：主版本、次版本、修订号
    - 显卡驱动：驱动版本、计算框架版本
    - 运行时环境：Docker 版本、Containerd 版本、Runc 版本

- 资源上报响应（ResourceInfoResponse）：
  - 处理结果：是否成功接收
  - 错误信息：失败时的错误说明
  - 下次上报时间：建议的下次上报时间

- 更新机制：
  - 注册时：完整采集并上报
  - 定期更新：默认 24 小时
  - 配置变更：立即上报
  - 手动触发：按需上报

- 数据存储：
  - 存储在节点状态中
  - 记录历史变更
  - 支持配置对比

- 错误处理：
  - 采集失败：返回错误，要求重试
  - 上报超时：允许重试，最多 3 次
  - 数据无效：返回错误，说明原因
  - 存储失败：记录错误日志

#### 1.2.1.3 节点心跳服务

- 功能描述：
  - 维护节点连接状态
  - 维护节点在线状态（online字段）
  - 检测节点存活状态
  - 处理节点离线情况

- 心跳请求（HeartbeatRequest）：
  - 节点 ID：节点的唯一标识符
  - 时间戳：心跳发送时间
  - 会话 ID：当前会话标识符

- 心跳响应（HeartbeatResponse）：
  - 状态码：心跳处理结果
  - 消息：可选的响应消息

- 心跳机制：
  - 默认间隔：15 秒
  - 超时时间：45 秒
  - 重试次数：3 次
  - 重试间隔：5 秒

- 在线状态维护：
  - 收到心跳：设置 online = true
  - 心跳超时：设置 online = false
  - 重连成功：设置 online = true
  - 连接断开：设置 online = false

- 状态管理：
  - 正常状态：正常处理心跳
  - 维护状态：正常处理心跳
  - 禁用状态：正常处理心跳

- 错误处理：
  - 心跳超时：标记节点离线
  - 连接断开：标记节点离线
  - 重连失败：保持离线状态
  - 会话无效：要求重新注册

#### 1.2.1.4 节点指标服务

- 功能描述：
  - 收集节点性能指标数据
  - 上报节点资源使用情况
  - 提供指标数据查询接口
  - 不处理监控告警等业务逻辑

- 指标上报请求（MetricsReportRequest）：
  - 节点 ID：目标节点标识符
  - 时间戳：指标采集时间
  - 指标数据：
    - 系统指标：
      - CPU：使用率、核心数、频率、温度
      - 内存：使用率、总量、可用量、交换区使用
      - 磁盘：使用率、IOPS、读写速度、剩余空间
      - 网络：带宽使用、连接数、丢包率、延迟
    - 业务指标：
      - GPU：使用率、温度、功耗
      - 显存：总量、已用、剩余、使用率
      - 容器：数量、状态、资源使用
    - 自定义指标：用户定义的业务指标

- 指标上报响应（MetricsReportResponse）：
  - 处理结果：是否成功接收
  - 错误信息：失败时的错误说明
  - 下次上报时间：建议的下次上报时间

- 采集机制：
  - 定时采集：默认 15 秒
  - 批量上报：默认 1 分钟
  - 阈值触发：实时上报
  - 异常触发：立即上报

- 数据存储：
  - 短期数据：内存缓存，保留 24 小时
  - 中期数据：时序数据库，保留 7 天
  - 长期数据：归档存储，保留 30 天

- 错误处理：
  - 数据格式错误：返回错误，要求重试
  - 上报超时：允许重试，最多 3 次
  - 存储失败：记录错误日志
  - 数据异常：记录警告日志

#### 1.2.1.5 节点维护状态变更服务

- 功能描述：
  - Server 主动通知 Agent 节点维护状态变更
  - 执行状态转换操作
  - 维护状态一致性
  - 确保 Agent 正确响应状态变更

- 状态变更请求（StateChangeRequest）：
  - 节点 ID：目标节点标识符
  - 目标状态：要变更的目标状态
  - 变更原因：状态变更的原因说明
  - 变更时间：状态变更的时间戳

- 状态变更响应（StateChangeResponse）：
  - 处理结果：是否成功处理
  - 错误信息：失败时的错误说明
  - 确认时间：Agent 确认变更的时间

- 状态转换流程：
  1. Server 发起状态变更请求
  2. Agent 接收并验证请求
  3. Agent 执行状态转换
  4. Agent 返回处理结果
  5. Server 确认状态变更

- 状态转换行为：
  - normal -> maintenance：
    - 停止接收新任务
    - 保持现有任务
    - 继续状态报告
  - normal -> disabled：
    - 停止所有任务
    - 终止现有任务
    - 只保持心跳
  - maintenance -> normal：
    - 恢复任务接收
    - 保持现有任务
    - 正常状态报告
  - maintenance -> disabled：
    - 停止所有任务
    - 终止现有任务
    - 只保持心跳
  - disabled -> normal：
    - 恢复所有功能
    - 可以接收任务
    - 正常状态报告

- 错误处理：
  - 请求超时：重试或标记失败
  - 节点离线：等待重连后重试
  - 状态冲突：返回错误信息
  - 处理失败：记录错误日志

- 安全控制：
  - 需要管理员权限
  - 记录审计日志
  - 通知相关组件
  - 保持状态一致性

  1.2.2 **节点监控**

- 采集节点 Metrics 数据
  - 系统指标：
    - CPU：使用率、核心数、频率、温度
    - 内存：使用率、总量、可用量、交换区使用
    - 磁盘：使用率、IOPS、读写速度、剩余空间
    - 网络：带宽使用、连接数、丢包率、延迟
  - 业务指标：
    - GPU：使用率、温度、功耗
    - 显存：总量、已用、剩余、使用率
  - 自定义指标：暂无
- 资源信息采集
  - 硬件资源：
    - CPU：型号、核心数、线程数、频率、缓存
    - 内存：总量、类型、频率、通道数
    - GPU：型号、显存大小、CUDA 核心数
    - 硬盘：型号、容量、类型、接口
  - 软件资源：
    - 操作系统：发行版、版本号、架构
    - 内核版本：主版本、次版本、修订号
    - 显卡驱动：驱动版本、CUDA 版本
    - 运行时环境：Docker 版本、Containerd 版本、Runc 版本
- 定时上报机制
  - Agent 定时采集（默认 15s）
  - 批量上报（默认 1min）
  - 异常阈值触发
- 资源信息更新时机

  - 节点注册时
  - 心跳包中携带
  - 资源变更时主动上报

  1.2.3 **Pipeline 任务管理**

- 创建和分发 Pipeline 任务
- 跟踪任务执行状态
- 处理任务取消请求
- 管理 Step 状态变更

同样对于 pipeline 任务，其状态由 GamePipelineGRPCService 负责维护，GameNodeServer 初始化时应传入 GamePipelineGRPCService 实例，用于管理 pipeline 数据。
注意任何不满足业务需求的数据结构都应该提出修改意见，并修改 GamePipeline 对象，这个对象的位置有设计文件专门管；

1.2.4 **状态同步**

- 流式获取 Pipeline.Step 的执行日志

一般来讲，Pipeline 会在 Step 执行完毕后，由 Agent 收集 Step 执行的日志，推送给 Server，但是想要实时查阅 Pipeline 的 Step 执行日志，则需要采用"流式获取 Pipeline.Step 的执行日志"方法。

### 1.3 系统架构

```mermaid
graph TD
    A[GameNodeServer] --> B[PipelineManager]
    A --> C[NodeManager]
    A --> D[LogManager]

    B --> E[GamePipelineGRPCService]
    C --> F[GameNodeService]

    B --> G[TaskDispatcher]
    C --> H[ConnectionManager]
    D --> I[LogStreamer]
```

## 2. 核心组件

### 2.1 PipelineManager

负责 Pipeline 任务的管理和执行：

```go
type PipelineManager struct {
    pipelineService *GamePipelineGRPCService
    taskDispatcher  *TaskDispatcher
}

// 任务分发器
type TaskDispatcher struct {
    agentConnections map[string]*AgentConnection
    mu              sync.RWMutex
}

// Agent连接
type AgentConnection struct {
    client  proto.GameNodeServiceClient
    conn    *grpc.ClientConn
    nodeID  string
}
```

#### 2.1.1 主要功能

1. **任务创建**

   - 验证任务参数
   - 通过 pipelineService 创建任务记录
   - 分配执行节点

2. **任务分发**

   - 选择目标节点
   - 序列化任务数据
   - 发送执行请求

3. **状态跟踪**

   - 接收状态更新
   - 通过 pipelineService 更新任务状态
   - 触发相关事件

### 2.2 NodeManager

负责节点连接管理：

```go
type NodeManager struct {
    nodeService    *GameNodeService
    connectionManager *ConnectionManager
}

// 连接管理器
type ConnectionManager struct {
    connections map[string]*AgentConnection
    mu          sync.RWMutex
}

// Agent连接
type AgentConnection struct {
    client  proto.GameNodeServiceClient
    conn    *grpc.ClientConn
    nodeID  string
}
```

#### 2.2.1 主要功能

1. **节点注册**

   - 验证节点信息
   - 通过 nodeService 处理注册数据
   - 创建 gRPC 连接

2. **心跳管理**

   - 接收心跳请求
   - 通过 nodeService 更新节点状态
   - 维护连接状态

3. **连接管理**

   - 维护 gRPC 连接
   - 处理连接异常
   - 清理断开的连接

### 2.3 LogManager

负责日志流管理：

```go
type LogManager struct {
    logStreamer *LogStreamer
}

// 日志流管理器
type LogStreamer struct {
    streams map[string]*LogStream
    mu      sync.RWMutex
}

// 日志流
type LogStream struct {
    pipelineID string
    stepID     string
    stream     proto.GameNodeService_StreamStepLogsClient
    status     StreamStatus
}
```

#### 2.3.1 主要功能

1. **日志流管理**

   - 创建日志流
   - 维护流状态
   - 处理流异常

2. **日志接收**

   - 接收实时日志
   - 处理日志格式
   - 转发日志数据

## 3. 接口设计

### 3.1 Proto 文件设计原则

3.1.1 **设计原则**

- 消息类型定义

  - 每个消息类型必须有明确的用途
  - 字段命名必须清晰且符合命名规范
  - 字段类型必须合适且符合最佳实践
  - 必须包含必要的元数据字段（如时间戳）

- 服务接口定义

  - 接口方法必须有明确的职责
  - 参数和返回值必须合理且完整
  - 必须考虑接口的扩展性
  - 必须考虑接口的兼容性

- 枚举类型定义

  - 枚举值必须有明确的含义
  - 枚举值必须保持向后兼容
  - 必须考虑枚举值的扩展性

    3.1.2 **变更流程**

- 变更申请

  - 提交变更申请，说明变更原因
  - 提供变更的详细设计
  - 评估变更的影响范围

- 变更评审

  - 评审变更的合理性
  - 评审变更的兼容性
  - 评审变更的实现方案

- 变更实施

  - 按照设计进行变更
  - 确保变更不影响现有功能
  - 更新相关文档和测试

- 变更验证

  - 验证变更的正确性
  - 验证变更的兼容性
  - 验证变更的性能影响

    3.1.3 **事件消息设计**

```protobuf
// 事件
message Event {
  string type = 1;      // 事件类型
  string node_id = 2;   // 节点ID
  string entity_id = 3; // 实体ID（如容器ID、流水线ID等）
  string status = 4;    // 事件状态
  string message = 5;   // 事件消息
  google.protobuf.Timestamp timestamp = 6; // 时间戳
  map<string, string> data = 7; // 额外数据字段
}
```

事件消息设计考虑：

1. 类型字段：用于区分不同的事件类型
2. 节点 ID：标识事件发生的节点
3. 实体 ID：标识事件相关的实体
4. 状态字段：表示事件的状态
5. 消息字段：提供事件的描述信息
6. 时间戳：记录事件发生的时间
7. 数据字段：存储事件的额外信息

### 3.2 gRPC 服务接口

```protobuf
service GameNodeService {
  // 节点管理
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  rpc ReportMetrics(MetricsReport) returns (ReportResponse);
  rpc UpdateResourceInfo(ResourceInfo) returns (UpdateResponse);

  // Pipeline管理
  rpc Execute(ExecuteRequest) returns (ExecuteResponse);
  rpc UpdatePipelineStatus(PipelineStatusUpdate) returns (UpdateResponse);
  rpc UpdateStepStatus(StepStatusUpdate) returns (UpdateResponse);
  rpc Cancel(PipelineCancelRequest) returns (CancelResponse);

  // 日志流
  rpc StreamLogs(LogRequest) returns (stream LogEntry);
}

// 节点指标报告
message MetricsReport {
  string node_id = 1;
  int64 timestamp = 2;
  repeated Metric metrics = 3;
}

message Metric {
  string name = 1;
  string type = 2;
  double value = 3;
  map<string, string> labels = 4;
}

// 资源信息更新
message ResourceInfo {
  string node_id = 1;
  int64 timestamp = 2;
  HardwareInfo hardware = 3;
  SoftwareInfo software = 4;
  NetworkInfo network = 5;
}

message HardwareInfo {
  CPUInfo cpu = 1;
  MemoryInfo memory = 2;
  GPUInfo gpu = 3;
  DiskInfo disk = 4;
}

// Step状态更新
message StepStatusUpdate {
  string pipeline_id = 1;
  string step_id = 2;
  StepStatus status = 3;
  int64 start_time = 4;
  int64 end_time = 5;
  string error_message = 6;
  bytes logs = 7;
}

enum StepStatus {
  PENDING = 0;
  RUNNING = 1;
  COMPLETED = 2;
  FAILED = 3;
  CANCELLED = 4;
}
```

### 3.3 内部接口

```go
// PipelineManager 接口
type PipelineManager interface {
    Create(ctx context.Context, req *CreateRequest) (*Pipeline, error)
    Execute(ctx context.Context, req *ExecuteRequest) error
    UpdateStatus(ctx context.Context, req *UpdateStatusRequest) error
    UpdateStepStatus(ctx context.Context, req *StepStatusUpdate) error
    Cancel(ctx context.Context, req *CancelRequest) error
}

// NodeManager 接口
type NodeManager interface {
    RegisterNode(ctx context.Context, req *RegisterRequest) error
    UpdateHeartbeat(ctx context.Context, req *HeartbeatRequest) error
    GetNodeStatus(ctx context.Context, nodeID string) (*NodeStatus, error)
    ReportMetrics(ctx context.Context, req *MetricsReport) error
    UpdateResourceInfo(ctx context.Context, req *ResourceInfo) error
}

// LogManager 接口
type LogManager interface {
    StreamLogs(ctx context.Context, req *StreamLogsRequest) (<-chan LogEntry, error)
}
```

## 4. 业务流程

### 4.1 Pipeline 执行流程

1. **任务创建**

   ```go
   func (pm *PipelineManager) Create(ctx context.Context, req *CreateRequest) (*Pipeline, error) {
       // 1. 验证请求参数
       if err := pm.validateRequest(req); err != nil {
           return nil, err
       }

       // 2. 通过 pipelineService 创建任务记录
       pipeline, err := pm.pipelineService.Create(ctx, req)
       if err != nil {
           return nil, err
       }

       return pipeline, nil
   }
   ```

2. **任务分发**

   ```go
   func (pm *PipelineManager) Execute(ctx context.Context, req *ExecuteRequest) error {
       // 1. 获取目标节点
       node, err := pm.nodeManager.GetNode(req.NodeID)
       if err != nil {
           return err
       }

       // 2. 序列化任务数据
       data, err := pm.serializePipeline(req.Pipeline)
       if err != nil {
           return err
       }

       // 3. 发送执行请求
       return pm.sendExecuteRequest(ctx, node, data)
   }
   ```

3. **状态更新**

   ```go
   func (pm *PipelineManager) UpdateStatus(ctx context.Context, req *UpdateStatusRequest) error {
       // 1. 通过 pipelineService 更新任务状态
       if err := pm.pipelineService.UpdateStatus(ctx, req); err != nil {
           return err
       }

       return nil
   }
   ```

### 4.2 节点管理流程

1. **节点注册**

   ```go
   func (nm *NodeManager) RegisterNode(ctx context.Context, req *RegisterRequest) error {
       // 1. 验证节点信息
       if err := nm.validateNodeInfo(req.NodeInfo); err != nil {
           return err
       }

       // 2. 通过 nodeService 处理注册数据
       return nm.nodeService.RegisterNode(ctx, req)
   }
   ```

2. **心跳处理**

   ```go
   func (nm *NodeManager) UpdateHeartbeat(ctx context.Context, req *HeartbeatRequest) error {
       // 1. 通过 nodeService 更新节点状态
       return nm.nodeService.UpdateHeartbeat(ctx, req)
   }
   ```

3. **指标上报**

   ```go
   func (nm *NodeManager) ReportMetrics(ctx context.Context, req *MetricsReport) error {
       // 1. 验证指标数据
       if err := nm.validateMetrics(req.Metrics); err != nil {
           return err
       }

       // 2. 通过 nodeService 更新指标数据
       return nm.nodeService.ReportMetrics(ctx, req)
   }
   ```

4. **资源信息更新**

   ```go
   func (nm *NodeManager) UpdateResourceInfo(ctx context.Context, req *ResourceInfo) error {
       // 1. 验证资源信息
       if err := nm.validateResourceInfo(req); err != nil {
           return err
       }

       // 2. 通过 nodeService 更新资源信息
       return nm.nodeService.UpdateResourceInfo(ctx, req)
   }
   ```

### 4.3 Pipeline 任务管理流程

1. **Step 状态更新**

   ```go
   func (pm *PipelineManager) UpdateStepStatus(ctx context.Context, req *StepStatusUpdate) error {
       // 1. 验证状态更新请求
       if err := pm.validateStepStatus(req); err != nil {
           return err
       }

       // 2. 根据状态类型处理日志
       if req.Status == StepStatus_COMPLETED || req.Status == StepStatus_FAILED {
           // 完整日志，保存到存储
           if err := pm.pipelineService.SaveStepLogs(ctx, req); err != nil {
               return err
           }
       } else if req.Status == StepStatus_CANCELLED {
           // 取消状态，不保存日志
           req.Logs = nil
       }

       // 3. 通过 pipelineService 更新 Step 状态
       return pm.pipelineService.UpdateStepStatus(ctx, req)
   }
   ```

## 5. 错误处理

### 5.1 错误类型

```go
type ServerError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

const (
    ErrInvalidRequest ErrorCode = iota
    ErrNodeNotFound
    ErrNodeOffline
    ErrPipelineNotFound
    ErrPipelineFailed
    ErrInternal
)
```

### 5.2 错误处理策略

1. **请求验证错误**

   - 返回详细的错误信息
   - 不进行重试

2. **节点错误**

   - 尝试重新连接
   - 更新节点状态
   - 通知相关组件

3. **Pipeline 错误**

   - 记录错误日志
   - 更新任务状态
   - 触发错误事件

4. **内部错误**

   - 记录详细日志
   - 尝试恢复状态
   - 通知管理员

## 6. 监控指标

### 6.1 系统指标

1. **节点指标**

   - 在线节点数
   - 节点心跳延迟
   - 节点状态分布

2. **任务指标**

   - 任务队列长度
   - 任务执行时间
   - 任务成功率

3. **资源指标**

   - CPU 使用率
   - 内存使用率
   - 网络流量

### 6.2 业务指标

1. **Pipeline 指标**

   - 创建数量
   - 执行数量
   - 完成数量
   - 失败数量

2. **事件指标**

   - 事件发布数
   - 事件订阅数
   - 事件处理延迟

3. **日志指标**

   - 日志收集量
   - 日志存储量
   - 日志查询延迟
