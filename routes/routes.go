package routes

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册所有路由
// 这是路由的统一入口，管理所有模块的路由注册
func RegisterRoutes(app *gin.Engine) {
	// 注册认证路由（登录、注册等）
	SetupAuthRoutes(app)

	// 注册业务路由
	api := app.Group("/api")
	ContactRoutes(api)
}
