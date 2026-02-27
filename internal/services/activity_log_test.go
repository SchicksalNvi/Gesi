package services

import (
	"testing"
	"time"

	"superview/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.ActivityLog{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestLogActivity(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	log := &models.ActivityLog{
		Level:     "INFO",
		Message:   "Test message",
		Action:    "test_action",
		Resource:  "test_resource",
		Target:    "test_target",
		Username:  "test_user",
		IPAddress: "127.0.0.1",
	}

	err := service.LogActivity(log)
	assert.NoError(t, err)
	assert.NotZero(t, log.ID)

	// 验证日志已保存
	var savedLog models.ActivityLog
	db.First(&savedLog, log.ID)
	assert.Equal(t, log.Message, savedLog.Message)
	assert.Equal(t, log.Username, savedLog.Username)
}

func TestGetActivityLogs(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// 创建测试数据
	testLogs := []*models.ActivityLog{
		{
			Level:     "INFO",
			Message:   "Log 1",
			Action:    "action1",
			Resource:  "resource1",
			Target:    "target1",
			Username:  "user1",
			IPAddress: "127.0.0.1",
		},
		{
			Level:     "ERROR",
			Message:   "Log 2",
			Action:    "action2",
			Resource:  "resource2",
			Target:    "target2",
			Username:  "user2",
			IPAddress: "127.0.0.2",
		},
		{
			Level:     "WARNING",
			Message:   "Log 3",
			Action:    "action3",
			Resource:  "resource3",
			Target:    "target3",
			Username:  "user1",
			IPAddress: "127.0.0.1",
		},
	}

	for _, log := range testLogs {
		service.LogActivity(log)
	}

	t.Run("Get all logs", func(t *testing.T) {
		logs, total, err := service.GetActivityLogs(1, 10, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, logs, 3)
	})

	t.Run("Filter by level", func(t *testing.T) {
		filters := map[string]interface{}{
			"level": "ERROR",
		}
		logs, total, err := service.GetActivityLogs(1, 10, filters)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, logs, 1)
		assert.Equal(t, "ERROR", logs[0].Level)
	})

	t.Run("Filter by username", func(t *testing.T) {
		filters := map[string]interface{}{
			"username": "user1",
		}
		logs, total, err := service.GetActivityLogs(1, 10, filters)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, logs, 2)
	})

	t.Run("Pagination", func(t *testing.T) {
		logs, total, err := service.GetActivityLogs(1, 2, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, logs, 2)

		logs, total, err = service.GetActivityLogs(2, 2, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, logs, 1)
	})
}

func TestGetRecentLogs(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// 创建测试数据
	for i := 0; i < 10; i++ {
		service.LogActivity(&models.ActivityLog{
			Level:     "INFO",
			Message:   "Test log",
			Action:    "test",
			Resource:  "test",
			Target:    "test",
			Username:  "admin",
			IPAddress: "127.0.0.1",
		})
	}

	t.Run("Get recent logs with limit", func(t *testing.T) {
		logs, err := service.GetRecentLogs(5)
		assert.NoError(t, err)
		assert.Len(t, logs, 5)
	})

	t.Run("Get all logs when limit exceeds total", func(t *testing.T) {
		logs, err := service.GetRecentLogs(20)
		assert.NoError(t, err)
		assert.Len(t, logs, 10)
	})
}

func TestGetLogStatistics(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// 创建测试数据
	levels := []string{"INFO", "WARNING", "ERROR"}
	actions := []string{"login", "logout", "start_process"}
	usernames := []string{"admin", "user1", "user2"}

	for _, level := range levels {
		for _, action := range actions {
			for _, username := range usernames {
				service.LogActivity(&models.ActivityLog{
					Level:     level,
					Message:   "Test log",
					Action:    action,
					Resource:  "test",
					Target:    "test",
					Username:  username,
					IPAddress: "127.0.0.1",
				})
			}
		}
	}

	stats, err := service.GetLogStatistics(7)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(27), stats["total_logs"])
	assert.NotZero(t, stats["info_count"])
	assert.NotZero(t, stats["warning_count"])
	assert.NotZero(t, stats["error_count"])
	assert.NotNil(t, stats["top_actions"])
	assert.NotNil(t, stats["top_users"])
}

func TestCleanOldLogs(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// 创建旧日志
	oldLog := &models.ActivityLog{
		Level:     "INFO",
		Message:   "Old log",
		Action:    "test",
		Resource:  "test",
		Target:    "test",
		Username:  "admin",
		IPAddress: "127.0.0.1",
		CreatedAt: time.Now().AddDate(0, 0, -100),
	}
	db.Create(oldLog)

	// 创建新日志
	newLog := &models.ActivityLog{
		Level:     "INFO",
		Message:   "New log",
		Action:    "test",
		Resource:  "test",
		Target:    "test",
		Username:  "admin",
		IPAddress: "127.0.0.1",
	}
	service.LogActivity(newLog)

	// 清理 90 天前的日志
	err := service.CleanOldLogs(90)
	assert.NoError(t, err)

	// 验证旧日志已删除
	var count int64
	db.Model(&models.ActivityLog{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// 验证新日志仍存在
	var remainingLog models.ActivityLog
	db.First(&remainingLog)
	assert.Equal(t, "New log", remainingLog.Message)
}

func TestExportLogs(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// 创建测试数据
	testLogs := []*models.ActivityLog{
		{
			Level:     "INFO",
			Message:   "Log 1",
			Action:    "action1",
			Resource:  "resource1",
			Target:    "target1",
			Username:  "user1",
			IPAddress: "127.0.0.1",
			Status:    "success",
			Duration:  100,
		},
		{
			Level:     "ERROR",
			Message:   "Log 2",
			Action:    "action2",
			Resource:  "resource2",
			Target:    "target2",
			Username:  "user2",
			IPAddress: "127.0.0.2",
			Status:    "error",
			Duration:  200,
		},
	}

	for _, log := range testLogs {
		service.LogActivity(log)
	}

	t.Run("Export all logs", func(t *testing.T) {
		csvData, err := service.ExportLogs(map[string]interface{}{})
		assert.NoError(t, err)
		assert.NotEmpty(t, csvData)
		
		csvString := string(csvData)
		assert.Contains(t, csvString, "ID,Created At,Level")
		assert.Contains(t, csvString, "Log 1")
		assert.Contains(t, csvString, "Log 2")
	})

	t.Run("Export filtered logs", func(t *testing.T) {
		filters := map[string]interface{}{
			"level": "ERROR",
		}
		csvData, err := service.ExportLogs(filters)
		assert.NoError(t, err)
		assert.NotEmpty(t, csvData)
		
		csvString := string(csvData)
		assert.Contains(t, csvString, "Log 2")
		assert.NotContains(t, csvString, "Log 1")
	})
}

func TestLogSystemEvent(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	err := service.LogSystemEvent("INFO", "process_stopped", "process", "node1:app", "Process stopped unexpectedly", nil)
	assert.NoError(t, err)

	// 验证系统事件已记录
	var log models.ActivityLog
	db.First(&log)
	assert.Equal(t, "system", log.Username)
	assert.Equal(t, "process_stopped", log.Action)
	assert.Equal(t, "Process stopped unexpectedly", log.Message)
}
