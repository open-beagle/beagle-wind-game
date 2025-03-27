package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/open-beagle/beagle-wind-game/internal/service"
	"github.com/open-beagle/beagle-wind-game/internal/types"
)

// GameNodePipelineHandler 处理游戏节点流水线相关的 HTTP 请求
type GameNodePipelineHandler struct {
	svc *service.GameNodePipelineService
}

// NewGameNodePipelineHandler 创建新的 GameNodePipelineHandler
func NewGameNodePipelineHandler(svc *service.GameNodePipelineService) *GameNodePipelineHandler {
	return &GameNodePipelineHandler{
		svc: svc,
	}
}

// RegisterRoutes 注册路由
func (h *GameNodePipelineHandler) RegisterRoutes(r *gin.Engine) {
	pipelines := r.Group("/api/v1/pipelines")
	{
		pipelines.GET("", h.ListPipelines)
		pipelines.GET("/:id", h.GetPipeline)
		pipelines.POST("/:id/cancel", h.CancelPipeline)
		pipelines.POST("/:id/delete", h.DeletePipeline)
	}
}

// ListPipelines 获取流水线列表
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
func (h *GameNodePipelineHandler) ListPipelines(c *gin.Context) {
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

	// 解析时间参数
	var startTime, endTime time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "开始时间格式错误，请使用 ISO 8601 格式",
				"error":   err.Error(),
			})
			return
		}
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "结束时间格式错误，请使用 ISO 8601 格式",
				"error":   err.Error(),
			})
			return
		}
	}

	// 验证时间范围
	if !startTime.IsZero() && !endTime.IsZero() && startTime.After(endTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "开始时间不能晚于结束时间",
		})
		return
	}

	// 构建查询参数
	params := types.PipelineListParams{
		Page:     page,
		PageSize: size,
		Status:   c.Query("status"),
	}

	// 调用服务层获取数据
	result, err := h.svc.ListPipelines(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取流水线列表失败",
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

// GetPipeline 获取流水线详情
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
func (h *GameNodePipelineHandler) GetPipeline(c *gin.Context) {
	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "流水线ID不能为空",
		})
		return
	}

	// 调用服务层获取数据
	pipeline, err := h.svc.GetPipeline(pipelineID)
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

// CancelPipeline 取消流水线
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
func (h *GameNodePipelineHandler) CancelPipeline(c *gin.Context) {
	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "流水线ID不能为空",
		})
		return
	}

	// 调用服务层取消流水线
	err := h.svc.CancelPipeline(c.Request.Context(), pipelineID)
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

// DeletePipeline 删除流水线
// @Summary 删除流水线
// @Description 删除指定的游戏节点流水线
// @Tags 游戏节点流水线
// @Accept json
// @Produce json
// @Param id path string true "流水线ID"
// @Param force query bool false "是否强制删除" default(false)
// @Success 200 {object} map[string]interface{} "操作结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "流水线不存在"
// @Failure 409 {object} map[string]interface{} "流水线状态不允许删除"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/pipelines/{id}/delete [post]
func (h *GameNodePipelineHandler) DeletePipeline(c *gin.Context) {
	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "流水线ID不能为空",
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

	// 调用服务层删除流水线
	err = h.svc.DeletePipeline(pipelineID, force)
	if err != nil {
		switch err.Error() {
		case "流水线不存在":
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "流水线不存在",
				"error":   err.Error(),
			})
		case "流水线状态不允许删除":
			c.JSON(http.StatusConflict, gin.H{
				"code":    http.StatusConflict,
				"message": "流水线状态不允许删除",
				"error":   err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "删除流水线失败",
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
