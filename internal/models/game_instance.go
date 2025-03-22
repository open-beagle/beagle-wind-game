package models

import "time"

// GameInstance 游戏实例模型
type GameInstance struct {
	ID          string    `json:"id"`
	NodeID      string    `json:"node_id"`     // 关联的游戏机ID
	PlatformID  string    `json:"platform_id"` // 关联的平台ID
	CardID      string    `json:"card_id"`     // 关联的游戏卡片ID
	Status      string    `json:"status"`      // 运行状态
	Resources   string    `json:"resources"`   // 资源占用
	Performance string    `json:"performance"` // 性能指标
	SaveData    string    `json:"save_data"`   // 存档数据
	Config      string    `json:"config"`      // 实例配置
	Backup      string    `json:"backup"`      // 备份数据
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	StartedAt   time.Time `json:"started_at"` // 启动时间
	StoppedAt   time.Time `json:"stopped_at"` // 停止时间
}

// TableName 返回表名
func (GameInstance) TableName() string {
	return "game_instances"
}
