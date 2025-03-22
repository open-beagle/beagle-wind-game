package main

import (
	"fmt"
	"log"

	"github.com/open-beagle/beagle-wind-game/internal/api"
	"github.com/open-beagle/beagle-wind-game/internal/config"
	"github.com/open-beagle/beagle-wind-game/pkg/middleware"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置路由
	r := api.SetupRouter()

	// 添加中间件
	r.Use(middleware.Logger())

	// 启动服务器
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
