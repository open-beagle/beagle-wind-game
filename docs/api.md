# API 设计文档

## 基础信息

- 基础路径: `/api/v1`
- 认证方式: JWT Token
- 响应格式: JSON
- 时间格式: ISO 8601

## 认证相关

### 登录

- 路径: `/auth/login`
- 方法: POST
- 请求体:

```json
{
  "username": "string",
  "password": "string"
}
```

- 响应:

```json
{
  "token": "string",
  "expires_at": "string",
  "user": {
    "id": "integer",
    "username": "string",
    "role": "string"
  }
}
```

### 登出

- 路径: `/auth/logout`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 响应: 204 No Content

## 游戏节点管理

### 获取节点列表

- 路径: `/nodes`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `page`: 页码 (默认: 1)
  - `size`: 每页数量 (默认: 20)
  - `status`: 节点状态 (可选)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "name": "string",
      "status": "string",
      "hardware": {
        "cpu": {
          "model": "string",
          "cores": "integer",
          "threads": "integer"
        },
        "memory": {
          "total": "integer",
          "used": "integer"
        },
        "gpu": {
          "model": "string",
          "memory": "integer"
        },
        "disk": {
          "total": "integer",
          "used": "integer"
        }
      },
      "network": {
        "ip": "string",
        "port": "integer",
        "bandwidth": "integer"
      },
      "created_at": "string",
      "updated_at": "string"
    }
  ]
}
```

### 获取节点详情

- 路径: `/nodes/{node_id}`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "id": "string",
  "name": "string",
  "status": "string",
  "hardware": {
    "cpu": {
      "model": "string",
      "cores": "integer",
      "threads": "integer",
      "usage": "float"
    },
    "memory": {
      "total": "integer",
      "used": "integer",
      "free": "integer"
    },
    "gpu": {
      "model": "string",
      "memory": "integer",
      "usage": "float"
    },
    "disk": {
      "total": "integer",
      "used": "integer",
      "free": "integer"
    }
  },
  "network": {
    "ip": "string",
    "port": "integer",
    "bandwidth": "integer",
    "current_usage": "integer"
  },
  "instances": [
    {
      "id": "string",
      "name": "string",
      "status": "string",
      "resources": {
        "cpu": "float",
        "memory": "integer",
        "disk": "integer"
      }
    }
  ],
  "created_at": "string",
  "updated_at": "string"
}
```

### 更新节点状态

- 路径: `/nodes/{node_id}/status`
- 方法: PUT
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "status": "string", // online/offline/maintenance
  "reason": "string" // 状态变更原因
}
```

- 响应: 204 No Content

### 获取节点资源使用情况

- 路径: `/nodes/{node_id}/resources`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `start_time`: 开始时间 (可选)
  - `end_time`: 结束时间 (可选)
  - `interval`: 时间间隔 (可选，默认: 1h)
- 响应:

```json
{
  "cpu": [
    {
      "timestamp": "string",
      "usage": "float"
    }
  ],
  "memory": [
    {
      "timestamp": "string",
      "used": "integer",
      "free": "integer"
    }
  ],
  "disk": [
    {
      "timestamp": "string",
      "used": "integer",
      "free": "integer"
    }
  ],
  "network": [
    {
      "timestamp": "string",
      "in": "integer",
      "out": "integer"
    }
  ]
}
```

### 获取节点日志

- 路径: `/nodes/{node_id}/logs`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `level`: 日志级别 (可选)
  - `start_time`: 开始时间 (可选)
  - `end_time`: 结束时间 (可选)
  - `page`: 页码 (默认: 1)
  - `size`: 每页数量 (默认: 20)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "timestamp": "string",
      "level": "string",
      "message": "string",
      "details": {}
    }
  ]
}
```

### 重启节点

- 路径: `/nodes/{node_id}/restart`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "task_id": "string",
  "status": "string"
}
```

## 游戏平台管理

### 获取平台列表

- 路径: `/platforms`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `page`: 页码 (默认: 1)
  - `size`: 每页数量 (默认: 20)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "name": "string",
      "version": "string",
      "type": "string",
      "status": "string",
      "created_at": "string",
      "updated_at": "string"
    }
  ]
}
```

### 获取平台详情

- 路径: `/platforms/{platform_id}`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "id": "string",
  "name": "string",
  "version": "string",
  "type": "string",
  "status": "string",
  "config": {
    "system": {},
    "user": {}
  },
  "created_at": "string",
  "updated_at": "string"
}
```

### 更新平台配置

- 路径: `/platforms/{platform_id}/config`
- 方法: PUT
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "config": {
    "system": {},
    "user": {}
  }
}
```

- 响应: 204 No Content

### 获取平台状态

- 路径: `/platforms/{platform_id}/status`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "status": "string",
  "last_check": "string",
  "resources": {
    "cpu": "float",
    "memory": "float",
    "disk": "float"
  }
}
```

## 游戏卡片管理

### 获取卡片列表

- 路径: `/cards`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `page`: 页码 (默认: 1)
  - `size`: 每页数量 (默认: 20)
  - `platform_id`: 平台 ID (可选)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "name": "string",
      "platform_id": "string",
      "version": "string",
      "status": "string",
      "created_at": "string",
      "updated_at": "string"
    }
  ]
}
```

### 获取卡片详情

- 路径: `/cards/{card_id}`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "id": "string",
  "name": "string",
  "platform_id": "string",
  "version": "string",
  "status": "string",
  "config": {},
  "created_at": "string",
  "updated_at": "string"
}
```

### 安装游戏卡片

- 路径: `/cards/{card_id}/install`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "task_id": "string",
  "status": "string"
}
```

### 卸载游戏卡片

- 路径: `/cards/{card_id}/uninstall`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "task_id": "string",
  "status": "string"
}
```

## 游戏实例管理

### 创建游戏实例

- 路径: `/instances`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "card_id": "string",
  "platform_id": "string",
  "config": {}
}
```

- 响应:

```json
{
  "id": "string",
  "status": "string",
  "task_id": "string"
}
```

### 获取实例状态

- 路径: `/instances/{instance_id}/status`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "id": "string",
  "status": "string",
  "resources": {
    "cpu": "float",
    "memory": "float",
    "disk": "float"
  },
  "created_at": "string",
  "updated_at": "string"
}
```

### 控制游戏实例

- 路径: `/instances/{instance_id}/control`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "action": "string" // start/stop/pause/resume
}
```

- 响应:

```json
{
  "status": "string",
  "task_id": "string"
}
```

### 删除游戏实例

- 路径: `/instances/{instance_id}`
- 方法: DELETE
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "task_id": "string",
  "status": "string"
}
```

## 任务管理

### 获取任务状态

- 路径: `/tasks/{task_id}`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "id": "string",
  "type": "string",
  "status": "string",
  "progress": "float",
  "result": {},
  "error": "string",
  "created_at": "string",
  "updated_at": "string"
}
```

## 错误响应

所有 API 在发生错误时都会返回以下格式：

```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": {}
  }
}
```

常见错误码：

- 400: 请求参数错误
- 401: 未认证
- 403: 权限不足
- 404: 资源不存在
- 409: 资源冲突
- 500: 服务器内部错误
