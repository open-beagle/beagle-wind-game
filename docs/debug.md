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
   VITE_API_BASE_URL=http://localhost:8080
   VITE_DEBUG=true
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
