# JWT 用户认证系统使用指南

## 📚 目录

1. [系统架构](#系统架构)
2. [底层原理](#底层原理)
3. [模块关联](#模块关联)
4. [API 使用](#api-使用)
5. [Redis 集成](#redis-集成)
6. [安全性说明](#安全性说明)

---

## 🏗️ 系统架构

### 整体流程图

```
┌─────────────┐
│   客户端    │
└──────┬──────┘
       │
       │ 1. POST /api/auth/login
       │    {username, password}
       ↓
┌─────────────────────────────────────┐
│          Gin 路由层                  │
│   routes/auth_routes.go             │
└──────────────┬──────────────────────┘
               │
               │ 2. 路由到 Handler
               ↓
┌─────────────────────────────────────┐
│        Handler 层                    │
│   handlers/auth_handlers.go         │
│   ├─ 验证参数                       │
│   ├─ 查询数据库                     │
│   └─ 调用工具函数                   │
└──────┬──────────────┬───────────────┘
       │              │
       │              │ 3. 验证密码
       │              ↓
       │     ┌────────────────────┐
       │     │   密码工具层        │
       │     │ util/password_util.go│
       │     │  bcrypt 加密/验证   │
       │     └────────────────────┘
       │
       │ 4. 生成 Token
       ↓
┌─────────────────────────────────────┐
│          JWT 工具层                  │
│     util/jwt_util.go                │
│  ├─ 创建 Claims                     │
│  ├─ 使用 HS256 签名                 │
│  └─ 返回 Token 字符串               │
└─────────────────────────────────────┘
       │
       │ 5. 返回 Token
       ↓
┌──────────────┐
│   客户端     │ 
│ 存储 Token   │
└──────┬───────┘
       │
       │ 6. 后续请求
       │    Header: Authorization Bearer xxx
       ↓
┌─────────────────────────────────────┐
│        认证中间件                    │
│  middleware/auth_middleware.go      │
│  ├─ 提取 Token                      │
│  ├─ 验证签名                        │
│  ├─ 检查过期                        │
│  └─ 提取 user_id 到上下文          │
└──────────────┬──────────────────────┘
               │
               │ 7. 通过验证，继续处理
               ↓
         ┌────────────┐
         │  业务逻辑   │
         └────────────┘
```

---

## 🔐 底层原理

### 1. JWT 结构详解

JWT (JSON Web Token) 由三部分组成：

#### Header（头部）
```json
{
  "alg": "HS256",    // 签名算法
  "typ": "JWT"       // Token 类型
}
```
经过 Base64Url 编码后得到第一部分

#### Payload（负载）
```json
{
  "user_id": 123,                    // 自定义：用户 ID
  "exp": 1704067200,                 // 标准：过期时间
  "iat": 1704063600,                 // 标准：签发时间
  "nbf": 1704063600                  // 标准：生效时间
}
```
经过 Base64Url 编码后得到第二部分

#### Signature（签名）
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret
)
```

**完整的 Token：**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsImV4cCI6MTcwNDA2NzIwMH0.signature_here
└────────── Header ──────────┘ └────────── Payload ──────────┘ └── Signature ──┘
```

### 2. 安全机制

#### 为什么 JWT 是安全的？

1. **签名验证**
   - Header 和 Payload 是 Base64 编码（任何人都能解码）
   - 但 Signature 需要密钥才能生成
   - 篡改 Payload 会导致签名不匹配

2. **实际攻击场景分析**

   **场景 1：攻击者修改 user_id**
   ```
   原始 Token:
   Header.Payload(user_id=123).Signature(valid)
   
   攻击者尝试：
   Header.Payload(user_id=999).Signature(valid)
   
   结果：
   ❌ 验证失败！因为 Signature 是用 user_id=123 计算的
   ```

   **场景 2：攻击者重新计算签名**
   ```
   攻击者没有 jwtSecret，无法计算正确的签名
   只有服务端知道 jwtSecret
   ```

3. **密码加密（bcrypt）**
   - 慢哈希算法，防暴力破解
   - 自动加盐，相同密码每次加密结果不同
   - 不可逆，无法从 hash 反推密码

### 3. 认证流程

#### 登录流程（生成 Token）

```go
// 1. 用户提交 username 和 password
POST /api/auth/login
{
  "username": "test",
  "password": "123456"
}

// 2. 服务端验证
func Login(c *gin.Context) {
  // a. 查询数据库，找到用户
  var user models.User
  db.Where("username = ?", username).First(&user)
  
  // b. 验证密码（bcrypt.CompareHashAndPassword）
  //    从 user.Password (hash) 中提取 salt
  //    用相同 salt 对输入密码加密
  //    比较两个 hash
  if !util.CheckPassword(password, user.Password) {
    return error
  }
  
  // c. 生成 JWT Token
  //    创建 Claims{user_id, exp}
  //    用 HS256 + jwtSecret 签名
  token := util.GenerateToken(user.ID, 24)
  
  // d. 返回 Token
  return {token: "xxx"}
}
```

#### 验证流程（验证 Token）

```go
// 1. 客户端携带 Token 发送请求
GET /api/some-protected-route
Header: Authorization Bearer eyJhbGc...

// 2. 中间件拦截
func AuthMiddleware() {
  // a. 提取 Token
  token := extractTokenFromHeader(c)
  
  // b. 解析 Token
  claims, err := util.ParseToken(token)
  
  // ParseToken 内部：
  //   ① 分离 Header、Payload、Signature
  //   ② 用 jwtSecret 重新计算签名
  //   ③ 比较签名是否一致
  //   ④ 检查 exp 是否过期
  
  // c. 验证通过，提取 user_id
  c.Set("user_id", claims.UserID)
  
  // d. 继续处理请求
  c.Next()
}

// 3. Handler 获取当前用户
func SomeHandler(c *gin.Context) {
  userID := c.Get("user_id")  // 从中间件设置的上下文获取
  // 业务逻辑...
}
```

---

## 🔗 模块关联

### 1. 数据流向

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP 请求                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│  routes/auth_routes.go - 路由注册                       │
│  ├─ 公开路由: /register, /login                        │
│  └─ 受保护路由: /logout, /me (需要中间件)              │
└────────────────────┬────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
  公开路由                   受保护路由
         │                       │
         ↓                       ↓
┌──────────────────┐   ┌──────────────────────┐
│  handlers/       │   │  middleware/         │
│  auth_handlers   │   │  auth_middleware     │
│  └─ Login()      │   │  └─ 验证 Token       │
│  └─ Register()   │   └──────────┬───────────┘
└────┬─────────────┘              │
     │                            │
     │ 调用工具                    │ 通过验证
     ↓                            ↓
┌─────────────────────────────────────────────┐
│  util/jwt_util.go                           │
│  ├─ GenerateToken() - 生成 Token           │
│  └─ ParseToken() - 验证 Token              │
└─────────────────────────────────────────────┘
     │
     │ 调用工具
     ↓
┌─────────────────────────────────────────────┐
│  util/password_util.go                      │
│  ├─ HashPassword() - 加密密码              │
│  └─ CheckPassword() - 验证密码             │
└─────────────────────────────────────────────┘
     │
     │ 数据库操作
     ↓
┌─────────────────────────────────────────────┐
│  models/user.go + config/db.go              │
│  └─ User 模型和数据库连接                   │
└─────────────────────────────────────────────┘
```

### 2. 依赖关系

```
main.go
  └─ routes/routes.go
       └─ routes/auth_routes.go
            ├─ handlers/auth_handlers.go
            │    ├─ util/jwt_util.go
            │    ├─ util/password_util.go
            │    ├─ models/user.go
            │    └─ config/db.go
            │
            └─ middleware/auth_middleware.go
                 └─ util/jwt_util.go
```

### 3. 关键模块职责

| 模块 | 职责 | 输入 | 输出 |
|------|------|------|------|
| `models/user.go` | 定义用户数据结构 | - | User 结构体 |
| `util/password_util.go` | 密码加密/验证 | 明文密码 | Hash / 验证结果 |
| `util/jwt_util.go` | JWT 生成/解析 | user_id | Token / Claims |
| `middleware/auth_middleware.go` | 请求拦截验证 | HTTP 请求 | 验证通过/拒绝 |
| `handlers/auth_handlers.go` | 业务逻辑处理 | HTTP 请求 | HTTP 响应 |
| `routes/auth_routes.go` | 路由注册 | - | 路由配置 |

---

## 🚀 API 使用

### 环境配置

1. 创建 `.env` 文件：
```bash
PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=yourdb

# 生成随机密钥：openssl rand -hex 32
JWT_SECRET=your_super_secret_key_min_32_chars
```

2. 启动服务：
```bash
go run main.go
```

### API 端点

#### 1. 用户注册

**请求：**
```bash
POST /api/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123",
  "email": "test@example.com",
  "nickname": "Test User"
}
```

**响应：**
```json
{
  "message": "注册成功",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "Test User",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 2. 用户登录

**请求：**
```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

**响应：**
```json
{
  "message": "登录成功",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

**Cookie：**
服务器会自动设置 `token` Cookie（HttpOnly）

#### 3. 获取当前用户信息（需要认证）

**请求方式 1：使用 Header**
```bash
GET /api/auth/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求方式 2：使用 Cookie**
```bash
GET /api/auth/me
Cookie: token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应：**
```json
{
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "Test User"
  }
}
```

#### 4. 登出

**请求：**
```bash
POST /api/auth/logout
Authorization: Bearer <token>
```

**响应：**
```json
{
  "message": "登出成功"
}
```

### 前端使用示例

#### JavaScript/Fetch

```javascript
// 登录
async function login() {
  const response = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      username: 'testuser',
      password: 'password123'
    })
  });
  
  const data = await response.json();
  
  // 方式 1：存储到 localStorage
  localStorage.setItem('token', data.token);
  
  // 方式 2：使用 Cookie（服务器已自动设置）
}

// 调用受保护的 API
async function getProfile() {
  const token = localStorage.getItem('token');
  
  const response = await fetch('http://localhost:8080/api/auth/me', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const data = await response.json();
  console.log(data.user);
}
```

#### Axios

```javascript
import axios from 'axios';

// 设置拦截器，自动添加 Token
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 登录
const login = async () => {
  const { data } = await axios.post('/api/auth/login', {
    username: 'testuser',
    password: 'password123'
  });
  localStorage.setItem('token', data.token);
};

// 获取用户信息
const getProfile = async () => {
  const { data } = await axios.get('/api/auth/me');
  return data.user;
};
```

---

## 🔴 Redis 集成（高级功能）

### 为什么需要 Redis？

1. **实现"踢人"功能**：从服务端主动让某个用户的 Token 失效
2. **Token 黑名单**：记录已登出但未过期的 Token
3. **限制单设备登录**：一个用户只能在一个地方登录
4. **提高性能**：减少数据库查询

### Redis 集成实现

#### 1. 安装 Redis 客户端

```bash
go get github.com/redis/go-redis/v9
```

#### 2. 创建 Redis 配置

创建 `config/redis.go`：

```go
package config

import (
  "context"
  "fmt"
  "os"

  "github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
  redisHost := os.Getenv("REDIS_HOST")
  redisPort := os.Getenv("REDIS_PORT")
  redisPassword := os.Getenv("REDIS_PASSWORD")

  RedisClient = redis.NewClient(&redis.Options{
    Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
    Password: redisPassword,
    DB:       0,
  })

  _, err := RedisClient.Ping(Ctx).Result()
  if err != nil {
    panic("Redis 连接失败: " + err.Error())
  }

  fmt.Println("✅ Redis 连接成功")
}
```

#### 3. 更新 JWT 工具

修改 `util/jwt_util.go`，添加 Redis 集成：

```go
package util

import (
  "api-postgre/config"
  "errors"
  "fmt"
  "os"
  "time"

  "github.com/golang-jwt/jwt/v5"
)

// GenerateTokenWithRedis 生成 Token 并存入 Redis
func GenerateTokenWithRedis(userID uint, expirationHours int) (string, error) {
  jwtSecret := os.Getenv("JWT_SECRET")
  if jwtSecret == "" {
    return "", errors.New("JWT_SECRET 未配置")
  }

  claims := Claims{
    UserID: userID,
    RegisteredClaims: jwt.RegisteredClaims{
      ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
      IssuedAt:  jwt.NewNumericDate(time.Now()),
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenString, err := token.SignedString([]byte(jwtSecret))
  if err != nil {
    return "", err
  }

  // 存入 Redis
  // Key: user:token:{user_id}
  // Value: token
  // TTL: 与 Token 过期时间一致
  redisKey := fmt.Sprintf("user:token:%d", userID)
  err = config.RedisClient.Set(
    config.Ctx,
    redisKey,
    tokenString,
    time.Hour*time.Duration(expirationHours),
  ).Err()

  if err != nil {
    return "", errors.New("Token 存入 Redis 失败")
  }

  return tokenString, nil
}

// ValidateTokenWithRedis 验证 Token（检查 Redis）
func ValidateTokenWithRedis(tokenString string) (*Claims, error) {
  // 1. 先解析 Token 获取 user_id
  claims, err := ParseToken(tokenString)
  if err != nil {
    return nil, err
  }

  // 2. 从 Redis 检查 Token 是否有效
  redisKey := fmt.Sprintf("user:token:%d", claims.UserID)
  storedToken, err := config.RedisClient.Get(config.Ctx, redisKey).Result()

  if err != nil {
    return nil, errors.New("Token 已失效或用户已被踢出")
  }

  // 3. 比对 Token 是否一致
  if storedToken != tokenString {
    return nil, errors.New("Token 不匹配，可能是旧 Token")
  }

  return claims, nil
}

// RevokeToken 撤销 Token（踢人）
func RevokeToken(userID uint) error {
  redisKey := fmt.Sprintf("user:token:%d", userID)
  err := config.RedisClient.Del(config.Ctx, redisKey).Err()
  if err != nil {
    return errors.New("撤销 Token 失败")
  }
  return nil
}
```

#### 4. 更新中间件

修改 `middleware/auth_middleware.go`：

```go
func AuthMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    token, err := c.Cookie("token")
    
    if err != nil || token == "" {
      authHeader := c.GetHeader("Authorization")
      if authHeader == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证 token"})
        c.Abort()
        return
      }
      parts := strings.SplitN(authHeader, " ", 2)
      if len(parts) != 2 || parts[0] != "Bearer" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "格式错误"})
        c.Abort()
        return
      }
      token = parts[1]
    }

    // 使用 Redis 验证（如果启用了 Redis）
    claims, err := util.ValidateTokenWithRedis(token)
    if err != nil {
      c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    c.Set("user_id", claims.UserID)
    c.Next()
  }
}
```

#### 5. 添加踢人接口

在 `handlers/auth_handlers.go` 添加：

```go
// KickUser 踢出用户（管理员功能）
func KickUser(c *gin.Context) {
  var req struct {
    UserID uint `json:"user_id" binding:"required"`
  }

  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
    return
  }

  // 从 Redis 删除该用户的 Token
  err := util.RevokeToken(req.UserID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "message": fmt.Sprintf("用户 %d 已被踢出", req.UserID),
  })
}
```

### Redis 使用场景

| 场景 | 实现方式 | Redis Key | 说明 |
|------|---------|-----------|------|
| 单点登录 | 每次登录覆盖旧 Token | `user:token:{user_id}` | 新登录会使旧 Token 失效 |
| 踢人 | 删除 Redis 中的 Token | `user:token:{user_id}` | 用户下次请求会被拒绝 |
| Token 黑名单 | 登出时加入黑名单 | `token:blacklist:{token}` | 记录已登出的 Token |
| 限流 | 记录请求次数 | `rate:limit:{user_id}` | 防止接口滥用 |

---

## 🛡️ 安全性说明

### 1. JWT Token 安全

**✅ 安全的做法：**
- Token 只包含必要信息（user_id、exp）
- 不在 Token 中存储敏感信息（密码、信用卡等）
- 设置合理的过期时间（24 小时）
- 使用强密钥（至少 32 字符）

**❌ 不安全的做法：**
- Token 包含用户密码
- Token 永不过期
- 密钥过短或使用简单字符串
- Token 存储在 URL 参数中

### 2. 密码存储安全

**✅ 使用 bcrypt：**
```go
// 加密时自动加盐
hash, _ := bcrypt.GenerateFromPassword([]byte("password"), 10)
// 每次结果都不同：
// $2a$10$N9qo8uLOickgx2ZMRZoMy.abcd...
// $2a$10$N9qo8uLOickgx2ZMRZoMy.efgh...
```

**❌ 不安全的做法：**
```go
// 直接存储明文
user.Password = "123456"

// 简单 MD5/SHA1（可以被彩虹表攻击）
hash := md5.Sum([]byte("123456"))
```

### 3. Token 传输安全

**✅ 推荐方式：**
1. **HTTP Header（Bearer Token）**
   ```
   Authorization: Bearer eyJhbGc...
   ```
   - 灵活性高
   - 支持跨域（CORS）
   - 不受 Cookie 限制

2. **HttpOnly Cookie**
   ```go
   c.SetCookie("token", token, 86400, "/", "", false, true)
   //                                              ↑     ↑
   //                                         secure  httpOnly
   ```
   - 防止 XSS 攻击（JS 无法读取）
   - 自动随请求发送

**❌ 不推荐：**
- 存储在 URL 参数中（会被日志记录）
- 存储在 localStorage 但未加密

### 4. HTTPS 的重要性

⚠️ **生产环境必须使用 HTTPS！**

```
HTTP (不安全)：
  客户端 ──→ Token (明文) ──→ 服务器
          ↑ 攻击者可以截获

HTTPS (安全)：
  客户端 ──→ Token (加密) ──→ 服务器
          ↑ 攻击者无法解密
```

### 5. 防御常见攻击

| 攻击类型 | 防御措施 | 实现位置 |
|---------|---------|---------|
| XSS | HttpOnly Cookie | auth_handlers.go |
| CSRF | SameSite Cookie | 需要配置 |
| SQL 注入 | GORM 参数化查询 | GORM 自动处理 |
| 暴力破解 | bcrypt 慢哈希 | password_util.go |
| Token 劫持 | HTTPS + 短期过期 | 生产环境配置 |

---

## 📝 总结

### 核心要点

1. **JWT 不加密，只签名** - 任何人都能看到内容，但无法篡改
2. **密钥至关重要** - `JWT_SECRET` 必须保密且足够复杂
3. **合理设置过期时间** - 平衡安全性和用户体验
4. **生产环境使用 HTTPS** - 防止 Token 被截获
5. **不存储敏感信息** - Token 只存 user_id 等必要信息

### 扩展阅读

- [JWT 官方网站](https://jwt.io/)
- [bcrypt 算法原理](https://en.wikipedia.org/wiki/Bcrypt)
- [OWASP 认证最佳实践](https://owasp.org/www-project-top-ten/)

---

## 🎯 快速测试

```bash
# 1. 注册用户
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","email":"test@example.com"}'

# 2. 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 3. 使用 Token 访问受保护接口
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

祝使用愉快！🎉

