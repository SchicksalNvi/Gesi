package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"superview/internal/errors"
	"superview/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// sendErrorResponse 发送错误响应（避免循环依赖）
func sendErrorResponse(c *gin.Context, appErr errors.AppError) {
	statusCode := getHTTPStatusCode(appErr.Code())
	
	c.JSON(statusCode, gin.H{
		"success": false,
		"error": gin.H{
			"code":    appErr.Code(),
			"message": appErr.Message(),
			"details": appErr.Details(),
		},
	})
}

// getHTTPStatusCode 根据错误代码获取 HTTP 状态码
func getHTTPStatusCode(code string) int {
	switch code {
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
	case "DATABASE_ERROR", "CONNECTION_ERROR", "INTERNAL_ERROR":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ErrorHandler 统一的错误处理中间件
// 捕获 panic 并转换为错误响应，记录错误日志
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈跟踪
				stack := string(debug.Stack())

				// 记录 panic 错误
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("stack", stack),
				)

				// 转换为内部错误
				appErr := errors.NewInternalError(
					fmt.Sprintf("Internal server error: %v", err),
					fmt.Errorf("%v", err),
				)

				// 返回错误响应
				sendErrorResponse(c, appErr)
			}
		}()

		// 执行下一个处理器
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last().Err

			// 记录错误日志
			logger.Error("Request error",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Int("status", c.Writer.Status()),
			)

			// 如果还没有写入响应，则写入错误响应
			if !c.Writer.Written() {
				// 尝试转换为 AppError
				if appErr, ok := err.(errors.AppError); ok {
					sendErrorResponse(c, appErr)
				} else {
					// 转换为内部错误
					appErr := errors.NewInternalError(err.Error(), err)
					sendErrorResponse(c, appErr)
				}
			}
		}
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始
		logger.Info("Request started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// 执行请求
		c.Next()

		// 记录请求完成
		logger.Info("Request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}

// Recovery 恢复中间件（简化版）
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈跟踪
				stack := string(debug.Stack())

				// 记录 panic
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", stack),
				)

				// 返回 500 错误
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": map[string]interface{}{
						"code":    "INTERNAL_ERROR",
						"message": "Internal server error",
					},
				})
			}
		}()

		c.Next()
	}
}

// NotFoundHandler 404 处理器
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		appErr := errors.NewNotFoundError("route", c.Request.URL.Path)
		sendErrorResponse(c, appErr)
	}
}

// MethodNotAllowedHandler 405 处理器
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		appErr := errors.NewValidationError("method", "Method not allowed for this route")
		sendErrorResponse(c, appErr)
	}
}
