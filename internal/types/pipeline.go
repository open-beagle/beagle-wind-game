package types

import (
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
)

// PipelineListParams Pipeline 列表查询参数
type PipelineListParams struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
	Status   string `form:"status"`
}

// PipelineListResult Pipeline 列表查询结果
type PipelineListResult struct {
	Total int64                  `json:"total"`
	Items []*models.GamePipeline `json:"items"`
}

// StepStatus 步骤状态
type StepStatus struct {
	ID        string    `json:"id" yaml:"id"`                 // 步骤ID
	Name      string    `json:"name" yaml:"name"`             // 步骤名称
	Status    string    `json:"status" yaml:"status"`         // 步骤状态
	StartTime time.Time `json:"start_time" yaml:"start_time"` // 开始时间
	EndTime   time.Time `json:"end_time" yaml:"end_time"`     // 结束时间
	Error     string    `json:"error" yaml:"error"`           // 错误信息
	Logs      []byte    `json:"logs" yaml:"logs"`             // 执行日志
	Progress  float64   `json:"progress" yaml:"progress"`     // 执行进度
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"` // 更新时间
}

// PipelineStatus 流水线状态
type PipelineStatus struct {
	ID          string       `json:"id" yaml:"id"`                     // 流水线ID
	NodeID      string       `json:"node_id" yaml:"node_id"`           // 节点ID
	Status      string       `json:"status" yaml:"status"`             // 流水线状态
	CurrentStep int32        `json:"current_step" yaml:"current_step"` // 当前步骤
	TotalSteps  int32        `json:"total_steps" yaml:"total_steps"`   // 总步骤数
	Progress    float64      `json:"progress" yaml:"progress"`         // 执行进度
	Error       string       `json:"error" yaml:"error"`               // 错误信息
	StartTime   time.Time    `json:"start_time" yaml:"start_time"`     // 开始时间
	EndTime     time.Time    `json:"end_time" yaml:"end_time"`         // 结束时间
	Steps       []StepStatus `json:"steps" yaml:"steps"`               // 步骤状态列表
	UpdatedAt   time.Time    `json:"updated_at" yaml:"updated_at"`     // 更新时间
}

// Pipeline 流水线接口
type Pipeline interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetStatus() *PipelineStatus
	UpdateStatus(status string)
	GetSteps() []StepStatus
	UpdateStepStatus(stepID string, status StepStatus)
	GetNodeID() string
	SetNodeID(nodeID string)
	GetProgress() float64
	SetProgress(progress float64)
	GetError() string
	SetError(error string)
	GetStartTime() time.Time
	SetStartTime(startTime time.Time)
	GetEndTime() time.Time
	SetEndTime(endTime time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(updatedAt time.Time)
}
