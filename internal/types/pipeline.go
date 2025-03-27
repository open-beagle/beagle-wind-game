package types

import (
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/gamenode"
)

// PipelineListParams 流水线列表查询参数
type PipelineListParams struct {
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
	Status    string    `json:"status"`
	NodeID    string    `json:"node_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	SortBy    string    `json:"sort_by"`
	SortOrder string    `json:"sort_order"`
}

// PipelineListResult 流水线列表查询结果
type PipelineListResult struct {
	Total int64                        `json:"total"`
	Items []*gamenode.GameNodePipeline `json:"items"`
}

// Pipeline 流水线接口
type Pipeline interface {
	GetName() string
	GetDescription() string
	GetStatus() *gamenode.PipelineStatus
	UpdateStatus(status gamenode.PipelineState)
}
