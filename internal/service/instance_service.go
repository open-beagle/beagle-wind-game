package service

import (
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
)

// 错误常量定义
var (
	ErrInstanceNotFound       = fmt.Errorf("实例不存在")
	ErrInstanceAlreadyRunning = fmt.Errorf("实例已经在运行中")
	ErrInstanceNotRunning     = fmt.Errorf("实例未在运行中")
	ErrNodeNotFound           = fmt.Errorf("节点不存在")
	ErrNodeNotReady           = fmt.Errorf("节点未就绪")
)

// InstanceService 游戏实例服务
type InstanceService struct {
	instanceStore store.InstanceStore
}

// NewInstanceService 创建游戏实例服务
func NewInstanceService(instanceStore store.InstanceStore) *InstanceService {
	return &InstanceService{
		instanceStore: instanceStore,
	}
}

// InstanceListParams 实例列表查询参数
type InstanceListParams struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword    string `form:"keyword" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty,oneof=running stopped paused error"`
	NodeID     string `form:"node_id" binding:"omitempty"`
	CardID     string `form:"card_id" binding:"omitempty"`
	PlatformID string `form:"platform_id" binding:"omitempty"`
}

// InstanceListResult 实例列表结果
type InstanceListResult struct {
	Total int                   `json:"total"`
	Items []models.GameInstance `json:"items"`
}

// ListInstances 获取游戏实例列表
func (s *InstanceService) ListInstances(params InstanceListParams) (InstanceListResult, error) {
	// 从存储获取实例列表
	instances, err := s.instanceStore.List()
	if err != nil {
		return InstanceListResult{}, err
	}

	// 过滤和分页
	var filteredInstances []models.GameInstance
	for _, instance := range instances {
		// 关键词过滤
		if params.Keyword != "" {
			// TODO: 实现关键词过滤
			continue
		}

		// 状态过滤
		if params.Status != "" && instance.Status != params.Status {
			continue
		}

		// 节点ID过滤
		if params.NodeID != "" && instance.NodeID != params.NodeID {
			continue
		}

		// 游戏卡片ID过滤
		if params.CardID != "" && instance.CardID != params.CardID {
			continue
		}

		// 平台ID过滤
		if params.PlatformID != "" && instance.PlatformID != params.PlatformID {
			continue
		}

		// 如果满足所有条件，添加到结果中
		filteredInstances = append(filteredInstances, instance)
	}

	// 处理分页
	total := len(filteredInstances)
	// 默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	// 计算分页范围
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start >= total {
		return InstanceListResult{
			Total: total,
			Items: []models.GameInstance{},
		}, nil
	}
	if end > total {
		end = total
	}

	return InstanceListResult{
		Total: total,
		Items: filteredInstances[start:end],
	}, nil
}

// GetInstance 获取实例详情
func (s *InstanceService) GetInstance(id string) (models.GameInstance, error) {
	return s.instanceStore.Get(id)
}

// CreateInstanceParams 创建实例参数
type CreateInstanceParams struct {
	NodeID     string `json:"node_id" binding:"required"`
	PlatformID string `json:"platform_id" binding:"required"`
	CardID     string `json:"card_id" binding:"required"`
	Config     string `json:"config,omitempty"` // 自定义配置
}

// CreateInstance 创建游戏实例
func (s *InstanceService) CreateInstance(params CreateInstanceParams) (string, error) {
	// 创建实例
	now := time.Now()
	instance := models.GameInstance{
		ID:         fmt.Sprintf("inst-%s-%s-%d", params.NodeID, params.CardID, now.Unix()),
		NodeID:     params.NodeID,
		PlatformID: params.PlatformID,
		CardID:     params.CardID,
		Status:     "starting", // 初始状态为starting
		Resources:  "",         // 待填充
		Config:     params.Config,
		CreatedAt:  now,
		UpdatedAt:  now,
		StartedAt:  now,
	}

	// 保存实例
	err := s.instanceStore.Add(instance)
	if err != nil {
		return "", fmt.Errorf("保存实例失败: %w", err)
	}

	return instance.ID, nil
}

// UpdateInstanceParams 更新实例参数
type UpdateInstanceParams struct {
	Status      string `json:"status,omitempty"`
	Resources   string `json:"resources,omitempty"`
	Performance string `json:"performance,omitempty"`
	SaveData    string `json:"save_data,omitempty"`
	Config      string `json:"config,omitempty"`
	Backup      string `json:"backup,omitempty"`
}

// UpdateInstance 更新游戏实例
func (s *InstanceService) UpdateInstance(id string, instance models.GameInstance) error {
	// 检查实例是否存在
	existingInstance, err := s.instanceStore.Get(id)
	if err != nil {
		return err
	}

	// 保留创建时间
	instance.CreatedAt = existingInstance.CreatedAt
	// 更新更新时间
	instance.UpdatedAt = time.Now()
	// 确保ID一致
	instance.ID = id

	// 更新实例
	return s.instanceStore.Update(instance)
}

// DeleteInstance 删除游戏实例
func (s *InstanceService) DeleteInstance(id string) error {
	return s.instanceStore.Delete(id)
}

// StartInstance 启动游戏实例
func (s *InstanceService) StartInstance(id string) error {
	// 获取实例
	instance, err := s.instanceStore.Get(id)
	if err != nil {
		return err
	}

	// 更新实例状态
	instance.Status = "running"
	instance.UpdatedAt = time.Now()
	instance.StartedAt = time.Now()

	// 保存更新
	return s.instanceStore.Update(instance)
}

// StopInstance 停止游戏实例
func (s *InstanceService) StopInstance(id string) error {
	// 获取实例
	instance, err := s.instanceStore.Get(id)
	if err != nil {
		return err
	}

	// 更新实例状态
	instance.Status = "stopped"
	instance.UpdatedAt = time.Now()
	instance.StoppedAt = time.Now()

	// 保存更新
	return s.instanceStore.Update(instance)
}
