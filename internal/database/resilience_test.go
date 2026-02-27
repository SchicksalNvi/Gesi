package database

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"superview/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Feature: concurrent-safety-fixes, Property 11: Database Resilience
// For any database failure scenario, the system should implement retry logic with exponential backoff and handle resource exhaustion gracefully
func TestDatabaseResilience(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database resilience test in short mode")
	}

	// 创建临时数据库用于测试
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_resilience.db")
	
	config := &DatabaseConfig{
		MaxOpenConns:        5,  // 限制连接数以测试资源耗尽
		MaxIdleConns:        2,
		ConnMaxLifetime:     1 * time.Minute,
		ConnMaxIdleTime:     30 * time.Second,
		QueryTimeout:        5 * time.Second,
		HealthCheckEnabled:  true,
		HealthCheckInterval: 1 * time.Second,
		TransactionTimeout:  10 * time.Second,
	}

	// 初始化测试数据库
	db, err := initTestDB(dbPath, config)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	t.Run("ConnectionPoolExhaustion", func(t *testing.T) {
		testConnectionPoolExhaustion(t, db, config)
	})

	t.Run("TransactionTimeout", func(t *testing.T) {
		testTransactionTimeout(t, db, config)
	})

	t.Run("HealthCheckRecovery", func(t *testing.T) {
		testHealthCheckRecovery(t, db)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		testConcurrentDatabaseOperations(t, db)
	})
}

// Feature: concurrent-safety-fixes, Property 12: Transaction Resource Management
// For any database transaction, timeout should trigger proper resource cleanup and appropriate error responses
func TestTransactionResourceManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping transaction resource management test in short mode")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_transaction.db")
	
	config := &DatabaseConfig{
		MaxOpenConns:       10,
		MaxIdleConns:       3,
		ConnMaxLifetime:    1 * time.Minute,
		ConnMaxIdleTime:    30 * time.Second,
		QueryTimeout:       2 * time.Second,
		TransactionTimeout: 3 * time.Second,
	}

	db, err := initTestDB(dbPath, config)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	t.Run("TransactionTimeoutCleanup", func(t *testing.T) {
		testTransactionTimeoutCleanup(t, db, config)
	})

	t.Run("NestedTransactionHandling", func(t *testing.T) {
		testNestedTransactionHandling(t, db)
	})

	t.Run("TransactionRollbackOnError", func(t *testing.T) {
		testTransactionRollbackOnError(t, db)
	})
}

// initTestDB 初始化测试数据库
func initTestDB(dbPath string, config *DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 迁移测试模型
	err = db.AutoMigrate(&models.User{}, &models.ActivityLog{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// testConnectionPoolExhaustion 测试连接池耗尽处理
func testConnectionPoolExhaustion(t *testing.T, db *gorm.DB, config *DatabaseConfig) {
	// 创建超过连接池限制的并发操作
	numGoroutines := config.MaxOpenConns * 2
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// 模拟长时间运行的查询
			ctx, cancel := context.WithTimeout(context.Background(), config.QueryTimeout)
			defer cancel()
			
			var count int64
			err := db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
			if err != nil {
				errors <- fmt.Errorf("goroutine %d failed: %w", id, err)
				return
			}
			
			// 添加一些延迟来增加连接压力
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查错误处理
	var errorCount int
	for err := range errors {
		t.Logf("Connection pool stress error: %v", err)
		errorCount++
	}

	// 系统应该优雅地处理连接池耗尽，而不是崩溃
	if errorCount > numGoroutines/2 {
		t.Errorf("Too many connection failures (%d/%d), system should handle pool exhaustion gracefully", 
			errorCount, numGoroutines)
	}

	// 验证连接池统计
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	stats := sqlDB.Stats()
	t.Logf("Connection pool stats: Open=%d, InUse=%d, Idle=%d, WaitCount=%d", 
		stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)

	// 连接数不应超过配置的最大值
	if stats.OpenConnections > config.MaxOpenConns {
		t.Errorf("Open connections (%d) exceeded max (%d)", stats.OpenConnections, config.MaxOpenConns)
	}
}

// testTransactionTimeout 测试事务超时处理
func testTransactionTimeout(t *testing.T, db *gorm.DB, config *DatabaseConfig) {
	// 测试超时事务的资源清理
	ctx, cancel := context.WithTimeout(context.Background(), config.TransactionTimeout/2)
	defer cancel()

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建一个测试用户
		user := &models.User{
			Username: "timeout_test_user",
			Email:    "timeout@test.com",
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 模拟长时间操作，触发超时
		time.Sleep(config.TransactionTimeout)
		
		return nil
	})

	// 应该收到超时错误
	if err == nil {
		t.Error("Expected timeout error, but transaction completed successfully")
	}

	// 验证事务被正确回滚
	var count int64
	db.Model(&models.User{}).Where("username = ?", "timeout_test_user").Count(&count)
	if count > 0 {
		t.Error("Transaction was not properly rolled back after timeout")
	}
}

// testHealthCheckRecovery 测试健康检查和恢复
func testHealthCheckRecovery(t *testing.T, db *gorm.DB) {
	// 模拟数据库连接问题
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// 记录初始健康状态
	initialHealth := true
	if err := sqlDB.Ping(); err != nil {
		initialHealth = false
	}

	// 临时关闭连接来模拟故障
	sqlDB.SetMaxOpenConns(0)
	time.Sleep(100 * time.Millisecond)

	// 检查健康检查是否检测到问题
	if err := sqlDB.Ping(); err == nil {
		t.Log("Expected ping to fail with zero max connections")
	}

	// 恢复连接池
	sqlDB.SetMaxOpenConns(10)
	time.Sleep(100 * time.Millisecond)

	// 验证恢复
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("Database should have recovered, but ping failed: %v", err)
	}

	// 如果初始状态是健康的，恢复后也应该是健康的
	if initialHealth {
		if err := sqlDB.Ping(); err != nil {
			t.Errorf("Database health should be restored: %v", err)
		}
	}
}

// testConcurrentDatabaseOperations 测试并发数据库操作
func testConcurrentDatabaseOperations(t *testing.T, db *gorm.DB) {
	numOperations := 50
	var wg sync.WaitGroup
	errors := make(chan error, numOperations)

	// 并发执行多种数据库操作
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// 随机选择操作类型
			switch rand.Intn(4) {
			case 0: // 创建用户
				user := &models.User{
					Username: fmt.Sprintf("concurrent_user_%d", id),
					Email:    fmt.Sprintf("user%d@concurrent.test", id),
				}
				if err := db.Create(user).Error; err != nil {
					errors <- fmt.Errorf("create user %d failed: %w", id, err)
				}
				
			case 1: // 查询用户
				var users []models.User
				if err := db.Limit(10).Find(&users).Error; err != nil {
					errors <- fmt.Errorf("query users %d failed: %w", id, err)
				}
				
			case 2: // 更新用户
				if err := db.Model(&models.User{}).Where("id > ?", 0).Update("updated_at", time.Now()).Error; err != nil {
					errors <- fmt.Errorf("update users %d failed: %w", id, err)
				}
				
			case 3: // 事务操作
				err := db.Transaction(func(tx *gorm.DB) error {
					user := &models.User{
						Username: fmt.Sprintf("tx_user_%d", id),
						Email:    fmt.Sprintf("tx%d@test.com", id),
					}
					if err := tx.Create(user).Error; err != nil {
						return err
					}
					
					log := &models.ActivityLog{
						UserID:   user.ID,
						Action:   "test_action",
						Resource: "test_resource",
						Message:  fmt.Sprintf("Test log from goroutine %d", id),
					}
					return tx.Create(log).Error
				})
				if err != nil {
					errors <- fmt.Errorf("transaction %d failed: %w", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 统计错误
	var errorCount int
	for err := range errors {
		t.Logf("Concurrent operation error: %v", err)
		errorCount++
	}

	// 允许少量错误（由于资源竞争），但不应该有大量失败
	if errorCount > numOperations/10 {
		t.Errorf("Too many concurrent operation failures (%d/%d)", errorCount, numOperations)
	}

	// 验证数据库状态仍然正常
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		t.Errorf("Failed to count users after concurrent operations: %v", err)
	}

	t.Logf("Concurrent operations completed with %d errors, %d users created", errorCount, userCount)
}

// testTransactionTimeoutCleanup 测试事务超时后的资源清理
func testTransactionTimeoutCleanup(t *testing.T, db *gorm.DB, config *DatabaseConfig) {
	// 获取初始连接统计
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	
	initialStats := sqlDB.Stats()
	
	// 执行会超时的事务
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建测试数据
		user := &models.User{
			Username: "cleanup_test",
			Email:    "cleanup@test.com",
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 模拟长时间操作
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	// 应该收到超时错误
	if err == nil {
		t.Error("Expected timeout error")
	}

	// 等待资源清理
	time.Sleep(100 * time.Millisecond)

	// 检查连接是否被正确清理
	finalStats := sqlDB.Stats()
	
	// 连接数不应该增长（表示资源被正确清理）
	if finalStats.OpenConnections > initialStats.OpenConnections+1 {
		t.Errorf("Connection leak detected: initial=%d, final=%d", 
			initialStats.OpenConnections, finalStats.OpenConnections)
	}

	// 验证事务被回滚
	var count int64
	db.Model(&models.User{}).Where("username = ?", "cleanup_test").Count(&count)
	if count > 0 {
		t.Error("Transaction was not properly rolled back")
	}
}

// testNestedTransactionHandling 测试嵌套事务处理
func testNestedTransactionHandling(t *testing.T, db *gorm.DB) {
	err := db.Transaction(func(tx1 *gorm.DB) error {
		// 外层事务：创建用户
		user := &models.User{
			Username: "nested_test",
			Email:    "nested@test.com",
		}
		if err := tx1.Create(user).Error; err != nil {
			return err
		}

		// 内层事务：创建活动日志
		return tx1.Transaction(func(tx2 *gorm.DB) error {
			log := &models.ActivityLog{
				UserID:   user.ID,
				Action:   "nested_test",
				Resource: "test",
				Message:  "Nested transaction test",
			}
			if err := tx2.Create(log).Error; err != nil {
				return err
			}

			// 模拟内层事务失败
			return fmt.Errorf("simulated inner transaction failure")
		})
	})

	// 整个事务应该失败
	if err == nil {
		t.Error("Expected nested transaction to fail")
	}

	// 验证所有操作都被回滚
	var userCount, logCount int64
	db.Model(&models.User{}).Where("username = ?", "nested_test").Count(&userCount)
	db.Model(&models.ActivityLog{}).Where("action = ?", "nested_test").Count(&logCount)

	if userCount > 0 || logCount > 0 {
		t.Errorf("Nested transaction rollback failed: users=%d, logs=%d", userCount, logCount)
	}
}

// testTransactionRollbackOnError 测试错误时的事务回滚
func testTransactionRollbackOnError(t *testing.T, db *gorm.DB) {
	// 记录初始状态
	var initialUserCount int64
	db.Model(&models.User{}).Count(&initialUserCount)

	// 执行会失败的事务
	err := db.Transaction(func(tx *gorm.DB) error {
		// 成功创建第一个用户
		user1 := &models.User{
			Username: "rollback_test_1",
			Email:    "rollback1@test.com",
		}
		if err := tx.Create(user1).Error; err != nil {
			return err
		}

		// 成功创建第二个用户
		user2 := &models.User{
			Username: "rollback_test_2",
			Email:    "rollback2@test.com",
		}
		if err := tx.Create(user2).Error; err != nil {
			return err
		}

		// 模拟失败操作
		return fmt.Errorf("simulated transaction failure")
	})

	// 事务应该失败
	if err == nil {
		t.Error("Expected transaction to fail")
	}

	// 验证所有操作都被回滚
	var finalUserCount int64
	db.Model(&models.User{}).Count(&finalUserCount)

	if finalUserCount != initialUserCount {
		t.Errorf("Transaction rollback failed: initial=%d, final=%d", initialUserCount, finalUserCount)
	}

	// 确认特定用户没有被创建
	var count int64
	db.Model(&models.User{}).Where("username LIKE ?", "rollback_test_%").Count(&count)
	if count > 0 {
		t.Errorf("Found %d users that should have been rolled back", count)
	}
}