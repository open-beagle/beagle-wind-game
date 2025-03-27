package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// GameNodeHandler 处理游戏节点相关的 HTTP 请求
type GameNodeHandler struct {
	svc *service.GameNodeService
}

// NewGameNodeHandler 创建新的 GameNodeHandler
func NewGameNodeHandler(svc *service.GameNodeService) *GameNodeHandler {
	return &GameNodeHandler{
		svc: svc,
	}
}

// RegisterRoutes 注册路由
func (h *GameNodeHandler) RegisterRoutes(r *gin.Engine) {
	nodes := r.Group("/api/v1/nodes")
	{
		nodes.GET("", h.ListNodes)
		nodes.GET("/:id", h.GetNode)
		nodes.POST("/:id/update", h.UpdateNode)
		nodes.POST("/:id/delete", h.DeleteNode)
	}
}

// ListNodes 获取节点列表
// @Summary 获取节点列表
// @Description 获取游戏节点列表，支持分页、搜索和状态筛选
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词"
// @Param status query string false "节点状态(online/offline/maintenance)"
// @Param type query string false "节点类型"
// @Param region query string false "区域"
// @Param sort_by query string false "排序字段(created_at/updated_at/status)"
// @Param sort_order query string false "排序方向(asc/desc)"
// @Success 200 {object} service.GameNodeListResult "节点列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/nodes [get]
func (h *GameNodeHandler) ListNodes(c *gin.Context) {
	// 解析查询参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "页码格式错误",
			"error":   err.Error(),
		})
		return
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "每页数量格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	// 构建查询参数
	params := service.GameNodeListParams{
		Page:      page,
		PageSize:  size,
		Keyword:   c.Query("keyword"),
		Status:    c.Query("status"),
		Type:      c.Query("type"),
		Region:    c.Query("region"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// 调用服务层获取数据
	result, err := h.svc.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取节点列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// GetNode 获取节点详情
// @Summary 获取节点详情
// @Description 根据节点ID获取游戏节点详情
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Param include_metrics query bool false "是否包含性能指标" default(false)
// @Success 200 {object} models.GameNode "节点详情"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "节点不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/nodes/{id} [get]
func (h *GameNodeHandler) GetNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "节点ID不能为空",
		})
		return
	}

	includeMetrics, err := strconv.ParseBool(c.DefaultQuery("include_metrics", "false"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "include_metrics 参数格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 调用服务层获取数据
	node, err := h.svc.Get(nodeID)
	if err != nil {
		if err.Error() == "节点不存在" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "节点不存在",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取节点详情失败",
			"error":   err.Error(),
		})
		return
	}

	// 如果不包含指标，则清空指标数据
	if !includeMetrics {
		node.Status.Metrics = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    node,
	})
}

// UpdateNode 更新节点信息
// @Summary 更新节点信息
// @Description 更新游戏节点的基本信息
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Param body body struct{Name string `json:"name" binding:"required"`;Model string `json:"model" binding:"required"`;Type string `json:"type" binding:"required,oneof=physical virtual"`;Location string `json:"location" binding:"required"`;Labels map[string]string `json:"labels"`;MaintenanceMode bool `json:"maintenance_mode,omitempty"`} true "节点信息"
// @Success 200 {object} models.GameNode "更新后的节点信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "节点不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/nodes/{id}/update [post]
func (h *GameNodeHandler) UpdateNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "节点ID不能为空",
		})
		return
	}

	var req struct {
		Name            string            `json:"name" binding:"required"`
		Model           string            `json:"model" binding:"required"`
		Type            string            `json:"type" binding:"required,oneof=physical virtual"`
		Location        string            `json:"location" binding:"required"`
		Labels          map[string]string `json:"labels"`
		MaintenanceMode bool              `json:"maintenance_mode,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取现有节点
	node, err := h.svc.Get(nodeID)
	if err != nil {
		if err.Error() == "节点不存在" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "节点不存在",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取节点失败",
			"error":   err.Error(),
		})
		return
	}

	// 更新节点信息
	node.Name = req.Name
	node.Model = req.Model
	node.Type = models.GameNodeType(req.Type)
	node.Location = req.Location
	node.Labels = req.Labels
	if req.MaintenanceMode {
		node.Status.State = models.GameNodeState("maintenance")
	}

	// 调用服务层更新数据
	err = h.svc.Update(nodeID, *node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "更新节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    node,
	})
}

// DeleteNode 删除节点
// @Summary 删除节点
// @Description 删除指定的游戏节点
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Param force query bool false "是否强制删除" default(false)
// @Success 200 {object} map[string]interface{} "操作结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "节点不存在"
// @Failure 409 {object} map[string]interface{} "节点状态不允许删除"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/nodes/{id}/delete [post]
func (h *GameNodeHandler) DeleteNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "节点ID不能为空",
		})
		return
	}

	force, err := strconv.ParseBool(c.DefaultQuery("force", "false"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "force 参数格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 调用服务层删除节点
	err = h.svc.Delete(nodeID, force)
	if err != nil {
		switch err.Error() {
		case "节点不存在":
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "节点不存在",
				"error":   err.Error(),
			})
		case "节点状态不允许删除":
			c.JSON(http.StatusConflict, gin.H{
				"code":    http.StatusConflict,
				"message": "节点状态不允许删除",
				"error":   err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "删除节点失败",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
