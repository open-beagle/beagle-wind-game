# 系统调试指南

## 前端调试

### 开发环境启动

```bash
# 安装所有依赖
bash ./scripts/install.sh

# 启动开发服务器
bash ./scripts/dev.sh
```

开发服务器将在 <http://localhost:5173> 启动，支持热重载。
如果需要从其他设备访问，可以使用主机的 IP 地址，例如 <http://192.168.1.xxx:5173>。

### 构建部署

```bash
# 构建生产环境代码
bash ./scripts/build.sh
```

构建完成后，生成的文件将位于 `frontend/dist` 目录下。

### 数据源调试

系统支持在前端开发过程中，使用 Mock 数据或实际 API 数据进行调试。这使得前端开发可以独立于后端进行，提高开发效率。

#### 通过界面切换数据源

应用顶部导航栏提供了数据源切换开关，可以实时在 Mock 数据和 API 数据之间切换：

- 打开开关：使用 Mock 数据（默认状态）
- 关闭开关：使用 API 数据

#### 使用命令行参数启动 Mock 数据模式

可以通过添加 `--mock` 参数来启动带有 Mock 数据的开发环境：

```bash
# 使用 Mock 数据启动开发环境
bash ./scripts/dev.sh --mock
```

这会设置必要的环境变量，使前端默认使用 Mock 数据源。

#### 编程方式切换数据源

在开发过程中，您也可以在代码中通过以下方式控制数据源：

```typescript
// 在组件或服务中导入配置
import { toggleDataSource } from "@/config";

// 切换到 Mock 数据
toggleDataSource(true);

// 切换到 API 数据
toggleDataSource(false);
```

#### Mock 数据调试技巧

1. 延迟模拟

   可以调整模拟的网络延迟，以测试加载状态和用户体验：

   ```typescript
   // 在 dataService.ts 中修改默认延迟时间（毫秒）
   protected async mockDelay(ms: number = 300): Promise<void> {
     if (this.useMock()) {
       await new Promise(resolve => setTimeout(resolve, ms));
     }
   }
   ```

2. 模拟错误

   在开发过程中测试错误处理：

   ```typescript
   // 在服务中模拟错误情况
   if (this.useMock() && Math.random() > 0.7) {
     throw new Error("模拟网络错误");
   }
   ```

3. 数据持久化

   在使用 Mock 数据时模拟数据持久化：

   ```typescript
   // 在 services 中实现基本的本地存储持久化
   private getLocalData<T>(key: string, defaultData: T[]): T[] {
     if (this.useMock()) {
       const saved = localStorage.getItem(`mock_${key}`);
       return saved ? JSON.parse(saved) : defaultData;
     }
     return defaultData;
   }

   private saveLocalData<T>(key: string, data: T[]): void {
     if (this.useMock()) {
       localStorage.setItem(`mock_${key}`, JSON.stringify(data));
     }
   }
   ```

### 调试工具

1. Vue DevTools

   - 安装 [Vue.js devtools](https://github.com/vuejs/vue-devtools) 浏览器扩展
   - 用于调试 Vue 组件、状态管理和性能分析

2. 浏览器开发者工具
   - 使用 Chrome DevTools 的 Vue 面板
   - 查看组件树和组件状态
   - 监控网络请求和性能

### 调试技巧

1. 组件调试

   ```js
   // 在组件中添加断点
   debugger;

   // 使用 console 输出
   console.log("组件状态:", state);
   console.warn("警告信息");
   console.error("错误信息");
   ```

2. 状态管理调试

   ```js
   // 在 store 中添加监听
   store.$subscribe((mutation, state) => {
     console.log("状态变更:", mutation, state);
   });
   ```

3. 路由调试

   ```js
   // 监听路由变化
   router.beforeEach((to, from, next) => {
     console.log("路由变化:", { to, from });
     next();
   });
   ```

## 后端调试

### 开发环境启动

```bash
# 进入后端目录
cd backend

# 启动开发服务器
go run main.go
```

### 调试工具

1. Delve

   ```bash
   # 安装 Delve
   go install github.com/go-delve/delve/cmd/dlv@latest

   # 启动调试
   dlv debug main.go
   ```

2. 日志调试

   ```go
   // 使用结构化日志
   log.WithFields(log.Fields{
     "component": "game_node",
     "action": "start",
   }).Info("启动游戏节点")
   ```

### 调试技巧

1. API 调试

   - 使用 Postman 或 curl 测试 API
   - 查看请求和响应日志
   - 使用中间件记录请求信息

2. 数据库调试

   ```go
   // 开启 SQL 日志
   db.Debug().Find(&users)
   ```

3. 性能调试
   - 使用 pprof 进行性能分析
   - 监控内存使用
   - 分析 CPU 使用情况

## 常见问题

### 前端问题

1. 热重载不生效

   - 检查 vite.config.ts 配置
   - 确认文件监听是否正确

2. 组件不更新

   - 检查响应式数据定义
   - 确认 props 传递正确

3. 数据源切换不生效
   - 检查 localStorage 是否可用
   - 检查 config/index.ts 是否正确导入
   - 在控制台执行 `localStorage.getItem('useMockData')` 检查设置

### 后端问题

1. 数据库连接失败

   - 检查数据库配置
   - 确认数据库服务状态

2. API 响应慢
   - 检查数据库查询
   - 分析中间件性能
   - 查看日志定位问题

## 调试环境配置

### 开发环境变量

1. 前端环境变量

   ```env
   VITE_DEBUG=true
   VITE_USE_MOCK=true  # 控制是否默认使用 Mock 数据
   ```

2. 后端环境变量

   ```env
   DEBUG=true
   DB_PATH=./data/game.db
   ```

### 日志配置

1. 前端日志

   ```typescript
   // 配置日志级别
   const logLevel = import.meta.env.VITE_DEBUG ? "debug" : "info";
   ```

2. 后端日志

   ```go
   // 配置日志输出
   log.SetLevel(log.DebugLevel)
   log.SetFormatter(&log.JSONFormatter{})
   ```

## Protobuf 更新

当修改了 `.proto` 文件后，需要重新生成 Go 代码：

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/agent/proto/agent.proto
```
