package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/client"

	"github.com/open-beagle/beagle-wind-game/internal/grpc"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

// startGRPCClient 启动gRPC客户端
func startGRPCClient(
	ctx context.Context,
	nodeID string,
	serverAddr string,
) (*grpc.GameNodeAgent, error) {
	// 组件初始化
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		utils.GetLogger().Warn("无法创建 Docker 客户端: %v", err)
		dockerClient = nil
	}

	// 初始化gRPC客户端
	grpcOpts := &grpc.GameNodeOptions{
		HeartbeatPeriod: 30 * time.Second, // 默认值
		MetricsInterval: 5 * time.Second,  // 默认值
	}

	grpcClient, err := grpc.NewGameNodeAgent(
		ctx,
		nodeID,
		serverAddr,
		dockerClient,
		grpcOpts,
	)
	if err != nil {
		utils.GetLogger().Error("初始化gRPC客户端失败: %v", err)
		return nil, err
	}

	return grpcClient, nil
}

// startHeartbeat 启动心跳服务
func startHeartbeat(
	ctx context.Context,
	grpcClient *grpc.GameNodeAgent,
	interval time.Duration,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if grpcClient == nil {
				continue
			}
			if err := grpcClient.SendHeartbeat(ctx); err != nil {
				utils.GetLogger().Error("心跳发送失败: %v", err)
			}
		}
	}
}

// startMetrics 启动指标收集服务
func startMetrics(
	ctx context.Context,
	grpcClient *grpc.GameNodeAgent,
	interval time.Duration,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if grpcClient == nil {
				continue
			}
			if err := grpcClient.ReportMetrics(ctx); err != nil {
				utils.GetLogger().Error("指标更新失败: %v", err)
			}
		}
	}
}

// startBusinessProcessing 启动业务处理服务
func startBusinessProcessing(
	ctx context.Context,
	grpcClient *grpc.GameNodeAgent,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	// TODO: 实现业务处理逻辑
}

func main() {
	// 解析命令行参数
	nodeID := flag.String("node-id", "", "节点ID")
	serverAddr := flag.String("server-addr", "", "服务器地址")
	logLevel := flag.String("log-level", "INFO", "日志级别: DEBUG, INFO, WARN, ERROR, FATAL")
	logFile := flag.String("log-file", "", "日志文件路径, 为空则只输出到控制台")
	logBoth := flag.Bool("log-both", false, "是否同时输出到文件和控制台")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Beagle Wind Game Agent v%s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		return
	}

	// 初始化日志框架
	logLevelMap := map[string]utils.LogLevel{
		"DEBUG": utils.DEBUG,
		"INFO":  utils.INFO,
		"WARN":  utils.WARN,
		"ERROR": utils.ERROR,
		"FATAL": utils.FATAL,
	}
	level, ok := logLevelMap[*logLevel]
	if !ok {
		level = utils.INFO
	}
	utils.InitLogger(*logFile, level, *logBoth)
	logger := utils.New("Agent")

	// 验证必要参数
	if *nodeID == "" {
		logger.Fatal("必须指定节点ID (--node-id)")
	}
	if *serverAddr == "" {
		logger.Fatal("必须指定服务器地址 (--server-addr)")
	}

	// 打印启动信息
	logger.Info("Agent 启动中，版本: %s, 节点 ID: %s", version, *nodeID)
	logger.Info("连接到服务器: %s", *serverAddr)

	// 创建上下文，添加取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化gRPC客户端
	grpcClient, err := startGRPCClient(
		ctx,
		*nodeID,
		*serverAddr,
	)
	if err != nil {
		logger.Fatal("gRPC客户端初始化失败: %v", err)
	}
	logger.Info("gRPC客户端初始化成功")

	// 启动gRPC+注册节点
	if err := grpcClient.Start(ctx); err != nil {
		logger.Fatal("节点注册失败: %v", err)
	}
	logger.Info("节点注册成功")

	// 上报资源信息
	if err := grpcClient.ReportResource(ctx); err != nil {
		logger.Fatal("资源信息上报失败: %v", err)
	}
	logger.Info("资源信息上报成功")

	// 定义服务间隔时间
	heartbeatInterval := 30 * time.Second
	metricsInterval := 5 * time.Second

	// 启动子线程
	var wg sync.WaitGroup

	// 根据维护状态启动不同的服务
	switch grpcClient.GetState() {
	case models.GameNodeStaticStateNormal:
		logger.Info("节点状态为正常，启动所有服务")
		// 正常状态：启动所有服务
		wg.Add(3) // 心跳、指标和业务服务
		go startHeartbeat(ctx, grpcClient, heartbeatInterval, &wg)
		go startMetrics(ctx, grpcClient, metricsInterval, &wg)
		go startBusinessProcessing(ctx, grpcClient, &wg)

	case models.GameNodeStaticStateMaintenance:
		logger.Info("节点状态为维护中，仅启动基础服务")
		// 维护状态：只启动基础服务
		wg.Add(2) // 心跳和指标服务
		go startHeartbeat(ctx, grpcClient, heartbeatInterval, &wg)
		go startMetrics(ctx, grpcClient, metricsInterval, &wg)

	case models.GameNodeStaticStateDisabled:
		logger.Info("节点状态为已禁用，仅保持心跳")
		// 禁用状态：只保持心跳
		wg.Add(1) // 只启动心跳服务
		go startHeartbeat(ctx, grpcClient, heartbeatInterval, &wg)

	default:
		logger.Fatal("无效的节点状态: %v", grpcClient.GetState())
	}

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case sig := <-sigCh:
		logger.Info("收到停止信号: %v", sig)
	case <-ctx.Done():
		logger.Info("上下文已取消")
	}

	// 执行清理
	logger.Info("正在停止 Agent...")
	cancel()
	wg.Wait()

	// 确保日志写入磁盘
	if logger, ok := logger.(interface{ Sync() error }); ok {
		logger.Sync()
	}
}
