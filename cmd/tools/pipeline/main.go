package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/pipeline"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

var (
	pipelineFile string
)

func init() {
	flag.StringVar(&pipelineFile, "pipeline", "test_pipeline.yaml", "Pipeline 定义文件路径")
	flag.Parse()
}

func main() {
	// 创建日志器
	logger, err := utils.NewWithConfig(utils.LoggerConfig{
		Level:  utils.DEBUG,
		Output: utils.CONSOLE,
		Module: "PipelineDebug",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建日志器失败: %v\n", err)
		os.Exit(1)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建执行引擎
	engine, err := pipeline.NewEngine()
	if err != nil {
		logger.Error("创建执行引擎失败: %v", err)
		os.Exit(1)
	}

	// 创建一个通道用于等待 Pipeline 完成
	pipelineDone := make(chan struct{})

	// 注册事件处理器
	engine.RegisterHandler(func(event pipeline.Event) {
		logger.Debug("收到事件: %s", event.Type)
		switch event.Type {
		case pipeline.PipelineStarted:
			logger.Info("Pipeline 开始执行: %s", event.Pipeline.Name)
		case pipeline.StepStarted:
			logger.Info("步骤开始执行: %s", event.Step.Name)
		case pipeline.StepCompleted:
			logger.Info("步骤执行完成: %s", event.Step.Name)
		case pipeline.StepFailed:
			logger.Error("步骤执行失败: %s, 错误: %s", event.Step.Name, event.Message)
		case pipeline.PipelineCompleted:
			logger.Info("Pipeline 执行完成: %s", event.Pipeline.Name)
			pipelineDone <- struct{}{}
		case pipeline.PipelineFailed:
			logger.Error("Pipeline 执行失败: %s, 错误: %s", event.Pipeline.Name, event.Message)
			pipelineDone <- struct{}{}
		default:
			logger.Debug("未知事件类型: %s", event.Type)
		}
	})

	// 启动执行引擎
	if err := engine.Start(ctx); err != nil {
		logger.Error("启动执行引擎失败: %v", err)
		os.Exit(1)
	}
	logger.Info("执行引擎启动成功")

	// 读取 Pipeline 定义文件
	pipelineData, err := os.ReadFile(pipelineFile)
	if err != nil {
		logger.Error("读取 Pipeline 定义文件失败: %v", err)
		os.Exit(1)
	}
	logger.Info("读取 Pipeline 定义文件成功")

	// 加载 Pipeline 定义
	pipeline, err := models.NewGamePipelineFromYAML(pipelineData)
	if err != nil {
		logger.Error("加载 Pipeline 定义失败: %v", err)
		os.Exit(1)
	}
	logger.Info("加载 Pipeline 定义成功")

	// 执行 Pipeline
	if err := engine.Execute(ctx, pipeline); err != nil {
		logger.Error("执行 Pipeline 失败: %v", err)
		os.Exit(1)
	}
	logger.Info("执行 Pipeline 成功")

	// 等待 Pipeline 完成或收到信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	logger.Debug("等待 Pipeline 完成或收到信号...")
	select {
	case <-pipelineDone:
		logger.Info("Pipeline 执行完成")
		// 等待一段时间，确保所有日志都被打印出来
		time.Sleep(time.Second)
	case <-sigCh:
		logger.Info("收到终止信号，正在停止...")
		if err := engine.Stop(ctx); err != nil {
			logger.Error("停止执行引擎失败: %v", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("上下文已取消，正在停止...")
	}

	// 等待所有 goroutine 完成
	logger.Debug("等待所有 goroutine 完成...")
	time.Sleep(2 * time.Second)
}
