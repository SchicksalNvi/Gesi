package logger

import (
	"go.uber.org/zap"
)

var (
	// globalLogger 全局日志实例
	globalLogger *zap.Logger
	// Sugar 全局Sugar日志实例，提供更简洁的API
	Sugar *zap.SugaredLogger
)

// Init 初始化日志系统（使用动态日志级别）
func Init() error {
	return InitDynamicLogger()
}

// Sync 同步日志缓冲区
func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}

// Close 关闭日志系统
func Close() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}

// Info 记录信息级别日志
func Info(msg string, fields ...zap.Field) {
	zap.L().Info(msg, fields...)
}

// Error 记录错误级别日志
func Error(msg string, fields ...zap.Field) {
	zap.L().Error(msg, fields...)
}

// Warn 记录警告级别日志
func Warn(msg string, fields ...zap.Field) {
	zap.L().Warn(msg, fields...)
}

// Debug 记录调试级别日志
func Debug(msg string, fields ...zap.Field) {
	zap.L().Debug(msg, fields...)
}

// Fatal 记录致命错误级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	zap.L().Fatal(msg, fields...)
}
