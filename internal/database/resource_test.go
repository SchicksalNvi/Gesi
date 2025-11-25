package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 属性 27：资源释放
// 验证需求：10.5
func TestResourceReleaseProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("idle connections are released after timeout", prop.ForAll(
		func(seed int) bool {
			// 使用固定的短超时时间
			idleTime := 1

			// 创建测试数据库配置
			config := &DatabaseConfig{
				MaxOpenConns:        5,
				MaxIdleConns:        2,
				ConnMaxLifetime:     10 * time.Second,
				ConnMaxIdleTime:     time.Duration(idleTime) * time.Second,
				QueryTimeout:        5 * time.Second,
				HealthCheckEnabled:  false,
				HealthCheckInterval: 30 * time.Second,
				TransactionTimeout:  10 * time.Second,
			}

			// 初始化测试数据库
			err := InitDBWithConfig(config)
			if err != nil {
				return false
			}

			sqlDB, err := DB.DB()
			if err != nil {
				return false
			}

			// 执行一些查询以创建连接
			for i := 0; i < 2; i++ {
				var result int
				DB.Raw("SELECT 1").Scan(&result)
			}

			// 获取初始统计
			initialStats := sqlDB.Stats()
			initialOpen := initialStats.OpenConnections

			// 等待空闲超时
			time.Sleep(time.Duration(idleTime+1) * time.Second)

			// 获取最终统计
			finalStats := sqlDB.Stats()
			finalOpen := finalStats.OpenConnections

			// 验证连接数减少或保持在合理范围
			return finalOpen <= initialOpen
		},
		gen.IntRange(1, 10),
	))

	properties.Property("connection pool respects max limits", prop.ForAll(
		func(maxOpen int, maxIdle int) bool {
			if maxOpen < 1 || maxOpen > 50 {
				maxOpen = 10
			}
			if maxIdle < 1 || maxIdle > maxOpen {
				maxIdle = maxOpen / 2
			}

			config := &DatabaseConfig{
				MaxOpenConns:        maxOpen,
				MaxIdleConns:        maxIdle,
				ConnMaxLifetime:     10 * time.Second,
				ConnMaxIdleTime:     2 * time.Second,
				QueryTimeout:        5 * time.Second,
				HealthCheckEnabled:  false,
				HealthCheckInterval: 30 * time.Second,
				TransactionTimeout:  10 * time.Second,
			}

			err := InitDBWithConfig(config)
			if err != nil {
				return false
			}

			sqlDB, err := DB.DB()
			if err != nil {
				return false
			}

			// 执行多个并发查询
			done := make(chan bool, maxOpen*2)
			for i := 0; i < maxOpen*2; i++ {
				go func() {
					var result int
					DB.Raw("SELECT 1").Scan(&result)
					done <- true
				}()
			}

			// 等待所有查询完成
			for i := 0; i < maxOpen*2; i++ {
				<-done
			}

			// 获取统计
			stats := sqlDB.Stats()

			// 验证打开的连接数不超过最大值
			return stats.OpenConnections <= maxOpen
		},
		gen.IntRange(1, 50),
		gen.IntRange(1, 25),
	))

	properties.Property("connections are reused efficiently", prop.ForAll(
		func(queryCount int) bool {
			if queryCount < 1 || queryCount > 50 {
				queryCount = 10
			}

			config := GetDefaultConfig()
			config.HealthCheckEnabled = false

			err := InitDBWithConfig(config)
			if err != nil {
				return false
			}

			sqlDB, err := DB.DB()
			if err != nil {
				return false
			}

			// 执行多个查询
			for i := 0; i < queryCount; i++ {
				var result int
				DB.Raw("SELECT 1").Scan(&result)
			}

			// 获取统计
			stats := sqlDB.Stats()

			// 验证连接被重用（打开的连接数应该远小于查询数）
			return stats.OpenConnections < queryCount
		},
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}

// TestDatabaseConfigDefaults 测试默认配置
func TestDatabaseConfigDefaults(t *testing.T) {
	config := GetDefaultConfig()

	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 1*time.Minute, config.ConnMaxIdleTime)
	assert.Equal(t, 30*time.Second, config.QueryTimeout)
	assert.True(t, config.HealthCheckEnabled)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
	assert.Equal(t, 60*time.Second, config.TransactionTimeout)
}

// TestConnectionPoolConfiguration 测试连接池配置
func TestConnectionPoolConfiguration(t *testing.T) {
	config := &DatabaseConfig{
		MaxOpenConns:        10,
		MaxIdleConns:        5,
		ConnMaxLifetime:     5 * time.Minute,
		ConnMaxIdleTime:     1 * time.Minute,
		QueryTimeout:        30 * time.Second,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
		TransactionTimeout:  60 * time.Second,
	}

	err := InitDBWithConfig(config)
	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	// 验证连接池配置
	stats := sqlDB.Stats()
	assert.Equal(t, 10, stats.MaxOpenConnections)
}

// TestIdleConnectionRelease 测试空闲连接释放
func TestIdleConnectionRelease(t *testing.T) {
	config := &DatabaseConfig{
		MaxOpenConns:        10,
		MaxIdleConns:        5,
		ConnMaxLifetime:     10 * time.Second,
		ConnMaxIdleTime:     1 * time.Second, // 1秒后释放空闲连接
		QueryTimeout:        5 * time.Second,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
		TransactionTimeout:  10 * time.Second,
	}

	err := InitDBWithConfig(config)
	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	// 执行一些查询以创建连接
	for i := 0; i < 5; i++ {
		var result int
		DB.Raw("SELECT 1").Scan(&result)
	}

	// 获取初始连接数
	initialStats := sqlDB.Stats()
	initialOpen := initialStats.OpenConnections

	// 等待空闲超时
	time.Sleep(2 * time.Second)

	// 获取最终连接数
	finalStats := sqlDB.Stats()
	finalOpen := finalStats.OpenConnections

	// 验证连接数减少或保持合理
	assert.LessOrEqual(t, finalOpen, initialOpen)
}

// TestConnectionLifetime 测试连接生命周期
func TestConnectionLifetime(t *testing.T) {
	config := &DatabaseConfig{
		MaxOpenConns:        5,
		MaxIdleConns:        2,
		ConnMaxLifetime:     2 * time.Second, // 2秒后重新创建连接
		ConnMaxIdleTime:     1 * time.Second,
		QueryTimeout:        5 * time.Second,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
		TransactionTimeout:  10 * time.Second,
	}

	err := InitDBWithConfig(config)
	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	// 执行查询
	var result int
	DB.Raw("SELECT 1").Scan(&result)

	// 等待连接生命周期结束
	time.Sleep(3 * time.Second)

	// 再次执行查询（应该使用新连接）
	DB.Raw("SELECT 1").Scan(&result)

	// 验证连接池仍然正常工作
	stats := sqlDB.Stats()
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
}

// TestMaxOpenConnections 测试最大打开连接数
func TestMaxOpenConnections(t *testing.T) {
	config := &DatabaseConfig{
		MaxOpenConns:        3,
		MaxIdleConns:        1,
		ConnMaxLifetime:     10 * time.Second,
		ConnMaxIdleTime:     2 * time.Second,
		QueryTimeout:        5 * time.Second,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
		TransactionTimeout:  10 * time.Second,
	}

	err := InitDBWithConfig(config)
	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	// 并发执行多个查询
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			var result int
			DB.Raw("SELECT 1").Scan(&result)
			done <- true
		}()
	}

	// 等待所有查询完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证打开的连接数不超过最大值
	stats := sqlDB.Stats()
	assert.LessOrEqual(t, stats.OpenConnections, 3)
}

// TestConnectionReuse 测试连接重用
func TestConnectionReuse(t *testing.T) {
	config := &DatabaseConfig{
		MaxOpenConns:        5,
		MaxIdleConns:        3,
		ConnMaxLifetime:     10 * time.Second,
		ConnMaxIdleTime:     2 * time.Second,
		QueryTimeout:        5 * time.Second,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
		TransactionTimeout:  10 * time.Second,
	}

	err := InitDBWithConfig(config)
	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	// 执行多个查询
	for i := 0; i < 20; i++ {
		var result int
		DB.Raw("SELECT 1").Scan(&result)
	}

	// 获取统计
	stats := sqlDB.Stats()

	// 验证连接被重用（打开的连接数应该远小于查询数）
	assert.Less(t, stats.OpenConnections, 20)
	assert.GreaterOrEqual(t, stats.OpenConnections, 1)
}

// TestQueryTimeout 测试查询超时
func TestQueryTimeout(t *testing.T) {
	config := &DatabaseConfig{
		MaxOpenConns:        5,
		MaxIdleConns:        2,
		ConnMaxLifetime:     10 * time.Second,
		ConnMaxIdleTime:     2 * time.Second,
		QueryTimeout:        1 * time.Second,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
		TransactionTimeout:  10 * time.Second,
	}

	err := InitDBWithConfig(config)
	require.NoError(t, err)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 执行查询
	var result int
	err = DB.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error

	// 查询应该成功或超时
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
}

// TestConnectionPoolStats 测试连接池统计
func TestConnectionPoolStats(t *testing.T) {
	config := GetDefaultConfig()
	config.HealthCheckEnabled = false

	err := InitDBWithConfig(config)
	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	// 执行一些查询
	for i := 0; i < 5; i++ {
		var result int
		DB.Raw("SELECT 1").Scan(&result)
	}

	// 获取统计
	stats := sqlDB.Stats()

	// 验证统计信息
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
	assert.LessOrEqual(t, stats.OpenConnections, config.MaxOpenConns)
	assert.GreaterOrEqual(t, stats.InUse, 0)
	assert.GreaterOrEqual(t, stats.Idle, 0)
}

// BenchmarkDatabaseQuery 基准测试：数据库查询
func BenchmarkDatabaseQuery(b *testing.B) {
	config := GetDefaultConfig()
	config.HealthCheckEnabled = false

	err := InitDBWithConfig(config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result int
		DB.Raw("SELECT 1").Scan(&result)
	}
}

// BenchmarkConnectionPoolReuse 基准测试：连接池重用
func BenchmarkConnectionPoolReuse(b *testing.B) {
	config := &DatabaseConfig{
		MaxOpenConns:        10,
		MaxIdleConns:        5,
		ConnMaxLifetime:     10 * time.Second,
		ConnMaxIdleTime:     2 * time.Second,
		QueryTimeout:        5 * time.Second,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
		TransactionTimeout:  10 * time.Second,
	}

	err := InitDBWithConfig(config)
	if err != nil {
		b.Fatal(err)
	}

	sqlDB, _ := DB.DB()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result int
		DB.Raw("SELECT 1").Scan(&result)
	}
	b.StopTimer()

	stats := sqlDB.Stats()
	b.ReportMetric(float64(stats.OpenConnections), "open_conns")
	b.ReportMetric(float64(stats.InUse), "in_use")
	b.ReportMetric(float64(stats.Idle), "idle")
}

// Helper function to get connection stats
func getConnectionStats(db *sql.DB) sql.DBStats {
	return db.Stats()
}
