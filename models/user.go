package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
// 存储用户的基本信息和认证凭据
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Username string `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Password string `json:"-" gorm:"not null"` // 存储加密后的密码，json 不返回
	Email    string `json:"email" gorm:"uniqueIndex;size:100"`
	Nickname string `json:"nickname" gorm:"size:50"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
