package handlers

import (
	"api-postgre/config"
	"api-postgre/models"
	"api-postgre/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 用户注册
//
// 流程：
//  1. 验证请求参数
//  2. 检查用户名是否已存在
//  3. 对密码进行 bcrypt 加密
//  4. 创建用户记录
//  5. 返回用户信息（不含密码）
func Register(c *gin.Context) {
	var req RegisterRequest

	// 1. 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数验证失败: " + err.Error(),
		})
		return
	}

	// 2. 检查用户名是否已存在
	var existingUser models.User
	if err := config.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "用户名已存在",
		})
		return
	}

	// 3. 对密码进行加密
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "密码加密失败",
		})
		return
	}

	// 4. 创建用户
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Nickname: req.Nickname,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建用户失败: " + err.Error(),
		})
		return
	}

	// 5. 返回用户信息（密码字段已通过 json:"-" 标签隐藏）
	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"user":    user,
	})
}

// Login 用户登录
//
// 流程：
//  1. 验证请求参数
//  2. 查询用户是否存在
//  3. 验证密码是否正确
//  4. 生成 JWT token
//  5. 返回 token（同时设置到 Cookie）
//
// Token 有效期：24 小时
func Login(c *gin.Context) {
	var req LoginRequest

	// 1. 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数验证失败: " + err.Error(),
		})
		return
	}

	// 2. 查询用户
	var user models.User
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误",
		})
		return
	}

	// 3. 验证密码
	if !util.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误",
		})
		return
	}

	// 4. 生成 JWT token（有效期 24 小时）
	token, err := util.GenerateToken(user.ID, 24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成 token 失败",
		})
		return
	}

	// 5. 设置 Cookie（可选）
	// 参数说明：
	//   - name: cookie 名称
	//   - value: cookie 值
	//   - maxAge: 过期时间（秒），24小时 = 86400秒
	//   - path: cookie 作用路径
	//   - domain: cookie 作用域名
	//   - secure: 是否只在 HTTPS 下传输
	//   - httpOnly: 是否禁止 JavaScript 访问（防止 XSS 攻击）
	c.SetCookie("token", token, 86400, "/", "", false, true)

	// 6. 返回 token
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user":    user,
	})
}

// Logout 用户登出
//
// 流程：
//  1. 清除 Cookie 中的 token
//  2. 返回登出成功消息
//
// 注意：如果使用 Redis 缓存，这里需要额外删除 Redis 中的 token
func Logout(c *gin.Context) {
	// 清除 Cookie（将 maxAge 设为 -1 表示立即删除）
	c.SetCookie("token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// GetCurrentUser 获取当前登录用户信息
//
// 这个 handler 需要在使用 AuthMiddleware 的路由组中
// 可以通过它测试认证是否工作正常
func GetCurrentUser(c *gin.Context) {
	// 从上下文中获取用户 ID（由 AuthMiddleware 设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证",
		})
		return
	}

	// 查询用户信息
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
