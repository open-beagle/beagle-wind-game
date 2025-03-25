# Pipeline 系统设计文档

## 1. 系统概述

Pipeline 系统是 Beagle Wind Game 平台的核心组件之一，负责 [GamePlatform](../README.md#gameplatform) 的部署和运行环境准备。作为游戏平台的重要组成部分，Pipeline 系统提供了基于容器的任务编排和执行能力，为 [GameInstance](../README.md#gameinstance) 的运行提供基础设施支持。

### 1.1 核心功能

- 多步骤任务编排：支持复杂的游戏平台部署流程
- 容器化任务执行：确保游戏运行环境的一致性
- 资源调度和管理：为游戏实例分配必要的系统资源
- 状态监控和同步：实时跟踪部署和运行状态
- 错误处理和恢复：确保游戏服务的可靠性

### 1.2 系统架构

Pipeline 系统采用主从架构，包含以下核心组件：

- Pipeline 服务器：负责任务调度和状态管理
- Agent 节点：负责容器执行和资源管理
- Pipeline 定义：使用 YAML 格式描述任务流程

### 1.3 与其他组件的关系

1. **与 GameNode 的关系**
   - Pipeline 系统通过 Agent 在游戏节点上执行任务
   - 利用游戏节点的硬件资源运行容器
   - 监控游戏节点的资源使用情况

2. **与 GamePlatform 的关系**
   - 负责游戏平台的部署和配置
   - 管理平台运行环境
   - 处理平台更新和升级

3. **与 GameInstance 的关系**
   - 为游戏实例提供运行环境
   - 管理实例的资源分配
   - 处理实例的生命周期

## 2. Pipeline 定义

Pipeline 使用 YAML 格式定义，包含以下主要部分：

```yaml
name: "示例 Pipeline"
description: "这是一个示例 Pipeline"
envs:
  - ENV_VAR1
  - ENV_VAR2
args:
  - ARG1
  - ARG2
steps:
  - name: "步骤1"
    type: "container"
    container:
      image: "ubuntu:latest"
      container_name: "container1"
      hostname: "host1"
      privileged: true
      deploy:
        resources:
          reservations:
            devices:
              - capabilities: [gpu]
      security_opt:
        - seccomp=unconfined
      cap_add:
        - SYS_RAWIO
      tmpfs:
        - /dev/shm:rw
      devices:
        - /dev/dri:/dev/dri
      volumes:
        - /dev/input:/dev/input
        - /data/nvidia:/data/nvidia
        - /host/path:/container/path
      ports:
        - "8080:8080"
      environment:
        KEY: "value"
      command:
        - "echo"
        - "Hello World"
```

### 2.1 字段说明

- `name`: Pipeline 名称
- `description`: Pipeline 描述
- `envs`: 环境变量列表
- `args`: 运行时参数列表
- `steps`: 步骤列表
  - `name`: 步骤名称
  - `type`: 步骤类型（目前仅支持 container）
  - `container`: 容器配置
    - `image`: 容器镜像
    - `container_name`: 容器名称
    - `hostname`: 主机名
    - `privileged`: 是否使用特权模式
    - `deploy`: 部署配置
      - `resources`: 资源限制
        - `reservations`: 资源预留
          - `devices`: 设备预留（如 GPU）
    - `security_opt`: 安全选项
    - `cap_add`: 添加的 Linux 能力
    - `tmpfs`: 临时文件系统挂载
    - `devices`: 设备映射
    - `volumes`: 卷挂载
    - `ports`: 端口映射
    - `environment`: 环境变量
    - `command`: 执行命令

## 3. 系统架构

### 3.1 组件交互

1. Pipeline 服务器接收任务请求
2. 服务器选择合适的 Agent 节点
3. Agent 节点执行容器任务
4. Agent 节点向服务器报告状态
5. 服务器更新任务状态

### 3.2 状态流转

Pipeline 状态包括：
- `pending`: 等待执行
- `running`: 正在执行
- `completed`: 执行完成
- `failed`: 执行失败
- `canceled`: 已取消

## 4. 使用示例

### 4.1 创建 Pipeline

```yaml
name: "游戏服务器部署"
description: "部署游戏服务器和相关服务"
envs:
  - BEAGLE_WIND_ROOT
  - BEAGLE_WIND_PASSWD
  - BEAGLE_WIND_TURN_HOST
  - BEAGLE_WIND_TURN_PORT
  - BEAGLE_WIND_TURN_PROTOCOL
  - BEAGLE_WIND_TURN_USERNAME
  - BEAGLE_WIND_TURN_PASSWORD
  - S3_ACCESS_KEY
  - S3_SECRET_KEY
  - S3_BUCKET
  - S3_URL
args:
  - PLATFORM
  - INSTANCE
  - IMAGE
  - PORT
  - HOSTNAME
steps:
  - name: "启动数据库"
    type: "container"
    container:
      image: "mysql:8.0"
      container_name: "game-db"
      environment:
        MYSQL_ROOT_PASSWORD: "secret"
      deploy:
        resources:
          reservations:
            devices:
              - capabilities: [gpu]
      volumes:
        - /data/mysql:/var/lib/mysql
      ports:
        - "3306:3306"

  - name: "启动游戏服务器"
    type: "container"
    container:
      image: "game-server:latest"
      container_name: "game-server"
      privileged: true
      deploy:
        resources:
          reservations:
            devices:
              - capabilities: [gpu]
      security_opt:
        - seccomp=unconfined
      cap_add:
        - SYS_RAWIO
      tmpfs:
        - /dev/shm:rw
      devices:
        - /dev/dri:/dev/dri
      volumes:
        - /dev/input:/dev/input
        - /data/nvidia:/data/nvidia
        - /data/game:/app/data
      ports:
        - "8080:8080"
      environment:
        DB_HOST: "localhost"
        DB_PORT: "3306"
        TZ: "Asia/Shanghai"
        DISPLAY_SIZEW: 1920
        DISPLAY_SIZEH: 1080
        DISPLAY_REFRESH: 60
        DISPLAY_DPI: 96
        DISPLAY_CDEPTH: 24
        SELKIES_ENCODER: nvh264enc
        SELKIES_VIDEO_BITRATE: 1000
        SELKIES_FRAMERATE: 30
        SELKIES_AUDIO_BITRATE: 24000
        SELKIES_ENABLE_RESIZE: "false"
        BEAGLE_ENABLE_DEBUG: "true"
        PASSWD: ${{ envs.BEAGLE_WIND_PASSWD }}
        SELKIES_BASIC_AUTH_PASSWORD: ${{ envs.BEAGLE_WIND_PASSWD }}
        SELKIES_ENABLE_HTTPS: "false"
        SELKIES_TURN_HOST: ${{ envs.BEAGLE_WIND_TURN_HOST }}
        SELKIES_TURN_PORT: ${{ envs.BEAGLE_WIND_TURN_PORT }}
        SELKIES_TURN_PROTOCOL: ${{ envs.BEAGLE_WIND_TURN_PROTOCOL }}
        SELKIES_TURN_USERNAME: ${{ envs.BEAGLE_WIND_TURN_USERNAME }}
        SELKIES_TURN_PASSWORD: ${{ envs.BEAGLE_WIND_TURN_PASSWORD }}
```

### 4.2 执行 Pipeline

```bash
# 启动 Agent
./agent -server localhost:50051

# 启动服务器
./server -listen :50051

# 执行 Pipeline
curl -X POST http://localhost:8080/api/pipelines \
  -H "Content-Type: application/json" \
  -d @pipeline.yaml
```

## 5. 注意事项

1. 资源限制
   - 确保节点有足够的资源执行容器
   - 合理设置 CPU 和内存限制
   - 注意 GPU 资源的分配

2. 网络配置
   - 正确配置端口映射
   - 注意网络安全性
   - 确保 TURN 服务器配置正确

3. 存储管理
   - 合理规划卷挂载
   - 注意数据持久化
   - 确保共享存储的访问权限

4. 错误处理
   - 设置合适的重试策略
   - 配置错误通知机制
   - 做好日志记录

## 6. 未来规划

1. 功能增强
   - 支持更多步骤类型
   - 添加条件分支
   - 支持并行执行
   - 增加资源预检

2. 性能优化
   - 优化资源调度
   - 改进状态同步
   - 提升执行效率
   - 优化容器启动时间

3. 可用性提升
   - 完善监控指标
   - 增强日志管理
   - 提供更多管理工具
   - 改进错误处理机制 