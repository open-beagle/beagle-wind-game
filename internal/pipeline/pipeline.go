package pipeline

import (
	"context"

	"github.com/open-beagle/beagle-wind-game/internal/models"
)

// GamePipelineEngine 定义了游戏Pipeline执行引擎的核心接口
type GamePipelineEngine interface {
	// Start 启动执行引擎
	Start(ctx context.Context) error

	// Stop 停止执行引擎
	Stop(ctx context.Context) error

	// Execute 执行Pipeline任务
	Execute(ctx context.Context, pipeline *models.GamePipeline) error

	// GetStatus 获取Pipeline执行状态
	GetStatus(pipelineID string) (*models.PipelineStatus, error)

	// CancelPipeline 取消Pipeline执行
	CancelPipeline(pipelineID string, reason string) error
}

// EventType 定义事件类型
type EventType string

const (
	// PipelineStarted Pipeline开始执行
	PipelineStarted EventType = "PipelineStarted"
	// StepStarted 单个步骤开始执行
	StepStarted EventType = "StepStarted"
	// StepCompleted 单个步骤执行完成
	StepCompleted EventType = "StepCompleted"
	// StepFailed 单个步骤执行失败
	StepFailed EventType = "StepFailed"
	// PipelineCompleted Pipeline执行完成
	PipelineCompleted EventType = "PipelineCompleted"
	// PipelineFailed Pipeline执行失败
	PipelineFailed EventType = "PipelineFailed"
	// StepLog 步骤执行日志
	StepLog EventType = "StepLog"
	// ErrorLog 错误日志
	ErrorLog EventType = "ErrorLog"
)

// Event 定义事件结构
type Event struct {
	Type      EventType
	Pipeline  *models.GamePipeline
	Step      *models.PipelineStep
	Message   string
	Timestamp int64
}

// EventHandler 定义事件处理函数类型
type EventHandler func(event Event)
