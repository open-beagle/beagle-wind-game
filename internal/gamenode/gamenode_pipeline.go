package gamenode

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
	State     StepState
	StartTime int64
	EndTime   int64
	Error     string
	Output    string
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
	State       PipelineState
	CurrentStep int32
	TotalSteps  int32
	Progress    float32
	StartTime   int64
	EndTime     int64

	// 步骤状态
	StepStatuses []*StepStatus
}

// GameNodePipeline 表示一个游戏节点流水线模板
type GameNodePipeline struct {
	// 静态信息（模板定义）
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Envs        []string       `yaml:"envs"`
	Args        []string       `yaml:"args"`
	Steps       []PipelineStep `yaml:"steps"`

	// 动态信息（执行状态）
	status *PipelineStatus
}

// NewGameNodePipelineFromYAML 从YAML创建新的游戏节点流水线模板
func NewGameNodePipelineFromYAML(data []byte) (*GameNodePipeline, error) {
	var pipeline GameNodePipeline
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}

	// 初始化状态
	totalSteps := int32(len(pipeline.Steps))
	pipeline.status = &PipelineStatus{
		State:        PipelineStatePending,
		TotalSteps:   totalSteps,
		StepStatuses: make([]*StepStatus, totalSteps),
	}

	// 初始化每个步骤的状态
	for i := range pipeline.status.StepStatuses {
		pipeline.status.StepStatuses[i] = &StepStatus{
			State: StepStatePending,
		}
	}

	return &pipeline, nil
}

// ToYAML 将流水线转换为YAML
func (p *GameNodePipeline) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}

// GetStatus 获取流水线状态
func (p *GameNodePipeline) GetStatus() map[string]interface{} {
	// 构建步骤状态列表
	stepStatuses := make([]map[string]interface{}, len(p.status.StepStatuses))
	for i, step := range p.status.StepStatuses {
		stepStatuses[i] = map[string]interface{}{
			"state":      step.State,
			"start_time": step.StartTime,
			"end_time":   step.EndTime,
			"error":      step.Error,
			"output":     step.Output,
		}
	}

	return map[string]interface{}{
		"state":         p.status.State,
		"current_step":  p.status.CurrentStep,
		"total_steps":   p.status.TotalSteps,
		"progress":      p.status.Progress,
		"start_time":    p.status.StartTime,
		"end_time":      p.status.EndTime,
		"step_statuses": stepStatuses,
	}
}

// UpdateStatus 更新流水线状态
func (p *GameNodePipeline) UpdateStatus(state PipelineState) {
	p.status.State = state
}

// UpdateProgress 更新流水线进度
func (p *GameNodePipeline) UpdateProgress(currentStep int32) {
	p.status.CurrentStep = currentStep
	p.status.Progress = float32(currentStep+1) / float32(p.status.TotalSteps)
}

// SetStartTime 设置开始时间
func (p *GameNodePipeline) SetStartTime(startTime int64) {
	p.status.StartTime = startTime
}

// SetEndTime 设置结束时间
func (p *GameNodePipeline) SetEndTime(endTime int64) {
	p.status.EndTime = endTime
}

// UpdateStepStatus 更新步骤状态
func (p *GameNodePipeline) UpdateStepStatus(stepIndex int32, state StepState) {
	if stepIndex < 0 || stepIndex >= int32(len(p.status.StepStatuses)) {
		return
	}

	step := p.status.StepStatuses[stepIndex]
	step.State = state

	switch state {
	case StepStateRunning:
		step.StartTime = time.Now().Unix()
	case StepStateCompleted, StepStateFailed, StepStateSkipped:
		step.EndTime = time.Now().Unix()
	}
}

// SetStepError 设置步骤错误信息
func (p *GameNodePipeline) SetStepError(stepIndex int32, err error) {
	if stepIndex < 0 || stepIndex >= int32(len(p.status.StepStatuses)) {
		return
	}

	step := p.status.StepStatuses[stepIndex]
	step.Error = err.Error()
}

// SetStepOutput 设置步骤输出信息
func (p *GameNodePipeline) SetStepOutput(stepIndex int32, output string) {
	if stepIndex < 0 || stepIndex >= int32(len(p.status.StepStatuses)) {
		return
	}

	step := p.status.StepStatuses[stepIndex]
	step.Output = output
}

// GetStepStatus 获取步骤状态
func (p *GameNodePipeline) GetStepStatus(stepIndex int32) *StepStatus {
	if stepIndex < 0 || stepIndex >= int32(len(p.status.StepStatuses)) {
		return nil
	}

	return p.status.StepStatuses[stepIndex]
}

// GetSteps 获取步骤列表
func (p *GameNodePipeline) GetSteps() []PipelineStep {
	return p.Steps
}

// GetEnvs 获取环境变量列表
func (p *GameNodePipeline) GetEnvs() []string {
	return p.Envs
}

// GetArgs 获取参数列表
func (p *GameNodePipeline) GetArgs() []string {
	return p.Args
}

// GetName 获取流水线名称
func (p *GameNodePipeline) GetName() string {
	return p.Name
}

// GetDescription 获取流水线描述
func (p *GameNodePipeline) GetDescription() string {
	return p.Description
}
