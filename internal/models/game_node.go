package models

import "time"

// GameNodeType 节点类型
type GameNodeType string

const (
	// NodeTypePhysical 物理节点
	NodeTypePhysical GameNodeType = "physical"
	// NodeTypeVirtual 虚拟节点
	NodeTypeVirtual GameNodeType = "virtual"
	// NodeTypeContainer 容器节点
	NodeTypeContainer GameNodeType = "container"
)

// GameNodeStatus 节点状态
type GameNodeStatus string

const (
	// NodeStatusOnline 在线
	NodeStatusOnline GameNodeStatus = "online"
	// NodeStatusOffline 离线
	NodeStatusOffline GameNodeStatus = "offline"
	// NodeStatusMaintenance 维护中
	NodeStatusMaintenance GameNodeStatus = "maintenance"
	// NodeStatusReady 准备就绪
	NodeStatusReady GameNodeStatus = "ready"
)

// GameNode 游戏节点
type GameNode struct {
	ID         string                 `json:"id" yaml:"id"`                   // 节点ID
	Name       string                 `json:"name" yaml:"name"`               // 节点名称
	Model      string                 `json:"model" yaml:"model"`             // 节点型号
	Type       string                 `json:"type" yaml:"type"`               // 节点类型（physical/virtual/container）
	Status     string                 `json:"status" yaml:"status"`           // 节点状态（online/offline/maintenance/ready）
	Location   string                 `json:"location" yaml:"location"`       // 节点地理位置
	Hardware   map[string]interface{} `json:"hardware" yaml:"hardware"`       // 硬件配置（CPU、RAM、GPU等）
	Network    map[string]interface{} `json:"network" yaml:"network"`         // 网络信息（IP、速度等）
	Resources  map[string]interface{} `json:"resources" yaml:"resources"`     // 资源使用情况
	Metrics    map[string]interface{} `json:"metrics" yaml:"metrics"`         // 监控指标
	Labels     map[string]string      `json:"labels" yaml:"labels"`           // 节点标签
	Online     bool                   `json:"online" yaml:"online"`           // 是否在线
	LastOnline time.Time              `json:"last_online" yaml:"last_online"` // 最后在线时间
	CreatedAt  time.Time              `json:"created_at" yaml:"created_at"`   // 创建时间
	UpdatedAt  time.Time              `json:"updated_at" yaml:"updated_at"`   // 更新时间
}

// TableName 返回表名
func (GameNode) TableName() string {
	return "game_nodes"
}
