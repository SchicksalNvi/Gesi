package main

import (
	"fmt"
	"go-cesi/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 获取新密码
	newPassword := "123456"
	if len(os.Args) > 1 {
		newPassword = os.Args[1]
	}
	
	fmt.Printf("Resetting admin password to: %s\n", newPassword)

	// 连接到数据库
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	dbPath := filepath.Join(cwd, "data", "cesi.db")
	fmt.Printf("Connecting to database: %s\n", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 查询admin用户
	var adminUser models.User
	if err := db.Where("username = ? AND is_admin = ?", "admin", true).First(&adminUser).Error; err != nil {
		log.Fatalf("Failed to find admin user: %v", err)
	}

	fmt.Printf("Found admin user: %s\n", adminUser.Username)
	fmt.Printf("Old password hash: %s\n", adminUser.Password)

	// 生成新的密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 更新密码 - 只更新密码字段
	if err := db.Model(&adminUser).Where("id = ?", adminUser.ID).Update("password", string(hashedPassword)).Error; err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}
	
	// 重新加载用户以获取更新后的数据
	if err := db.Where("username = ? AND is_admin = ?", "admin", true).First(&adminUser).Error; err != nil {
		log.Fatalf("Failed to reload admin user: %v", err)
	}

	fmt.Printf("New password hash: %s\n", adminUser.Password)
	fmt.Println("✅ Password reset successfully!")
	
	// 验证新密码
	if adminUser.VerifyPassword(newPassword) {
		fmt.Println("✅ Password verification successful!")
	} else {
		fmt.Println("❌ Password verification failed!")
	}
}
