package middleware

import (
	"api-postgre/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
//
// 工作流程：
//  1. 从请求中提取 token（支持 Cookie 和 Header 两种方式）
//  2. 验证 token 的有效性
//  3. 将用户 ID 存入上下文，供后续 handler 使用
//  4. 如果验证失败，直接返回 401 错误，不继续处理请求
//
// 使用方式：
//
//	需要认证的路由组：router.Use(middleware.AuthMiddleware())
//	不需要认证的路由：不添加此中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 尝试从 Cookie 中获取 token
		token, err := c.Cookie("token")

		// 2. 如果 Cookie 中没有，尝试从 Authorization Header 中获取
		if err != nil || token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "未提供认证 token",
				})
				c.Abort() // 终止请求处理链
				return
			}

			// Authorization Header 格式：Bearer <token>
			// 需要提取出 token 部分
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Authorization Header 格式错误，应为 'Bearer <token>'",
				})
				c.Abort()
				return
			}
			token = parts[1]
		}

		// 3. 验证 token
		claims, err := util.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的 token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 4. 将用户 ID 存入上下文
		// 后续的 handler 可以通过 c.Get("user_id") 获取当前用户 ID
		c.Set("user_id", claims.UserID)

		// 5. 继续处理请求
		c.Next()
	}
}

// GetCurrentUserID 从上下文中获取当前用户 ID
// 这是一个辅助函数，方便在 handler 中获取用户 ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}
