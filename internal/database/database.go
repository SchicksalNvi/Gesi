package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go-cesi/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
	healthCheckTicker *time.Ticker
	healthCheckStop chan struct{}
	healthCheckMutex sync.RWMutex
	isHealthy bool = true
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MaxOpenConns       int           // 最大打开连接数
	MaxIdleConns       int           // 最大空闲连接数
	ConnMaxLifetime    time.Duration // 连接最大生命周期
	ConnMaxIdleTime    time.Duration // 连接最大空闲时间
	QueryTimeout       time.Duration // 查询超时时间
	HealthCheckEnabled bool          // 是否启用健康检查
	HealthCheckInterval time.Duration // 健康检查间隔
	TransactionTimeout time.Duration // 事务超时时间
}

// GetDefaultConfig 获取默认数据库配置
func GetDefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MaxOpenConns:        25,                // SQLite建议较少的连接数
		MaxIdleConns:        5,                 // 保持少量空闲连接
		ConnMaxLifetime:     5 * time.Minute,   // 连接5分钟后重新创建
		ConnMaxIdleTime:     1 * time.Minute,   // 空闲1分钟后关闭
		QueryTimeout:        30 * time.Second,  // 查询超时30秒
		HealthCheckEnabled:  true,              // 启用健康检查
		HealthCheckInterval: 30 * time.Second,  // 每30秒检查一次
		TransactionTimeout:  60 * time.Second,  // 事务超时60秒
	}
}

func InitDB() error {
	return InitDBWithConfig(GetDefaultConfig())
}

func InitDBWithConfig(config *DatabaseConfig) error {
	// 获取当前工作目录作为项目根目录
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	// 确保数据库目录存在
	dataDir := filepath.Join(projectRoot, "data")
	log.Printf("Database directory: %s", dataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		PrepareStmt: true, // 启用预编译语句缓存
	}

	// 连接SQLite数据库
	dbPath := filepath.Join(dataDir, "cesi.db")
	dsn := fmt.Sprintf("%s?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	// 获取底层sql.DB实例并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 测试数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), config.QueryTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// 自动迁移数据库模式
	err = db.AutoMigrate(
		&models.User{},
		&models.ActivityLog{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.NodeAccess{},
		&models.AlertRule{},
		&models.Alert{},
		&models.NotificationChannel{},
		&models.Notification{},
		&models.AlertRuleNotificationChannel{},
		&models.SystemMetric{},
		&models.ProcessGroup{},
		&models.ProcessGroupItem{},
		&models.ProcessDependency{},
		&models.ScheduledTask{},
		&models.TaskExecution{},
		&models.ProcessTemplate{},
		&models.ProcessBackup{},
		&models.ProcessMetrics{},
		&models.Configuration{},
		&models.EnvironmentVariable{},
		&models.ConfigurationHistory{},
		&models.ConfigurationBackup{},
		&models.ConfigurationTemplate{},
		&models.ConfigurationValidation{},
		&models.ConfigurationAudit{},
		&models.LogEntry{},
		&models.LogAnalysisRule{},
		&models.LogStatistics{},
		&models.LogAlert{},
		&models.LogFilter{},
		&models.LogExport{},
		&models.LogRetentionPolicy{},
		&models.BackupRecord{},
		&models.DataExportRecord{},
		&models.DataImportRecord{},
		&models.SystemSettings{},
		&models.UserPreferences{},
		&models.WebhookConfig{},
		&models.WebhookLog{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate models: %v", err)
	}

	DB = db

	// 启动健康检查
	if config.HealthCheckEnabled {
		StartHealthCheck(config.HealthCheckInterval)
	}

	log.Printf("Database initialized successfully with connection pool (max_open: %d, max_idle: %d)", 
		config.MaxOpenConns, config.MaxIdleConns)
	return nil
}

// WithTransaction 执行事务操作
func WithTransaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// WithTransactionContext 带上下文的事务操作
func WithTransactionContext(ctx context.Context, fn func(*gorm.DB) error) error {
	return DB.WithContext(ctx).Transaction(fn)
}

// WithTransactionTimeout 带超时的事务操作
func WithTransactionTimeout(timeout time.Duration, fn func(*gorm.DB) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WithTransactionContext(ctx, fn)
}

// StartHealthCheck 启动数据库健康检查
func StartHealthCheck(interval time.Duration) {
	StopHealthCheck() // 确保之前的检查已停止
	
	healthCheckStop = make(chan struct{})
	healthCheckTicker = time.NewTicker(interval)
	
	go func() {
		for {
			select {
			case <-healthCheckTicker.C:
				performHealthCheck()
			case <-healthCheckStop:
				return
			}
		}
	}()
	
	log.Printf("Database health check started with interval: %v", interval)
}

// StopHealthCheck 停止数据库健康检查
func StopHealthCheck() {
	if healthCheckTicker != nil {
		healthCheckTicker.Stop()
		healthCheckTicker = nil
	}
	if healthCheckStop != nil {
		close(healthCheckStop)
		healthCheckStop = nil
	}
}

// performHealthCheck 执行健康检查
func performHealthCheck() {
	err := HealthCheck()
	
	healthCheckMutex.Lock()
	defer healthCheckMutex.Unlock()
	
	if err != nil {
		if isHealthy {
			log.Printf("Database health check failed: %v", err)
			isHealthy = false
		}
	} else {
		if !isHealthy {
			log.Printf("Database health check recovered")
			isHealthy = true
		}
	}
}

// IsHealthy 检查数据库是否健康
func IsHealthy() bool {
	healthCheckMutex.RLock()
	defer healthCheckMutex.RUnlock()
	return isHealthy
}

// GetHealthStatus 获取详细的健康状态信息
func GetHealthStatus() map[string]interface{} {
	healthCheckMutex.RLock()
	healthy := isHealthy
	healthCheckMutex.RUnlock()
	
	stats, err := GetConnectionStats()
	if err != nil {
		return map[string]interface{}{
			"healthy": false,
			"error":   err.Error(),
		}
	}
	
	return map[string]interface{}{
		"healthy":           healthy,
		"open_connections":  stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
		"wait_count":       stats.WaitCount,
		"wait_duration":    stats.WaitDuration.String(),
		"max_idle_closed":  stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

// GetConnectionStats 获取连接池统计信息
func GetConnectionStats() (sql.DBStats, error) {
	sqlDB, err := DB.DB()
	if err != nil {
		return sql.DBStats{}, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}
	return sqlDB.Stats(), nil
}

// Close 关闭数据库连接
func Close() error {
	// 停止健康检查
	StopHealthCheck()
	
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}
	return sqlDB.Close()
}
