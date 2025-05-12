package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// Engine 实现 GamePipelineEngine 接口
type Engine struct {
	logger       utils.Logger
	handlers     []EventHandler
	mu           sync.RWMutex
	runningPipes map[string]*models.GamePipeline
	containerMgr *ContainerManager
	eventQueue   chan Event
	done         chan struct{}
}

// NewEngine 创建新的执行引擎
func NewEngine() (*Engine, error) {
	containerMgr, err := NewContainerManager()
	if err != nil {
		return nil, fmt.Errorf("创建容器管理器失败: %w", err)
	}

	// 创建日志器
	logger, err := utils.NewWithConfig(utils.LoggerConfig{
		Level:  utils.DEBUG,
		Output: utils.CONSOLE,
		Module: "PipelineEngine",
	})
	if err != nil {
		return nil, fmt.Errorf("创建日志器失败: %w", err)
	}

	engine := &Engine{
		logger:       logger,
		handlers:     make([]EventHandler, 0),
		runningPipes: make(map[string]*models.GamePipeline),
		containerMgr: containerMgr,
		eventQueue:   make(chan Event, 1000), // 缓冲通道，避免阻塞
		done:         make(chan struct{}),
	}

	// 启动事件处理循环
	go engine.eventLoop()

	return engine, nil
}

// Start 启动执行引擎
func (e *Engine) Start(ctx context.Context) error {
	e.logger.Info("启动 Pipeline 执行引擎...")
	return nil
}

// Stop 停止执行引擎
func (e *Engine) Stop(ctx context.Context) error {
	e.logger.Info("停止 Pipeline 执行引擎...")
	close(e.done) // 停止事件处理循环
	return e.containerMgr.Close()
}

// Execute 执行Pipeline任务
func (e *Engine) Execute(ctx context.Context, pipeline *models.GamePipeline) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查Pipeline是否已经在运行
	if _, exists := e.runningPipes[pipeline.ID]; exists {
		e.logger.Error("Pipeline %s 已经在运行中", pipeline.ID)
		return fmt.Errorf("pipeline %s is already running", pipeline.ID)
	}

	// 初始化Pipeline状态
	now := time.Now()
	pipeline.Status = &models.PipelineStatus{
		State:      models.PipelineStateRunning,
		TotalSteps: int32(len(pipeline.Steps)),
		StartTime:  &now,
		Steps:      make([]models.StepStatus, len(pipeline.Steps)),
	}

	// 初始化每个步骤的状态
	for i := range pipeline.Status.Steps {
		pipeline.Status.Steps[i] = models.StepStatus{
			State: models.StepStatePending,
		}
		e.logger.Debug("初始化步骤 %d 状态为 Pending", i+1)
	}

	// 添加到运行中的Pipeline列表
	e.runningPipes[pipeline.ID] = pipeline
	e.logger.Debug("Pipeline %s 已添加到运行列表", pipeline.ID)

	// 发送Pipeline开始事件
	e.logger.Debug("准备发送 PipelineStarted 事件")
	e.emitEvent(Event{
		Type:      PipelineStarted,
		Pipeline:  pipeline,
		Timestamp: time.Now().Unix(),
	})
	e.logger.Debug("PipelineStarted 事件已发送")

	// 启动Pipeline执行
	e.logger.Info("启动 Pipeline %s 的异步执行", pipeline.ID)
	go func() {
		e.logger.Debug("开始执行 Pipeline goroutine")
		e.executePipeline(ctx, pipeline)
		e.logger.Debug("Pipeline goroutine 执行完成")
	}()

	return nil
}

// GetStatus 获取Pipeline执行状态
func (e *Engine) GetStatus(pipelineID string) (*models.PipelineStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pipeline, exists := e.runningPipes[pipelineID]
	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", pipelineID)
	}

	return pipeline.Status, nil
}

// CancelPipeline 取消Pipeline执行
func (e *Engine) CancelPipeline(pipelineID string, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	pipeline, exists := e.runningPipes[pipelineID]
	if !exists {
		return fmt.Errorf("pipeline %s not found", pipelineID)
	}

	// 更新Pipeline状态
	now := time.Now()
	pipeline.Status.State = models.PipelineStateFailed
	pipeline.Status.ErrorMessage = reason
	pipeline.Status.EndTime = &now

	// 发送Pipeline失败事件
	e.emitEvent(Event{
		Type:      PipelineFailed,
		Pipeline:  pipeline,
		Message:   reason,
		Timestamp: now.Unix(),
	})

	// 从运行中的Pipeline列表中移除
	delete(e.runningPipes, pipelineID)

	return nil
}

// RegisterHandler 注册事件处理器
func (e *Engine) RegisterHandler(handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers = append(e.handlers, handler)
}

// eventLoop 事件处理循环
func (e *Engine) eventLoop() {
	for {
		select {
		case event := <-e.eventQueue:
			e.mu.RLock()
			handlers := make([]EventHandler, len(e.handlers))
			copy(handlers, e.handlers)
			e.mu.RUnlock()

			// 异步处理每个事件处理器
			for _, handler := range handlers {
				go func(h EventHandler, evt Event) {
					defer func() {
						if r := recover(); r != nil {
							e.logger.Error("事件处理器发生panic: %v", r)
						}
					}()
					h(evt)
				}(handler, event)
			}
		case <-e.done:
			e.logger.Info("事件处理循环已停止")
			return
		}
	}
}

// emitEvent 发送事件
func (e *Engine) emitEvent(event Event) {
	select {
	case e.eventQueue <- event:
		// 事件已成功发送到队列
	case <-e.done:
		e.logger.Warn("尝试发送事件时引擎已停止")
	default:
		// 如果队列已满，记录警告但继续执行
		e.logger.Warn("事件队列已满，丢弃事件: %s", event.Type)
	}
}

// executePipeline 执行Pipeline
func (e *Engine) executePipeline(ctx context.Context, pipeline *models.GamePipeline) {
	e.logger.Info("开始执行 Pipeline %s 的步骤", pipeline.ID)
	defer func() {
		e.mu.Lock()
		delete(e.runningPipes, pipeline.ID)
		e.mu.Unlock()
		e.logger.Info("Pipeline %s 已从运行列表中移除", pipeline.ID)
	}()

	for i := range pipeline.Steps {
		step := &pipeline.Steps[i]
		e.logger.Info("准备执行步骤 %d/%d: %s", i+1, len(pipeline.Steps), step.Name)
		// 更新当前步骤
		pipeline.Status.CurrentStep = int32(i + 1)

		// 发送步骤开始事件
		e.emitEvent(Event{
			Type:      StepStarted,
			Pipeline:  pipeline,
			Step:      step,
			Timestamp: time.Now().Unix(),
		})

		// 执行步骤
		e.logger.Debug("开始执行步骤 %s", step.Name)
		err := e.executeStep(ctx, pipeline, step)
		now := time.Now()
		if err != nil {
			e.logger.Error("步骤 %s 执行失败: %v", step.Name, err)
			// 更新步骤状态为失败
			pipeline.Status.Steps[i].State = models.StepStateFailed
			pipeline.Status.Steps[i].Error = err.Error()
			pipeline.Status.Steps[i].EndTime = &now

			// 发送步骤失败事件
			e.emitEvent(Event{
				Type:      StepFailed,
				Pipeline:  pipeline,
				Step:      step,
				Message:   err.Error(),
				Timestamp: now.Unix(),
			})

			// 等待一小段时间，确保 StepFailed 事件被处理
			time.Sleep(100 * time.Millisecond)

			// 更新Pipeline状态为失败
			pipeline.Status.State = models.PipelineStateFailed
			pipeline.Status.ErrorMessage = err.Error()
			pipeline.Status.EndTime = &now

			// 发送Pipeline失败事件
			e.emitEvent(Event{
				Type:      PipelineFailed,
				Pipeline:  pipeline,
				Message:   err.Error(),
				Timestamp: now.Unix(),
			})

			return
		}

		e.logger.Info("步骤 %s 执行成功", step.Name)
		// 更新步骤状态为完成
		pipeline.Status.Steps[i].State = models.StepStateCompleted
		pipeline.Status.Steps[i].EndTime = &now

		// 发送步骤完成事件
		e.emitEvent(Event{
			Type:      StepCompleted,
			Pipeline:  pipeline,
			Step:      step,
			Timestamp: now.Unix(),
		})
	}

	// 更新Pipeline状态为完成
	now := time.Now()
	pipeline.Status.State = models.PipelineStateCompleted
	pipeline.Status.EndTime = &now
	e.logger.Info("Pipeline %s 所有步骤执行完成", pipeline.ID)

	// 发送Pipeline完成事件
	e.emitEvent(Event{
		Type:      PipelineCompleted,
		Pipeline:  pipeline,
		Timestamp: now.Unix(),
	})
}

// executeStep 执行单个步骤
func (e *Engine) executeStep(ctx context.Context, pipeline *models.GamePipeline, step *models.PipelineStep) error {
	e.logger.Debug("执行步骤 %s, 类型: %s", step.Name, step.Type)
	// 检查步骤类型
	switch step.Type {
	case "container":
		// 执行容器步骤
		e.logger.Debug("准备执行容器步骤: %s, 镜像: %s", step.Name, step.Container.Image)
		return e.containerMgr.RunContainer(ctx, step)
	default:
		e.logger.Error("不支持的步骤类型: %s", step.Type)
		return fmt.Errorf("不支持的步骤类型: %s", step.Type)
	}
}
