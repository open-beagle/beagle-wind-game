package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建默认的gin引擎
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 游戏机管理
		gameNode := v1.Group("/game-nodes")
		{
			gameNode.GET("", listGameNodes)
			gameNode.POST("", createGameNode)
			gameNode.GET("/:id", getGameNode)
			gameNode.PUT("/:id", updateGameNode)
			gameNode.DELETE("/:id", deleteGameNode)
		}

		// 游戏平台管理
		platform := v1.Group("/platforms")
		{
			platform.GET("", listPlatforms)
			platform.POST("", createPlatform)
			platform.GET("/:id", getPlatform)
			platform.PUT("/:id", updatePlatform)
			platform.DELETE("/:id", deletePlatform)
		}

		// 游戏卡片管理
		gameCard := v1.Group("/game-cards")
		{
			gameCard.GET("", listGameCards)
			gameCard.POST("", createGameCard)
			gameCard.GET("/:id", getGameCard)
			gameCard.PUT("/:id", updateGameCard)
			gameCard.DELETE("/:id", deleteGameCard)
		}

		// 游戏实例管理
		instance := v1.Group("/instances")
		{
			instance.GET("", listInstances)
			instance.POST("", createInstance)
			instance.GET("/:id", getInstance)
			instance.PUT("/:id", updateInstance)
			instance.DELETE("/:id", deleteInstance)
		}
	}

	return r
}

// 游戏机管理处理函数
func listGameNodes(c *gin.Context)  { c.JSON(200, gin.H{"message": "list game nodes"}) }
func createGameNode(c *gin.Context) { c.JSON(200, gin.H{"message": "create game node"}) }
func getGameNode(c *gin.Context)    { c.JSON(200, gin.H{"message": "get game node"}) }
func updateGameNode(c *gin.Context) { c.JSON(200, gin.H{"message": "update game node"}) }
func deleteGameNode(c *gin.Context) { c.JSON(200, gin.H{"message": "delete game node"}) }

// 游戏平台管理处理函数
func listPlatforms(c *gin.Context)  { c.JSON(200, gin.H{"message": "list platforms"}) }
func createPlatform(c *gin.Context) { c.JSON(200, gin.H{"message": "create platform"}) }
func getPlatform(c *gin.Context)    { c.JSON(200, gin.H{"message": "get platform"}) }
func updatePlatform(c *gin.Context) { c.JSON(200, gin.H{"message": "update platform"}) }
func deletePlatform(c *gin.Context) { c.JSON(200, gin.H{"message": "delete platform"}) }

// 游戏卡片管理处理函数
func listGameCards(c *gin.Context)  { c.JSON(200, gin.H{"message": "list game cards"}) }
func createGameCard(c *gin.Context) { c.JSON(200, gin.H{"message": "create game card"}) }
func getGameCard(c *gin.Context)    { c.JSON(200, gin.H{"message": "get game card"}) }
func updateGameCard(c *gin.Context) { c.JSON(200, gin.H{"message": "update game card"}) }
func deleteGameCard(c *gin.Context) { c.JSON(200, gin.H{"message": "delete game card"}) }

// 游戏实例管理处理函数
func listInstances(c *gin.Context)  { c.JSON(200, gin.H{"message": "list instances"}) }
func createInstance(c *gin.Context) { c.JSON(200, gin.H{"message": "create instance"}) }
func getInstance(c *gin.Context)    { c.JSON(200, gin.H{"message": "get instance"}) }
func updateInstance(c *gin.Context) { c.JSON(200, gin.H{"message": "update instance"}) }
func deleteInstance(c *gin.Context) { c.JSON(200, gin.H{"message": "delete instance"}) }
