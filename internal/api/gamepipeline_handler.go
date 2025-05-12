package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// GamePipelineHandler 处理游戏节点流水线相关的 HTTP 请求
type GamePipelineHandler struct {
	svc *service.GamePipelineService
}

// NewGamePipelineHandler 创建新的 GamePipelineHandler
func NewGamePipelineHandler(svc *service.GamePipelineService) *GamePipelineHandler {
	return &GamePipelineHandler{
		svc: svc,
	}
}

// RegisterRoutes 注册路由
func (h *GamePipelineHandler) RegisterRoutes(r *gin.Engine) {
	pipelines := r.Group("/api/v1/pipelines")
	{
		pipelines.GET("", h.List)
		pipelines.GET("/:id", h.Get)
		pipelines.POST("/:id/cancel", h.Cancel)
		pipelines.POST("/:id/delete", h.Delete)
	}
}

// List 获取流水线列表
// @Summary 获取流水线列表
// @Description 获取游戏节点流水线列表，支持分页、状态筛选和时间范围查询
// @Tags 游戏节点流水线
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param status query string false "流水线状态(pending/running/completed/failed/canceled)"
// @Param node_id query string false "节点ID"
// @Param start_time query string false "开始时间(ISO 8601格式)"
// @Param end_time query string false "结束时间(ISO 8601格式)"
// @Param sort_by query string false "排序字段(created_at/updated_at/status)" Enums(created_at,updated_at,status)
// @Param sort_order query string false "排序方向" Enums(asc,desc) default(desc)
// @Success 200 {object} service.PipelineListResult "流水线列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/pipelines [get]
func (h *GamePipelineHandler) List(c *gin.Context) {
	pipelines, err := h.svc.List(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, pipelines)
}

// Get 获取流水线详情
// @Summary 获取流水线详情
// @Description 根据流水线ID获取游戏节点流水线详情
// @Tags 游戏节点流水线
// @Accept json
// @Produce json
// @Param id path string true "流水线ID"
// @Success 200 {object} models.Pipeline "流水线详情"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "流水线不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/pipelines/{id} [get]
func (h *GamePipelineHandler) Get(c *gin.Context) {
	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "流水线ID不能为空",
		})
		return
	}

	// 调用服务层获取数据
	pipeline, err := h.svc.Get(c, pipelineID)
	if err != nil {
		if err.Error() == "流水线不存在" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "流水线不存在",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取流水线详情失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pipeline,
	})
}

// Cancel 取消流水线
// @Summary 取消流水线
// @Description 取消指定的游戏节点流水线
// @Tags 游戏节点流水线
// @Accept json
// @Produce json
// @Param id path string true "流水线ID"
// @Success 200 {object} map[string]interface{} "操作结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "流水线不存在"
// @Failure 409 {object} map[string]interface{} "流水线状态不允许取消"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/pipelines/{id}/cancel [post]
func (h *GamePipelineHandler) Cancel(c *gin.Context) {
	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "流水线ID不能为空",
		})
		return
	}

	// 调用服务层取消流水线
	err := h.svc.Cancel(c.Request.Context(), pipelineID)
	if err != nil {
		switch err.Error() {
		case "流水线不存在":
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "流水线不存在",
				"error":   err.Error(),
			})
		case "流水线状态不允许取消":
			c.JSON(http.StatusConflict, gin.H{
				"code":    http.StatusConflict,
				"message": "流水线状态不允许取消",
				"error":   err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "取消流水线失败",
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

// Delete 删除流水线
// @Summary 删除流水线
// @Description 删除指定的游戏节点流水线
// @Tags 游戏节点流水线
// @Accept json
// @Produce json
// @Param id path string true "流水线ID"
// @Success 200 {object} map[string]interface{} "操作结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "流水线不存在"
// @Failure 409 {object} map[string]interface{} "流水线状态不允许删除"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/pipelines/{id}/delete [post]
func (h *GamePipelineHandler) Delete(c *gin.Context) {
	// 获取目标流水线ID
	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(400, gin.H{"error": "无效的流水线ID"})
		return
	}

	// 删除流水线
	err := h.svc.Delete(c, pipelineID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "流水线已删除"})
}
