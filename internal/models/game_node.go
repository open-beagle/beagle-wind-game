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
)

// GameNode 游戏节点
type GameNode struct {
	ID          string                 `json:"id" yaml:"id"`                               // 节点ID
	Name        string                 `json:"name" yaml:"name"`                           // 节点名称
	Description string                 `json:"description" yaml:"description"`             // 节点描述
	Type        string                 `json:"type" yaml:"type"`                           // 节点类型（physical/virtual/container）
	Status      string                 `json:"status" yaml:"status"`                       // 节点状态（online/offline/maintenance）
	Resources   map[string]interface{} `json:"resources" yaml:"resources"`                 // 资源信息
	Metrics     map[string]interface{} `json:"metrics" yaml:"metrics"`                     // 监控指标
	Labels      map[string]string      `json:"labels" yaml:"labels"`                       // 节点标签
	Network     map[string]interface{} `json:"network,omitempty" yaml:"network,omitempty"` // 网络信息
	CreatedAt   time.Time              `json:"created_at"`                                 // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`                                 // 更新时间
}

// TableName 返回表名
func (GameNode) TableName() string {
	return "game_nodes"
}
