# GamePipelineService 服务设计文档

## 1. 系统概述

GamePipelineService 服务是 Beagle Wind Game 平台的核心服务之一，负责管理游戏节点流水线的生命周期。该服务采用 gRPC 作为通信协议，支持流水线的创建、更新、状态管理和监控等功能。

### 1.1 核心功能

- 流水线生命周期管理：创建、更新、删除流水线
- 状态管理：跟踪和更新流水线及步骤状态
- 日志管理：收集和存储流水线执行日志
- 自动调度：自动选择合适的服务器执行流水线

### 1.2 系统架构

```
            +-------------------+
            | GamePipelineGRPC  |
            | Service(Proto)    |
            +-------------------+
                      ↑
                      |
+-------------------+     +-------------------+     +-------------------+     +-------------------+
| GamePipelineAgent | --> | GamePipelineServer| --> | GamePipelineService| --> | GamePipelineHandler|
+-------------------+     +-------------------+     +-------------------+     +-------------------+
                                                              ↑
                                                              |
                                                    +-------------------+
                                                    | GamePipelineStore |
                                                    +-------------------+
```

系统组件说明：

1. **GamePipelineAgent**
   - 运行在游戏节点上的代理服务
   - 通过 gRPC 与服务器通信
   - 负责执行流水线任务
   - 遵循 GamePipelineGRPCService 协议

2. **GamePipelineGRPCService**
   - 定义 gRPC 服务接口
   - 包含所有流水线相关的 RPC 方法
   - 使用 Protocol Buffers 定义消息格式
   - 同时约束 Agent 和 Server 的通信

3. **GamePipelineServer**
   - 实现 gRPC 服务接口
   - 处理来自 Agent 的请求
   - 管理流水线生命周期
   - 遵循 GamePipelineGRPCService 协议

4. **GamePipelineService**
   - 实现核心业务逻辑
   - 处理流水线状态管理
   - 实现自动调度机制
   - 使用 GamePipelineStore 进行数据操作

5. **GamePipelineHandler**
   - 提供 HTTP REST API
   - 处理外部系统请求
   - 转发请求到 Service 层

6. **GamePipelineStore**
   - 专门服务于 GamePipelineService
   - 负责数据持久化
   - 管理流水线状态
   - 提供数据查询接口

### 1.3 通信流程

1. **Agent 到 Server**
   - Agent 通过 gRPC 与 Server 建立长连接
   - 定期发送心跳保持连接
   - 接收和执行流水线任务
   - 双方都遵循 GamePipelineGRPCService 协议

2. **Server 到 Service**
   - Server 接收请求后转发给 Service
   - Service 处理业务逻辑
   - 结果通过 Server 返回给 Agent

3. **Handler 到 Service**
   - Handler 接收 HTTP 请求
   - 转换为内部请求格式
   - 调用 Service 处理业务逻辑

4. **Service 到 Store**
   - Service 使用 Store 进行数据操作
   - Store 专门服务于 Service 的数据需求
   - 所有数据访问都通过 Store 进行

## 2. 接口设计

### 2.1 存储接口

```go
type PipelineStore interface {
    List(params types.PipelineListParams) (*types.PipelineListResult, error)
    Get(id string) (*models.GamePipeline, error)
    Add(pipeline *models.GamePipeline) error
    Update(pipeline *models.GamePipeline) error
    UpdateStatus(id string, status string) error
    Delete(id string, force bool) error
    Cleanup() error
}
```

### 2.2 服务接口

```go
type GamePipelineGRPCService struct {
    store  store.GamePipelineStore
    logger utils.Logger
}
```

## 3. 状态管理

### 3.1 流水线状态

```go
type PipelineState string

const (
    PipelineStateNotStarted PipelineState = "not_started" // 未开始
    PipelineStatePending    PipelineState = "pending"     // 等待中
    PipelineStateRunning    PipelineState = "running"     // 运行中
    PipelineStateCompleted  PipelineState = "completed"   // 已完成
    PipelineStateFailed     PipelineState = "failed"      // 失败
    PipelineStateCanceled   PipelineState = "canceled"    // 取消
)
```

### 3.2 步骤状态

```go
type StepState string

const (
    StepStatePending   StepState = "pending"   // 等待中
    StepStateRunning   StepState = "running"   // 运行中
    StepStateCompleted StepState = "completed" // 已完成
    StepStateFailed    StepState = "failed"    // 失败
    StepStateSkipped   StepState = "skipped"   // 已跳过
)
```

## 4. 自动调度机制

### 4.1 调度流程

1. 流水线创建后自动进入调度队列
2. 调度器根据以下因素选择合适的执行节点：
   - 节点资源状态（CPU、内存、GPU等）
   - 节点当前负载
   - 流水线资源需求
   - 节点地理位置
   - 网络延迟

### 4.2 状态转换

```
创建 -> 未开始 -> 等待中 -> 运行中 -> 已完成/失败/取消
```

## 5. 主要方法

### 5.1 创建流水线

```go
func (s *GamePipelineGRPCService) Create(ctx context.Context, pipeline *models.GamePipeline) error
```

- 验证流水线信息
- 设置初始状态
- 保存到存储
- 触发自动调度

### 5.2 更新状态

```go
func (s *GamePipelineGRPCService) UpdateStatus(ctx context.Context, id string, status *models.PipelineStatus) error
```

- 更新流水线状态
- 触发相关事件
- 更新存储

### 5.3 更新步骤状态

```go
func (s *GamePipelineGRPCService) UpdateStepStatus(ctx context.Context, pipelineID string, stepID string, status *models.StepStatus) error
```

- 验证状态转换
- 更新步骤状态
- 更新流水线进度
- 检查流水线状态

### 5.4 取消流水线

```go
func (s *GamePipelineGRPCService) Cancel(ctx context.Context, id string) error
```

- 更新流水线状态为取消
- 停止相关资源
- 清理临时数据
