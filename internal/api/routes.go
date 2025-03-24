package api

import (
	"github.com/gin-gonic/gin"
	"github.com/open-beagle/beagle-wind-game/internal/service"
)

// SetupRouter 设置路由
func SetupRouter(platformService *service.PlatformService, nodeService *service.NodeService,
	gameCardService *service.GameCardService, instanceService *service.InstanceService) *gin.Engine {
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
	platformHandler := NewPlatformHandler(platformService)
	nodeHandler := NewNodeHandler(nodeService)
	gameCardHandler := NewGameCardHandler(gameCardService)
	instanceHandler := NewInstanceHandler(instanceService)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 游戏节点管理
		nodes := v1.Group("/nodes")
		{
			nodes.GET("", nodeHandler.ListNodes)
			nodes.POST("", nodeHandler.CreateNode)
			nodes.GET("/:id", nodeHandler.GetNode)
			nodes.PUT("/:id", nodeHandler.UpdateNode)
			nodes.DELETE("/:id", nodeHandler.DeleteNode)
			// 更新节点状态
			nodes.PUT("/:id/status", nodeHandler.UpdateNodeStatus)
		}

		// 游戏平台管理
		platforms := v1.Group("/platforms")
		{
			platforms.GET("", platformHandler.ListPlatforms)
			platforms.POST("", platformHandler.CreatePlatform)
			platforms.GET("/:id", platformHandler.GetPlatform)
			platforms.PUT("/:id", platformHandler.UpdatePlatform)
			platforms.DELETE("/:id", platformHandler.DeletePlatform)
			// 远程访问
			platforms.GET("/:id/access", platformHandler.GetPlatformAccess)
			platforms.POST("/:id/access/refresh", platformHandler.RefreshPlatformAccess)
		}

		// 游戏卡片管理
		cards := v1.Group("/cards")
		{
			cards.GET("", gameCardHandler.ListGameCards)
			cards.POST("", gameCardHandler.CreateGameCard)
			cards.GET("/:id", gameCardHandler.GetGameCard)
			cards.PUT("/:id", gameCardHandler.UpdateGameCard)
			cards.DELETE("/:id", gameCardHandler.DeleteGameCard)
		}

		// 游戏实例管理
		instances := v1.Group("/instances")
		{
			instances.GET("", instanceHandler.ListInstances)
			instances.POST("", instanceHandler.CreateInstance)
			instances.GET("/:id", instanceHandler.GetInstance)
			instances.PUT("/:id", instanceHandler.UpdateInstance)
			instances.DELETE("/:id", instanceHandler.DeleteInstance)
			// 实例操作
			instances.POST("/:id/start", instanceHandler.StartInstance)
			instances.POST("/:id/stop", instanceHandler.StopInstance)
		}
	}

	// 前端路由支持
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	return r
}
