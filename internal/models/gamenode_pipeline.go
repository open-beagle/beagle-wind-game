package models

import (
	"time"

	"gopkg.in/yaml.v3"
)

// PipelineState 表示流水线状态
type PipelineState string

const (
	PipelineStatePending   PipelineState = "pending"
	PipelineStateRunning   PipelineState = "running"
	PipelineStateCompleted PipelineState = "completed"
	PipelineStateFailed    PipelineState = "failed"
	PipelineStateCanceled  PipelineState = "canceled"
)

// StepState 表示步骤状态
type StepState string

const (
	StepStatePending   StepState = "pending"
	StepStateRunning   StepState = "running"
	StepStateCompleted StepState = "completed"
	StepStateFailed    StepState = "failed"
	StepStateSkipped   StepState = "skipped"
)

// StepStatus 步骤状态信息
type StepStatus struct {
	ID        string    `json:"id" yaml:"id"`                 // 步骤ID
	Name      string    `json:"name" yaml:"name"`             // 步骤名称
	State     StepState `json:"status" yaml:"status"`         // 步骤状态
	StartTime time.Time `json:"start_time" yaml:"start_time"` // 开始时间
	EndTime   time.Time `json:"end_time" yaml:"end_time"`     // 结束时间
	Error     string    `json:"error" yaml:"error"`           // 错误信息
	Output    string    `json:"-" yaml:"-"`                   // 执行输出
	Logs      []byte    `json:"logs" yaml:"logs"`             // 执行日志
	Progress  float64   `json:"progress" yaml:"progress"`     // 执行进度
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"` // 更新时间
}

// ContainerConfig 容器配置
type ContainerConfig struct {
	Image         string            `yaml:"image"`
	ContainerName string            `yaml:"container_name"`
	Hostname      string            `yaml:"hostname"`
	Privileged    bool              `yaml:"privileged"`
	Deploy        DeployConfig      `yaml:"deploy"`
	SecurityOpt   []string          `yaml:"security_opt"`
	CapAdd        []string          `yaml:"cap_add"`
	Tmpfs         []string          `yaml:"tmpfs"`
	Devices       []string          `yaml:"devices"`
	Volumes       []string          `yaml:"volumes"`
	Ports         []string          `yaml:"ports"`
	Environment   map[string]string `yaml:"environment"`
	Command       []string          `yaml:"command"`
}

// DeployConfig 部署配置
type DeployConfig struct {
	Resources ResourcesConfig `yaml:"resources"`
}

// ResourcesConfig 资源配置
type ResourcesConfig struct {
	Reservations ReservationsConfig `yaml:"reservations"`
}

// ReservationsConfig 资源预留配置
type ReservationsConfig struct {
	Devices []DeviceConfig `yaml:"devices"`
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	Capabilities []string `yaml:"capabilities"`
}

// PipelineStep 流水线步骤
type PipelineStep struct {
	Name      string          `yaml:"name"`
	Type      string          `yaml:"type"`
	Container ContainerConfig `yaml:"container"`
}

// PipelineStatus 流水线状态信息
type PipelineStatus struct {
	ID           string        `json:"id" yaml:"id"`                     // 流水线ID
	NodeID       string        `json:"id" yaml:"id"`                     // 节点ID
	State        PipelineState `json:"status" yaml:"status"`             // 流水线状态
	CurrentStep  int32         `json:"current_step" yaml:"current_step"` // 当前步骤
	TotalSteps   int32         `json:"total_steps" yaml:"total_steps"`   // 总步骤数
	Progress     float64       `json:"progress" yaml:"progress"`         // 执行进度
	StartTime    time.Time     `json:"start_time" yaml:"start_time"`     // 开始时间
	EndTime      time.Time     `json:"end_time" yaml:"end_time"`         // 结束时间
	Steps        []StepStatus  `json:"steps" yaml:"steps"`               // 步骤状态列表
	ErrorMessage string        `json:"error" yaml:"error"`               // 错误信息
	UpdatedAt    time.Time     `json:"updated_at" yaml:"updated_at"`     // 更新时间
}

// GameNodePipeline 表示一个游戏节点流水线模板
type GameNodePipeline struct {
	// 静态信息（模板定义）
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Envs        []string       `yaml:"envs"`
	Args        []string       `yaml:"args"`
	Steps       []PipelineStep `yaml:"steps"`
	ID          string         `yaml:"id"` // 节点ID

	// 动态信息（执行状态）
	Status *PipelineStatus `yaml:"status"`
}

// NewGameNodePipelineFromYAML 从YAML创建新的游戏节点流水线模板
func NewGameNodePipelineFromYAML(data []byte) (*GameNodePipeline, error) {
	var pipeline GameNodePipeline
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}

	// 初始化状态
	totalSteps := int32(len(pipeline.Steps))
	pipeline.Status = &PipelineStatus{
		State:      PipelineStatePending,
		TotalSteps: totalSteps,
		Steps:      make([]StepStatus, totalSteps),
	}

	// 初始化每个步骤的状态
	for i := range pipeline.Status.Steps {
		pipeline.Status.Steps[i] = StepStatus{
			State: StepStatePending,
		}
	}

	return &pipeline, nil
}

// ToYAML 将流水线转换为YAML
func (p *GameNodePipeline) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}
