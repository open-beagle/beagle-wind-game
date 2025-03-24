package models

import "time"

// GameCard 游戏卡片模型
type GameCard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SortName    string    `json:"sort_name"`    // 排序用的英文名
	SlugName    string    `json:"slug_name"`    // 用于文件系统的名称
	Type        string    `json:"type"`         // 游戏类型
	PlatformID  string    `json:"platform_id"`  // 关联的平台ID
	Description string    `json:"description"`  // 游戏描述
	Cover       string    `json:"cover"`        // 封面图片URL
	Category    string    `json:"category"`     // 游戏分类
	ReleaseDate string    `json:"release_date"` // 发行日期
	Tags        string    `json:"tags"`         // 游戏标签
	Files       string    `json:"files"`        // 游戏文件信息
	Updates     string    `json:"updates"`      // 更新包信息
	Patches     string    `json:"patches"`      // 补丁包信息
	Params      string    `json:"params"`       // 运行参数
	Settings    string    `json:"settings"`     // 游戏设置
	Permissions string    `json:"permissions"`  // 权限控制
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 返回表名
func (GameCard) TableName() string {
	return "game_cards"
}
