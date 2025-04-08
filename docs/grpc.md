# gRPC 服务设计规范

## 1. 服务定义规范

### 1.1 命名规范

1. **服务命名**

   - 使用 PascalCase 命名服务
   - 服务名应该清晰表达其功能
   - 服务名应该以 Service 结尾

   ```protobuf
   service GameNodeService {}
   service PipelineService {}
   ```

2. **方法命名**

   - 使用 PascalCase 命名方法
   - 方法名应该清晰表达其功能
   - 方法名应该使用动词开头

   ```protobuf
   rpc Register(RegisterRequest) returns (RegisterResponse);
   rpc UpdateStatus(UpdateStatusRequest) returns (UpdateStatusResponse);
   ```

3. **消息命名**

   - 使用 PascalCase 命名消息
   - 请求消息以 Request 结尾
   - 响应消息以 Response 结尾

   ```protobuf
   message RegisterRequest {}
   message RegisterResponse {}
   ```

### 1.2 接口设计

1. **方法设计**

   - 每个方法应该只做一件事
   - 方法应该是幂等的
   - 避免过于复杂的方法签名
   - 使用标准的错误码

2. **消息设计**

   - 消息应该包含必要的字段
   - 使用合适的字段类型
   - 添加字段注释
   - 使用枚举类型表示状态

3. **流式设计**

   - 服务器流：用于推送数据
   - 客户端流：用于上传数据
   - 双向流：用于实时交互

## 2. 错误处理规范

### 2.1 错误码

1. **标准错误码**

   ```protobuf
   enum ErrorCode {
     OK = 0;
     INVALID_ARGUMENT = 1;
     NOT_FOUND = 2;
     ALREADY_EXISTS = 3;
     PERMISSION_DENIED = 4;
     UNAUTHENTICATED = 5;
     FAILED_PRECONDITION = 6;
     ABORTED = 7;
     OUT_OF_RANGE = 8;
     UNIMPLEMENTED = 9;
     INTERNAL = 10;
     UNAVAILABLE = 11;
     DATA_LOSS = 12;
   }
   ```

2. **错误消息**

   - 错误消息应该清晰明确
   - 包含错误原因和解决方案
   - 避免暴露敏感信息

### 2.2 错误处理策略

1. **重试策略**

   - 使用指数退避算法
   - 设置最大重试次数
   - 区分可重试和不可重试错误

2. **错误恢复**
   - 实现优雅降级
   - 提供回滚机制
   - 保持数据一致性

## 3. 安全规范

### 3.1 认证授权

1. **认证机制**

   - 使用 TLS 加密
   - 实现身份验证
   - 管理会话状态

2. **授权控制**

   - 基于角色的访问控制
   - 细粒度的权限管理
   - 操作审计日志

### 3.2 数据安全

1. **传输安全**

   - 使用 TLS 1.2+
   - 实现数据加密
   - 防止重放攻击

2. **存储安全**

   - 敏感数据加密
   - 安全的数据存储
   - 定期数据备份

## 4. 性能规范

### 4.1 连接管理

1. **连接池**

   - 复用连接
   - 动态调整池大小
   - 定期清理空闲连接

2. **负载均衡**

   - 实现负载均衡
   - 健康检查
   - 故障转移

3. **客户端连接最佳实践**

   - **使用 `grpc.NewClient` 而非过时的方法**

     从 gRPC v1.50.0 开始，官方推荐使用 `grpc.NewClient` 替代过时的 `grpc.Dial` 和 `grpc.DialContext` 方法。

     ```go
     // 过时的方式 - 不推荐
     conn, err := grpc.Dial(target, opts...)
     // 或
     conn, err := grpc.DialContext(ctx, target, opts...)

     // 推荐的方式
     conn, err := grpc.NewClient(target, opts...)
     ```

     **优势：**

     - 使用"DNS"作为默认名称解析器，而不是"passthrough"
     - 更好的连接延迟初始化
     - 更清晰的错误处理模型
     - 移除了过时的连接选项，如 `WithBlock()`

   - **正确使用连接状态管理**

     ```go
     // 推荐的连接状态检查方式
     state := conn.GetState()
     if state == connectivity.Ready {
       // 连接就绪，可以使用
     } else if state == connectivity.TransientFailure || state == connectivity.Shutdown {
       // 连接状态异常，需要处理
     }

     // 等待连接状态变化
     if !conn.WaitForStateChange(ctx, state) {
       // 超时或上下文取消
     }
     ```

   - **优雅的重试策略**

     实现指数退避算法进行重试：

     ```go
     maxRetries := 5
     initialBackoff := 500 * time.Millisecond
     maxBackoff := 10 * time.Second

     for i := 0; i < maxRetries; i++ {
       if i > 0 {
         backoff := initialBackoff * time.Duration(1<<uint(i-1))
         if backoff > maxBackoff {
           backoff = maxBackoff
         }
         time.Sleep(backoff)
       }

       // 尝试连接
       conn, err := grpc.NewClient(target, opts...)
       if err == nil {
         // 连接成功
         return conn, nil
       }
     }
     ```

   - **资源清理**

     确保在不再需要连接时关闭它：

     ```go
     defer conn.Close()
     ```

4. **与服务集成的最佳实践**

   - **使用服务特定的客户端**

     在建立连接后，使用生成的服务客户端而不是直接使用连接：

     ```go
     conn, err := grpc.NewClient(target, opts...)
     if err != nil {
       return err
     }

     // 创建服务特定的客户端
     client := pb.NewMyServiceClient(conn)

     // 使用客户端调用方法
     response, err := client.MyMethod(ctx, request)
     ```

   - **连接和客户端生命周期管理**

     注意连接和客户端的生命周期管理，避免过早关闭或资源泄漏。

### 4.2 资源管理

1. **内存管理**

   - 控制消息大小
   - 实现内存限制
   - 优化垃圾回收

2. **并发控制**

   - 限制并发请求
   - 实现背压机制
   - 防止资源耗尽

## 5. 监控规范

### 5.1 指标收集

1. **系统指标**

   - 请求延迟
   - 错误率
   - 并发数
   - 资源使用

2. **业务指标**

   - 业务成功率
   - 业务延迟
   - 业务量统计

### 5.2 日志管理

1. **日志记录**

   - 结构化日志
   - 分级日志
   - 关键信息记录

2. **日志处理**

   - 日志聚合
   - 日志分析
   - 告警机制

## 6. 测试规范

### 6.1 单元测试

1. **测试覆盖**

   - 核心功能测试
   - 边界条件测试
   - 错误场景测试

2. **测试质量**

   - 测试代码质量
   - 测试可维护性
   - 测试可读性

### 6.2 集成测试

1. **测试环境**

   - 模拟依赖
   - 环境隔离
   - 数据准备

2. **测试流程**

   - 端到端测试
   - 性能测试
   - 压力测试

## 7. 文档规范

### 7.1 接口文档

1. **文档内容**

   - 接口说明
   - 参数说明
   - 返回值说明
   - 错误码说明

2. **文档格式**

   - 使用 Markdown
   - 包含示例代码
   - 保持文档更新

### 7.2 示例代码

1. **代码示例**

   - 基本用法
   - 高级用法
   - 错误处理

2. **示例质量**

   - 代码可运行
   - 注释完整
   - 最佳实践

## 8. gRPC 版本兼容性指南

### 8.1 Go gRPC 客户端 API 变更

gRPC Go 客户端 API 在 1.50.0 版本中引入了几项重要变更：

1. **`grpc.NewClient` 的引入**

   - 在 gRPC v1.50.0+ 中，`grpc.NewClient` 被引入作为创建 gRPC 客户端连接的推荐方式
   - `grpc.Dial` 和 `grpc.DialContext` 被标记为过时，但会在整个 1.x 版本中继续支持
   - 项目应该迁移到 `grpc.NewClient` 以获得更好的默认行为和未来兼容性

2. **已弃用的选项和 API**

   - `WithBlock()` 选项已弃用，应使用 `Connect()` 和 `WaitForStateChange()` 代替
   - `DialOption` 中的 `WithTimeout` 和 `WithReturnConnectionError` 在 `NewClient` 中会被忽略

3. **行为差异**

   | 功能       | `grpc.Dial`         | `grpc.NewClient`                    |
   | ---------- | ------------------- | ----------------------------------- |
   | 默认解析器 | passthrough         | dns                                 |
   | 连接启动   | 立即启动            | 延迟到首次 RPC 调用或显式 Connect() |
   | 错误处理   | 混合模式            | 更清晰的错误模型                    |
   | 阻塞行为   | 通过 WithBlock 选项 | 通过 Connect 和 WaitForStateChange  |

4. **迁移指南**

   - 对于使用 `WithBlock(true)` 的代码：

     ```go
     // 旧代码
     conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithBlock())

     // 新代码
     conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
     if err != nil {
       return err
     }
     state := conn.GetState()
     if state != connectivity.Ready {
       conn.Connect()
       if !conn.WaitForStateChange(ctx, state) {
         return ctx.Err()
       }
     }
     ```

   - 对于简单连接的代码：

     ```go
     // 旧代码
     conn, err := grpc.Dial(target, grpc.WithInsecure())

     // 新代码
     conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
     ```

### 8.2 版本检查与兼容性保证

在项目中实施 gRPC 客户端更改时，应注意以下事项：

1. **检查当前环境的 gRPC 版本**

   通过 `go list -m google.golang.org/grpc` 命令确定项目中使用的 gRPC 版本。

2. **测试兼容性**

   在迁移到 `grpc.NewClient` 前，应在测试环境中验证新 API 是否可用及其行为。

3. **渐进式迁移**

   - 在新代码中使用 `NewClient`
   - 分阶段更新现有代码
   - 确保有适当的回滚机制

4. **监控与审计**

   迁移后监控连接成功率、延迟和错误率，以确保新 API 按预期工作。
