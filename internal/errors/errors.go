package errors

import (
	"fmt"
)

// AppError 应用错误接口
type AppError interface {
	error
	Code() string
	Message() string
	Details() interface{}
	Cause() error
	WithContext(key string, value interface{}) AppError
}

// baseError 基础错误实现
type baseError struct {
	code    string
	message string
	details interface{}
	cause   error
	context map[string]interface{}
}

func (e *baseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *baseError) Code() string {
	return e.code
}

func (e *baseError) Message() string {
	return e.message
}

func (e *baseError) Details() interface{} {
	return e.details
}

func (e *baseError) Cause() error {
	return e.cause
}

func (e *baseError) WithContext(key string, value interface{}) AppError {
	if e.context == nil {
		e.context = make(map[string]interface{})
	}
	e.context[key] = value
	return e
}

// ValidationError 验证错误
type ValidationError struct {
	*baseError
	Field string
}

func NewValidationError(field string, message string) AppError {
	if message == "" {
		message = fmt.Sprintf("validation failed for field: %s", field)
	}
	return &ValidationError{
		baseError: &baseError{
			code:    "VALIDATION_ERROR",
			message: message,
			details: map[string]interface{}{"field": field},
		},
		Field: field,
	}
}

// NotFoundError 资源未找到错误
type NotFoundError struct {
	*baseError
	Resource string
	ID       string
}

func NewNotFoundError(resource string, id string) AppError {
	return &NotFoundError{
		baseError: &baseError{
			code:    "NOT_FOUND",
			message: fmt.Sprintf("%s not found: %s", resource, id),
			details: map[string]interface{}{
				"resource": resource,
				"id":       id,
			},
		},
		Resource: resource,
		ID:       id,
	}
}

// ConflictError 冲突错误
type ConflictError struct {
	*baseError
	Resource string
}

func NewConflictError(resource string, message string) AppError {
	if message == "" {
		message = fmt.Sprintf("conflict with resource: %s", resource)
	}
	return &ConflictError{
		baseError: &baseError{
			code:    "CONFLICT",
			message: message,
			details: map[string]interface{}{"resource": resource},
		},
		Resource: resource,
	}
}

// InternalError 内部错误
type InternalError struct {
	*baseError
}

func NewInternalError(message string, cause error) AppError {
	if message == "" {
		message = "internal server error"
	}
	return &InternalError{
		baseError: &baseError{
			code:    "INTERNAL_ERROR",
			message: message,
			cause:   cause,
			details: map[string]interface{}{},
		},
	}
}

// UnauthorizedError 未授权错误
type UnauthorizedError struct {
	*baseError
}

func NewUnauthorizedError(message string) AppError {
	if message == "" {
		message = "unauthorized access"
	}
	return &UnauthorizedError{
		baseError: &baseError{
			code:    "UNAUTHORIZED",
			message: message,
			details: map[string]interface{}{},
		},
	}
}

// ForbiddenError 禁止访问错误
type ForbiddenError struct {
	*baseError
	RequiredPermission string
}

func NewForbiddenError(message string) AppError {
	if message == "" {
		message = "access forbidden"
	}
	return &ForbiddenError{
		baseError: &baseError{
			code:    "FORBIDDEN",
			message: message,
			details: map[string]interface{}{},
		},
	}
}

// DatabaseError 数据库错误
type DatabaseError struct {
	*baseError
	Operation string
}

func NewDatabaseError(operation string, cause error) AppError {
	message := fmt.Sprintf("database operation failed: %s", operation)
	return &DatabaseError{
		baseError: &baseError{
			code:    "DATABASE_ERROR",
			message: message,
			cause:   cause,
			details: map[string]interface{}{"operation": operation},
		},
		Operation: operation,
	}
}

// ConnectionError 连接错误
type ConnectionError struct {
	*baseError
	Target string
}

func NewConnectionError(target string, cause error) AppError {
	message := fmt.Sprintf("connection failed: %s", target)
	return &ConnectionError{
		baseError: &baseError{
			code:    "CONNECTION_ERROR",
			message: message,
			cause:   cause,
			details: map[string]interface{}{"target": target},
		},
		Target: target,
	}
}

// IsAppError 检查是否为 AppError
func IsAppError(err error) bool {
	_, ok := err.(AppError)
	return ok
}

// GetAppError 获取 AppError，如果不是则包装为 InternalError
func GetAppError(err error) AppError {
	if appErr, ok := err.(AppError); ok {
		return appErr
	}
	return NewInternalError("unexpected error", err)
}

// 类型检查函数

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// IsConflictError 检查是否为冲突错误
func IsConflictError(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}

// IsInternalError 检查是否为内部错误
func IsInternalError(err error) bool {
	_, ok := err.(*InternalError)
	return ok
}

// IsUnauthorizedError 检查是否为未授权错误
func IsUnauthorizedError(err error) bool {
	_, ok := err.(*UnauthorizedError)
	return ok
}

// IsForbiddenError 检查是否为禁止错误
func IsForbiddenError(err error) bool {
	_, ok := err.(*ForbiddenError)
	return ok
}

// IsDatabaseError 检查是否为数据库错误
func IsDatabaseError(err error) bool {
	_, ok := err.(*DatabaseError)
	return ok
}

// IsConnectionError 检查是否为连接错误
func IsConnectionError(err error) bool {
	_, ok := err.(*ConnectionError)
	return ok
}
