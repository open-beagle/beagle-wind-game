# GamePipeline 执行引擎设计

## 概况

### 概要设计

数据载体，models.GamePipeline，GamePipelineAgent 将 models.GamePipeline 给到 GamePipelineEngine 来执行任务，
GamePipelineEngine，负责启动容器执行 models.GamePipeline 任务，并收集更新 models.GamePipeline 的 Status 属性。
GamePipelineAgent 将订阅 GamePipelineEngine 的事件来更新 models.GamePipeline 的状态与取消 models.GamePipeline 的执行。

#### 事件类型

GamePipelineEngine 将产生以下类型的事件：

1. **状态更新事件**
   - PipelineStarted：Pipeline 开始执行
   - StepStarted：单个步骤开始执行
   - StepCompleted：单个步骤执行完成
   - StepFailed：单个步骤执行失败
   - PipelineCompleted：Pipeline 执行完成
   - PipelineFailed：Pipeline 执行失败

2. **日志事件**
   - StepLog：步骤执行日志
   - ErrorLog：错误日志

#### 状态更新机制

1. **实时状态更新**
   - GamePipelineEngine 通过事件总线发送状态更新事件
   - GamePipelineAgent 订阅并处理这些事件
   - 状态更新采用乐观锁机制，确保并发安全

2. **状态持久化**
   - 关键状态变更实时持久化
   - 定期状态快照保存
   - 支持状态恢复机制

#### 错误处理策略

1. **Docker 命令执行失败**
   - 原因：Docker 命令本身执行失败（如 docker run 失败）
   - 处理策略：
     - 记录详细错误日志
     - 释放相关资源
     - 更新 Pipeline 状态为失败
     - 触发错误告警

2. **Step 执行失败**
   - 原因：Docker 容器内命令执行返回非零状态码
   - 处理策略：
     - 收集容器日志
     - 分析失败原因
     - 根据配置决定是否重试
     - 更新 Step 状态为失败
     - 触发 Pipeline 失败处理流程

3. **Pipeline 执行超时**
   - 原因：整体执行时间超过配置的超时时间
   - 处理策略：
     - 强制终止所有运行中的容器
     - 清理所有相关资源
     - 更新 Pipeline 状态为超时
     - 记录超时原因和上下文
     - 触发超时告警

### 开发目标

设计并实现一个 GamePipeline 执行引擎，该引擎作为 GamePipelineAgent 的核心组件，负责管理和执行 GamePipeline 任务。主要职责包括：

- Pipeline 生命周期的完整管理
- 容器化步骤的执行控制
- 执行状态的实时监控
- 执行日志的采集与管理
- 异常情况的处理与恢复

### 文件结构

考虑到代码的内聚性和可维护性，建议在 internal 目录下新建 `pipeline` 目录，具体结构如下：

```
internal/
  └── pipeline/
      ├── pipeline.go       // Pipeline执行引擎核心接口定义
      ├── engine.go         // Pipeline执行引擎实现
      ├── container.go      // 容器运行时管理
      ├── status.go         // 状态管理
      └── logger.go         // 日志管理
```

### 核心接口

建议在 `pipeline.go` 中定义核心接口，在 `engine.go` 中实现：

```go
// GamePipelineEngine 定义了游戏Pipeline执行引擎的核心接口
type GamePipelineEngine interface {
    // Start 启动执行引擎
    Start(ctx context.Context) error

    // Stop 停止执行引擎
    Stop(ctx context.Context) error

    // Execute 执行Pipeline任务
    Execute(ctx context.Context, pipeline *models.GamePipeline) error

    // GetStatus 获取Pipeline执行状态
    GetStatus(pipelineID string) (*models.PipelineStatus, error)

    // CancelPipeline 取消Pipeline执行
    CancelPipeline(pipelineID string, reason string) error
}
```
