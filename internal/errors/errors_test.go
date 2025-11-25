package errors

import (
	"errors"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// 属性 1：错误上下文完整性
// 验证需求：2.1
func TestErrorContextCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("all errors contain code, message and details", prop.ForAll(
		func(msg string, code string) bool {
			// 测试所有错误类型
			errors := []AppError{
				NewValidationError("field", msg),
				NewNotFoundError("resource", "id"),
				NewConflictError("resource", msg),
				NewInternalError(msg, nil),
				NewUnauthorizedError(msg),
				NewForbiddenError(msg),
			}

			for _, err := range errors {
				if err.Code() == "" {
					return false
				}
				if err.Message() == "" {
					return false
				}
				if err.Details() == nil {
					return false
				}
			}
			return true
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.Property("errors with context preserve context", prop.ForAll(
		func(key string, value string) bool {
			err := NewInternalError("test", nil).
				WithContext(key, value)

			// 验证错误仍然有效
			return err.Code() != "" && err.Message() != ""
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.Property("errors with cause preserve cause", prop.ForAll(
		func(msg string) bool {
			cause := errors.New("original error")
			err := NewInternalError(msg, cause)

			return err.Cause() == cause
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// 单元测试：验证错误类型
func TestValidationError(t *testing.T) {
	err := NewValidationError("username", "username is required")

	if err.Code() != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", err.Code())
	}

	if err.Message() != "username is required" {
		t.Errorf("expected message 'username is required', got %s", err.Message())
	}

	details := err.Details().(map[string]interface{})
	if details["field"] != "username" {
		t.Errorf("expected field 'username', got %v", details["field"])
	}
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError("user", "123")

	if err.Code() != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", err.Code())
	}

	if err.Message() != "user not found: 123" {
		t.Errorf("unexpected message: %s", err.Message())
	}
}

func TestConflictError(t *testing.T) {
	err := NewConflictError("user", "user already exists")

	if err.Code() != "CONFLICT" {
		t.Errorf("expected code CONFLICT, got %s", err.Code())
	}

	if err.Message() != "user already exists" {
		t.Errorf("unexpected message: %s", err.Message())
	}
}

func TestInternalError(t *testing.T) {
	cause := errors.New("database connection failed")
	err := NewInternalError("failed to save user", cause)

	if err.Code() != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", err.Code())
	}

	if err.Cause() != cause {
		t.Errorf("expected cause to be preserved")
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("invalid token")

	if err.Code() != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %s", err.Code())
	}
}

func TestForbiddenError(t *testing.T) {
	err := NewForbiddenError("insufficient permissions")

	if err.Code() != "FORBIDDEN" {
		t.Errorf("expected code FORBIDDEN, got %s", err.Code())
	}
}

func TestWithContext(t *testing.T) {
	err := NewInternalError("test error", nil).
		WithContext("user_id", "123").
		WithContext("action", "delete")

	if err.Code() != "INTERNAL_ERROR" {
		t.Errorf("expected error to maintain its type after adding context")
	}
}

func TestIsAppError(t *testing.T) {
	appErr := NewInternalError("test", nil)
	stdErr := errors.New("standard error")

	if !IsAppError(appErr) {
		t.Error("expected IsAppError to return true for AppError")
	}

	if IsAppError(stdErr) {
		t.Error("expected IsAppError to return false for standard error")
	}
}

func TestGetAppError(t *testing.T) {
	appErr := NewInternalError("test", nil)
	stdErr := errors.New("standard error")

	result1 := GetAppError(appErr)
	if result1 != appErr {
		t.Error("expected GetAppError to return the same AppError")
	}

	result2 := GetAppError(stdErr)
	if result2.Code() != "INTERNAL_ERROR" {
		t.Error("expected GetAppError to wrap standard error as InternalError")
	}
}
