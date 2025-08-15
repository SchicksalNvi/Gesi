package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// 定义错误类型
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	ErrorTypeForbidden    ErrorType = "forbidden"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeDatabase     ErrorType = "database"
	ErrorTypeNetwork      ErrorType = "network"
	ErrorTypeTimeout      ErrorType = "timeout"
)

// AppError 应用程序错误结构
type AppError struct {
	Type    ErrorType `json:"type"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Cause   error     `json:"-"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// HTTPStatus 返回对应的HTTP状态码
func (e *AppError) HTTPStatus() int {
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// 预定义的错误变量
var (
	// 通用错误
	ErrInvalidInput     = errors.New("invalid input")
	ErrInternalServer   = errors.New("internal server error")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrNotFound         = errors.New("resource not found")
	ErrConflict         = errors.New("resource conflict")
	ErrTimeout          = errors.New("operation timeout")

	// 用户相关错误
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooWeak  = errors.New("password too weak")

	// 节点相关错误
	ErrNodeNotFound     = errors.New("node not found")
	ErrNodeNotConnected = errors.New("node not connected")
	ErrNodeExists       = errors.New("node already exists")

	// 进程相关错误
	ErrProcessNotFound  = errors.New("process not found")
	ErrProcessRunning   = errors.New("process is running")
	ErrProcessStopped   = errors.New("process is stopped")
	ErrOperationFailed  = errors.New("operation failed")

	// 数据库相关错误
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseQuery      = errors.New("database query failed")
	ErrDatabaseTransaction = errors.New("database transaction failed")
)

// 错误构造函数

// NewValidationError 创建验证错误
func NewValidationError(code, message string, details ...string) *AppError {
	err := &AppError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// NewNotFoundError 创建资源未找到错误
func NewNotFoundError(resource string, identifier string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Code:    "RESOURCE_NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Details: fmt.Sprintf("identifier: %s", identifier),
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Code:    "FORBIDDEN",
		Message: message,
	}
}

// NewConflictError 创建冲突错误
func NewConflictError(resource string, message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Code:    "RESOURCE_CONFLICT",
		Message: fmt.Sprintf("%s conflict: %s", resource, message),
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Code:    "INTERNAL_ERROR",
		Message: message,
		Cause:   cause,
	}
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(operation string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeDatabase,
		Code:    "DATABASE_ERROR",
		Message: fmt.Sprintf("database %s failed", operation),
		Cause:   cause,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(operation string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeNetwork,
		Code:    "NETWORK_ERROR",
		Message: fmt.Sprintf("network %s failed", operation),
		Cause:   cause,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(operation string) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Code:    "TIMEOUT_ERROR",
		Message: fmt.Sprintf("%s timeout", operation),
	}
}

// NewConnectionError 创建连接错误
func NewConnectionError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeNetwork,
		Code:    "CONNECTION_ERROR",
		Message: message,
	}
}

// 错误检查辅助函数

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeValidation
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeNotFound
}

// IsUnauthorizedError 检查是否为未授权错误
func IsUnauthorizedError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeUnauthorized
}

// IsForbiddenError 检查是否为禁止访问错误
func IsForbiddenError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeForbidden
}

// IsConflictError 检查是否为冲突错误
func IsConflictError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeConflict
}

// IsInternalError 检查是否为内部错误
func IsInternalError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeInternal
}

// IsDatabaseError 检查是否为数据库错误
func IsDatabaseError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeDatabase
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeNetwork
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeTimeout
}

// IsConnectionError 检查是否为连接错误
func IsConnectionError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Type == ErrorTypeNetwork {
		return true
	}
	// 也检查预定义的连接错误
	return errors.Is(err, ErrNodeNotConnected)
}

// WrapError 包装现有错误为AppError
func WrapError(err error, errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   err,
	}
}
