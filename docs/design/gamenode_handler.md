# GameNodeHandler 设计文档

GameNodeHandler，的实现在这里[gamenode_handler.go](../../internal/api/gamenode_handler.go)

## 1. 概述

GameNodeHandler 是系统的 HTTP API 服务实体，负责处理前端交互。它不直接参与节点的核心业务逻辑，而是作为一个轻量级的 API 网关，提供必要的查询和管理功能。

## 2. 职责边界

### 2.1 核心职责

1. 提供节点基础信息的查询和管理
2. 提供节点状态查询功能

### 2.2 非职责范围

1. 节点注册：由 GameNodeAgent 通过 gRPC 实现
2. 节点心跳：由 GameNodeAgent 通过 gRPC 实现
3. Pipeline 管理：由 GamePipelineHandler 负责，详见 [GamePipelineHandler 设计文档](gamepipeline_handler.md)
4. 容器管理：暂缓实现
5. 日志管理：暂缓实现
6. 事件流管理：由 Event Handler 负责，详见 [Event Handler 设计文档](event_handler.md)

## 3. API 设计

### 3.1 节点管理接口

```http
# 节点生命周期管理
GET    /api/v1/nodes                # 获取节点列表
GET    /api/v1/nodes/{id}           # 获取节点详情
POST   /api/v1/nodes/{id}/update    # 编辑节点
POST   /api/v1/nodes/{id}/delete    # 删除节点
```

#### 3.1.1 获取节点列表

- 请求参数：
  - page: 页码（默认：1）
  - size: 每页数量（默认：20）
  - keyword: 搜索关键词
  - status: 节点状态过滤
  - type: 节点类型过滤
  - region: 区域过滤
  - sort_by: 排序字段（created_at/updated_at/status）
  - sort_order: 排序方向（asc/desc）
- 响应：分页的节点列表

#### 3.1.2 获取节点详情

- 请求参数：
  - id: 节点 ID
  - include_metrics: 是否包含性能指标（可选，默认：false）
- 响应：节点详细信息

#### 3.1.3 编辑节点

- 请求参数：
  - id: 节点 ID
  - name: 节点名称
  - model: 节点型号
  - type: 节点类型
  - location: 节点位置
  - labels: 节点标签
  - description: 节点描述（可选）
  - maintenance_mode: 维护模式（可选）
- 响应：更新后的节点信息

#### 3.1.4 删除节点

- 请求参数：
  - id: 节点 ID
  - force: 是否强制删除（可选，默认：false）
- 响应：操作结果

## 4. 实现细节

### 4.1 依赖关系

- GameNodeService：支撑 GameNodeHandler，内部实现 GameNode 业务的服务，[gamenode_service.go](../../internal/service/gamenode_service.go)
- GameNodeStore：存储 GameNode 的实现，管理 GameNode 的 Yaml，[gamenode_store.go](../../internal/store/gamenode_store.go)
- Event Handler：事件处理服务

### 4.2 存储设计

#### 4.2.1 GameNodeStore 接口

```go
// GameNodeStore 节点存储接口
type GameNodeStore interface {
    // 节点基础信息管理
    List() ([]*GameNode, error)
    Get(id string) (*GameNode, error)
    Update(node *GameNode) error
    Delete(id string) error

    // 节点状态管理
    UpdateStatus(id string, status string) error
    ReportMetrics(id string, metrics *NodeMetrics) error
}
```

#### 4.2.2 数据模型

```go
// GameNode 游戏节点
type GameNode struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Model     string            `json:"model"`
    Type      string            `json:"type"`
    Location  string            `json:"location"`
    Labels    map[string]string `json:"labels"`
    Status    string            `json:"status"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
    Metrics   *NodeMetrics      `json:"metrics,omitempty"`
}

// NodeMetrics 节点性能指标
type NodeMetrics struct {
    CPUUsage    float64   `json:"cpu_usage"`
    MemoryUsage float64   `json:"memory_usage"`
    DiskUsage   float64   `json:"disk_usage"`
    NetworkIn   int64     `json:"network_in"`
    NetworkOut  int64     `json:"network_out"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### 4.2.3 存储实现

- 使用 YAML 文件存储节点配置
- 使用内存缓存存储节点状态和指标
- 定期将状态和指标持久化到磁盘

### 4.3 性能优化

- 节点查询性能优化
- 节点状态缓存策略
- 节点列表分页优化

## 5. 待办事项

### 5.1 暂缓功能

1. 容器管理接口

   - 需要明确具体需求
   - 评估实现复杂度
   - 确定优先级

2. 日志管理接口
   - 需要明确具体需求
   - 评估实现复杂度
   - 确定优先级

### 5.2 优化方向

1. 节点查询
   - 添加更多过滤条件
   - 优化查询性能
   - 添加排序功能
