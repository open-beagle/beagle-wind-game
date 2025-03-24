package models

import "time"

// GameCard 游戏卡片模型
type GameCard struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	SortName    string    `json:"sort_name" yaml:"sort_name"`       // 排序用的英文名
	SlugName    string    `json:"slug_name" yaml:"slug_name"`       // 用于文件系统的名称
	Type        string    `json:"type" yaml:"type"`                 // 游戏类型
	PlatformID  string    `json:"platform_id" yaml:"platform_id"`   // 关联的平台ID
	Description string    `json:"description" yaml:"description"`   // 游戏描述
	Cover       string    `json:"cover" yaml:"cover"`               // 封面图片URL
	Category    string    `json:"category" yaml:"category"`         // 游戏分类
	ReleaseDate string    `json:"release_date" yaml:"release_date"` // 发行日期
	Tags        string    `json:"tags" yaml:"tags"`                 // 游戏标签
	Files       string    `json:"files" yaml:"files"`               // 游戏文件信息
	Updates     string    `json:"updates" yaml:"updates"`           // 更新包信息
	Patches     string    `json:"patches" yaml:"patches"`           // 补丁包信息
	Params      string    `json:"params" yaml:"params"`             // 运行参数
	Settings    string    `json:"settings" yaml:"settings"`         // 游戏设置
	Permissions string    `json:"permissions" yaml:"permissions"`   // 权限控制
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
}

// TableName 返回表名
func (GameCard) TableName() string {
	return "game_cards"
}
