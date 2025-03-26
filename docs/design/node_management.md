# 游戏节点管理设计

## 1. 核心功能

### 节点信息管理

- 节点基本信息（ID、名称、型号、类型、状态）管理
- 硬件配置（CPU、内存、GPU 等）记录
- 网络信息（IP、端口、带宽）维护
- 节点标签管理（区域、机架等）

### 节点 Agent 管理

- 每个节点都要部署一个 Agent，这个 Agent 负责节点和平台的所有通讯工作
- 实现业务流程，我将发一个 pipeline 给 Agent，其按照我给的业务流程下载镜像，配置数据与网络环境，最终启动一个容器
- 报告节点状态信息，监控 pipeline 的执行情况，获取日志，像平台反馈进度信息等。
- 报告节点监控信息，当前运行了多少游戏实例，节点的 CPU、GPU、硬盘、网络等使用量。

## 2. 数据模型

根据已有的`GameNode`模型：

```go
// GameNode 游戏节点
type GameNode struct {
    ID         string                 `json:"id" yaml:"id"`                   // 节点ID
    Name       string                 `json:"name" yaml:"name"`               // 节点名称
    Model      string                 `json:"model" yaml:"model"`             // 节点型号
    Type       string                 `json:"type" yaml:"type"`               // 节点类型（physical/virtual/container）
    Status     string                 `json:"status" yaml:"status"`           // 节点状态（online/offline/maintenance/ready）
    Location   string                 `json:"location" yaml:"location"`       // 节点地理位置
    Hardware   map[string]interface{} `json:"hardware" yaml:"hardware"`       // 硬件配置（CPU、RAM、GPU等）
    Network    map[string]interface{} `json:"network" yaml:"network"`         // 网络信息（IP、速度等）
    Resources  map[string]interface{} `json:"resources" yaml:"resources"`     // 资源使用情况
    Metrics    map[string]interface{} `json:"metrics" yaml:"metrics"`         // 监控指标
    Labels     map[string]string      `json:"labels" yaml:"labels"`           // 节点标签
    Online     bool                   `json:"online" yaml:"online"`           // 是否在线
    LastOnline time.Time              `json:"last_online" yaml:"last_online"` // 最后在线时间
    CreatedAt  time.Time              `json:"created_at" yaml:"created_at"`   // 创建时间
    UpdatedAt  time.Time              `json:"updated_at" yaml:"updated_at"`   // 更新时间
}
```

## 3. API 设计

已实现的 API：

- `GET /api/v1/nodes` - 获取节点列表
- `GET /api/v1/nodes/{node_id}` - 获取节点详情
- `POST /api/v1/nodes` - 创建新节点
- `PUT /api/v1/nodes/{node_id}` - 更新节点信息
- `DELETE /api/v1/nodes/{node_id}` - 删除节点
- `PUT /api/v1/nodes/{node_id}/status` - 更新节点状态

## 4. 存储层

节点信息存储在 YAML 文件中，通过`NodeStore`进行管理：

```go
// NodeStore 游戏节点存储
type NodeStore struct {
    dataFile string           // 节点数据文件路径
    nodes    []models.GameNode // 节点列表
    mu       sync.RWMutex      // 读写锁
}
```

存储层操作：

- `Load()` - 从文件加载节点数据
- `Save()` - 保存节点数据到文件
- `List()` - 获取所有节点
- `Get(id)` - 获取指定 ID 的节点
- `Add(node)` - 添加节点
- `Update(node)` - 更新节点
- `Delete(id)` - 删除节点

## 5. 服务层

`NodeService`提供节点管理的核心业务逻辑：

```go
// NodeService 游戏节点服务
type NodeService struct {
    nodeStore     *store.NodeStore
    instanceStore *store.InstanceStore
}
```

服务层方法：

- `ListNodes(params)` - 获取节点列表
- `GetNode(id)` - 获取节点详情
- `CreateNode(node)` - 创建节点
- `UpdateNode(id, node)` - 更新节点
- `DeleteNode(id)` - 删除节点
- `UpdateNodeStatus(id, status, reason)` - 更新节点状态

## 6. 控制层

`NodeHandler`处理 HTTP 请求并调用服务层方法：

```go
// NodeHandler 节点API处理器
type NodeHandler struct {
    nodeService *service.NodeService
}
```

控制层方法：

- `ListNodes(c)` - 处理获取节点列表请求
- `GetNode(c)` - 处理获取节点详情请求
- `CreateNode(c)` - 处理创建节点请求
- `UpdateNode(c)` - 处理更新节点请求
- `DeleteNode(c)` - 处理删除节点请求
- `UpdateNodeStatus(c)` - 处理更新节点状态请求

## 7. 开发计划

1. [x] **节点信息管理**

   - [x] 完善节点结构定义
   - [x] 增强节点信息验证
   - [x] 优化节点信息展示

2. **节点 Agent 管理**

   - [x] 设计 Agent 通信协议和接口
     - [x] 定义 gRPC 服务和消息格式
     - [x] 设计双向通信机制
     - [x] 制定错误处理和重试策略
   - [x] 开发服务端通信组件
     - [x] 实现 gRPC 服务端功能
     - [x] 开发节点连接管理
     - [-] 实现消息分发和处理
   - [-] 设计节点 Agent 核心功能
     - [-] 设计 pipeline 的数据结构
       config/pipeline/platform-start.yaml
   - [] 开发节点 Agent 核心功能
     - [] 实现 pipeline 执行引擎
     - [] 开发资源监控和上报模块
     - [] 开发日志收集和传输功能
     - [] 实现容器生命周期管理

3. **接口扩展与优化**

   - [] 增加 Agent 相关 API
     - [] 添加 pipeline 部署接口
     - [] 设计节点监控数据查询接口
     - [] 开发日志检索和查询功能
   - [] 性能优化与安全加固
     - [] 优化通信效率
     - [] 增强认证和授权机制
     - [] 实现数据传输加密

4. **存储层单元测试**

   - [] 基础存储接口测试
     - [] 测试`Store`接口的基本CRUD操作
     - [] 测试并发读写操作
     - [] 测试错误处理和边界情况
   - [] 节点存储测试
     - [] 测试`NodeStore`的节点CRUD操作
     - [] 测试节点状态更新
     - [] 测试节点查询和过滤
     - [] 测试节点标签管理
   - [] 实例存储测试
     - [] 测试`InstanceStore`的实例CRUD操作
     - [] 测试实例状态转换
     - [] 测试实例资源分配
     - [] 测试实例查询和过滤
   - [] 游戏卡存储测试
     - [] 测试`GameCardStore`的游戏卡CRUD操作
     - [] 测试游戏卡状态管理
     - [] 测试游戏卡查询和过滤
     - [] 测试游戏卡版本管理
   - [] 平台存储测试
     - [] 测试`PlatformStore`的平台CRUD操作
     - [] 测试平台配置管理
     - [] 测试平台状态管理
     - [] 测试平台查询和过滤
   - [] 集成测试
     - [] 测试存储层与文件系统的交互
     - [] 测试存储层的并发性能
     - [] 测试存储层的错误恢复能力

## 8. 已完成工作

### 通信协议设计

我们设计并实现了基于 gRPC 的节点 Agent 通信协议，包括：

1. **通信架构设计**：客户端-服务器架构，使用 Protocol Buffers 定义消息格式
2. **服务接口定义**：包括节点注册、心跳维持、Pipeline 执行、容器管理等功能
3. **双向通信机制**：支持请求-响应、服务器流、客户端流和双向流四种通信模式
4. **错误处理策略**：包括指数退避重试、会话恢复和任务幂等性设计

### 服务端组件实现

实现了节点 Agent 的服务端组件，包括：

1. **节点连接管理**：跟踪节点的注册和心跳状态
2. **节点注册与会话**：处理节点的注册请求，生成和管理会话 ID
3. **心跳管理**：处理节点心跳并监控节点在线状态
4. **事件订阅机制**：支持客户端订阅服务端事件的推送

### 客户端组件实现

实现了节点 Agent 的客户端组件，包括：

1. **连接管理**：自动连接服务端并保持会话
2. **心跳机制**：定期发送心跳保持连接
3. **重连机制**：连接断开时自动重连
4. **状态上报**：上报节点资源使用情况
5. **事件订阅**：接收服务端推送的事件通知

### 示例程序

创建了一个完整的示例程序，展示客户端和服务端的基本使用方法，包括：

1. **多模式支持**：可以单独运行服务端、客户端或同时运行两者
2. **命令行参数**：支持自定义配置服务器地址和节点 ID
3. **信号处理**：优雅地处理终止信号
