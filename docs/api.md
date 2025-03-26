# API 设计文档

## 设计规范

设计 HTTP API 接口时应遵守一下规范： 1.不允许使用 PUT、Patch、DELETE 方法，仅支持 GET、POST、OPTIONS 方法； 2.对应的 get、post、put、patch、delete 操作应遵循以下示例的要求：

- 查询操作: Get /instance/:id
- 新建操作: Post /instance/:id/create
- 更新操作: Post /instance/:id/update
- 删除操作: Post /instance/:id/delete

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
  - `keyword`: 搜索关键词 (可选)
  - `status`: 节点状态 (可选: online/offline/maintenance)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "type": "string",
      "status": "string",
      "network": {
        "ip": "string",
        "port": "integer",
        "protocol": "string",
        "bandwidth": "integer"
      },
      "resources": {
        "cpu": "integer",
        "memory": "integer",
        "storage": "integer",
        "network": "integer",
        "gpu": {
          "model": "string",
          "memory": "integer"
        }
      },
      "metrics": {
        "cpuUsage": "float",
        "memoryUsage": "float",
        "storageUsage": "float",
        "networkUsage": "float",
        "gpuUsage": "float"
      },
      "labels": {
        "region": "string",
        "zone": "string",
        "rack": "string"
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
  "description": "string",
  "type": "string",
  "status": "string",
  "network": {
    "ip": "string",
    "port": "integer",
    "protocol": "string",
    "bandwidth": "integer"
  },
  "resources": {
    "cpu": "integer",
    "memory": "integer",
    "storage": "integer",
    "network": "integer",
    "gpu": {
      "model": "string",
      "memory": "integer"
    }
  },
  "metrics": {
    "cpuUsage": "float",
    "memoryUsage": "float",
    "storageUsage": "float",
    "networkUsage": "float",
    "gpuUsage": "float",
    "uptime": "integer",
    "fps": "float",
    "instanceCount": "integer",
    "playerCount": "integer"
  },
  "instances": [
    {
      "id": "string",
      "name": "string",
      "status": "string",
      "gameCard": {
        "id": "string",
        "name": "string"
      }
    }
  ],
  "labels": {
    "region": "string",
    "zone": "string",
    "rack": "string"
  },
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
  ],
  "gpu": [
    {
      "timestamp": "string",
      "usage": "float"
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
  - `keyword`: 搜索关键词 (可选)
  - `status`: 平台状态 (可选: active/maintenance/inactive)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "name": "string",
      "version": "string",
      "os": "string",
      "status": "string",
      "description": "string",
      "image": "string",
      "created_at": "string",
      "updated_at": "string"
    }
  ]
}
```

### 创建平台

- 路径: `/platforms`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "name": "string",
  "version": "string",
  "os": "string",
  "description": "string",
  "image": "string",
  "bin": "string",
  "data": "string",
  "files": [
    {
      "type": "string",
      "url": "string"
    }
  ],
  "features": ["string"],
  "config": {},
  "installer": [
    {
      "command": "string"
    },
    {
      "move": {
        "src": "string",
        "dst": "string"
      }
    },
    {
      "chmodx": "string"
    },
    {
      "extract": {
        "file": "string",
        "dst": "string"
      }
    }
  ]
}
```

- 响应:

```json
{
  "id": "string",
  "name": "string",
  "status": "string"
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
  "os": "string",
  "status": "string",
  "description": "string",
  "image": "string",
  "bin": "string",
  "data": "string",
  "files": [
    {
      "id": "string",
      "type": "string",
      "url": "string"
    }
  ],
  "features": ["string"],
  "config": {
    "wine": "string",
    "dxvk": "string",
    "vkd3d": "string",
    "python": "string",
    "proton": "string",
    "shader-cache": "string",
    "remote-play": "string",
    "broadcast": "string",
    "mode": "string",
    "resolution": "string",
    "wifi": "string",
    "bluetooth": "string"
  },
  "installer": [
    {
      "command": "string"
    },
    {
      "move": {
        "src": "string",
        "dst": "string"
      }
    },
    {
      "chmodx": "string"
    },
    {
      "extract": {
        "file": "string",
        "dst": "string"
      }
    }
  ],
  "created_at": "string",
  "updated_at": "string"
}
```

### 更新平台

- 路径: `/platforms/{platform_id}`
- 方法: PUT
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "name": "string",
  "version": "string",
  "os": "string",
  "status": "string",
  "description": "string",
  "image": "string",
  "bin": "string",
  "data": "string",
  "files": [
    {
      "id": "string",
      "type": "string",
      "url": "string"
    }
  ],
  "features": ["string"],
  "config": {},
  "installer": [
    {
      "command": "string"
    },
    {
      "move": {
        "src": "string",
        "dst": "string"
      }
    },
    {
      "chmodx": "string"
    },
    {
      "extract": {
        "file": "string",
        "dst": "string"
      }
    }
  ]
}
```

- 响应: 204 No Content

### 删除平台

- 路径: `/platforms/{platform_id}`
- 方法: DELETE
- 请求头: `Authorization: Bearer <token>`
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

### 获取远程访问链接

- 路径: `/platforms/{platform_id}/access`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "link": "string",
  "expires_at": "string"
}
```

### 刷新远程访问链接

- 路径: `/platforms/{platform_id}/access/refresh`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "link": "string",
  "expires_at": "string"
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
  - `keyword`: 搜索关键词 (可选)
  - `platform_id`: 平台 ID (可选)
  - `status`: 状态 (可选: draft/published/archived)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "name": "string",
      "platform": {
        "id": "string",
        "name": "string"
      },
      "description": "string",
      "coverImage": "string",
      "status": "string",
      "created_at": "string"
    }
  ]
}
```

### 创建游戏卡片

- 路径: `/cards`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "name": "string",
  "platformId": "string",
  "description": "string",
  "coverImage": "string",
  "status": "string"
}
```

- 响应:

```json
{
  "id": "string",
  "name": "string",
  "status": "string"
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
  "platform": {
    "id": "string",
    "name": "string",
    "version": "string",
    "os": "string",
    "status": "string"
  },
  "description": "string",
  "coverImage": "string",
  "status": "string",
  "created_at": "string"
}
```

### 更新游戏卡片

- 路径: `/cards/{card_id}`
- 方法: PUT
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "name": "string",
  "description": "string",
  "coverImage": "string",
  "status": "string"
}
```

- 响应: 204 No Content

### 删除游戏卡片

- 路径: `/cards/{card_id}`
- 方法: DELETE
- 请求头: `Authorization: Bearer <token>`
- 响应: 204 No Content

### 安装游戏卡片

- 路径: `/cards/{card_id}/install`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "node_id": "string"
}
```

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
- 请求体:

```json
{
  "node_id": "string"
}
```

- 响应:

```json
{
  "task_id": "string",
  "status": "string"
}
```

## 游戏实例管理

### 获取实例列表

- 路径: `/instances`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `page`: 页码 (默认: 1)
  - `size`: 每页数量 (默认: 20)
  - `keyword`: 搜索关键词 (可选)
  - `game_card_id`: 游戏卡片 ID (可选)
  - `node_id`: 节点 ID (可选)
  - `status`: 状态 (可选: running/stopped/error)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "gameCard": {
        "id": "string",
        "name": "string"
      },
      "node": {
        "id": "string",
        "name": "string"
      },
      "status": "string",
      "created_at": "string",
      "updated_at": "string"
    }
  ]
}
```

### 创建游戏实例

- 路径: `/instances`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 请求体:

```json
{
  "gameCardId": "string",
  "nodeId": "string",
  "config": {
    "maxPlayers": "integer",
    "port": "integer",
    "settings": {
      "map": "string",
      "difficulty": "string"
    }
  }
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

### 获取实例详情

- 路径: `/instances/{instance_id}`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 响应:

```json
{
  "id": "string",
  "gameCard": {
    "id": "string",
    "name": "string",
    "platform": {
      "id": "string",
      "name": "string",
      "version": "string",
      "os": "string"
    }
  },
  "node": {
    "id": "string",
    "name": "string",
    "status": "string"
  },
  "status": "string",
  "config": {
    "maxPlayers": "integer",
    "port": "integer",
    "settings": {
      "map": "string",
      "difficulty": "string"
    }
  },
  "metrics": {
    "cpuUsage": "float",
    "memoryUsage": "float",
    "storageUsage": "float",
    "gpuUsage": "float",
    "networkUsage": "float",
    "uptime": "integer",
    "fps": "float",
    "playerCount": "integer"
  },
  "logs": ["string"],
  "created_at": "string",
  "updated_at": "string"
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
  "metrics": {
    "cpuUsage": "float",
    "memoryUsage": "float",
    "storageUsage": "float",
    "gpuUsage": "float",
    "networkUsage": "float",
    "uptime": "integer",
    "fps": "float",
    "playerCount": "integer"
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
  "action": "string" // start/stop/restart/pause/resume
}
```

- 响应:

```json
{
  "status": "string",
  "task_id": "string"
}
```

### 获取实例日志

- 路径: `/instances/{instance_id}/logs`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `lines`: 行数 (默认: 100)
  - `since`: 开始时间 (可选)
- 响应:

```json
{
  "logs": ["string"]
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

### 获取任务列表

- 路径: `/tasks`
- 方法: GET
- 请求头: `Authorization: Bearer <token>`
- 查询参数:
  - `page`: 页码 (默认: 1)
  - `size`: 每页数量 (默认: 20)
  - `status`: 任务状态 (可选: pending/running/completed/failed)
- 响应:

```json
{
  "total": "integer",
  "items": [
    {
      "id": "string",
      "type": "string",
      "status": "string",
      "progress": "float",
      "created_at": "string",
      "updated_at": "string"
    }
  ]
}
```

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

### 取消任务

- 路径: `/tasks/{task_id}/cancel`
- 方法: POST
- 请求头: `Authorization: Bearer <token>`
- 响应: 204 No Content

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
