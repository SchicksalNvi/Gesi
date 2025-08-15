package loggers

import (
	"sync"
)

var (
	activityLogService *ActivityLogService
	once               sync.Once
)

// GetActivityLogService 获取活动日志服务实例
func GetActivityLogService() *ActivityLogService {
	return activityLogService
}
