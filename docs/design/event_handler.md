# Event Handler 设计文档

Event Handler，的实现在这里[event_handler.go](../../internal/api/event_handler.go)

## 1. 概述

Event Handler 是系统的全局事件流管理器，负责处理所有实时事件的推送。它通过 Server-Sent Events (SSE) 机制，为客户端提供实时的事件订阅服务。

## 2. 核心职责

1. 事件流管理
   - 事件流的创建和维护
   - 事件的分发和推送
   - 连接的生命周期管理

2. 事件过滤
   - 基于事件类型的过滤
   - 基于资源ID的过滤
   - 基于时间范围的过滤

3. 事件存储
   - 事件历史记录
   - 事件重放支持
   - 事件清理策略

## 3. API 设计

### 3.1 事件订阅接口

```http
# 事件管理
GET    /api/v1/events                         # 订阅事件
```

#### 3.1.1 请求参数

- type: 事件类型（可选）
  - node_status: 节点状态变更
  - node_metrics: 节点指标更新
  - pipeline_status: Pipeline 状态变更
  - pipeline_progress: Pipeline 进度更新
  - container_status: 容器状态变更
  - container_metrics: 容器指标更新
  - system_alert: 系统告警
  - system_log: 系统日志

- resource_id: 资源ID（可选）
  - node_id: 节点ID
  - pipeline_id: Pipeline ID
  - container_id: 容器ID

- start_time: 开始时间（可选）
- end_time: 结束时间（可选）
- limit: 历史事件数量限制（可选）

#### 3.1.2 响应格式

```json
{
  "event": "string",      // 事件类型
  "timestamp": "string",  // 事件时间
  "resource_id": "string", // 资源ID
  "data": {}             // 事件数据
}
```

## 4. 事件类型设计

### 4.1 节点事件

#### 4.1.1 节点状态变更 (node_status)

```json
{
  "event": "node_status",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "node-001",
  "data": {
    "old_state": "offline",
    "new_state": "online",
    "reason": "heartbeat_received",
    "details": {
      "ip": "192.168.1.100",
      "last_heartbeat": "2024-03-20T10:00:00Z"
    }
  }
}
```

#### 4.1.2 节点指标更新 (node_metrics)

```json
{
  "event": "node_metrics",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "node-001",
  "data": {
    "cpu": {
      "usage": 45.5,
      "cores": 8,
      "temperature": 65
    },
    "memory": {
      "total": 32768,
      "used": 16384,
      "free": 16384
    },
    "disk": {
      "total": 512000,
      "used": 256000,
      "free": 256000
    },
    "network": {
      "in_bytes": 1024000,
      "out_bytes": 512000,
      "connections": 100
    },
    "gpu": {
      "usage": 30.5,
      "memory_used": 4096,
      "temperature": 75
    }
  }
}
```

### 4.2 Pipeline 事件

#### 4.2.1 Pipeline 状态变更 (pipeline_status)

```json
{
  "event": "pipeline_status",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "pipeline-001",
  "data": {
    "old_status": "running",
    "new_status": "completed",
    "node_id": "node-001",
    "start_time": "2024-03-20T09:00:00Z",
    "end_time": "2024-03-20T10:00:00Z",
    "error": null
  }
}
```

#### 4.2.2 Pipeline 进度更新 (pipeline_progress)

```json
{
  "event": "pipeline_progress",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "pipeline-001",
  "data": {
    "current_step": 3,
    "total_steps": 5,
    "progress": 60,
    "step_status": {
      "name": "build",
      "status": "running",
      "start_time": "2024-03-20T09:55:00Z",
      "details": {
        "stage": "compiling",
        "files_processed": 100
      }
    }
  }
}
```

### 4.3 容器事件

#### 4.3.1 容器状态变更 (container_status)

```json
{
  "event": "container_status",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "container-001",
  "data": {
    "old_state": "running",
    "new_state": "stopped",
    "node_id": "node-001",
    "pipeline_id": "pipeline-001",
    "reason": "task_completed",
    "exit_code": 0,
    "started_at": "2024-03-20T09:00:00Z",
    "finished_at": "2024-03-20T10:00:00Z"
  }
}
```

#### 4.3.2 容器指标更新 (container_metrics)

```json
{
  "event": "container_metrics",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "container-001",
  "data": {
    "cpu": {
      "usage": 35.5,
      "limit": 100,
      "period": 100000
    },
    "memory": {
      "usage": 2048,
      "limit": 4096,
      "swap": 0
    },
    "network": {
      "rx_bytes": 512000,
      "tx_bytes": 256000,
      "rx_packets": 1000,
      "tx_packets": 500
    },
    "block_io": {
      "read_bytes": 1024000,
      "write_bytes": 512000
    }
  }
}
```

### 4.4 系统事件

#### 4.4.1 系统告警 (system_alert)

```json
{
  "event": "system_alert",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "alert-001",
  "data": {
    "level": "warning",
    "category": "resource",
    "title": "High CPU Usage",
    "message": "Node node-001 CPU usage is above 80%",
    "source": "node-001",
    "details": {
      "threshold": 80,
      "current_value": 85,
      "duration": "5m"
    },
    "actions": [
      {
        "type": "scale_up",
        "target": "node-001",
        "params": {
          "cpu_limit": 100
        }
      }
    ]
  }
}
```

#### 4.4.2 系统日志 (system_log)

```json
{
  "event": "system_log",
  "timestamp": "2024-03-20T10:00:00Z",
  "resource_id": "log-001",
  "data": {
    "level": "info",
    "component": "pipeline",
    "message": "Pipeline pipeline-001 started",
    "source": "pipeline-001",
    "trace_id": "trace-001",
    "context": {
      "user_id": "user-001",
      "action": "start_pipeline",
      "params": {
        "node_id": "node-001",
        "template": "game-build"
      }
    }
  }
}
```

## 5. 实现细节

### 5.1 性能优化

1. 连接管理
   - 连接池复用
   - 心跳检测
   - 自动重连

2. 数据处理
   - 事件批处理
   - 数据压缩
   - 缓存策略

3. 资源控制
   - 连接数限制
   - 带宽控制
   - 内存管理

### 5.2 可靠性保证

1. 错误处理
   - 连接异常处理
   - 数据丢失重传
   - 服务降级策略

2. 监控告警
   - 连接数监控
   - 延迟监控
   - 错误率监控

## 6. 待办事项

### 6.1 功能增强

1. 事件过滤
   - 增加更多过滤条件
   - 支持复杂过滤规则
   - 动态过滤规则

2. 事件存储
   - 优化存储策略
   - 支持事件重放
   - 数据清理策略

### 6.2 性能优化

1. 连接管理
   - 优化连接池
   - 改进心跳机制
   - 优化重连策略

2. 数据处理
   - 优化批处理
   - 改进压缩算法
   - 优化缓存策略 