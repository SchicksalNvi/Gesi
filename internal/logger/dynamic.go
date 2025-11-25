package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// dynamicLevel 动态日志级别控制器
	dynamicLevel zap.AtomicLevel
	// levelMutex 保护日志级别变更的互斥锁
	levelMutex sync.RWMutex
	// levelHistory 日志级别变更历史
	levelHistory []LevelChange
	// maxHistorySize 最大历史记录数量
	maxHistorySize = 100
)

// LevelChange 日志级别变更记录
type LevelChange struct {
	Timestamp time.Time     `json:"timestamp"`
	OldLevel  zapcore.Level `json:"old_level"`
	NewLevel  zapcore.Level `json:"new_level"`
	ChangedBy string        `json:"changed_by"`
	Reason    string        `json:"reason,omitempty"`
}

// LogLevelInfo 日志级别信息
type LogLevelInfo struct {
	CurrentLevel    string        `json:"current_level"`
	AvailableLevels []string      `json:"available_levels"`
	History         []LevelChange `json:"history"`
	LastChanged     *time.Time    `json:"last_changed,omitempty"`
}

// InitDynamicLogger 初始化支持动态级别调整的日志系统
func InitDynamicLogger() error {
	// 创建动态级别控制器
	dynamicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

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

	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 创建文件输出
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	appLogFile, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	errorLogFile, err := os.OpenFile(filepath.Join(logDir, "error.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// 创建控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleOutput := zapcore.AddSync(os.Stdout)

	// 创建核心
	core := zapcore.NewTee(
		// 控制台输出（所有级别）
		zapcore.NewCore(consoleEncoder, consoleOutput, dynamicLevel),
		// 应用日志文件（Info及以上）
		zapcore.NewCore(fileEncoder, zapcore.AddSync(appLogFile), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel && dynamicLevel.Enabled(lvl)
		})),
		// 错误日志文件（Error及以上）
		zapcore.NewCore(fileEncoder, zapcore.AddSync(errorLogFile), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel && dynamicLevel.Enabled(lvl)
		})),
	)

	// 创建logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	globalLogger = logger
	Sugar = logger.Sugar()
	zap.ReplaceGlobals(logger)

	// 记录初始化日志
	Info("Dynamic logger initialized", zap.String("initial_level", dynamicLevel.Level().String()))

	return nil
}

// SetLogLevel 设置日志级别
func SetLogLevel(level string, changedBy string, reason string) error {
	levelMutex.Lock()
	defer levelMutex.Unlock()

	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	// 获取当前级别
	oldLevel := dynamicLevel.Level()

	// 设置新级别
	dynamicLevel.SetLevel(zapLevel)

	// 记录级别变更
	levelChange := LevelChange{
		Timestamp: time.Now(),
		OldLevel:  oldLevel,
		NewLevel:  zapLevel,
		ChangedBy: changedBy,
		Reason:    reason,
	}

	// 添加到历史记录
	levelHistory = append(levelHistory, levelChange)
	if len(levelHistory) > maxHistorySize {
		levelHistory = levelHistory[1:]
	}

	// 记录级别变更日志
	Info("Log level changed",
		zap.String("old_level", oldLevel.String()),
		zap.String("new_level", zapLevel.String()),
		zap.String("changed_by", changedBy),
		zap.String("reason", reason),
	)

	return nil
}

// GetLogLevel 获取当前日志级别
func GetLogLevel() string {
	levelMutex.RLock()
	defer levelMutex.RUnlock()
	return dynamicLevel.Level().String()
}

// GetLogLevelInfo 获取日志级别详细信息
func GetLogLevelInfo() LogLevelInfo {
	levelMutex.RLock()
	defer levelMutex.RUnlock()

	info := LogLevelInfo{
		CurrentLevel:    dynamicLevel.Level().String(),
		AvailableLevels: []string{"debug", "info", "warn", "error", "fatal"},
		History:         make([]LevelChange, len(levelHistory)),
	}

	// 复制历史记录
	copy(info.History, levelHistory)

	// 设置最后变更时间
	if len(levelHistory) > 0 {
		lastChange := levelHistory[len(levelHistory)-1].Timestamp
		info.LastChanged = &lastChange
	}

	return info
}

// ResetLogLevel 重置日志级别到默认值
func ResetLogLevel(changedBy string) error {
	return SetLogLevel("info", changedBy, "Reset to default level")
}

// SetTemporaryLogLevel 设置临时日志级别（指定时间后自动恢复）
func SetTemporaryLogLevel(level string, duration time.Duration, changedBy string, reason string) error {
	// 记录当前级别
	currentLevel := GetLogLevel()

	// 设置新级别
	if err := SetLogLevel(level, changedBy, fmt.Sprintf("Temporary: %s", reason)); err != nil {
		return err
	}

	// 启动定时器恢复原级别
	go func() {
		time.Sleep(duration)
		SetLogLevel(currentLevel, "system", fmt.Sprintf("Auto-restore after %v", duration))
	}()

	return nil
}

// GetAvailableLogLevels 获取可用的日志级别列表
func GetAvailableLogLevels() []string {
	return []string{"debug", "info", "warn", "error", "fatal"}
}

// ClearLevelHistory 清空日志级别变更历史
func ClearLevelHistory(changedBy string) {
	levelMutex.Lock()
	defer levelMutex.Unlock()

	levelHistory = []LevelChange{}
	Info("Log level history cleared", zap.String("cleared_by", changedBy))
}
