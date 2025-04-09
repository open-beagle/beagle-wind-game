# GameNodePipelineHandler 设计文档

GameNodePipelineHandler[gamenode_pipeline_handler.go](../../internal/api/gamenode_pipeline_handler.go)

## 1. 概述

GameNodePipelineHandler 是系统的 HTTP API 服务实体，负责处理 Pipeline 相关的 HTTP 请求。它不直接参与 Pipeline 的执行逻辑，而是作为一个轻量级的 API 网关，提供 Pipeline 的查询和管理功能。

## 2. 职责边界

### 2.1 核心职责

1. 提供 Pipeline 状态查询功能
2. 提供 Pipeline 取消功能
3. 提供 Pipeline 历史记录管理

### 2.2 非职责范围

1. Pipeline 创建：由 GameNodeService 负责
2. Pipeline 执行：由 GameNodeService 负责
3. Pipeline 状态更新：由 GameNodeService 负责
4. Pipeline 事件推送：由 Event Handler 负责

## 3. API 设计

### 3.1 Pipeline 查询接口

```http
# Pipeline 查询
GET    /api/v1/pipelines               # 获取所有 Pipelines
GET    /api/v1/pipelines/{id}          # 查询 Pipeline
```

#### 3.1.1 获取 Pipeline 列表

- 请求参数：
  - page: 页码（默认：1）
  - size: 每页数量（默认：20）
  - status: Pipeline 状态过滤
  - node_id: 节点 ID 过滤
  - start_time: 开始时间
  - end_time: 结束时间
  - sort_by: 排序字段（created_at/updated_at/status）
  - sort_order: 排序方向（asc/desc）
- 响应：分页的 Pipeline 列表

#### 3.1.2 查询 Pipeline

- 请求参数：
  - id: Pipeline ID
- 响应：Pipeline 详细信息

### 3.2 Pipeline 管理接口

```http
# Pipeline 管理
POST   /api/v1/pipelines/{id}/cancel  # 取消 Pipeline
POST   /api/v1/pipelines/{id}/delete  # 删除 Pipeline
```

#### 3.2.1 取消 Pipeline

- 请求参数：
  - id: Pipeline ID
  - reason: 取消原因（可选）
- 响应：操作结果

#### 3.2.2 删除 Pipeline

- 请求参数：
  - id: Pipeline ID
  - force: 是否强制删除（可选，默认：false）
- 响应：操作结果

## 4. 实现细节

### 4.1 依赖关系

- GameNodePipelineService：支撑 GameNodePipelineHandler，内部实现 GameNodePipeline 业务的服务，[gamenode_pipeline_service.go](../../internal/service/gamenode_pipeline_service.go)
- GameNodePipelineStore：存储 GameNodePipeline 的实现，管理 GameNodePipeline 的 Yaml，[gamenode_pipeline_store.go](../../internal/store/gamenode_pipeline_store.go)
- Event Handler：事件处理服务

### 4.2 存储设计

#### 4.2.1 GameNodePipelineStore 接口

```go
// GameNodePipelineStore Pipeline 存储接口
type GameNodePipelineStore interface {
    // Pipeline 基础信息管理
    List(query *PipelineQuery) ([]*gamenode.GameNodePipeline, int64, error)
    Get(id string) (*gamenode.GameNodePipeline, error)
    Delete(id string, force bool) error

    // Pipeline 状态管理
    UpdateStatus(id string, status string) error
    UpdateProgress(id string, currentStep int, totalSteps int, progress float64) error
    UpdateError(id string, error string) error
    UpdateStartTime(id string, startTime time.Time) error
    UpdateEndTime(id string, endTime time.Time) error
}
```

#### 4.2.2 数据模型

Pipeline 数据模型使用 `models.GameNodePipeline`，定义在 [gamenode_pipeline.go](../../internal/models/gamenode_pipeline.go) 中：
