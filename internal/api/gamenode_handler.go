package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// NodeHandler 节点API处理器
type NodeHandler struct {
	nodeService *service.GameNodeService
}

// NewNodeHandler 创建节点API处理器
func NewNodeHandler(nodeService *service.GameNodeService) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
	}
}

// ListNodes 获取节点列表
// @Summary 获取节点列表
// @Description 获取游戏节点列表，支持分页、搜索和状态筛选
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Param status query string false "节点状态(online/offline/maintenance)"
// @Success 200 {object} service.GameNodeListResult "节点列表"
// @Router /api/v1/game-nodes [get]
func (h *NodeHandler) ListNodes(c *gin.Context) {
	var params service.GameNodeListParams
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

	result, err := h.nodeService.ListNodes(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetNode 获取节点详情
// @Summary 获取节点详情
// @Description 根据节点ID获取游戏节点详情
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Success 200 {object} models.GameNode "节点详情"
// @Router /api/v1/game-nodes/{id} [get]
func (h *NodeHandler) GetNode(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "节点ID不能为空"})
		return
	}

	node, err := h.nodeService.GetNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if node.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}

	c.JSON(http.StatusOK, node)
}

// UpdateNodeStatus 更新节点状态
// @Summary 更新节点状态
// @Description 更新游戏节点的状态
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Param body body struct{Status string `json:"status"`;Reason string `json:"reason,omitempty"`} true "状态信息"
// @Success 204 "无内容"
// @Router /api/v1/game-nodes/{id}/status [put]
func (h *NodeHandler) UpdateNodeStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "节点ID不能为空"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=online offline maintenance"`
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证节点是否存在
	node, err := h.nodeService.GetNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if node.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}

	err = h.nodeService.UpdateNodeStatus(id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateNode 创建节点
// @Summary 创建节点
// @Description 创建新的游戏节点
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param body body models.GameNode true "节点信息"
// @Success 201 {object} gin.H "包含新创建的节点ID"
// @Router /api/v1/game-nodes [post]
func (h *NodeHandler) CreateNode(c *gin.Context) {
	var node models.GameNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := h.nodeService.CreateNode(node)
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdateNode 更新节点
// @Summary 更新节点
// @Description 更新现有的游戏节点
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Param body body models.GameNode true "节点信息"
// @Success 204 "无内容"
// @Router /api/v1/game-nodes/{id} [put]
func (h *NodeHandler) UpdateNode(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "节点ID不能为空"})
		return
	}

	// 验证节点是否存在
	node, err := h.nodeService.GetNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if node.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}

	var updatedNode models.GameNode
	if err := c.ShouldBindJSON(&updatedNode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 确保ID一致
	updatedNode.ID = id

	err = h.nodeService.UpdateNode(id, updatedNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteNode 删除节点
// @Summary 删除节点
// @Description 删除指定的游戏节点
// @Tags 游戏节点
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Success 204 "无内容"
// @Router /api/v1/game-nodes/{id} [delete]
func (h *NodeHandler) DeleteNode(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "节点ID不能为空"})
		return
	}

	// 验证节点是否存在
	node, err := h.nodeService.GetNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if node.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}

	err = h.nodeService.DeleteNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
