package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go-cesi/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 获取当前工作目录
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// 确保数据库目录存在
	dataDir := filepath.Join(projectRoot, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 连接数据库
	dbPath := filepath.Join(dataDir, "cesi.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("开始清理模拟数据...")

	// 清理测试用户数据（保留admin用户）
	result := db.Where("username != ?", "admin").Delete(&models.User{})
	if result.Error != nil {
		log.Printf("清理测试用户失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试用户\n", result.RowsAffected)
	}

	// 清理测试日志数据（保留最近7天的真实日志）
	result = db.Where("created_at < datetime('now', '-7 days')").Delete(&models.LogEntry{})
	if result.Error != nil {
		log.Printf("清理旧日志失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 条旧日志\n", result.RowsAffected)
	}

	// 清理测试配置数据（保留真实的数据库配置等）
	result = db.Where("key LIKE ? OR key LIKE ?", "test_%", "mock_%").Delete(&models.Configuration{})
	if result.Error != nil {
		log.Printf("清理测试配置失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试配置\n", result.RowsAffected)
	}

	// 清理测试节点数据（保留真实的supervisord节点）
	result = db.Where("name LIKE ? OR name LIKE ?", "test-%", "mock-%").Delete(&models.Node{})
	if result.Error != nil {
		log.Printf("清理测试节点失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试节点\n", result.RowsAffected)
	}

	// 清理测试进程组数据
	result = db.Where("name LIKE ? OR name LIKE ?", "test-%", "mock-%").Delete(&models.ProcessGroup{})
	if result.Error != nil {
		log.Printf("清理测试进程组失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试进程组\n", result.RowsAffected)
	}

	// 清理测试计划任务
	result = db.Where("name LIKE ? OR name LIKE ?", "test-%", "mock-%").Delete(&models.ScheduledTask{})
	if result.Error != nil {
		log.Printf("清理测试计划任务失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试计划任务\n", result.RowsAffected)
	}

	// 清理测试备份记录
	result = db.Where("name LIKE ? OR name LIKE ?", "test-%", "mock-%").Delete(&models.BackupRecord{})
	if result.Error != nil {
		log.Printf("清理测试备份记录失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试备份记录\n", result.RowsAffected)
	}

	// 清理测试导出记录
	result = db.Where("name LIKE ? OR name LIKE ?", "test-%", "mock-%").Delete(&models.DataExportRecord{})
	if result.Error != nil {
		log.Printf("清理测试导出记录失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试导出记录\n", result.RowsAffected)
	}

	// 清理测试导入记录
	result = db.Where("name LIKE ? OR name LIKE ?", "test-%", "mock-%").Delete(&models.DataImportRecord{})
	if result.Error != nil {
		log.Printf("清理测试导入记录失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 个测试导入记录\n", result.RowsAffected)
	}

	fmt.Println("模拟数据清理完成！")
	fmt.Println("注意：已保留admin用户和真实的连接配置数据")
}
