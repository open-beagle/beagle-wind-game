# Beagle Wind Game

## 项目简介

Beagle Wind Game 是一个基于 Go 语言开发的游戏云平台，支持多平台游戏运行和管理。

## 命名规范

### 核心业务领域命名

项目采用领域驱动设计(DDD)的思想，所有命名都应反映核心业务领域：

1. 游戏节点相关

   - 类型命名：以 `GameNode` 开头，如 `GameNode`、`GameNodeType`、`GameNodeState`
   - 常量命名：以 `GameNode` 开头，如 `GameNodeTypePhysical`、`GameNodeStateOnline`
   - 表名：使用小写下划线，如 `game_nodes`

2. 游戏平台相关

   - 类型命名：以 `GamePlatform` 开头，如 `GamePlatform`、`GamePlatformType`
   - 常量命名：以 `GamePlatform` 开头，如 `GamePlatformTypeSteam`
   - 表名：使用小写下划线，如 `game_platforms`

3. 游戏卡相关

   - 类型命名：以 `GameCard` 开头，如 `GameCard`、`GameCardType`
   - 常量命名：以 `GameCard` 开头，如 `GameCardTypeNormal`
   - 表名：使用小写下划线，如 `game_cards`

### 命名原则

1. 所有核心业务类型必须使用领域前缀（GameNode/GamePlatform/GameCard）
2. 避免使用通用名称（如 Node、Platform、Card）
3. 数据库表名统一使用小写下划线命名法
4. 常量值使用领域前缀 + 类型 + 具体值的形式
5. 接口命名应反映其业务职责，如 `AgentServerManager`、`GamePlatformService`

## 系统架构

### 核心组件

1. GameNode（游戏节点）

   - 静态属性
     - 基础属性：id，名称 name，模版 model，所属机房 location
     - 硬件配置 Hardware：CPU、RAM、GPU、Storage、Network
     - 网络配置 Network：IP、Speed
     - 通用属性：创建时间 CreatedAt，更新时间 UpdatedAt
     - 其他属性：节点标签、节点类型
   - 动态属性 Status
     - 节点状态：在线还是维护中，最后在线时间，累积在线时长
     - 节点资源：可以创建几个游戏实例，CPU 啊，内存啊，这些资源总量有多少
     - 节点指标：metrics 数据

2. GamePlatform（游戏平台）[详细设计](docs/design/gameplatform.md)

   - 静态属性
     - 基础属性：id，名称 name，版本 version，类型 type ，特性列表
     - 容器属性：镜像 Image，启动路径 Bin，操作系统 OS
     - 平台配置：wine , dxvk , vkd3d 等
     - 通用属性：创建时间 CreatedAt，更新时间 UpdatedAt
     - 文件与安装器
   - 动态属性 Status
     - 运行的游戏实例相关信息

3. GameCard（游戏卡片）[详细设计](docs/design/gamecard.md)

   - 静态属性
     - 游戏信息管理
     - 资源文件管理
     - 更新包管理
     - 权限控制
   - 动态属性 Status 暂无

4. GamePlatformUser (用户拥有的游戏平台配置)

   - 用户拥有自己的游戏平台配置

5. GameCardUser (用户拥有的游戏信息记录)

   - 用户拥有自己的游戏记录信息
     如：最后运行时间、游戏时长
   - 游戏存档
   - 游戏 Mods

6. GameInstance（游戏实例）[详细设计](docs/design/gameinstance.md)

   - 静态属性
     - 实例生命周期管理
     - 资源分配
     - 状态监控
     - 数据备份
   - 动态属性 Status 暂无

7. GameNodePipeline（流程管理）[详细设计](docs/design/gamenode_pipeline.md)

   - 静态属性
     - 基础属性：name、description
     - 参数定义：args、envs
     - 执行步骤：steps
   - 动态属性 Status
     - 执行状态：pending、running、completed、failed、canceled
   - 核心功能
     - 多步骤任务编排
     - 容器化任务执行
     - 资源调度和管理
     - 状态监控和同步
     - 错误处理和恢复

### 技术架构

1. 后端

   - Go + Gin
   - SQLite
   - WebSocket
   - Docker

2. 前端

   - Vue3 + TypeScript
   - Element Plus
   - Pinia
   - Vite

## 开发计划

### 第一阶段：基础框架搭建

1. 后端框架搭建

   - [x] 项目结构设计
   - [x] 数据库设计
   - [x] 基础中间件
   - [x] API 路由设计

2. 前端框架搭建

   - [x] 项目初始化
   - [x] 组件库集成
   - [x] 路由配置
   - [x] 状态管理

3. 数据模型设计

   - [x] 核心模型定义
   - [x] 数据关系设计
   - [x] 配置管理设计
   - [x] 数据验证规则

### 第二阶段：核心功能开发

1. 用户系统

   - [ ] 认证授权
   - [ ] 权限管理
   - [ ] 会话管理
   - [ ] 操作日志

2. 游戏管理

   - [x] 游戏平台管理
     - [x] 平台列表
     - [x] 平台配置
     - [x] 平台详情
     - [ ] 平台监控
   - [x] 游戏卡片管理
     - [x] 卡片列表
     - [x] 卡片配置
     - [x] 卡片详情
     - [ ] 卡片监控
   - [x] 游戏实例管理
     - [x] 实例列表
     - [x] 实例控制
     - [x] 实例详情
     - [ ] 实例监控
   - [ ] 游戏资源管理
   - [ ] 版本控制
   - [ ] 更新管理

3. 运行环境

   - [x] 节点管理
     - [x] 节点列表
     - [x] 节点配置
     - [x] 节点详情
     - [ ] 节点监控
   - [ ] 容器管理
   - [ ] 资源调度
   - [ ] 状态监控
   - [ ] 日志收集

### 第三阶段：功能完善

1. 监控系统

   - [ ] 性能监控
   - [ ] 资源监控
   - [ ] 告警系统
   - [ ] 统计分析

2. 运维工具

   - [ ] 部署脚本
   - [ ] 备份工具
   - [ ] 诊断工具
   - [ ] 维护工具

3. 文档完善

   - [x] 项目结构文档
   - [x] 前端开发文档
   - [x] 详情页开发文档
   - [x] 调试指南
   - [x] API 文档 (进行中)
   - [ ] 部署文档
   - [ ] 使用手册
   - [x] 开发指南 (进行中)

## 当前开发重点

1. 监控功能开发

   - [ ] 节点监控系统
   - [ ] 实例监控系统
   - [ ] 资源使用统计
   - [ ] 性能指标展示

2. 用户系统实现

   - [ ] 登录认证
   - [ ] 权限控制
   - [ ] 用户管理
   - [ ] 操作审计

3. API 对接

   - [x] 接口规范定义
   - [ ] 请求响应封装
   - [ ] 错误处理机制
   - [ ] 数据验证
   - [x] 后端 API 实现与前端联调 (进行中)
     - [ ] **高优先级**: 游戏平台 API（平台列表、详情、远程访问）
     - [ ] **高优先级**: 游戏节点 API（节点列表、详情）
     - [ ] **中优先级**: 游戏卡片 API（卡片列表、详情）
     - [ ] **中优先级**: 游戏实例 API（实例列表、控制、状态）
     - [ ] **低优先级**: 用户认证 API（登录、登出）
     - [ ] **低优先级**: 任务管理 API（任务列表、状态）

4. 性能优化

   - [ ] 前端构建优化
   - [ ] 组件懒加载
   - [ ] 数据缓存策略
   - [ ] 页面加载优化

## 文档

### 系统设计

- [系统总体设计](docs/design/summary.md)
- [节点 Agent 通信设计](docs/design/gamenode_communication.md)
- [游戏平台设计](docs/design/gameplatform.md)
- [游戏卡片设计](docs/design/gamecard.md)
- [游戏实例设计](docs/design/gameinstance.md)
- [Pipeline 系统设计](docs/design/gamenode_pipeline.md)
- [节点管理设计](docs/design/node_management.md)
- [数据模型设计](docs/models.md)
- [工作流程设计](docs/workflow.md)

### API 文档

- [API 接口文档](docs/api.md)
- [数据验证规则](docs/validation.md)

### 开发文档

- [前端开发文档](docs/frontend.md)
- [详情页开发文档](docs/detail-pages.md)
- [数据处理文档](docs/data.md)
- [数据库设计](docs/database.yaml)

### 指南文档

- [开发指南](docs/development.md)
- [系统调试指南](docs/debug.md)

### 其他文档

- [Logo 设计](docs/logo-design.md)
- [原型反馈](docs/prototype-feedback.md)

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交变更
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 项目结构

```txt
.
├── cmd/                    # 命令行入口
│   ├── server/            # 服务器入口
│   └── agent/             # 代理节点入口
├── internal/              # 内部包
│   ├── api/              # HTTP API 处理
│   ├── gamenode/         # 游戏节点核心逻辑
│   ├── models/           # 数据模型
│   └── store/            # 数据存储
├── pkg/                   # 公共包
│   ├── event/            # 事件系统
│   └── log/              # 日志系统
├── docs/                  # 文档
│   ├── design/           # 组件设计文档
│   └── grpc.md           # gRPC 服务设计规范
└── README.md             # 项目说明
```

## 文档结构说明

项目文档采用分层结构：

1. **根目录文档**

   - `README.md`: 项目整体说明
   - `docs/grpc.md`: 通用技术规范（如 gRPC 服务设计规范）

2. **设计文档**
   - `docs/design/`: 存放具体组件的设计文档
   - 每个组件都有独立的设计文档，如 `gamenode_service.md`、`pipeline_service.md` 等

这种结构的优势：

- 通用技术规范与具体组件设计分离，便于维护和查找
- 设计文档集中管理，便于团队协作
- 清晰的层次结构，符合直觉的文档组织方式
