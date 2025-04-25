package models

import (
	"time"

	"gopkg.in/yaml.v3"
)

// PipelineModel 表示流水线模板类型
type PipelineModel int

const (
	PipelineModelUnknown PipelineModel = iota
	PipelineModelStartPlatform
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
	ID          string     `json:"id" yaml:"id"`                                         // 步骤ID
	Name        string     `json:"name" yaml:"name"`                                     // 步骤名称
	State       StepState  `json:"status" yaml:"status"`                                 // 步骤状态
	ContainerID string     `json:"container_id,omitempty" yaml:"container_id,omitempty"` // 容器ID
	StartTime   *time.Time `json:"start_time,omitempty" yaml:"start_time,omitempty"`     // 开始时间
	EndTime     *time.Time `json:"end_time,omitempty" yaml:"end_time,omitempty"`         // 结束时间
	Error       string     `json:"error,omitempty" yaml:"error,omitempty"`               // 错误信息
	Output      string     `json:"-" yaml:"-"`                                           // 执行输出
	Logs        []byte     `json:"logs,omitempty" yaml:"logs,omitempty"`                 // 执行日志
	Progress    float64    `json:"progress,omitempty" yaml:"progress,omitempty"`         // 执行进度
	UpdatedAt   *time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`     // 更新时间
}

// ContainerConfig 容器配置
type ContainerConfig struct {
	Image         string            `json:"image" yaml:"image"`
	ContainerName string            `json:"container_name,omitempty" yaml:"container_name,omitempty"`
	Hostname      string            `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Privileged    bool              `json:"privileged" yaml:"privileged"`
	Deploy        DeployConfig      `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	SecurityOpt   []string          `json:"security_opt,omitempty" yaml:"security_opt,omitempty"`
	CapAdd        []string          `json:"cap_add,omitempty" yaml:"cap_add,omitempty"`
	Tmpfs         []string          `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`
	Devices       []string          `json:"devices,omitempty" yaml:"devices,omitempty"`
	Volumes       []string          `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Ports         []string          `json:"ports,omitempty" yaml:"ports,omitempty"`
	Environment   map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Command       []string          `json:"command,omitempty" yaml:"command,omitempty"`
}

// DeployConfig 部署配置
type DeployConfig struct {
	Resources ResourcesConfig `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// ResourcesConfig 资源配置
type ResourcesConfig struct {
	Reservations ReservationsConfig `json:"reservations,omitempty" yaml:"reservations,omitempty"`
}

// ReservationsConfig 资源预留配置
type ReservationsConfig struct {
	Devices []DeviceConfig `json:"devices,omitempty" yaml:"devices,omitempty"`
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	Capabilities []string `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

// PipelineStep 流水线步骤
type PipelineStep struct {
	Name      string          `json:"name" yaml:"name"`
	Type      string          `json:"type" yaml:"type"`
	Container ContainerConfig `json:"container,omitempty" yaml:"container,omitempty"`
}

// PipelineStatus 流水线状态信息
type PipelineStatus struct {
	NodeID       string        `json:"node_id" yaml:"node_id"`                           // 节点ID
	State        PipelineState `json:"status" yaml:"status"`                             // 流水线状态
	CurrentStep  int32         `json:"current_step" yaml:"current_step"`                 // 当前步骤
	TotalSteps   int32         `json:"total_steps" yaml:"total_steps"`                   // 总步骤数
	StartTime    *time.Time    `json:"start_time,omitempty" yaml:"start_time,omitempty"` // 开始时间
	EndTime      *time.Time    `json:"end_time,omitempty" yaml:"end_time,omitempty"`     // 结束时间
	Steps        []StepStatus  `json:"steps,omitempty" yaml:"steps,omitempty"`           // 步骤状态列表
	ErrorMessage string        `json:"error,omitempty" yaml:"error,omitempty"`           // 错误信息
	UpdatedAt    *time.Time    `json:"updated_at,omitempty" yaml:"updated_at,omitempty"` // 更新时间
}

// GamePipeline 表示一个游戏节点流水线模板
type GamePipeline struct {
	ID    string        `json:"id" yaml:"id"`       // 实例ID
	Model PipelineModel `json:"model" yaml:"model"` // 实例模板

	// 静态信息（模板定义）
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Envs        []string       `json:"envs,omitempty" yaml:"envs,omitempty"`
	Args        []string       `json:"args,omitempty" yaml:"args,omitempty"`
	Steps       []PipelineStep `json:"steps,omitempty" yaml:"steps,omitempty"`

	// 动态信息（执行状态）
	Status *PipelineStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// NewGamePipelineFromYAML 从YAML创建新的游戏节点流水线模板
func NewGamePipelineFromYAML(data []byte) (*GamePipeline, error) {
	var pipeline GamePipeline
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
func (p *GamePipeline) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}

// String 返回 PipelineModel 的字符串表示
func (p PipelineModel) String() string {
	switch p {
	case PipelineModelStartPlatform:
		return "start-platform"
	default:
		return "unknown"
	}
}

// MarshalYAML 实现 yaml.Marshaler 接口
func (p PipelineModel) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

// UnmarshalYAML 实现 yaml.Unmarshaler 接口
func (p *PipelineModel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	switch s {
	case "start-platform":
		*p = PipelineModelStartPlatform
	default:
		*p = PipelineModelUnknown
	}
	return nil
}
