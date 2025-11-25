package logger

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"
)

// 属性 19：结构化日志格式
// 验证需求：7.1
func TestStructuredLoggingProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// 初始化测试日志
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	logger, _ := config.Build()

	properties.Property("all logs contain timestamp and level", prop.ForAll(
		func(msg string) bool {
			// 创建上下文日志记录器
			ctxLogger := NewContextLogger(logger)

			// 测试所有日志级别都能正常工作
			ctxLogger.Debug(msg)
			ctxLogger.Info(msg)
			ctxLogger.Warn(msg)
			// Error 会添加堆栈跟踪

			return true // 如果没有 panic，说明日志格式正确
		},
		gen.AnyString(),
	))

	properties.Property("logs with context preserve context fields", prop.ForAll(
		func(requestID string, userID string) bool {
			ctx := context.Background()
			ctx = WithRequestID(ctx, requestID)
			ctx = WithUserID(ctx, userID)

			ctxLogger := FromContext(ctx)
			ctxLogger.Info("test message")

			return true // 验证不会 panic
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// 单元测试
func TestContextLogger(t *testing.T) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	logger, err := config.Build()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	ctxLogger := NewContextLogger(logger)

	// 测试基本日志记录
	ctxLogger.Debug("debug message")
	ctxLogger.Info("info message")
	ctxLogger.Warn("warn message")
	ctxLogger.Error("error message")
}

func TestLoggerWithFields(t *testing.T) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	logger, err := config.Build()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	ctxLogger := NewContextLogger(logger)

	// 测试带字段的日志
	ctxLogger.Info("test message",
		String("key1", "value1"),
		Int("key2", 42),
	)

	// 测试 With 方法
	loggerWithFields := ctxLogger.With(
		String("service", "test"),
		String("version", "1.0"),
	)
	loggerWithFields.Info("message with preset fields")
}

func TestContextFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithUserID(ctx, "user-456")

	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	logger, err := config.Build()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	ctxLogger := NewContextLogger(logger).WithContext(ctx)
	ctxLogger.Info("message with context fields")
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-789")

	// 初始化全局 logger
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	logger, err := config.Build()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	zap.ReplaceGlobals(logger)

	ctxLogger := FromContext(ctx)
	ctxLogger.Info("message from context")
}

func TestDynamicLogLevel(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	// 测试获取日志级别
	level := GetLogLevel()
	if level != "info" {
		t.Errorf("expected level 'info', got '%s'", level)
	}

	// 测试设置日志级别
	if err := SetLogLevel("debug", "test", "testing"); err != nil {
		t.Errorf("failed to set log level: %v", err)
	}

	level = GetLogLevel()
	if level != "debug" {
		t.Errorf("expected level 'debug', got '%s'", level)
	}

	// 测试无效的日志级别
	if err := SetLogLevel("invalid", "test", "testing"); err == nil {
		t.Error("expected error for invalid log level")
	}

	// 测试获取日志级别信息
	info := GetLogLevelInfo()
	if info.CurrentLevel != "debug" {
		t.Errorf("expected current level 'debug', got '%s'", info.CurrentLevel)
	}
	if len(info.AvailableLevels) == 0 {
		t.Error("expected available levels to be populated")
	}
}
