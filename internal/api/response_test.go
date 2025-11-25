package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go-cesi/internal/errors"
)

// 属性 15：响应格式统一
// 属性 16：错误响应标准化
// 验证需求：6.2, 6.3
func TestResponseFormatProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	properties.Property("all success responses contain status and data", prop.ForAll(
		func(data string) bool {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Success(c, map[string]string{"result": data})

			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				return false
			}

			return response.Status == "success" && response.Data != nil
		},
		gen.AnyString(),
	))

	properties.Property("all error responses contain status and error info", prop.ForAll(
		func(msg string) bool {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			err := errors.NewInternalError(msg, nil)
			Error(c, err)

			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				return false
			}

			return response.Status == "error" &&
				response.Error != nil &&
				response.Error.Code != "" &&
				response.Error.Message != ""
		},
		gen.AnyString(),
	))

	properties.Property("error responses have correct HTTP status codes", prop.ForAll(
		func() bool {
			testCases := []struct {
				err            errors.AppError
				expectedStatus int
			}{
				{errors.NewValidationError("field", "invalid"), http.StatusBadRequest},
				{errors.NewNotFoundError("resource", "id"), http.StatusNotFound},
				{errors.NewConflictError("resource", "conflict"), http.StatusConflict},
				{errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized},
				{errors.NewForbiddenError("forbidden"), http.StatusForbidden},
				{errors.NewInternalError("internal", nil), http.StatusInternalServerError},
			}

			for _, tc := range testCases {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				Error(c, tc.err)

				if w.Code != tc.expectedStatus {
					return false
				}
			}
			return true
		},
	))

	properties.TestingRun(t)
}

// 单元测试
func TestSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"key": "value"}
	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", response.Status)
	}

	if response.Data == nil {
		t.Error("expected data to be present")
	}
}

func TestSuccessWithMessageResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SuccessWithMessage(c, "operation completed", map[string]string{"id": "123"})

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Message != "operation completed" {
		t.Errorf("expected message 'operation completed', got '%s'", response.Message)
	}
}

func TestErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := errors.NewNotFoundError("user", "123")
	Error(c, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("expected status 'error', got '%s'", response.Status)
	}

	if response.Error == nil {
		t.Fatal("expected error info to be present")
	}

	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code 'NOT_FOUND', got '%s'", response.Error.Code)
	}
}

func TestValidationErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ValidationError(c, "email", "invalid email format")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code 'VALIDATION_ERROR', got '%s'", response.Error.Code)
	}
}

func TestValidationErrorsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errs := map[string]string{
		"email":    "invalid email",
		"password": "password too short",
	}
	ValidationErrors(c, errs)

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Error.Details == nil {
		t.Fatal("expected details to be present")
	}
}
