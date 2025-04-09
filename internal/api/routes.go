package api

import (
	"github.com/gin-gonic/gin"

	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// SetupRouter 设置路由
func SetupRouter(gameplatformService *service.GamePlatformService, gamenodeService *service.GameNodeService,
	gameCardService *service.GameCardService, gameinstanceService *service.GameInstanceService,
	gamenodePipelineService *service.GameNodePipelineService) *gin.Engine {
	// 创建默认的gin引擎
	r := gin.Default()

	// 添加CORS中间件
	r.Use(SetupCORS())

	// 提供静态文件服务
	r.Static("/assets", "./static/assets")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 创建处理器
	gameplatformHandler := NewGamePlatformHandler(gameplatformService)
	gamenodeHandler := NewGameNodeHandler(gamenodeService)
	gameCardHandler := NewGameCardHandler(gameCardService)
	gameinstanceHandler := NewGameInstanceHandler(gameinstanceService)
	gamenodePipelineHandler := NewGameNodePipelineHandler(gamenodePipelineService)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 游戏节点管理
		nodes := v1.Group("/nodes")
		{
			nodes.GET("", gamenodeHandler.ListNodes)
			nodes.GET("/:id", gamenodeHandler.GetNode)
			nodes.POST("/:id/update", gamenodeHandler.UpdateNode)
			nodes.POST("/:id/delete", gamenodeHandler.DeleteNode)
		}

		// 游戏平台管理
		platforms := v1.Group("/platforms")
		{
			platforms.GET("", gameplatformHandler.List)
			platforms.GET("/:id", gameplatformHandler.Get)
			platforms.POST("", gameplatformHandler.Create)
			platforms.POST("/:id/update", gameplatformHandler.Update)
			platforms.POST("/:id/delete", gameplatformHandler.Delete)
		}

		// 游戏卡片管理
		cards := v1.Group("/cards")
		{
			cards.GET("", gameCardHandler.List)
			cards.GET("/:id", gameCardHandler.Get)
			cards.POST("", gameCardHandler.Create)
			cards.POST("/:id/update", gameCardHandler.Update)
			cards.POST("/:id/delete", gameCardHandler.Delete)
		}

		// 游戏实例管理
		instances := v1.Group("/instances")
		{
			instances.GET("", gameinstanceHandler.List)
			instances.GET("/:id", gameinstanceHandler.Get)
			instances.POST("", gameinstanceHandler.Create)
			instances.POST("/:id/update", gameinstanceHandler.Update)
			instances.POST("/:id/delete", gameinstanceHandler.Delete)
			// 实例操作
			instances.POST("/:id/start", gameinstanceHandler.Start)
			instances.POST("/:id/stop", gameinstanceHandler.Stop)
		}

		// 游戏节点流水线管理
		pipelines := v1.Group("/pipelines")
		{
			pipelines.GET("", gamenodePipelineHandler.List)
			pipelines.GET("/:id", gamenodePipelineHandler.Get)
			pipelines.POST("/:id/cancel", gamenodePipelineHandler.Cancel)
			pipelines.POST("/:id/delete", gamenodePipelineHandler.Delete)
		}
	}

	// 前端路由支持
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	return r
}
