package models

import "time"

// GamePlatform 游戏平台模型
type GamePlatform struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Type         string    `json:"type"`         // 平台类型（PC、主机、移动端等）
	Features     string    `json:"features"`     // 平台特性
	Requirements string    `json:"requirements"` // 系统要求
	Config       string    `json:"config"`       // 运行环境配置
	Network      string    `json:"network"`      // 网络配置
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 返回表名
func (GamePlatform) TableName() string {
	return "game_platforms"
}
