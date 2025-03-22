package models

import "time"

// GameNode 游戏机模型
type GameNode struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	Hardware   string    `json:"hardware"`  // 硬件配置信息
	Network    string    `json:"network"`   // 网络状态
	Location   string    `json:"location"`  // 地理位置
	Status     string    `json:"status"`    // 运行状态
	Resources  string    `json:"resources"` // 资源使用情况
	Online     bool      `json:"online"`    // 在线状态
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastOnline time.Time `json:"last_online"` // 最后在线时间
}

// TableName 返回表名
func (GameNode) TableName() string {
	return "game_nodes"
}
