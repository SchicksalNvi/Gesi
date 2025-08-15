package loggers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go-cesi/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ActivityLog struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Level     string `gorm:"size:16"`   // INFO, WARN, ERROR
	Action    string `gorm:"size:64"`   // 操作类型
	Resource  string `gorm:"size:128"`  // 操作资源
	Username  string `gorm:"size:64"`   // 操作用户
	Message   string `gorm:"size:1024"` // 日志消息
	Data      string `gorm:"type:text"` // 额外数据(JSON格式)
}

type ActivityLogService struct {
	db         *gorm.DB
	fileLogger *FileLogger
	mu         sync.Mutex
}

func InitActivityLogService(db *gorm.DB) {
	once.Do(func() {
		activityLogService = newActivityLogService(db)
	})
}

func newActivityLogService(db *gorm.DB) *ActivityLogService {
	// 初始化文件日志记录器
	logPath := filepath.Join("logs", "activity.log")
	fileLogger, err := NewFileLogger(logPath, 10, 5) // 10MB大小限制，保留5个备份
	if err != nil {
		// 如果文件日志初始化失败，只使用数据库日志
		return &ActivityLogService{db: db}
	}

	// 确保日志目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		logger.Error("Failed to create logs directory", zap.Error(err))
		return &ActivityLogService{db: db}
	}

	return &ActivityLogService{
		db:         db,
		fileLogger: fileLogger,
	}
}

func (s *ActivityLogService) Log(level, action, resource, username, message string, data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 写入数据库日志
	logEntry := ActivityLog{
		Level:     level,
		Action:    action,
		Resource:  resource,
		Username:  username,
		Message:   message,
		Data:      toJSONString(data),
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&logEntry).Error; err != nil {
		return err
	}

	// 写入文件日志
	if s.fileLogger != nil {
		logData := fmt.Sprintf("%s | %s | %s | %s | %v",
			action, resource, username, message, data)
		s.fileLogger.WriteLog(level, logData)
	}

	return nil
}

func (s *ActivityLogService) GetActivityLogs(page, pageSize int, filters map[string]interface{}) ([]ActivityLog, int64, error) {
	var logs []ActivityLog
	var total int64

	query := s.db.Model(&ActivityLog{})

	// 应用过滤条件
	if level, ok := filters["level"]; ok && level != "" {
		query = query.Where("level = ?", level)
	}
	if action, ok := filters["action"]; ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"]; ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if username, ok := filters["username"]; ok && username != "" {
		query = query.Where("username = ?", username)
	}
	if startTime, ok := filters["start_time"]; ok && startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["end_time"]; ok && endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (s *ActivityLogService) GetRecentLogs(limit int) ([]ActivityLog, error) {
	var logs []ActivityLog

	if err := s.db.Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func (s *ActivityLogService) Close() error {
	if s.fileLogger != nil {
		return s.fileLogger.Close()
	}
	return nil
}

func toJSONString(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}

	jsonStr := "{"
	first := true
	for k, v := range data {
		if !first {
			jsonStr += ","
		}
		jsonStr += fmt.Sprintf("\"%s\":\"%v\"", k, v)
		first = false
	}
	jsonStr += "}"

	return jsonStr
}
