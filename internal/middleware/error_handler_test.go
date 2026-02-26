package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appErrors "superview/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)
}

// TestErrorHandlerPanic 测试 panic 恢复
func TestErrorHandlerPanic(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
}

// TestErrorHandlerAppError 测试 AppError 处理
func TestErrorHandlerAppError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/error", func(c *gin.Context) {
		err := appErrors.NewNotFoundError("resource", "123")
		c.Error(err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
}

// TestErrorHandlerStandardError 测试标准错误处理
func TestErrorHandlerStandardError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/error", func(c *gin.Context) {
		c.Error(errors.New("standard error"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
}

// TestErrorHandlerNoError 测试无错误情况
func TestErrorHandlerNoError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "success"}})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/success", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestRequestLogger 测试请求日志中间件
func TestRequestLogger(t *testing.T) {
	router := gin.New()
	router.Use(RequestLogger())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRecovery 测试恢复中间件
func TestRecovery(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
}

// TestNotFoundHandler 测试 404 处理器
func TestNotFoundHandler(t *testing.T) {
	router := gin.New()
	router.NoRoute(NotFoundHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
}

// TestMethodNotAllowedHandler 测试 405 处理器
func TestMethodNotAllowedHandler(t *testing.T) {
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.NoMethod(MethodNotAllowedHandler())

	// 只允许 GET 方法
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 尝试使用 POST 方法
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

// TestErrorHandlerMultipleErrors 测试多个错误
func TestErrorHandlerMultipleErrors(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/errors", func(c *gin.Context) {
		c.Error(errors.New("error 1"))
		c.Error(errors.New("error 2"))
		c.Error(errors.New("error 3"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/errors", nil)
	router.ServeHTTP(w, req)

	// 应该只处理最后一个错误
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error 3")
}

// TestErrorHandlerWithResponse 测试已经写入响应的情况
func TestErrorHandlerWithResponse(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/response", func(c *gin.Context) {
		// 先写入响应
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
		// 然后添加错误（不应该覆盖响应）
		c.Error(errors.New("error after response"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/response", nil)
	router.ServeHTTP(w, req)

	// 响应应该保持为 200
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
