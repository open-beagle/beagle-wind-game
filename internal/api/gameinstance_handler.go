package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// InstanceHandler 游戏实例API处理器
type InstanceHandler struct {
	instanceService *service.GameInstanceService
}

// NewInstanceHandler 创建游戏实例API处理器
func NewInstanceHandler(instanceService *service.GameInstanceService) *InstanceHandler {
	return &InstanceHandler{
		instanceService: instanceService,
	}
}

// ListInstances 获取实例列表
// @Summary 获取实例列表
// @Description 获取游戏实例列表，支持分页和筛选
// @Tags 游戏实例
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Param status query string false "实例状态(running/stopped/paused/error)"
// @Param node_id query string false "节点ID"
// @Param card_id query string false "游戏卡片ID"
// @Param platform_id query string false "平台ID"
// @Success 200 {object} service.InstanceListResult "实例列表"
// @Router /api/v1/instances [get]
func (h *InstanceHandler) ListInstances(c *gin.Context) {
	var params service.InstanceListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	result, err := h.instanceService.ListInstances(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetInstance 获取实例详情
// @Summary 获取实例详情
// @Description, 根据实例ID获取游戏实例详情
// @Tags 游戏实例
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 200 {object} models.GameInstance "实例详情"
// @Router /api/v1/instances/{id} [get]
func (h *InstanceHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实例ID不能为空"})
		return
	}

	instance, err := h.instanceService.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if instance.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// CreateInstance 创建实例
// @Summary 创建实例
// @Description 创建新的游戏实例
// @Tags 游戏实例
// @Accept json
// @Produce json
// @Param body body service.CreateInstanceParams true "创建实例参数"
// @Success 201 {object} gin.H "包含新创建的实例ID"
// @Router /api/v1/instances [post]
func (h *InstanceHandler) CreateInstance(c *gin.Context) {
	var params service.CreateInstanceParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.instanceService.CreateInstance(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdateInstance 更新实例
// @Summary 更新实例
// @Description 更新游戏实例信息
// @Tags 游戏实例
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Param body body service.UpdateInstanceParams true "更新实例参数"
// @Success 204 "无内容"
// @Router /api/v1/instances/{id} [put]
func (h *InstanceHandler) UpdateInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实例ID不能为空"})
		return
	}

	// 验证实例是否存在
	instance, err := h.instanceService.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if instance.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	var params service.UpdateInstanceParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将参数转换为 GameInstance
	instance = models.GameInstance{
		ID:          id,
		Status:      params.Status,
		Resources:   params.Resources,
		Performance: params.Performance,
		SaveData:    params.SaveData,
		Config:      params.Config,
		Backup:      params.Backup,
	}

	err = h.instanceService.UpdateInstance(id, instance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteInstance 删除实例
// @Summary 删除实例
// @Description 删除指定的游戏实例
// @Tags 游戏实例
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 204 "无内容"
// @Router /api/v1/instances/{id} [delete]
func (h *InstanceHandler) DeleteInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实例ID不能为空"})
		return
	}

	// 验证实例是否存在
	instance, err := h.instanceService.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if instance.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	err = h.instanceService.DeleteInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartInstance 启动实例
// @Summary 启动实例
// @Description 启动指定的游戏实例
// @Tags 游戏实例
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 204 "无内容"
// @Router /api/v1/instances/{id}/start [post]
func (h *InstanceHandler) StartInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实例ID不能为空"})
		return
	}

	// 验证实例是否存在
	instance, err := h.instanceService.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if instance.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	err = h.instanceService.StartInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// StopInstance 停止实例
// @Summary 停止实例
// @Description 停止指定的游戏实例
// @Tags 游戏实例
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 204 "无内容"
// @Router /api/v1/instances/{id}/stop [post]
func (h *InstanceHandler) StopInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实例ID不能为空"})
		return
	}

	// 验证实例是否存在
	instance, err := h.instanceService.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if instance.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	err = h.instanceService.StopInstance(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
