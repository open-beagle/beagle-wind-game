# GamePipeline 系统设计文档

## 1. 系统概述

GamePipeline 简称 Pipeline 系统是 Beagle Wind Game 平台的控制 GameNode 上运行流程的核心组件之一。
GameNode 流程管理核心组件之一，注意 Pipeline 是流程执行的模版，其提供 Status 属性来存储流程运行的进度与状态信息；
生命周期开始，注意 GamePipelineServer 发布 Pipeline 任务时，其生命周期开始的特征点；
生命周期结束，Pipeline 执行完毕，则其生命周期结束；注意虽然 Pipeline 最后一个执行步骤是启动一个容器，而这个容器的生命周期可能长达数小时，或数天，但是其不影响 Pipeline 进入生命周期的终点。

### 1.1 核心功能

- 多步骤任务编排：支持复杂的游戏平台部署流程
- 容器化任务执行：确保游戏运行环境的一致性
- 资源调度和管理：为游戏实例分配必要的系统资源
- 状态监控和同步：实时跟踪部署和运行状态
- 错误处理和恢复：确保游戏服务的可靠性

### 1.2 系统架构

Pipeline 是 GameNode 流程管理核心组件之一，GameNode 流程管理包含以下核心组件：

- GamePipelineServer - Pipeline 服务器：负责任务调度和状态管理
- GamePipelineAgent - Pipeline 客户端：负责容器执行和资源管理
- GamePipeline - Pipeline 定义模版：使用 YAML 格式描述任务流程

### 1.3 主要业务

注意 Pipeline 本身是一个流程模版，运行的 Pipeline 是一个流程实例，是其主要业务执行的信息载体，但是不负责执行具体业务，所以不要给 Pipeline 定义或实现管理主要业务生命周期的方法。

1. **GamePlatform 相关任务**

   - 启动一个 GamePlatform 供系统管理员维护
   - 启动一个 GamePlatform 供普通用户维护：如登录 Steam
   - 完成维护后上传本地数据，删除本地容器；

2. **GameInstance 相关任务**

   - 启动 GameInstance
   - 回收 GameInstance

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

### 3.1 系统构成

1. Pipeline 类型结构
   Pipeline 类型结构，由架构师在项目开始时设计并实现，主要包含以下内容

- 基础属性，name , description
- 参数定义，args，envs
- 执行步骤，steps
- 执行状态, status

2. Pipeline 的模版实例
   Pipeline 的模版实例存储在 config/pipeline 目录中，将有架构师在运行时根据业务不断的迭代。

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
./bin/agent -server localhost:50051

# 启动服务器
./bin/server -listen :50051

# 执行 Pipeline
curl -X POST http://localhost:8080/api/pipelines \
  -H "Content-Type: application/json" \
  -d @pipeline.yaml
```

## 2. 通信设计

### 2.1 gRPC 服务定义

Pipeline 系统使用 gRPC 进行通信，主要包含以下服务：

```protobuf
service GamePipelineGRPCService {
    // Pipeline 流式服务
    // 1. Agent 通过此服务保持与 Server 的连接
    // 2. Server 通过此服务下发 Pipeline 任务
    // 3. Server 通过此服务发送取消命令
    rpc PipelineStream(stream PipelineStreamRequest) returns (stream PipelineStreamResponse);
    
    // 更新流水线状态
    rpc UpdatePipelineStatus(UpdatePipelineStatusRequest) returns (UpdatePipelineStatusResponse);
    
    // 更新步骤状态
    rpc UpdateStepStatus(UpdateStepStatusRequest) returns (UpdateStepStatusResponse);
}
```

### 2.2 通信流程

1. **Agent 连接建立**：
   - Agent 启动后，通过 `PipelineStream` 建立与 Server 的长连接
   - 定期发送心跳消息保持连接活跃

2. **任务下发**：
   - Server 通过 `PipelineStream` 下发 Pipeline 任务
   - Agent 接收任务后开始执行

3. **状态上报**：
   - Agent 通过 `UpdatePipelineStatus` 上报整体状态
   - Agent 通过 `UpdateStepStatus` 上报步骤状态

4. **任务取消**：
   - Server 通过 `PipelineStream` 发送取消命令
   - Agent 接收命令后停止执行

### 2.3 故障恢复

1. **Server 重启**：
   - Server 从持久化存储中恢复 Pipeline 状态
   - 等待 Agent 重新连接

2. **Agent 重连**：
   - Agent 重新建立 `PipelineStream` 连接
   - 通过 `UpdatePipelineStatus` 上报当前状态
   - Server 根据状态更新恢复任务管理

## 3. 数据模型

### 3.1 Pipeline 模板

```protobuf
message GamePipeline {
    string id = 1;
    PipelineModel model = 2;
    
    // 静态信息（模板定义）
    string name = 3;
    string description = 4;
    repeated string envs = 5;
    repeated string args = 6;
    repeated PipelineStep steps = 7;
    
    // 动态信息（执行状态）
    PipelineStatus status = 8;
}
```

### 3.2 状态定义

```protobuf
// Pipeline 状态
enum PipelineState {
    PIPELINE_STATE_NOT_STARTED = 0;  // 未开始
    PIPELINE_STATE_PENDING = 1;      // 等待中
    PIPELINE_STATE_RUNNING = 2;      // 运行中
    PIPELINE_STATE_COMPLETED = 3;    // 已完成
    PIPELINE_STATE_FAILED = 4;       // 失败
    PIPELINE_STATE_CANCELED = 5;     // 取消
}

// 步骤状态
enum StepState {
    STEP_STATE_PENDING = 0;    // 等待中
    STEP_STATE_RUNNING = 1;    // 运行中
    STEP_STATE_COMPLETED = 2;  // 已完成
    STEP_STATE_FAILED = 3;     // 失败
    STEP_STATE_SKIPPED = 4;    // 已跳过
}
```

## 4. 后续工作

1. **实现 Server 端**：
   - 实现 `PipelineStream` 服务
   - 实现状态管理
   - 实现持久化存储

2. **实现 Agent 端**：
   - 实现 `PipelineStream` 客户端
   - 实现状态上报
   - 实现任务执行

3. **测试验证**：
   - 测试正常流程
   - 测试故障恢复
   - 测试性能指标
