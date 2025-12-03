package routes

import (
	"api-postgre/handlers"
	"api-postgre/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes 设置认证相关路由
//
// 路由划分：
//   - 公开路由：不需要认证，任何人都可以访问
//   - 受保护路由：需要 JWT 认证，只有登录用户才能访问
func SetupAuthRoutes(router *gin.Engine) {
	// 公开路由组（不需要认证）
	public := router.Group("/api/auth")
	{
		public.POST("/register", handlers.Register) // 注册
		public.POST("/login", handlers.Login)       // 登录
	}

	// 受保护路由组（需要认证）
	protected := router.Group("/api/auth")
	protected.Use(middleware.AuthMiddleware()) // 应用认证中间件
	{
		protected.POST("/logout", handlers.Logout)    // 登出
		protected.GET("/me", handlers.GetCurrentUser) // 获取当前用户信息
	}
}
