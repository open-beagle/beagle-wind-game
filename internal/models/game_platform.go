package models

import "time"

// GamePlatform 游戏平台
type GamePlatform struct {
	ID        string              `json:"id" yaml:"id"`               // 平台ID
	Name      string              `json:"name" yaml:"name"`           // 平台名称
	Version   string              `json:"version" yaml:"version"`     // 平台版本
	Type      string              `json:"type" yaml:"type"`           // 平台类型
	OS        string              `json:"os" yaml:"os"`               // 操作系统
	Image     string              `json:"image" yaml:"image"`         // 容器镜像
	Bin       string              `json:"bin" yaml:"bin"`             // 启动路径
	Features  []string            `json:"features" yaml:"features"`   // 平台特性列表
	Config    map[string]string   `json:"config" yaml:"config"`       // 平台配置键值对
	Files     []PlatformFile      `json:"files" yaml:"files"`         // 平台文件
	Installer []PlatformInstaller `json:"installer" yaml:"installer"` // 安装步骤
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// PlatformFile 平台文件
type PlatformFile struct {
	ID   string `json:"id" yaml:"id"`     // 文件ID
	Type string `json:"type" yaml:"type"` // 文件类型
	URL  string `json:"url" yaml:"url"`   // 文件URL
}

// PlatformInstaller 安装步骤
type PlatformInstaller struct {
	Move    *InstallerMove    `json:"move,omitempty" yaml:"move,omitempty"`       // 移动文件
	Chmodx  string            `json:"chmodx,omitempty" yaml:"chmodx,omitempty"`   // 添加执行权限
	Extract *InstallerExtract `json:"extract,omitempty" yaml:"extract,omitempty"` // 解压文件
	Command string            `json:"command,omitempty" yaml:"command,omitempty"` // 执行命令
}

// InstallerMove 移动文件配置
type InstallerMove struct {
	Dst string `json:"dst" yaml:"dst"` // 目标路径
	Src string `json:"src" yaml:"src"` // 源文件
}

// InstallerExtract 解压文件配置
type InstallerExtract struct {
	Dst  string `json:"dst" yaml:"dst"`   // 目标路径
	File string `json:"file" yaml:"file"` // 源文件
}

// TableName 返回表名
func (GamePlatform) TableName() string {
	return "game_platforms"
}
