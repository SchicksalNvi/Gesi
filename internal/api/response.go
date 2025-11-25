package api

import (
	"net/http"

	"go-cesi/internal/errors"

	"github.com/gin-gonic/gin"
)

// Response 标准响应格式
type Response struct {
	Status  string      `json:"status"`            // "success" or "error"
	Message string      `json:"message,omitempty"` // 可选的消息
	Data    interface{} `json:"data,omitempty"`    // 成功时的数据
	Error   *ErrorInfo  `json:"error,omitempty"`   // 错误时的错误信息
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string                 `json:"code"`              // 错误代码
	Message string                 `json:"message"`           // 错误消息
	Details map[string]interface{} `json:"details,omitempty"` // 详细信息
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, err error) {
	// 转换为 AppError
	appErr := errors.GetAppError(err)

	// 根据错误类型确定 HTTP 状态码
	statusCode := getHTTPStatusCode(appErr)

	// 构建错误信息
	errorInfo := &ErrorInfo{
		Code:    appErr.Code(),
		Message: appErr.Message(),
	}

	// 添加详细信息
	if details := appErr.Details(); details != nil {
		if detailsMap, ok := details.(map[string]interface{}); ok {
			errorInfo.Details = detailsMap
		}
	}

	c.JSON(statusCode, Response{
		Status: "error",
		Error:  errorInfo,
	})
}

// ValidationError 返回验证错误响应
func ValidationError(c *gin.Context, field string, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Status: "error",
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Details: map[string]interface{}{
				"field": field,
			},
		},
	})
}

// ValidationErrors 返回多个验证错误
func ValidationErrors(c *gin.Context, errs map[string]string) {
	c.JSON(http.StatusBadRequest, Response{
		Status: "error",
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: map[string]interface{}{
				"errors": errs,
			},
		},
	})
}

// getHTTPStatusCode 根据错误类型返回 HTTP 状态码
func getHTTPStatusCode(err errors.AppError) int {
	switch err.Code() {
	case "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "INTERNAL_ERROR":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// 旧的响应类型（保持向后兼容）
// 注意：新代码应使用 Response 类型
type ErrorResponseLegacy struct {
	Error string `json:"error"`
}

type SuccessResponseLegacy struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
