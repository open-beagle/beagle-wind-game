# GameNodeAgent 设计文档

## 1. 系统概述

### 1.1 设计目标

GameNodeAgent 是游戏节点系统的核心组件，负责在物理节点上执行具体的任务和资源管理。主要目标包括：

1. **节点生命周期管理**

   - 节点注册和初始化
   - 状态维护和同步
   - 资源监控和报告
   - 任务执行和管理

2. **资源管理**

   - 系统资源监控
   - 容器资源管理
   - 资源使用优化

3. **任务执行**

   - Pipeline 任务执行
   - 容器生命周期管理
   - 状态同步和报告

### 1.2 核心职责

1. **节点注册与初始化**

   - 采集节点信息
   - 向服务器注册
   - 初始化系统组件
   - 建立心跳连接

2. **资源监控**

   - 系统指标采集
   - 资源使用监控
   - 性能数据收集
   - 异常情况报告

3. **任务执行**

   - Pipeline 任务接收
   - 容器生命周期管理
   - 任务状态同步
   - 错误处理和恢复

4. **状态维护**

   - 心跳检测
   - 状态同步
   - 断线重连
   - 日志收集

## 2. 对象设计

### 2.1 核心结构

```go
// GameNodeAgent 游戏节点代理
type GameNodeAgent struct {
    // 基础信息
    id          string            // 节点ID
    serverAddr  string            // 服务器地址
    config      *AgentConfig      // 代理配置

    // 系统组件
    pipelineExecutor *PipelineExecutor  // Pipeline执行器
    containerManager *ContainerManager   // 容器管理器
    metricsCollector *MetricsCollector   // 指标收集器
    eventManager     *EventManager       // 事件管理器

    // 客户端
    grpcClient  proto.GameNodeServiceClient  // gRPC客户端
    dockerClient *docker.Client              // Docker客户端

    // 状态管理
    status      *AgentStatus      // 代理状态
    mu          sync.RWMutex      // 状态锁
}

// AgentConfig 代理配置
   type AgentConfig struct {
       HeartbeatPeriod time.Duration  // 心跳间隔
       RetryCount      int            // 重试次数
       RetryDelay      time.Duration  // 重试延迟
       MetricsInterval time.Duration  // 指标采集间隔
   }

// AgentStatus 代理状态
type AgentStatus struct {
    State       string            // 状态
    LastHeartbeat time.Time       // 最后心跳时间
    Resources    *ResourceInfo    // 资源信息
    Metrics      *MetricsReport   // 性能指标
}
```

### 2.2 核心方法

   ```go
// NewGameNodeAgent 创建新的游戏节点代理
func NewGameNodeAgent(
    id string,                    // 节点ID
    serverAddr string,            // 服务器地址
    config *AgentConfig,          // 代理配置
) (*GameNodeAgent, error)

// Start 启动代理
func (a *GameNodeAgent) Start(ctx context.Context) error

// Stop 停止代理
func (a *GameNodeAgent) Stop() error

// Register 注册节点
func (a *GameNodeAgent) Register(ctx context.Context) error

// Heartbeat 发送心跳
func (a *GameNodeAgent) Heartbeat(ctx context.Context) error

// CollectMetrics 收集指标
func (a *GameNodeAgent) CollectMetrics() (*MetricsReport, error)

// ExecutePipeline 执行Pipeline
func (a *GameNodeAgent) ExecutePipeline(ctx context.Context, pipeline *Pipeline) error
```

### 2.3 状态管理

1. **状态定义**

   ```go
   const (
       StateInitializing = "initializing"
       StateRunning      = "running"
       StateStopping     = "stopping"
       StateStopped      = "stopped"
       StateError        = "error"
   )
   ```

2. **状态转换**

   ```go
   // 状态转换规则
   StateInitializing -> StateRunning
   StateRunning -> StateStopping
   StateStopping -> StateStopped
   AnyState -> StateError
   ```

### 2.4 错误处理

1. **错误类型**

   ```go
   type AgentError struct {
       Code    string
       Message string
       Cause   error
   }
   ```

2. **错误处理策略**

   - 网络错误：重试机制
   - 资源错误：资源释放
   - 任务错误：任务清理
   - 系统错误：优雅退出

## 3. 组件交互

### 3.1 与服务器交互

- 注册和认证
- 心跳维护
- 任务接收
- 状态同步

### 3.2 与容器交互

- 容器创建
- 容器监控
- 资源管理
- 日志收集

### 3.3 与系统交互

- 资源监控
- 性能采集
- 日志管理
- 事件处理

## 4. 配置管理

### 4.1 配置项

```go
type AgentConfig struct {
    // 基础配置
    NodeID          string
    ServerAddr      string

    // 心跳配置
    HeartbeatPeriod time.Duration
    RetryCount      int
    RetryDelay      time.Duration

    // 监控配置
    MetricsInterval time.Duration
    LogLevel        string

    // 资源限制
    MaxContainers   int
    ResourceQuota   ResourceQuota
}
```

### 4.2 配置加载

- 命令行参数
- 配置文件
- 环境变量
- 默认值
