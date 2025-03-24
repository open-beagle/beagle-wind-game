package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupCORS 配置CORS中间件
func SetupCORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// ResponseError 错误响应
func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// ResponseOK 成功响应
func ResponseOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// ResponseCreated 创建成功响应
func ResponseCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}
