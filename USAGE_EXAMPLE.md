# 实际使用示例

## 如何保护你的业务接口

### 示例：保护 Contact 接口

假设你有一个 Contact 相关的业务接口，只允许登录用户访问。

#### 原始代码（routes/contact_routes.go）

```go
func ContactRoutes(router *gin.RouterGroup) {
  // 所有人都可以访问
  router.GET("/contacts", handlers.GetAllContacts)
  router.POST("/contacts", handlers.CreateContact)
  router.PUT("/contacts/:id", handlers.UpdateContact)
  router.DELETE("/contacts/:id", handlers.DeleteContact)
}
```

#### 修改后（添加认证保护）

```go
package routes

import (
  "api-postgre/handlers"
  "api-postgre/middleware"
  
  "github.com/gin-gonic/gin"
)

func ContactRoutes(router *gin.RouterGroup) {
  // 方式 1：给所有 Contact 接口添加认证
  contacts := router.Group("/contacts")
  contacts.Use(middleware.AuthMiddleware())  // 应用认证中间件
  {
    contacts.GET("", handlers.GetAllContacts)
    contacts.POST("", handlers.CreateContact)
    contacts.PUT("/:id", handlers.UpdateContact)
    contacts.DELETE("/:id", handlers.DeleteContact)
  }
  
  // 方式 2：部分接口需要认证
  router.GET("/contacts", handlers.GetAllContacts)  // 公开：任何人可查看
  
  protected := router.Group("/contacts")
  protected.Use(middleware.AuthMiddleware())  // 需要登录
  {
    protected.POST("", handlers.CreateContact)      // 需要登录才能创建
    protected.PUT("/:id", handlers.UpdateContact)   // 需要登录才能更新
    protected.DELETE("/:id", handlers.DeleteContact) // 需要登录才能删除
  }
}
```

### 在 Handler 中获取当前用户

#### 修改 Contact Handler（handlers/contact_handlers.go）

```go
package handlers

import (
  "api-postgre/config"
  "api-postgre/middleware"
  "api-postgre/models"
  "net/http"

  "github.com/gin-gonic/gin"
)

// CreateContact 创建联系人（需要登录）
func CreateContact(c *gin.Context) {
  // 1. 获取当前登录用户 ID
  userID, exists := middleware.GetCurrentUserID(c)
  if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
    return
  }

  // 2. 绑定请求数据
  var contact models.Contact
  if err := c.ShouldBindJSON(&contact); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  // 3. 关联到当前用户（如果 Contact 模型有 UserID 字段）
  // contact.UserID = userID

  // 4. 保存到数据库
  if err := config.DB.Create(&contact).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
    return
  }

  // 5. 返回结果
  c.JSON(http.StatusCreated, gin.H{
    "message": "创建成功",
    "contact": contact,
    "created_by_user_id": userID,  // 可以返回是哪个用户创建的
  })
}

// GetAllContacts 获取当前用户的所有联系人
func GetAllContacts(c *gin.Context) {
  // 获取当前用户 ID
  userID, exists := middleware.GetCurrentUserID(c)
  if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
    return
  }

  // 只查询当前用户的联系人
  var contacts []models.Contact
  if err := config.DB.Where("user_id = ?", userID).Find(&contacts).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "contacts": contacts,
  })
}

// UpdateContact 更新联系人（需要验证所有权）
func UpdateContact(c *gin.Context) {
  // 1. 获取当前用户 ID
  userID, exists := middleware.GetCurrentUserID(c)
  if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
    return
  }

  // 2. 获取联系人 ID
  contactID := c.Param("id")

  // 3. 查询联系人
  var contact models.Contact
  if err := config.DB.First(&contact, contactID).Error; err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "联系人不存在"})
    return
  }

  // 4. 验证所有权（确保是当前用户的联系人）
  if contact.UserID != userID {
    c.JSON(http.StatusForbidden, gin.H{"error": "无权修改此联系人"})
    return
  }

  // 5. 更新数据
  if err := c.ShouldBindJSON(&contact); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if err := config.DB.Save(&contact).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "message": "更新成功",
    "contact": contact,
  })
}
```

### 更新 Contact 模型（添加用户关联）

如果你想让每个 Contact 属于某个用户，需要修改模型：

```go
// models/contact.go
package models

import "gorm.io/gorm"

type Contact struct {
  gorm.Model
  
  UserID     uint   `json:"user_id" gorm:"not null;index"`  // 添加用户 ID
  FirstName  string `json:"first_name" gorm:"not null"`
  SecondName string `json:"second_name" gorm:"not null"`
  Email      string `json:"email" gorm:"not null"`
  Phone      string `json:"phone" gorm:"not null"`
  
  // 外键关联
  User User `json:"user" gorm:"foreignKey:UserID"`
}
```

### 测试流程

```bash
# 1. 注册用户
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "123456",
    "email": "john@example.com"
  }'

# 2. 登录获取 Token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "123456"
  }'

# 响应：
# {
#   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "user": {...}
# }

# 3. 使用 Token 创建联系人
curl -X POST http://localhost:8080/api/contacts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "first_name": "Alice",
    "second_name": "Smith",
    "email": "alice@example.com",
    "phone": "1234567890"
  }'

# 4. 获取当前用户的所有联系人
curl -X GET http://localhost:8080/api/contacts \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 5. 不带 Token 访问（会失败）
curl -X POST http://localhost:8080/api/contacts \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Bob",
    "second_name": "Jones",
    "email": "bob@example.com",
    "phone": "0987654321"
  }'

# 响应：
# {
#   "error": "未提供认证 token"
# }
```

## 权限控制示例

### 实现管理员和普通用户权限

#### 1. 更新 User 模型

```go
// models/user.go
type User struct {
  ID        uint           `json:"id" gorm:"primaryKey"`
  CreatedAt time.Time      `json:"created_at"`
  UpdatedAt time.Time      `json:"updated_at"`
  DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
  
  Username string `json:"username" gorm:"uniqueIndex;not null"`
  Password string `json:"-" gorm:"not null"`
  Email    string `json:"email" gorm:"uniqueIndex"`
  Nickname string `json:"nickname"`
  
  // 添加角色字段
  Role string `json:"role" gorm:"default:'user'"`  // user, admin
}
```

#### 2. 创建管理员中间件

```go
// middleware/admin_middleware.go
package middleware

import (
  "api-postgre/config"
  "api-postgre/models"
  "net/http"

  "github.com/gin-gonic/gin"
)

// AdminMiddleware 管理员权限中间件
// 必须在 AuthMiddleware 之后使用
func AdminMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    // 从上下文获取用户 ID
    userID, exists := GetCurrentUserID(c)
    if !exists {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
      c.Abort()
      return
    }

    // 查询用户角色
    var user models.User
    if err := config.DB.First(&user, userID).Error; err != nil {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
      c.Abort()
      return
    }

    // 检查是否为管理员
    if user.Role != "admin" {
      c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
      c.Abort()
      return
    }

    c.Next()
  }
}
```

#### 3. 使用权限中间件

```go
// routes/auth_routes.go
func SetupAuthRoutes(router *gin.Engine) {
  public := router.Group("/api/auth")
  {
    public.POST("/register", handlers.Register)
    public.POST("/login", handlers.Login)
  }

  // 普通用户接口
  user := router.Group("/api/auth")
  user.Use(middleware.AuthMiddleware())
  {
    user.POST("/logout", handlers.Logout)
    user.GET("/me", handlers.GetCurrentUser)
  }

  // 管理员接口
  admin := router.Group("/api/admin")
  admin.Use(middleware.AuthMiddleware())     // 先验证登录
  admin.Use(middleware.AdminMiddleware())    // 再验证管理员权限
  {
    admin.POST("/kick-user", handlers.KickUser)        // 踢出用户
    admin.GET("/users", handlers.GetAllUsers)          // 获取所有用户
    admin.DELETE("/users/:id", handlers.DeleteUser)    // 删除用户
  }
}
```

## 完整的错误处理

### 统一错误响应格式

```go
// middleware/error_middleware.go
package middleware

import (
  "net/http"

  "github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
  Code    int    `json:"code"`
  Message string `json:"message"`
  Details string `json:"details,omitempty"`
}

// ErrorHandler 全局错误处理中间件
func ErrorHandler() gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Next()

    // 检查是否有错误
    if len(c.Errors) > 0 {
      err := c.Errors.Last()
      
      // 根据错误类型返回不同的状态码
      c.JSON(http.StatusInternalServerError, ErrorResponse{
        Code:    http.StatusInternalServerError,
        Message: "服务器内部错误",
        Details: err.Error(),
      })
    }
  }
}
```

### 在 main.go 中应用

```go
func main() {
  godotenv.Load()
  config.ConnectDB()
  
  app := gin.Default()
  
  // 应用全局中间件
  app.Use(middleware.ErrorHandler())
  
  // CORS 中间件（如果需要跨域）
  app.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if c.Request.Method == "OPTIONS" {
      c.AbortWithStatus(204)
      return
    }
    
    c.Next()
  })
  
  routes.RegisterRoutes(app)
  app.Run(":8080")
}
```

## 总结

### 认证流程

```
1. 用户注册/登录 → 获得 Token
2. 客户端存储 Token（localStorage 或 Cookie）
3. 后续请求携带 Token
4. 服务端中间件验证 Token
5. 验证通过 → 提取 user_id → 执行业务逻辑
6. 验证失败 → 返回 401 错误
```

### 权限层级

```
无需认证 → 任何人都可以访问
    ↓
需要登录 → 使用 AuthMiddleware
    ↓
需要管理员 → 使用 AuthMiddleware + AdminMiddleware
    ↓
自定义权限 → 在 Handler 中检查具体权限
```

### 最佳实践

1. **敏感操作必须验证所有权**
   - 修改/删除资源前，检查是否属于当前用户

2. **Token 有效期设置合理**
   - 一般应用：24 小时
   - 高安全应用：1-2 小时 + 刷新 Token 机制

3. **使用 HTTPS**
   - 生产环境必须使用 HTTPS 传输 Token

4. **日志记录**
   - 记录登录、登出、权限错误等关键操作

5. **限流保护**
   - 防止暴力破解：登录接口添加限流

祝你使用愉快！🎉

