package logger

import (
	"context"

	"go.uber.org/zap"
)

// StructuredLogger 结构化日志接口
type StructuredLogger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) StructuredLogger
	WithContext(ctx context.Context) StructuredLogger
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// String 创建字符串字段
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Error 创建错误字段
func ErrorField(key string, err error) Field {
	return Field{Key: key, Value: err}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// contextLogger 带上下文的日志记录器
type contextLogger struct {
	logger *zap.Logger
	fields []zap.Field
}

// NewContextLogger 创建带上下文的日志记录器
func NewContextLogger(logger *zap.Logger) StructuredLogger {
	return &contextLogger{
		logger: logger,
		fields: []zap.Field{},
	}
}

// Debug 记录调试日志
func (l *contextLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, l.convertFields(fields)...)
}

// Info 记录信息日志
func (l *contextLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, l.convertFields(fields)...)
}

// Warn 记录警告日志
func (l *contextLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, l.convertFields(fields)...)
}

// Error 记录错误日志
func (l *contextLogger) Error(msg string, fields ...Field) {
	// 自动添加堆栈跟踪
	zapFields := l.convertFields(fields)
	zapFields = append(zapFields, zap.Stack("stacktrace"))
	l.logger.Error(msg, zapFields...)
}

// With 添加字段
func (l *contextLogger) With(fields ...Field) StructuredLogger {
	zapFields := l.convertFields(fields)
	return &contextLogger{
		logger: l.logger.With(zapFields...),
		fields: append(l.fields, zapFields...),
	}
}

// WithContext 从上下文添加字段
func (l *contextLogger) WithContext(ctx context.Context) StructuredLogger {
	// 从上下文提取字段
	fields := extractFieldsFromContext(ctx)
	return l.With(fields...)
}

// convertFields 转换字段格式
func (l *contextLogger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(f.Key, v))
		case int:
			zapFields = append(zapFields, zap.Int(f.Key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		default:
			zapFields = append(zapFields, zap.Any(f.Key, v))
		}
	}
	return zapFields
}

// 上下文键类型
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// WithRequestID 添加请求 ID 到上下文
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID 添加用户 ID 到上下文
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// FromContext 从上下文获取日志记录器
func FromContext(ctx context.Context) StructuredLogger {
	fields := extractFieldsFromContext(ctx)
	return NewContextLogger(zap.L()).With(fields...)
}

// extractFieldsFromContext 从上下文提取字段
func extractFieldsFromContext(ctx context.Context) []Field {
	fields := []Field{}

	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		fields = append(fields, String("request_id", requestID))
	}

	if userID, ok := ctx.Value(userIDKey).(string); ok {
		fields = append(fields, String("user_id", userID))
	}

	return fields
}
