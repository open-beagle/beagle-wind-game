package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// GamePlatformHandler 平台API处理器
type GamePlatformHandler struct {
	platformService *service.GamePlatformService
}

// NewPlatformHandler 创建平台API处理器
func NewPlatformHandler(platformService *service.GamePlatformService) *GamePlatformHandler {
	return &GamePlatformHandler{
		platformService: platformService,
	}
}

// ListPlatforms 获取平台列表
// @Summary 获取平台列表
// @Description 获取游戏平台列表，支持分页、搜索和状态筛选
// @Tags 游戏平台
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Param status query string false "平台状态(active/maintenance/inactive)"
// @Success 200 {object} service.PlatformListResult "平台列表"
// @Router /api/v1/platforms [get]
func (h *GamePlatformHandler) ListPlatforms(c *gin.Context) {
	var params service.PlatformListParams
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

	result, err := h.platformService.ListPlatforms(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPlatform 获取平台详情
// @Summary 获取平台详情
// @Description 根据平台ID获取游戏平台详情
// @Tags 游戏平台
// @Accept json
// @Produce json
// @Param id path string true "平台ID"
// @Success 200 {object} models.GamePlatform "平台详情"
// @Router /api/v1/platforms/{id} [get]
func (h *GamePlatformHandler) GetPlatform(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "平台ID不能为空"})
		return
	}

	platform, err := h.platformService.GetPlatform(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if platform.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "平台不存在"})
		return
	}

	c.JSON(http.StatusOK, platform)
}

// GetPlatformAccess 获取平台远程访问链接
// @Summary 获取平台远程访问链接
// @Description 获取平台远程访问的WebRTC链接
// @Tags 游戏平台
// @Accept json
// @Produce json
// @Param id path string true "平台ID"
// @Success 200 {object} service.PlatformAccessResult "访问链接"
// @Router /api/v1/platforms/{id}/access [get]
func (h *GamePlatformHandler) GetPlatformAccess(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "平台ID不能为空"})
		return
	}

	// 验证平台是否存在
	platform, err := h.platformService.GetPlatform(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if platform.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "平台不存在"})
		return
	}

	result, err := h.platformService.GetPlatformAccess(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RefreshPlatformAccess 刷新平台远程访问链接
// @Summary 刷新平台远程访问链接
// @Description 刷新平台远程访问的WebRTC链接
// @Tags 游戏平台
// @Accept json
// @Produce json
// @Param id path string true "平台ID"
// @Success 200 {object} service.PlatformAccessResult "访问链接"
// @Router /api/v1/platforms/{id}/access/refresh [post]
func (h *GamePlatformHandler) RefreshPlatformAccess(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "平台ID不能为空"})
		return
	}

	// 验证平台是否存在
	platform, err := h.platformService.GetPlatform(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if platform.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "平台不存在"})
		return
	}

	result, err := h.platformService.RefreshPlatformAccess(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreatePlatform 创建平台
// @Summary 创建平台
// @Description 创建新的游戏平台
// @Tags 游戏平台
// @Accept json
// @Produce json
// @Param body body models.GamePlatform true "平台信息"
// @Success 201 {object} gin.H "包含新创建的平台ID"
// @Router /api/v1/platforms [post]
func (h *GamePlatformHandler) CreatePlatform(c *gin.Context) {
	var platform models.GamePlatform
	if err := c.ShouldBindJSON(&platform); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证平台数据
	if platform.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "平台ID不能为空"})
		return
	}

	if platform.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "平台名称不能为空"})
		return
	}

	// 创建平台
	id, err := h.platformService.CreatePlatform(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdatePlatform 更新平台
// @Summary 更新平台
// @Description 更新现有的游戏平台
// @Tags 游戏平台
// @Accept json
// @Produce json
// @Param id path string true "平台ID"
// @Param body body models.GamePlatform true "平台信息"
// @Success 204 "无内容"
// @Router /api/v1/platforms/{id} [put]
func (h *GamePlatformHandler) UpdatePlatform(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "平台ID不能为空"})
		return
	}

	// 验证平台是否存在
	_, err := h.platformService.GetPlatform(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 绑定请求数据
	var platform models.GamePlatform
	if err := c.ShouldBindJSON(&platform); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新平台
	err = h.platformService.UpdatePlatform(id, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeletePlatform 删除平台
// @Summary 删除平台
// @Description 删除指定的游戏平台
// @Tags 游戏平台
// @Accept json
// @Produce json
// @Param id path string true "平台ID"
// @Success 204 "无内容"
// @Router /api/v1/platforms/{id} [delete]
func (h *GamePlatformHandler) DeletePlatform(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "平台ID不能为空"})
		return
	}

	// 删除平台
	err := h.platformService.DeletePlatform(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
