package config

import (
	"api-postgre/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局变量,gorm.DB是gorm库中的一个结构体,用于表示数据库连接。

// ConnectDB 连接数据库并执行自动迁移
func ConnectDB() {
	var err error
	fmt.Println("📦 正在连接数据库...")

	// 从环境变量读取数据库配置
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// 构建数据库连接字符串
	dbURL := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	// 连接数据库
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 自动迁移模型
	// AutoMigrate 会根据结构体定义自动创建/更新表结构
	// 它不会删除未使用的列，是安全的操作
	fmt.Println("🔄 执行数据库迁移...")
	err = DB.AutoMigrate(
		&models.User{},
		&models.Contact{},
	)
	if err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}

	fmt.Println("✅ 数据库迁移完成")
}
