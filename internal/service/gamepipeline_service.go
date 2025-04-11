package service

import (
	"context"
	"fmt"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/open-beagle/beagle-wind-game/internal/types"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// PipelineStore Pipeline 存储接口
type PipelineStore interface {
	List(params types.PipelineListParams) (*types.PipelineListResult, error)
	Get(id string) (*models.GamePipeline, error)
	Add(pipeline *models.GamePipeline) error
	Update(pipeline *models.GamePipeline) error
	UpdateStatus(id string, status string) error
	Delete(id string, force bool) error
	Cleanup() error
}

// GamePipelineGRPCService 游戏节点流水线服务
type GamePipelineGRPCService struct {
	store  store.GamePipelineStore
	logger utils.Logger
}

// NewGamePipelineService 创建新的游戏节点流水线服务
func NewGamePipelineService(store store.GamePipelineStore) *GamePipelineGRPCService {
	logger := utils.New("GamePipelineGRPCService")
	return &GamePipelineGRPCService{
		store:  store,
		logger: logger,
	}
}

// List 获取流水线列表
func (s *GamePipelineGRPCService) List(ctx context.Context) ([]*models.GamePipeline, error) {
	s.logger.Debug("获取流水线列表")
	pipelines, err := s.store.List(ctx)
	if err != nil {
		s.logger.Error("获取流水线列表失败: %v", err)
		return nil, fmt.Errorf("获取流水线列表失败: %w", err)
	}
	s.logger.Debug("成功获取流水线列表，共 %d 个流水线", len(pipelines))
	return pipelines, nil
}

// Get 获取流水线详情
func (s *GamePipelineGRPCService) Get(ctx context.Context, id string) (*models.GamePipeline, error) {
	s.logger.Debug("获取流水线详情: %s", id)
	pipeline, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取流水线详情失败: %v", err)
		return nil, fmt.Errorf("获取流水线详情失败: %w", err)
	}
	if pipeline == nil {
		s.logger.Error("流水线不存在: %s", id)
		return nil, fmt.Errorf("流水线不存在: %s", id)
	}
	s.logger.Debug("成功获取流水线详情: %s", id)
	return pipeline, nil
}

// Create 创建流水线
func (s *GamePipelineGRPCService) Create(ctx context.Context, pipeline *models.GamePipeline) error {
	s.logger.Debug("创建流水线: %s", pipeline.ID)

	// 1. 验证Pipeline信息
	if err := s.validatePipeline(pipeline); err != nil {
		return err
	}

	// 2. 设置初始状态
	if pipeline.Status == nil {
		pipeline.Status = &models.PipelineStatus{
			State:        models.PipelineStatePending,
			TotalSteps:   int32(len(pipeline.Steps)),
			Steps:        make([]models.StepStatus, len(pipeline.Steps)),
			CurrentStep:  0,
			StartTime:    time.Time{},
			EndTime:      time.Time{},
			ErrorMessage: "",
			UpdatedAt:    time.Now(),
		}
	}

	// 3. 保存到存储
	if err := s.store.Add(ctx, pipeline); err != nil {
		s.logger.Error("创建流水线失败: %v", err)
		return fmt.Errorf("创建流水线失败: %w", err)
	}

	s.logger.Info("成功创建流水线: %s", pipeline.ID)
	return nil
}

// validatePipeline 验证Pipeline信息
func (s *GamePipelineGRPCService) validatePipeline(pipeline *models.GamePipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline is nil")
	}

	if pipeline.ID == "" {
		return fmt.Errorf("pipeline id is required")
	}

	if len(pipeline.Steps) == 0 {
		return fmt.Errorf("pipeline steps is empty")
	}

	for i, step := range pipeline.Steps {
		if step.Name == "" {
			return fmt.Errorf("step[%d] name is required", i)
		}
		if step.Type == "" {
			return fmt.Errorf("step[%d] type is required", i)
		}
	}

	return nil
}

// Execute 执行流水线
func (s *GamePipelineGRPCService) Execute(ctx context.Context, id string) error {
	s.logger.Debug("执行流水线: %s", id)
	// 获取流水线
	pipeline, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取流水线失败: %v", err)
		return fmt.Errorf("获取流水线失败: %w", err)
	}
	if pipeline == nil {
		s.logger.Error("流水线不存在: %s", id)
		return fmt.Errorf("流水线不存在: %s", id)
	}

	// 更新状态
	pipeline.Status.State = models.PipelineStateRunning
	err = s.store.Update(ctx, pipeline)
	if err != nil {
		s.logger.Error("更新流水线状态失败: %v", err)
		return fmt.Errorf("更新流水线状态失败: %w", err)
	}
	s.logger.Info("成功执行流水线: %s, 状态: %s", id, pipeline.Status.State)
	return nil
}

// UpdateStatus 更新流水线状态
func (s *GamePipelineGRPCService) UpdateStatus(ctx context.Context, id string, status *models.PipelineStatus) error {
	s.logger.Debug("更新流水线状态: %s, 新状态: %s", id, status.State)
	// 获取流水线
	pipeline, err := s.store.Get(ctx, id)
	if err != nil {
		s.logger.Error("获取流水线失败: %v", err)
		return fmt.Errorf("获取流水线失败: %w", err)
	}
	if pipeline == nil {
		s.logger.Error("流水线不存在: %s", id)
		return fmt.Errorf("流水线不存在: %s", id)
	}

	// 更新状态
	pipeline.Status = status
	err = s.store.Update(ctx, pipeline)
	if err != nil {
		s.logger.Error("更新流水线状态失败: %v", err)
		return fmt.Errorf("更新流水线状态失败: %w", err)
	}
	s.logger.Info("成功更新流水线状态: %s, 状态: %s", id, status.State)
	return nil
}

// UpdateStepStatus 更新步骤状态
func (s *GamePipelineGRPCService) UpdateStepStatus(ctx context.Context, pipelineID string, stepID string, status *models.StepStatus) error {
	s.logger.Debug("更新流水线步骤状态: 流水线ID: %s, 步骤ID: %s, 状态: %s", pipelineID, stepID, status.State)

	// 获取流水线
	pipeline, err := s.store.Get(ctx, pipelineID)
	if err != nil {
		s.logger.Error("获取流水线失败: %v", err)
		return fmt.Errorf("获取流水线失败: %w", err)
	}
	if pipeline == nil {
		s.logger.Error("流水线不存在: %s", pipelineID)
		return fmt.Errorf("流水线不存在: %s", pipelineID)
	}

	// 验证状态转换是否合法
	if err := s.validateStepStateTransition(pipeline, stepID, status.State); err != nil {
		return err
	}

	// 更新步骤状态
	stepFound := false
	for i := range pipeline.Status.Steps {
		if pipeline.Status.Steps[i].ID == stepID {
			// 只更新状态字段，保留其他信息
			pipeline.Status.Steps[i].State = status.State
			pipeline.Status.Steps[i].Error = status.Error
			pipeline.Status.Steps[i].UpdatedAt = time.Now()
			stepFound = true
			break
		}
	}

	if !stepFound {
		s.logger.Error("流水线步骤不存在: 流水线ID: %s, 步骤ID: %s", pipelineID, stepID)
		return fmt.Errorf("流水线步骤不存在: 流水线ID: %s, 步骤ID: %s", pipelineID, stepID)
	}

	// 更新流水线进度
	s.updatePipelineProgress(pipeline)

	// 检查是否需要更新流水线状态
	s.updatePipelineState(pipeline)

	// 保存更新
	err = s.store.Update(ctx, pipeline)
	if err != nil {
		s.logger.Error("更新流水线步骤状态失败: %v", err)
		return fmt.Errorf("更新流水线步骤状态失败: %w", err)
	}
	s.logger.Info("成功更新流水线步骤状态: 流水线ID: %s, 步骤ID: %s, 状态: %s", pipelineID, stepID, status.State)
	return nil
}

// validateStepStateTransition 验证步骤状态转换是否合法
func (s *GamePipelineGRPCService) validateStepStateTransition(pipeline *models.GamePipeline, stepID string, newState models.StepState) error {
	// 获取当前步骤
	var currentStep *models.StepStatus
	for i := range pipeline.Status.Steps {
		if pipeline.Status.Steps[i].ID == stepID {
			currentStep = &pipeline.Status.Steps[i]
			break
		}
	}
	if currentStep == nil {
		return fmt.Errorf("步骤不存在: %s", stepID)
	}

	// 状态转换规则
	switch currentStep.State {
	case models.StepStatePending:
		// 只能转换为运行中
		if newState != models.StepStateRunning {
			return fmt.Errorf("无效的状态转换: 从 %s 到 %s", currentStep.State, newState)
		}
	case models.StepStateRunning:
		// 可以转换为完成、失败或跳过
		if newState != models.StepStateCompleted && newState != models.StepStateFailed && newState != models.StepStateSkipped {
			return fmt.Errorf("无效的状态转换: 从 %s 到 %s", currentStep.State, newState)
		}
	case models.StepStateCompleted, models.StepStateFailed, models.StepStateSkipped:
		// 终态不能改变
		return fmt.Errorf("步骤已处于终态 %s，不能改变状态", currentStep.State)
	}

	return nil
}

// updatePipelineProgress 更新流水线进度
func (s *GamePipelineGRPCService) updatePipelineProgress(pipeline *models.GamePipeline) {
	completedSteps := 0

	for _, step := range pipeline.Status.Steps {
		if step.State == models.StepStateCompleted || step.State == models.StepStateFailed || step.State == models.StepStateSkipped {
			completedSteps++
		}
	}

}

// updatePipelineState 更新流水线状态
func (s *GamePipelineGRPCService) updatePipelineState(pipeline *models.GamePipeline) {
	// 检查所有步骤是否都已完成
	allStepsCompleted := true
	hasFailedStep := false
	hasSkippedStep := false

	for _, step := range pipeline.Status.Steps {
		if step.State == models.StepStateRunning {
			allStepsCompleted = false
			break
		}
		if step.State == models.StepStateFailed {
			hasFailedStep = true
		}
		if step.State == models.StepStateSkipped {
			hasSkippedStep = true
		}
	}

	if allStepsCompleted {
		if hasFailedStep {
			pipeline.Status.State = models.PipelineStateFailed
		} else if hasSkippedStep {
			pipeline.Status.State = models.PipelineStateCanceled
		} else {
			pipeline.Status.State = models.PipelineStateCompleted
		}
		pipeline.Status.EndTime = time.Now()
	}
}

// SaveStepLogs 保存步骤日志
func (s *GamePipelineGRPCService) SaveStepLogs(ctx context.Context, pipelineID string, stepID string, logs []byte) error {
	s.logger.Debug("保存流水线步骤日志: 流水线ID: %s, 步骤ID: %s, 日志大小: %d字节", pipelineID, stepID, len(logs))
	// 获取流水线
	pipeline, err := s.store.Get(ctx, pipelineID)
	if err != nil {
		s.logger.Error("获取流水线失败: %v", err)
		return fmt.Errorf("获取流水线失败: %w", err)
	}
	if pipeline == nil {
		s.logger.Error("流水线不存在: %s", pipelineID)
		return fmt.Errorf("流水线不存在: %s", pipelineID)
	}

	// 更新步骤日志
	stepFound := false
	for i := range pipeline.Status.Steps {
		if pipeline.Status.Steps[i].ID == stepID {
			pipeline.Status.Steps[i].Logs = logs
			stepFound = true
			break
		}
	}

	if !stepFound {
		s.logger.Error("流水线步骤不存在: 流水线ID: %s, 步骤ID: %s", pipelineID, stepID)
		return fmt.Errorf("流水线步骤不存在: 流水线ID: %s, 步骤ID: %s", pipelineID, stepID)
	}

	err = s.store.Update(ctx, pipeline)
	if err != nil {
		s.logger.Error("保存流水线步骤日志失败: %v", err)
		return fmt.Errorf("保存流水线步骤日志失败: %w", err)
	}
	s.logger.Info("成功保存流水线步骤日志: 流水线ID: %s, 步骤ID: %s", pipelineID, stepID)
	return nil
}

// Cancel 取消流水线
func (s *GamePipelineGRPCService) Cancel(ctx context.Context, id string) error {
	s.logger.Debug("取消流水线: %s", id)

	// 获取当前流水线状态
	pipeline, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}

	// 创建新的状态对象
	status := &models.PipelineStatus{
		NodeID:       pipeline.Status.NodeID,
		State:        models.PipelineStateCanceled,
		CurrentStep:  pipeline.Status.CurrentStep,
		TotalSteps:   pipeline.Status.TotalSteps,
		StartTime:    pipeline.Status.StartTime,
		EndTime:      time.Now(),
		Steps:        pipeline.Status.Steps,
		ErrorMessage: "Pipeline was canceled",
		UpdatedAt:    time.Now(),
	}

	return s.UpdateStatus(ctx, id, status)
}

// Delete 删除流水线
func (s *GamePipelineGRPCService) Delete(ctx context.Context, id string) error {
	s.logger.Debug("删除流水线: %s", id)
	err := s.store.Delete(ctx, id)
	if err != nil {
		s.logger.Error("删除流水线失败: %v", err)
		return fmt.Errorf("删除流水线失败: %w", err)
	}
	s.logger.Info("成功删除流水线: %s", id)
	return nil
}

// Update 更新流水线
func (s *GamePipelineGRPCService) Update(ctx context.Context, pipeline *models.GamePipeline) error {
	s.logger.Debug("更新流水线: %s", pipeline.ID)

	// 1. 验证Pipeline信息
	if err := s.validatePipeline(pipeline); err != nil {
		return err
	}

	// 2. 更新到存储
	if err := s.store.Update(ctx, pipeline); err != nil {
		s.logger.Error("更新流水线失败: %v", err)
		return fmt.Errorf("更新流水线失败: %w", err)
	}

	s.logger.Info("成功更新流水线: %s", pipeline.ID)
	return nil
}
