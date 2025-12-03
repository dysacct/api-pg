package main

import (
	"api-postgre/config"
	"api-postgre/routes"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	godotenv.Load()

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080" // 默认端口
	}

	// 连接数据库并自动迁移
	config.ConnectDB()

	// 创建 Gin 实例
	app := gin.Default()

	// 注册所有路由
	routes.RegisterRoutes(app)

	// 启动服务器
	fmt.Printf("🚀 服务器启动在端口 %s\n", PORT)
	app.Run(fmt.Sprintf(":%s", PORT))
}
