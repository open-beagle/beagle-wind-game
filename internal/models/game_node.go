package models

import "time"

// GameNodeType 游戏节点类型
type GameNodeType string

const (
	GameNodeTypePhysical GameNodeType = "physical" // 物理节点
	GameNodeTypeVirtual  GameNodeType = "virtual"  // 虚拟节点
)

// GameNodeState 游戏节点状态
type GameNodeState string

const (
	GameNodeStateOffline GameNodeState = "offline" // 离线
	GameNodeStateOnline  GameNodeState = "online"  // 在线
	GameNodeStateReady   GameNodeState = "ready"   // 就绪
	GameNodeStateBusy    GameNodeState = "busy"    // 忙碌
	GameNodeStateError   GameNodeState = "error"   // 错误
)

// GameNodeStatus 节点状态信息
type GameNodeStatus struct {
	State      GameNodeState          `json:"state" yaml:"state"`             // 节点状态
	Online     bool                   `json:"online" yaml:"online"`           // 是否在线
	LastOnline time.Time              `json:"last_online" yaml:"last_online"` // 最后在线时间
	UpdatedAt  time.Time              `json:"updated_at" yaml:"updated_at"`   // 状态更新时间
	Resources  map[string]string      `json:"resources" yaml:"resources"`     // 资源使用情况
	Metrics    map[string]interface{} `json:"metrics" yaml:"metrics"`         // 性能指标
}

// GameNode 游戏节点
type GameNode struct {
	ID        string            `json:"id" yaml:"id"`                 // 节点ID
	Name      string            `json:"name" yaml:"name"`             // 节点名称
	Model     string            `json:"model" yaml:"model"`           // 节点型号
	Type      GameNodeType      `json:"type" yaml:"type"`             // 节点类型
	Location  string            `json:"location" yaml:"location"`     // 节点位置
	Hardware  map[string]string `json:"hardware" yaml:"hardware"`     // 硬件配置
	Network   map[string]string `json:"network" yaml:"network"`       // 网络配置
	Labels    map[string]string `json:"labels" yaml:"labels"`         // 标签
	Status    GameNodeStatus    `json:"status" yaml:"status"`         // 节点状态信息
	CreatedAt time.Time         `json:"created_at" yaml:"created_at"` // 创建时间
	UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"` // 更新时间
}

// TableName 返回表名
func (GameNode) TableName() string {
	return "game_nodes"
}
