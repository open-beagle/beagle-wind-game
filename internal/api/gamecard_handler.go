package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// GameCardHandler 游戏卡片API处理器
type GameCardHandler struct {
	gameCardService *service.GameCardService
}

// NewGameCardHandler 创建游戏卡片API处理器
func NewGameCardHandler(gameCardService *service.GameCardService) *GameCardHandler {
	return &GameCardHandler{
		gameCardService: gameCardService,
	}
}

// ListGameCards 获取游戏卡片列表
// @Summary 获取游戏卡片列表
// @Description 获取游戏卡片列表，支持分页、搜索和筛选
// @Tags 游戏卡片
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Param type query string false "游戏类型"
// @Param category query string false "游戏分类"
// @Success 200 {object} service.GameCardListResult "游戏卡片列表"
// @Router /api/v1/game-cards [get]
func (h *GameCardHandler) ListGameCards(c *gin.Context) {
	var params service.GameCardListParams
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

	result, err := h.gameCardService.ListGameCards(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGameCard 获取游戏卡片详情
// @Summary 获取游戏卡片详情
// @Description 根据游戏卡片ID获取详情
// @Tags 游戏卡片
// @Accept json
// @Produce json
// @Param id path string true "游戏卡片ID"
// @Success 200 {object} models.GameCard "游戏卡片详情"
// @Router /api/v1/game-cards/{id} [get]
func (h *GameCardHandler) GetGameCard(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "游戏卡片ID不能为空"})
		return
	}

	card, err := h.gameCardService.GetGameCard(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if card.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "游戏卡片不存在"})
		return
	}

	c.JSON(http.StatusOK, card)
}

// CreateGameCard 创建游戏卡片
// @Summary 创建游戏卡片
// @Description 创建新的游戏卡片
// @Tags 游戏卡片
// @Accept json
// @Produce json
// @Param body body models.GameCard true "游戏卡片信息"
// @Success 201 {object} gin.H "包含新创建的游戏卡片ID"
// @Router /api/v1/game-cards [post]
func (h *GameCardHandler) CreateGameCard(c *gin.Context) {
	var card models.GameCard
	if err := c.ShouldBindJSON(&card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.gameCardService.CreateGameCard(card)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdateGameCard 更新游戏卡片
// @Summary 更新游戏卡片
// @Description 更新现有的游戏卡片
// @Tags 游戏卡片
// @Accept json
// @Produce json
// @Param id path string true "游戏卡片ID"
// @Param body body models.GameCard true "游戏卡片信息"
// @Success 204 "无内容"
// @Router /api/v1/game-cards/{id} [put]
func (h *GameCardHandler) UpdateGameCard(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "游戏卡片ID不能为空"})
		return
	}

	// 验证卡片是否存在
	card, err := h.gameCardService.GetGameCard(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if card.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "游戏卡片不存在"})
		return
	}

	var updatedCard models.GameCard
	if err := c.ShouldBindJSON(&updatedCard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 确保ID一致
	updatedCard.ID = id

	err = h.gameCardService.UpdateGameCard(id, updatedCard)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteGameCard 删除游戏卡片
// @Summary 删除游戏卡片
// @Description 删除指定的游戏卡片
// @Tags 游戏卡片
// @Accept json
// @Produce json
// @Param id path string true "游戏卡片ID"
// @Success 204 "无内容"
// @Router /api/v1/game-cards/{id} [delete]
func (h *GameCardHandler) DeleteGameCard(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "游戏卡片ID不能为空"})
		return
	}

	// 验证卡片是否存在
	card, err := h.gameCardService.GetGameCard(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if card.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "游戏卡片不存在"})
		return
	}

	err = h.gameCardService.DeleteGameCard(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
