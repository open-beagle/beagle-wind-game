# 节点 Agent 通信设计

## 1. 通信架构

### 1.1 总体架构

节点 Agent 通信系统是 [GameNode](../README.md#gamenode) 的核心组件之一，负责游戏节点的远程管理和控制。系统采用客户端-服务器架构，基于 gRPC 实现。系统包含以下主要组件：

1. **中央服务器**：作为 gRPC 服务端，提供 API 接口供节点 Agent 调用
2. **节点 Agent**：部署在每个游戏节点上，作为 gRPC 客户端与中央服务器通信
3. **通信协议**：使用 Protocol Buffers 定义消息格式和服务接口

### 1.2 通信流程

```txt
┌─────────────┐                           ┌─────────────┐
│             │  1. 节点注册(Register)      │             │
│             │ ─────────────────────────> │             │
│             │                           │             │
│             │  2. 心跳(Heartbeat)        │             │
│  节点Agent   │ <────────────────────────> │ 中央服务器  │
│             │                           │             │
│             │  3. 命令执行(Execute*)      │             │
│             │ <────────────────────────> │             │
│             │                           │             │
│             │  4. 事件流(Events)         │             │
│             │ <─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │             │
└─────────────┘                           └─────────────┘
```

### 1.3 核心功能

1. **节点注册与心跳**：Agent 启动时向服务器注册，并定期发送心跳保持连接
2. **Pipeline 执行**：服务器向 Agent 下发 Pipeline，Agent 执行并反馈状态
3. **容器管理**：服务器可远程控制 Agent 上的容器生命周期
4. **状态监控**：Agent 定期上报节点资源使用情况和容器状态
5. **日志流**：Agent 实时上传节点和容器日志
6. **事件流**：服务器订阅 Agent 的各类事件通知

## 2. Pipeline 系统

Pipeline 系统是节点 Agent 的核心功能之一，负责执行容器化任务。它是 [GamePlatform](../README.md#gameplatform) 的重要组成部分，用于游戏平台的部署和运行环境准备。详细设计请参考 [Pipeline 系统设计文档](../pipeline.md)。

### 2.1 Pipeline 管理接口

Pipeline 系统提供以下管理接口：

1. **执行 Pipeline**

   ```protobuf
   rpc ExecutePipeline(ExecutePipelineRequest) returns (ExecutePipelineResponse)
   ```

2. **获取状态**

   ```protobuf
   rpc GetPipelineStatus(PipelineStatusRequest) returns (PipelineStatusResponse)
   ```

3. **取消执行**

   ```protobuf
   rpc CancelPipeline(PipelineCancelRequest) returns (PipelineCancelResponse)
   ```

## 3. 双向通信机制

### 3.1 请求-响应模式

用于简单的命令下发和状态查询：

1. **注册(Register)**：Agent 启动时向服务器注册，获取 session_id
2. **心跳(Heartbeat)**：Agent 定期发送心跳，附带基本状态信息
3. **命令执行**：如启动容器(StartContainer)、执行 Pipeline(ExecutePipeline)等

### 3.2 服务器流模式

用于服务器持续接收 Agent 发送的数据流：

1. **日志流(StreamNodeLogs, StreamContainerLogs)**：Agent 将日志实时发送给服务器
2. **监控指标流**：Agent 定期上报详细的资源使用情况

### 3.3 客户端流模式

用于 Agent 持续向服务器发送数据：

1. **批量事件上报**：Agent 将多个事件批量发送给服务器
2. **文件上传**：Agent 将大文件分块上传到服务器

### 3.4 双向流模式

用于实时交互的场景：

1. **事件订阅(SubscribeEvents)**：服务器订阅 Agent 的事件，Agent 实时推送事件通知
2. **远程控制会话**：服务器与 Agent 建立长连接交互会话

## 4. 错误处理与重试策略

### 4.1 通信错误分类

1. **临时性错误**：网络抖动、服务器暂时不可用等
2. **持久性错误**：认证失败、权限不足、参数无效等
3. **超时错误**：请求超过预定时间未得到响应

### 4.2 重试策略

1. **指数退避算法**：

   - 初始重试等待时间：100ms
   - 最大重试等待时间：30s
   - 随机抖动因子：0.2 (±20%)
   - 最大重试次数：5 次

2. **临时性错误处理**：

   - 应用指数退避重试
   - 错误日志记录
   - 重试计数

3. **持久性错误处理**：

   - 不进行重试
   - 错误日志记录
   - 错误上报给服务器

4. **超时错误处理**：
   - 针对幂等操作应用重试策略
   - 非幂等操作需谨慎重试，确认操作状态

### 4.3 错误恢复机制

1. **会话恢复**：

   - Agent 在连接断开后尝试重连
   - 使用上一次的 session_id 尝试恢复会话
   - 如果会话无效，则重新注册

2. **状态同步**：

   - 重连成功后，Agent 发送完整的状态报告
   - 服务器与 Agent 进行状态对比和同步
   - 对已下发但未确认完成的任务进行状态查询

3. **任务幂等性**：
   - 所有重要操作都设计为幂等操作
   - 使用唯一标识符防止重复执行
   - 操作执行前检查状态，避免重复操作

## 5. 安全机制

### 5.1 认证与授权

1. **TLS 加密**：

   - 所有通信使用 TLS 加密
   - 证书验证确保服务端身份

2. **节点认证**：

   - 每个节点分配唯一的节点 ID 和认证令牌
   - 注册时进行身份验证
   - 心跳消息包含认证信息

3. **访问控制**：
   - 基于角色的权限控制
   - 操作审计日志
   - 敏感操作需要额外授权

### 5.2 数据保护

1. **敏感信息处理**：

   - 不在日志中记录敏感信息
   - 密码和令牌等敏感数据加密存储
   - 传输中的敏感数据加密

2. **数据完整性**：
   - 消息摘要验证
   - 校验和验证文件完整性
   - 版本控制防止数据覆盖

## 6. 性能优化

### 6.1 连接管理

1. **连接池**：

   - 复用 gRPC 连接
   - 动态调整连接数量
   - 定期清理空闲连接

2. **流控制**：
   - 客户端限制并发请求数
   - 服务端限制请求处理速率
   - 背压机制防止过载

### 6.2 数据优化

1. **批处理**：

   - 批量发送小消息
   - 合并相似请求
   - 异步处理非关键消息

2. **压缩**：

   - 对日志和大数据包启用压缩
   - 选择合适的压缩算法
   - 动态调整压缩级别

3. **缓存**：
   - 缓存频繁请求的响应
   - 状态变更增量更新
   - 预取可能需要的数据

## Docker 客户端依赖关系

在 Agent 系统中，有三个主要组件：AgentServer、Agent 和 Pipeline。关于 Docker 客户端的依赖关系，我们做了如下设计：

### 1. AgentServer

- **职责**：提供 gRPC 服务，管理节点连接，处理 Pipeline 请求
- **特点**：不直接执行 Docker 操作
- **结论**：**不需要**从外部传入 dockerClient
- **原因**：
  - AgentServer 只负责服务管理和请求转发
  - 不涉及具体的 Docker 操作

### 2. Agent

- **职责**：作为客户端连接到 AgentServer，执行 Pipeline 任务
- **特点**：需要直接执行 Docker 操作（创建容器、管理容器等）
- **结论**：**需要**从外部传入 dockerClient
- **原因**：
  - 需要执行实际的 Docker 操作
  - 需要支持测试时注入 mock 客户端
  - 遵循依赖倒置原则
  - 直接依赖 Docker 功能
  - 最适合管理 Docker 客户端生命周期

### 3. Pipeline

- **职责**：执行具体的容器操作步骤
- **特点**：被 Agent 调用，执行具体的容器操作
- **结论**：**不需要**从外部传入 dockerClient
- **原因**：
  - Pipeline 是 Agent 的内部实现细节
  - Pipeline 的 Docker 操作应该通过 Agent 的 dockerClient 执行
  - 保持 Pipeline 的简单性，让它专注于步骤执行逻辑

## 设计原则

这个设计遵循以下原则：

1. **单一职责原则**

   - 每个组件都有明确的职责
   - 避免职责重叠

2. **依赖倒置原则**

   - 高层模块不依赖低层模块的具体实现
   - 通过依赖注入实现解耦

3. **接口隔离原则**

   - 组件之间通过清晰的接口通信
   - 避免不必要的依赖

4. **可测试性**
   - 支持单元测试
   - 可以方便地注入 mock 对象

## 实现建议

1. Agent 的初始化：

```go
func NewAgent(opts AgentOptions, dockerClient *client.Client) *Agent {
    // 使用外部传入的 dockerClient
}
```

2. Pipeline 的创建：

```go
func (a *Agent) executePipeline(req *proto.ExecutePipelineRequest) error {
    // 使用 Agent 的 dockerClient
    pipeline := NewPipeline(req, a.dockerClient)
}
```

3. 测试用例：

```go
func TestAgent(t *testing.T) {
    mockDockerClient := NewMockDockerClient()
    agent := NewAgent(opts, mockDockerClient)
    // 进行测试...
}
```
