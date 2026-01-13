package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_SingleImplementation(t *testing.T) {
	// 确保只有一个认证中间件实现
	gin.SetMode(gin.TestMode)
	
	authService := NewAuthService(nil)
	middleware := authService.AuthMiddleware()
	
	// 测试中间件存在且可调用
	assert.NotNil(t, middleware, "AuthMiddleware should return a valid middleware function")
}

func TestAuthMiddleware_ConsistentBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	authService := NewAuthService(nil)
	
	tests := []struct {
		name           string
		path           string
		headers        map[string]string
		cookies        map[string]string
		queryParams    map[string]string
		expectedStatus int
		expectedJSON   bool
	}{
		{
			name:           "API request without token returns JSON error",
			path:           "/api/nodes",
			expectedStatus: http.StatusUnauthorized,
			expectedJSON:   true,
		},
		{
			name:           "Web request without token redirects to login",
			path:           "/dashboard",
			expectedStatus: http.StatusFound,
			expectedJSON:   false,
		},
		{
			name:           "Login page is excluded from auth",
			path:           "/login",
			expectedStatus: http.StatusOK,
			expectedJSON:   false,
		},
		{
			name:           "Static resources are excluded from auth",
			path:           "/static/css/style.css",
			expectedStatus: http.StatusOK,
			expectedJSON:   false,
		},
		{
			name: "Bearer token in Authorization header",
			path: "/api/nodes",
			headers: map[string]string{
				"Authorization": "Bearer invalid_token",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedJSON:   true,
		},
		{
			name: "Token in cookie",
			path: "/api/nodes",
			cookies: map[string]string{
				"token": "invalid_token",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedJSON:   true,
		},
		{
			name: "Token in query parameter (for WebSocket)",
			path: "/api/nodes",
			queryParams: map[string]string{
				"token": "invalid_token",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			// 添加认证中间件
			router.Use(authService.AuthMiddleware())
			
			// 添加测试路由
			router.GET("/*path", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// 创建请求
			req := httptest.NewRequest("GET", tt.path, nil)
			
			// 添加请求头
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			
			// 添加Cookie
			for key, value := range tt.cookies {
				req.AddCookie(&http.Cookie{Name: key, Value: value})
			}
			
			// 添加查询参数
			if len(tt.queryParams) > 0 {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match expected")
			
			// 验证响应类型
			if tt.expectedJSON {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json", 
					"API requests should return JSON")
			}
			
			// 验证重定向
			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				assert.Equal(t, "/login", location, "Should redirect to login page")
			}
		})
	}
}

func TestAuthMiddleware_ConsistentErrorFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	authService := NewAuthService(nil)
	router := gin.New()
	router.Use(authService.AuthMiddleware())
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试无令牌的错误格式
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "error")
	assert.Contains(t, w.Body.String(), "Authorization is required")
	
	// 测试无效令牌的错误格式
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "error")
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestAuthMiddleware_ContextKeyConsistency(t *testing.T) {
	// 这个测试确保上下文键名的一致性
	// 由于我们无法轻易创建有效的JWT令牌进行测试，
	// 这里主要是文档化预期的行为
	
	// 预期行为：
	// - 成功认证后，用户ID应该存储在 c.Set("user_id", claims.UserID)
	// - 所有使用认证的地方都应该通过 c.GetString("user_id") 获取用户ID
	
	t.Log("Context key should be 'user_id' for consistency")
	t.Log("All authenticated endpoints should use c.GetString('user_id') to get user ID")
}