package util

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 的 Payload 部分
// 包含我们需要在 token 中携带的信息
type Claims struct {
	UserID               uint `json:"user_id"`
	jwt.RegisteredClaims      // 嵌入标准声明（如过期时间）
}

// GenerateToken 生成 JWT Token
//
// 流程：
//  1. 创建 Claims 结构（包含 user_id 和过期时间）
//  2. 使用 HS256 算法创建 token 对象
//  3. 用 jwtSecret 对 token 进行签名
//  4. 返回签名后的字符串
//
// 参数：
//
//	userID: 用户 ID
//	expirationHours: token 有效期（小时）
func GenerateToken(userID uint, expirationHours int) (string, error) {
	// 从环境变量获取密钥
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET 未配置")
	}

	// 创建 Claims
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 使用 HS256 算法创建 token
	// NewWithClaims 创建一个未签名的 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名并获得完整的编码后的字符串
	// 这一步会生成 Header.Payload.Signature
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析和验证 JWT Token
//
// 流程：
//  1. 用 jwtSecret 验证签名
//  2. 检查 token 是否过期
//  3. 提取 Claims 中的信息
//
// 安全性：
//   - 如果 token 被篡改，签名验证会失败
//   - 如果 token 过期，jwt 库会自动返回错误
//   - 只有签名正确且未过期的 token 才能通过验证
func ParseToken(tokenString string) (*Claims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET 未配置")
	}

	// 解析 token
	// jwt.ParseWithClaims 会：
	//   1. 解码 Header 和 Payload
	//   2. 用提供的密钥重新计算签名
	//   3. 比较签名是否一致
	//   4. 检查过期时间等标准声明
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名算法")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// 提取 Claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的 token")
}
