# 数据验证规则

## 通用规则

### 时间格式

- 所有时间字段必须使用 ISO 8601 格式
- 示例：`2024-03-20T10:30:00Z`

### 分页参数

- `page`: 正整数，默认值 1
- `size`: 正整数，范围 1-100，默认值 20

### ID 格式

- 节点 ID: 字母数字组合，长度 8-32 字符
- 平台 ID: 小写字母，长度 4-16 字符
- 卡片 ID: 平台 ID-发行日期（YYYYMMDD）
- 实例 ID: UUID v4 格式

## 游戏节点验证

### 节点基本信息

```json
{
  "name": {
    "type": "string",
    "required": true,
    "minLength": 2,
    "maxLength": 50,
    "pattern": "^[a-zA-Z0-9_-]+$"
  },
  "status": {
    "type": "string",
    "required": true,
    "enum": ["online", "offline", "maintenance"]
  }
}
```

### 硬件信息

```json
{
  "hardware": {
    "cpu": {
      "model": {
        "type": "string",
        "required": true
      },
      "cores": {
        "type": "integer",
        "required": true,
        "minimum": 1
      },
      "threads": {
        "type": "integer",
        "required": true,
        "minimum": 1
      }
    },
    "memory": {
      "total": {
        "type": "integer",
        "required": true,
        "minimum": 1024
      },
      "used": {
        "type": "integer",
        "required": true,
        "minimum": 0
      }
    },
    "gpu": {
      "model": {
        "type": "string",
        "required": true
      },
      "memory": {
        "type": "integer",
        "required": true,
        "minimum": 1024
      }
    },
    "disk": {
      "total": {
        "type": "integer",
        "required": true,
        "minimum": 10240
      },
      "used": {
        "type": "integer",
        "required": true,
        "minimum": 0
      }
    }
  }
}
```

### 网络信息

```json
{
  "network": {
    "ip": {
      "type": "string",
      "required": true,
      "format": "ipv4"
    },
    "port": {
      "type": "integer",
      "required": true,
      "minimum": 1,
      "maximum": 65535
    },
    "bandwidth": {
      "type": "integer",
      "required": true,
      "minimum": 0
    }
  }
}
```

## 游戏平台验证

### 平台基本信息

```json
{
  "name": {
    "type": "string",
    "required": true,
    "minLength": 2,
    "maxLength": 50
  },
  "version": {
    "type": "string",
    "required": true,
    "pattern": "^\\d+\\.\\d+\\.\\d+$"
  },
  "type": {
    "type": "string",
    "required": true,
    "enum": ["wine", "steam", "switch"]
  }
}
```

### 平台配置

```json
{
  "config": {
    "system": {
      "type": "object",
      "required": true,
      "additionalProperties": true
    },
    "user": {
      "type": "object",
      "required": true,
      "additionalProperties": true
    }
  }
}
```

## 游戏卡片验证

### 卡片基本信息

```json
{
  "name": {
    "type": "string",
    "required": true,
    "minLength": 2,
    "maxLength": 100
  },
  "platform_id": {
    "type": "string",
    "required": true,
    "pattern": "^[a-z]{4,16}$"
  },
  "version": {
    "type": "string",
    "required": true,
    "pattern": "^\\d+\\.\\d+\\.\\d+$"
  }
}
```

### 卡片配置

```json
{
  "config": {
    "type": "object",
    "required": true,
    "additionalProperties": true
  }
}
```

## 游戏实例验证

### 实例创建

```json
{
  "card_id": {
    "type": "string",
    "required": true,
    "pattern": "^[a-z]{4,16}-\\d{8}$"
  },
  "platform_id": {
    "type": "string",
    "required": true,
    "pattern": "^[a-z]{4,16}$"
  },
  "config": {
    "type": "object",
    "required": true,
    "additionalProperties": true
  }
}
```

### 实例控制

```json
{
  "action": {
    "type": "string",
    "required": true,
    "enum": ["start", "stop", "pause", "resume"]
  }
}
```

## 任务验证

### 任务状态

```json
{
  "status": {
    "type": "string",
    "required": true,
    "enum": ["pending", "running", "completed", "failed"]
  },
  "progress": {
    "type": "number",
    "required": true,
    "minimum": 0,
    "maximum": 100
  }
}
```

## 错误响应验证

### 错误格式

```json
{
  "error": {
    "code": {
      "type": "string",
      "required": true,
      "pattern": "^[A-Z_]+$"
    },
    "message": {
      "type": "string",
      "required": true,
      "minLength": 1
    },
    "details": {
      "type": "object",
      "required": true,
      "additionalProperties": true
    }
  }
}
```
