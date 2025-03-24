package models

import "time"

// GameInstance 游戏实例模型
type GameInstance struct {
	ID          string    `json:"id" yaml:"id"`
	NodeID      string    `json:"node_id" yaml:"node_id"`         // 关联的游戏机ID
	PlatformID  string    `json:"platform_id" yaml:"platform_id"` // 关联的平台ID
	CardID      string    `json:"card_id" yaml:"card_id"`         // 关联的游戏卡片ID
	Status      string    `json:"status" yaml:"status"`           // 运行状态
	Resources   string    `json:"resources" yaml:"resources"`     // 资源占用
	Performance string    `json:"performance" yaml:"performance"` // 性能指标
	SaveData    string    `json:"save_data" yaml:"save_data"`     // 存档数据
	Config      string    `json:"config" yaml:"config"`           // 实例配置
	Backup      string    `json:"backup" yaml:"backup"`           // 备份数据
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
	StartedAt   time.Time `json:"started_at" yaml:"started_at"` // 启动时间
	StoppedAt   time.Time `json:"stopped_at" yaml:"stopped_at"` // 停止时间
}

// TableName 返回表名
func (GameInstance) TableName() string {
	return "game_instances"
}
