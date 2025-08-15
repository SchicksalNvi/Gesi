package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appErrors "go-cesi/internal/errors"
)

// ErrorResponse 统一错误响应格式
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse 统一成功响应格式
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// parseIDParam 解析URL参数中的ID
func parseIDParam(c *gin.Context, paramName string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(paramName), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// getUserID 从上下文中获取用户ID
func getUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// handleError 统一错误处理
func handleError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}

// handleAppError 处理应用程序错误
func handleAppError(c *gin.Context, err error) {
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus(), ErrorResponse{Error: appErr.Message})
	} else {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}
}

// handleSuccess 统一成功响应
func handleSuccess(c *gin.Context, message string, data interface{}) {
	response := SuccessResponse{
		Status:  "success",
		Message: message,
	}
	if data != nil {
		response.Data = data
	}
	c.JSON(http.StatusOK, response)
}

// handleInvalidID 处理无效ID错误
func handleInvalidID(c *gin.Context, resourceType string) {
	err := appErrors.NewValidationError("INVALID_ID", "Invalid "+resourceType+" ID")
	handleAppError(c, err)
}

// handleUnauthorized 处理未授权错误
func handleUnauthorized(c *gin.Context) {
	err := appErrors.NewUnauthorizedError("User not authenticated")
	handleAppError(c, err)
}

// handleInternalError 处理内部服务器错误
func handleInternalError(c *gin.Context, err error) {
	appErr := appErrors.NewInternalError("Internal server error", err)
	handleAppError(c, appErr)
}

// handleBadRequest 处理请求参数错误
func handleBadRequest(c *gin.Context, err error) {
	appErr := appErrors.NewValidationError("BAD_REQUEST", "Invalid request parameters", err.Error())
	handleAppError(c, appErr)
}

// handleNotFound 处理资源未找到错误
func handleNotFound(c *gin.Context, resource, identifier string) {
	err := appErrors.NewNotFoundError(resource, identifier)
	handleAppError(c, err)
}

// handleConflict 处理资源冲突错误
func handleConflict(c *gin.Context, resource, message string) {
	err := appErrors.NewConflictError(resource, message)
	handleAppError(c, err)
}

// handleForbidden 处理禁止访问错误
func handleForbidden(c *gin.Context, message string) {
	err := appErrors.NewForbiddenError(message)
	handleAppError(c, err)
}

// parseAndValidateID 解析并验证ID参数的通用函数
func parseAndValidateID(c *gin.Context, paramName, resourceType string) (uint, bool) {
	id, err := parseIDParam(c, paramName)
	if err != nil {
		handleInvalidID(c, resourceType)
		return 0, false
	}
	return id, true
}

// validateUserAuth 验证用户认证的通用函数
func validateUserAuth(c *gin.Context) (uint, bool) {
	userID, exists := getUserID(c)
	if !exists {
		handleUnauthorized(c)
		return 0, false
	}
	return userID, true
}