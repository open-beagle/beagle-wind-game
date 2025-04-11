# 日志框架规范

## 1. 概述

本规范旨在统一 server 和 agent 的日志输出，确保日志格式一致、级别明确、输出目标可配置，便于问题排查和系统监控。

**实现位置**: 日志框架实现位于 `internal/utils/log.go`

**技术选型**:

- 日志库: [zap](https://github.com/uber-go/zap) - Uber 开发的高性能、结构化、分级日志库
- 日志轮转: [lumberjack](https://github.com/natefinch/lumberjack) - 支持日志文件轮转的 Go 库

**应用范围**:

- **Agent**: 完全使用日志框架进行日志输出
- **Server**:
  - Gin (HTTP) 服务: 使用 Gin 自带的日志系统，不由日志框架管理
  - 其他所有组件和服务: 全部使用日志框架进行日志输出，包括但不限于：
    - gRPC 服务
    - 存储层 (store)
    - 服务层 (service)
    - 事件系统
    - 后台任务
    - 启动和关闭流程

## 2. 日志级别

日志级别从低到高依次为：

| 级别  | 说明     | 使用场景                                         |
| ----- | -------- | ------------------------------------------------ |
| DEBUG | 调试信息 | 详细的系统运行信息，仅在开发和调试环境开启       |
| INFO  | 一般信息 | 系统正常运行的状态信息，如启动、关闭、配置加载等 |
| WARN  | 警告信息 | 不影响系统运行但需要关注的异常情况               |
| ERROR | 错误信息 | 系统运行出现错误，但不影响整体功能               |
| FATAL | 严重错误 | 导致系统无法正常运行的严重错误                   |

## 3. 日志配置

### 3.1 环境变量配置

服务启动时可以读取环境变量，然后通过 `InitLogger` 方法设置配置：

```bash
# 可在服务启动脚本中设置环境变量
export LOG_LEVEL=INFO
export LOG_FILE_PATH=./logs/app.log
```

### 3.2 命令行配置

Agent 和 Server 均支持通过命令行参数配置日志行为：

```bash
# Agent 示例
./agent --node-id=node1 --server-addr=127.0.0.1:50051 --log-level=INFO --log-file=./logs/agent.log --log-both=true

# Server 示例
./server --http=:8080 --grpc=:50051 --log-level=INFO --log-file=./logs/server.log --log-both=true
```

可用的命令行参数：

| 参数        | 说明                                           | 默认值   |
| ----------- | ---------------------------------------------- | -------- |
| `log-level` | 日志级别: DEBUG, INFO, WARN, ERROR, FATAL      | INFO     |
| `log-file`  | 日志文件路径，为空则只输出到控制台             | (空)     |
| `log-both`  | 是否同时输出到控制台和文件                     | false    |

### 3.3 代码初始化配置

通过调用 `InitLogger` 方法在代码中初始化日志配置：

```go
// 参数说明:
// logFile: 日志文件路径，为空则仅输出到控制台
// level: 日志级别
// enableBoth: 是否同时输出到控制台和文件，仅当logFile不为空时有效

// 示例1: 仅输出到控制台 (默认行为)
utils.InitLogger("", utils.INFO, false)

// 示例2: 仅输出到文件
utils.InitLogger("./logs/app.log", utils.INFO, false)

// 示例3: 同时输出到控制台和文件
utils.InitLogger("./logs/app.log", utils.DEBUG, true)
```

### 3.4 日志输出目标

日志支持以下输出目标：

1. 控制台（标准输出/标准错误）- 默认模式
2. 日志文件
3. 同时输出到控制台和文件（需通过 `enableBoth` 参数或 `--log-both` 命令行参数开启）

输出逻辑：
- 如果未指定日志文件路径，则仅输出到控制台
- 如果指定了日志文件路径，默认仅输出到文件
- 如果指定了日志文件路径且开启 `enableBoth`，则同时输出到控制台和文件

## 4. 日志格式

### 4.1 基本格式

```txt
[时间戳] [级别] [模块] [请求ID] [消息]
```

示例：

```txt
[2023-05-16T14:30:45.123Z] [INFO] [UserService] [req-123456] 用户登录成功，用户ID: 10001
```

### 4.2 日志字段说明

| 字段    | 说明                                       | 示例                         |
| ------- | ------------------------------------------ | ---------------------------- |
| 时间戳  | ISO 8601 格式的时间，精确到毫秒            | 2023-05-16T14:30:45.123Z     |
| 级别    | 日志级别                                   | INFO                         |
| 模块    | 产生日志的模块或服务名称                   | UserService                  |
| 请求 ID | 用于追踪单次请求的唯一标识，多条日志可关联 | req-123456                   |
| 消息    | 具体的日志内容                             | 用户登录成功，用户 ID: 10001 |

## 5. 日志文件管理

### 5.1 日志轮转

使用 lumberjack 库实现以下日志轮转功能：

- 按大小轮转：单个日志文件超过指定大小（默认 100MB）时创建新文件
- 备份数量限制：默认保留最近 5 个备份文件
- 压缩备份：可选择是否压缩旧日志文件以节省空间

### 5.2 日志保留策略

- 默认保留最近 30 天的日志文件
- 通过 `LoggerConfig` 配置可详细定制文件轮转策略

## 6. 实现规范

### 6.1 接口定义

日志接口应包含以下方法：

```go
// Go语言示例
type Logger interface {
    Debug(format string, args ...interface{})
    Info(format string, args ...interface{})
    Warn(format string, args ...interface{})
    Error(format string, args ...interface{})
    Fatal(format string, args ...interface{})
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
    WithRequestID(requestID string) Logger
    Sync() error
}
```

### 6.2 使用方式

```go
// Go语言示例
import "github.com/open-beagle/beagle-wind-game/internal/utils"

func main() {
    // 初始化日志配置
    utils.InitLogger("./logs/app.log", utils.INFO, true)
    
    // 获取默认日志器
    log := utils.GetLogger()
    
    // 创建特定模块的日志器
    moduleLog := utils.New("ModuleName")
    
    // 基本用法
    log.Info("系统启动成功")
    
    // 带参数
    log.Info("用户 %s 登录成功", username)
    
    // 带请求ID
    requestLog := log.WithRequestID("req-123456")
    requestLog.Info("处理请求")
    
    // 带额外字段
    log.WithField("user_id", 10001).Info("用户操作")
    
    // 使用自定义配置创建日志器
    customConfig := utils.LoggerConfig{
        Level:      utils.INFO,
        Output:     utils.BOTH,
        FilePath:   "./logs/custom.log",
        Module:     "CustomModule",
        MaxSize:    50,    // 50MB
        MaxAge:     7,     // 7天
        MaxBackups: 3,     // 保留3个备份
        Compress:   true,  // 压缩旧日志
    }
    customLog, err := utils.NewWithConfig(customConfig)
    if err != nil {
        panic(err)
    }
    customLog.Info("使用自定义配置的日志器")
}
```

### 6.3 代码集成

#### Agent 集成

在 `cmd/agent/main.go` 中，已完全替换标准日志库为日志框架：

```go
// 示例摘录
func main() {
    // 解析命令行参数
    nodeID := flag.String("node-id", "", "节点ID")
    serverAddr := flag.String("server-addr", "", "服务器地址")
    logLevel := flag.String("log-level", "INFO", "日志级别: DEBUG, INFO, WARN, ERROR, FATAL")
    logFile := flag.String("log-file", "", "日志文件路径, 为空则只输出到控制台")
    logBoth := flag.Bool("log-both", false, "是否同时输出到文件和控制台")
    
    // 初始化日志框架
    logLevelMap := map[string]utils.LogLevel{...}
    level, ok := logLevelMap[*logLevel]
    if !ok {
        level = utils.INFO
    }
    utils.InitLogger(*logFile, level, *logBoth)
    logger := utils.New("Agent")
    
    // 使用日志
    logger.Info("Agent 启动中，版本: %s, 节点 ID: %s", version, *nodeID)
}
```

#### Server 集成

在 `cmd/server/main.go` 中，除了 Gin HTTP 服务器外，所有组件都应使用日志框架：

```go
// 示例摘录
func main() {
    // 解析命令行参数
    httpAddr := flag.String("http", ":8080", "HTTP服务器监听地址")
    grpcAddr := flag.String("grpc", ":50051", "gRPC服务器监听地址")
    logLevel := flag.String("log-level", "INFO", "日志级别: DEBUG, INFO, WARN, ERROR, FATAL")
    logFile := flag.String("log-file", "", "日志文件路径, 为空则只输出到控制台")
    logBoth := flag.Bool("log-both", false, "是否同时输出到文件和控制台")
    
    // 初始化日志框架
    utils.InitLogger(*logFile, level, *logBoth)
    logger := utils.New("Server")
    
    // 初始化存储
    logger.Info("初始化存储...")
    gamenodeStore, GamePipelineStore, gamePlatformStore, gameCardStore, gameInstanceStore, err := initStores()
    if err != nil {
        logger.Fatal("初始化存储失败: %v", err)
    }
    
    // Gin服务使用自己的日志系统
    router := gin.Default()
    
    // gRPC服务使用我们的日志框架
    grpcLogger := utils.New("gRPC")
    grpcLogger.Info("gRPC服务器开始监听 %s", *grpcAddr)
    
    // 所有其他组件也应使用日志框架
    storeLogger := utils.New("Store")
    serviceLogger := utils.New("Service")
}
```

#### 服务组件集成

每个服务组件都应创建自己的命名日志器实例：

```go
// 服务层示例
type GameService struct {
    store  store.GameStore
    logger utils.Logger
}

func NewGameService(store store.GameStore) *GameService {
    return &GameService{
        store:  store,
        logger: utils.New("GameService"),
    }
}

func (s *GameService) GetGame(id string) (*models.Game, error) {
    s.logger.Info("获取游戏，ID: %s", id)
    game, err := s.store.GetGame(id)
    if err != nil {
        s.logger.Error("获取游戏失败: %v", err)
        return nil, err
    }
    return game, nil
}
```

## 7. 最佳实践

1. 合理使用日志级别，避免过多 DEBUG 日志在生产环境输出
2. 错误日志应包含足够的上下文信息，便于问题排查
3. 敏感信息（如密码、token）不应在日志中明文显示
4. 结构化日志优于纯文本日志，便于后期分析
5. 在处理请求时，应在入口处生成请求 ID 并传递给后续处理流程
6. 利用 zap 的结构化日志功能，通过 WithField/WithFields 添加结构化数据
7. 在长时间运行的应用中定期调用 Sync() 确保日志写入磁盘
8. 在 server 和 agent 启动时调用 InitLogger 进行日志初始化
9. 在程序退出前调用 logger.Sync() 确保所有日志都已写入磁盘
10. 在服务组件（service, store等）内部创建命名日志器，便于识别日志来源
11. 确保除了Gin HTTP服务外，server中所有组件都使用日志框架，禁止使用标准库的log或fmt进行日志输出

## 8. 注意事项

1. 日志输出不应影响系统性能，高并发场景应考虑异步日志
2. 避免在热点代码路径中使用字符串拼接，应使用格式化方法
3. 日志文件应定期归档和清理，避免占用过多磁盘空间
4. 生产环境中的 ERROR 级别日志应配置告警机制
5. 在程序正常退出前应调用 Sync() 方法确保所有日志都写入磁盘
6. Gin HTTP 服务继续使用自身的日志系统，不需要引入我们的日志框架
7. 禁止在server代码中使用标准库的log包和fmt包直接输出日志，所有非Gin组件都必须使用日志框架
8. 各组件应根据其功能领域创建有意义的命名日志器，如"Store"、"Service"、"gRPC"等
