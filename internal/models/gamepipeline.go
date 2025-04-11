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
	Image         string            `json:"image" yaml:"image"`
	ContainerName string            `json:"container_name" yaml:"container_name"`
	Hostname      string            `json:"hostname" yaml:"hostname"`
	Privileged    bool              `json:"privileged" yaml:"privileged"`
	Deploy        DeployConfig      `json:"deploy" yaml:"deploy"`
	SecurityOpt   []string          `json:"security_opt" yaml:"security_opt"`
	CapAdd        []string          `json:"cap_add" yaml:"cap_add"`
	Tmpfs         []string          `json:"tmpfs" yaml:"tmpfs"`
	Devices       []string          `json:"devices" yaml:"devices"`
	Volumes       []string          `json:"volumes" yaml:"volumes"`
	Ports         []string          `json:"ports" yaml:"ports"`
	Environment   map[string]string `json:"environment" yaml:"environment"`
	Command       []string          `json:"command" yaml:"command"`
}

// DeployConfig 部署配置
type DeployConfig struct {
	Resources ResourcesConfig `json:"resources" yaml:"resources"`
}

// ResourcesConfig 资源配置
type ResourcesConfig struct {
	Reservations ReservationsConfig `json:"reservations" yaml:"reservations"`
}

// ReservationsConfig 资源预留配置
type ReservationsConfig struct {
	Devices []DeviceConfig `json:"devices" yaml:"devices"`
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	Capabilities []string `json:"capabilities" yaml:"capabilities"`
}

// PipelineStep 流水线步骤
type PipelineStep struct {
	Name      string          `json:"name" yaml:"name"`
	Type      string          `json:"type" yaml:"type"`
	Container ContainerConfig `json:"container" yaml:"container"`
}

// PipelineStatus 流水线状态信息
type PipelineStatus struct {
	NodeID       string        `json:"node_id" yaml:"node_id"`           // 节点ID
	State        PipelineState `json:"status" yaml:"status"`             // 流水线状态
	CurrentStep  int32         `json:"current_step" yaml:"current_step"` // 当前步骤
	TotalSteps   int32         `json:"total_steps" yaml:"total_steps"`   // 总步骤数
	StartTime    time.Time     `json:"start_time" yaml:"start_time"`     // 开始时间
	EndTime      time.Time     `json:"end_time" yaml:"end_time"`         // 结束时间
	Steps        []StepStatus  `json:"steps" yaml:"steps"`               // 步骤状态列表
	ErrorMessage string        `json:"error" yaml:"error"`               // 错误信息
	UpdatedAt    time.Time     `json:"updated_at" yaml:"updated_at"`     // 更新时间
}

// GamePipeline 表示一个游戏节点流水线模板
type GamePipeline struct {
	ID    string        `json:"id" yaml:"id"`       // 实例ID
	Model PipelineModel `json:"model" yaml:"model"` // 实例模板

	// 静态信息（模板定义）
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	Envs        []string       `json:"envs" yaml:"envs"`
	Args        []string       `json:"args" yaml:"args"`
	Steps       []PipelineStep `json:"steps" yaml:"steps"`

	// 动态信息（执行状态）
	Status *PipelineStatus `json:"status" yaml:"status"`
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
