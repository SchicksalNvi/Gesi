package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// Sugar 全局Sugar日志实例，提供更简洁的API
	Sugar *zap.SugaredLogger
)

// Init 初始化日志系统
func Init() error {
	// 获取当前工作目录
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	// 确保日志目录存在
	logDir := filepath.Join(projectRoot, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 配置日志
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{
		"stdout",
		filepath.Join(logDir, "app.log"),
	}
	config.ErrorOutputPaths = []string{
		"stderr",
		filepath.Join(logDir, "error.log"),
	}

	// 创建logger
	logger, err := config.Build()
	if err != nil {
		return err
	}

	Logger = logger
	Sugar = logger.Sugar()

	return nil
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Close 关闭日志系统
func Close() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Info 记录信息级别日志
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Error 记录错误级别日志
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Warn 记录警告级别日志
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Debug 记录调试级别日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Fatal 记录致命错误级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}
