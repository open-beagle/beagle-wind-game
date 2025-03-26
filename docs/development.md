# 开发指南

## 后端 API 开发计划（与前端联调）

基于对项目结构的分析，我们需要实现以下 API 以与前端进行联调：

### 第一阶段：基础架构调整（1 天）

1. **创建服务层**

   - 在 internal 目录下创建 service 目录
   - 实现各模块的服务层代码
   - 实现数据访问层

2. **完善数据模型**

   - 根据前端模型更新内部模型
   - 确保模型字段与 API 文档一致
   - 添加必要的验证规则

3. **实现基础工具函数**
   - 分页处理
   - 错误处理
   - 数据格式转换

### 第二阶段：核心 API 实现（3-4 天）

#### 1. 游戏平台 API（高优先级）

- **实现平台列表 API**

  ```go
  // internal/api/platform_handler.go
  func List(c *gin.Context) {
      // 实现分页、搜索等功能
  }
  ```

- **实现平台详情 API**

  ```go
  // internal/api/platform_handler.go
  func getPlatform(c *gin.Context) {
      // 获取平台ID并返回详情
  }
  ```

- **实现远程访问 API**

  ```go
  // internal/api/platform_handler.go
  func GetAccess(c *gin.Context) {
      // 生成远程访问链接
  }

  func refreshPlatformAccess(c *gin.Context) {
      // 刷新远程访问链接
  }
  ```

#### 2. 游戏节点 API（高优先级）

- **实现节点列表 API**

  ```go
  // internal/api/node_handler.go
  func listGameNodes(c *gin.Context) {
      // 实现分页、搜索等功能
  }
  ```

- **实现节点详情 API**

  ```go
  // internal/api/node_handler.go
  func getGameNode(c *gin.Context) {
      // 获取节点ID并返回详情
  }
  ```

#### 3. 游戏卡片 API（中优先级）

- **实现卡片列表 API**

  ```go
  // internal/api/card_handler.go
  func listGameCards(c *gin.Context) {
      // 实现分页、搜索等功能
  }
  ```

- **实现卡片详情 API**

  ```go
  // internal/api/card_handler.go
  func getGameCard(c *gin.Context) {
      // 获取卡片ID并返回详情
  }
  ```

#### 4. 游戏实例 API（中优先级）

- **实现实例列表 API**

  ```go
  // internal/api/instance_handler.go
  func listInstances(c *gin.Context) {
      // 实现分页、搜索等功能
  }
  ```

- **实现实例控制 API**

  ```go
  // internal/api/instance_handler.go
  func controlInstance(c *gin.Context) {
      // 控制游戏实例（启动、停止等）
  }
  ```

### 第三阶段：辅助功能 API（2 天）

#### 1. 用户认证 API（低优先级）

- **实现登录 API**

  ```go
  // internal/api/auth_handler.go
  func login(c *gin.Context) {
      // 验证用户并生成token
  }
  ```

- **实现登出 API**

  ```go
  // internal/api/auth_handler.go
  func logout(c *gin.Context) {
      // 清除用户token
  }
  ```

#### 2. 任务管理 API（低优先级）

- **实现任务列表 API**

  ```go
  // internal/api/task_handler.go
  func listTasks(c *gin.Context) {
      // 获取任务列表
  }
  ```

- **实现任务状态 API**

  ```go
  // internal/api/task_handler.go
  func getTaskStatus(c *gin.Context) {
      // 获取任务状态
  }
  ```

### 第四阶段：联调与优化（2 天）

1. **与前端进行接口联调**

   - 解决数据格式问题
   - 调整接口参数
   - 修复兼容性问题

2. **优化接口性能**
   - 添加缓存机制
   - 优化数据查询
   - 加强错误处理

## 工作计划安排

### 第 1 天：

- 完善项目架构，添加缺失的服务层
- 调整数据模型以匹配前端模型
- 实现游戏平台列表 API 和详情 API

### 第 2 天：

- 实现游戏平台远程访问 API
- 实现游戏节点列表 API 和详情 API
- 进行基础联调测试

### 第 3 天：

- 实现游戏卡片 API
- 实现游戏实例列表 API
- 解决联调过程中的问题

### 第 4 天：

- 实现实例控制 API
- 实现用户认证 API（如有必要）
- 优化 API 响应格式

### 第 5 天：

- 实现任务管理 API（如有必要）
- 完成全部接口联调
- 编写 API 使用文档

## API 开发进度跟踪

| API              | 优先级 | 状态   | 负责人 | 完成时间 |
| ---------------- | ------ | ------ | ------ | -------- |
| 游戏平台列表     | 高     | 待开发 | -      | -        |
| 游戏平台详情     | 高     | 待开发 | -      | -        |
| 游戏平台远程访问 | 高     | 待开发 | -      | -        |
| 游戏节点列表     | 高     | 待开发 | -      | -        |
| 游戏节点详情     | 高     | 待开发 | -      | -        |
| 游戏卡片列表     | 中     | 待开发 | -      | -        |
| 游戏卡片详情     | 中     | 待开发 | -      | -        |
| 游戏实例列表     | 中     | 待开发 | -      | -        |
| 游戏实例控制     | 中     | 待开发 | -      | -        |
| 用户认证登录     | 低     | 待开发 | -      | -        |
| 用户认证登出     | 低     | 待开发 | -      | -        |
| 任务列表         | 低     | 待开发 | -      | -        |
| 任务状态         | 低     | 待开发 | -      | -        |

## 开发规范

### 命名规范

- **文件命名**：使用下划线命名法，如`platform_handler.go`
- **函数命名**：使用驼峰命名法，如`List`
- **变量命名**：使用驼峰命名法，如`platformID`
- **常量命名**：使用全大写下划线命名法，如`MAX_PAGE_SIZE`

### 代码格式

- 使用`gofmt`格式化代码
- 使用`golint`检查代码风格
- 使用`go vet`检查代码错误

### 注释规范

- 每个函数都需要添加注释
- 每个 API 处理函数都需要添加路由、参数和返回值的注释
- 复杂逻辑需要添加详细注释

### 错误处理

- 使用统一的错误响应格式
- 记录详细的错误日志
- 不在 API 响应中泄露敏感信息

### 测试规范

- 编写单元测试
- 使用模拟数据进行测试
- 测试覆盖率至少达到 80%
