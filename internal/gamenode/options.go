package gamenode

import (
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
)

// 额外定义的常量
const (
	GameNodeTypeUnknown models.GameNodeType = "unknown" // 未知节点类型
)

// Options 游戏节点代理选项
type Options struct {
	// 节点标识
	ID string
	// 节点名称
	Name string
	// 节点命名空间
	Namespace string
	// 节点类型
	NodeType models.GameNodeType
	// 节点IP
	IP string
	// 节点标签
	Labels map[string]string
	// 服务器地址
	ServerAddr string
	// 心跳周期
	HeartbeatPeriod time.Duration
	// 指标收集间隔
	MetricsInterval time.Duration
}

// Option 选项函数类型
type Option func(*Options)

// defaultOptions 默认选项
func defaultOptions() *Options {
	return &Options{
		ID:              "node-" + time.Now().Format("20060102150405"),
		Name:            "unknown",
		Namespace:       "default",
		NodeType:        GameNodeTypeUnknown,
		IP:              "127.0.0.1",
		Labels:          make(map[string]string),
		ServerAddr:      "localhost:8080",
		HeartbeatPeriod: 30 * time.Second,
		MetricsInterval: 5 * time.Second,
	}
}

// applyOptions 应用选项
func applyOptions(opts ...Option) *Options {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithID 设置节点ID
func WithID(id string) Option {
	return func(o *Options) {
		o.ID = id
	}
}

// WithName 设置节点名称
func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// WithNamespace 设置节点命名空间
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

// WithNodeType 设置节点类型
func WithNodeType(nodeType models.GameNodeType) Option {
	return func(o *Options) {
		o.NodeType = nodeType
	}
}

// WithIP 设置节点IP
func WithIP(ip string) Option {
	return func(o *Options) {
		o.IP = ip
	}
}

// WithLabels 设置节点标签
func WithLabels(labels map[string]string) Option {
	return func(o *Options) {
		o.Labels = labels
	}
}

// WithServerAddr 设置服务器地址
func WithServerAddr(serverAddr string) Option {
	return func(o *Options) {
		o.ServerAddr = serverAddr
	}
}

// WithHeartbeatPeriod 设置心跳周期
func WithHeartbeatPeriod(period time.Duration) Option {
	return func(o *Options) {
		o.HeartbeatPeriod = period
	}
}

// WithMetricsInterval 设置指标收集间隔
func WithMetricsInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.MetricsInterval = interval
	}
}
