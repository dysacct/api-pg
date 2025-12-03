package util

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行 bcrypt 加密
// bcrypt 是一种慢哈希算法，专门为密码存储设计
// 优点：
//  1. 自动加盐（salt），每次加密同样的密码结果都不同
//  2. 计算慢，防止暴力破解
//  3. 有成本因子（cost），可以随硬件升级调整计算复杂度
func HashPassword(password string) (string, error) {
	// cost=10 表示计算复杂度，数字越大越安全但也越慢
	// 10 是推荐的平衡值
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CheckPassword 验证密码是否正确
// bcrypt.CompareHashAndPassword 会：
//  1. 从 hash 中提取 salt
//  2. 用相同的 salt 对 password 进行加密
//  3. 比较两个 hash 是否一致
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
