package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"superview/internal/models"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// isSecureRequest 检查请求是否通过 HTTPS
func isSecureRequest(c *gin.Context) bool {
	// 检查 X-Forwarded-Proto（反向代理场景）
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		return true
	}
	// 检查请求 scheme
	return c.Request.TLS != nil
}

// setCookie 设置 Cookie，自动根据请求协议设置 Secure 标志
func (s *AuthService) setCookie(c *gin.Context, name, value string, maxAge int) {
	secure := isSecureRequest(c)
	c.SetCookie(name, value, maxAge, "/", "", secure, true)
}

func (s *AuthService) Login(c *gin.Context) {
	type loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request format",
		})
		return
	}

	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Invalid username/password",
		})
		return
	}

	// 验证密码
	passwordValid := user.VerifyPassword(req.Password)

	if !passwordValid {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Invalid username/password",
		})
		return
	}

	// 生成JWT令牌
	token, err := GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate token",
		})
		return
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Update("last_login", now)

	// 设置Cookie（自动检测 HTTPS 并设置 Secure 标志）
	s.setCookie(c, "token", token, 3600*24)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":         user.ID,
				"username":   user.Username,
				"email":      user.Email,
				"full_name":  user.FullName,
				"is_admin":   user.IsAdmin,
				"is_active":  user.IsActive,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			},
		},
	})
}

func (s *AuthService) Logout(c *gin.Context) {
	// 清除Cookie（自动检测 HTTPS 并设置 Secure 标志）
	s.setCookie(c, "token", "", -1)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logout successful",
	})
}

func (s *AuthService) GetCurrentUser(c *gin.Context) {
	// 从请求头或Cookie获取令牌
	var tokenString string

	// 先尝试从Authorization头获取
	auth := c.GetHeader("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}

	// 如果请求头中没有令牌，尝试从Cookie获取
	if tokenString == "" {
		token, err := c.Cookie("token")
		if err == nil {
			tokenString = token
		}
	}

	// 如果都没有找到令牌
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Not authenticated",
		})
		return
	}

	// 验证令牌
	claims, err := ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Invalid or expired token",
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get user info",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user": gin.H{
				"id":        user.ID,
				"username":  user.Username,
				"email":     user.Email,
				"full_name": user.FullName,
				"is_active": user.IsActive,
				"is_admin":  user.IsAdmin,
			},
		},
	})
}

func (s *AuthService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 排除登录页面和静态资源
		if c.Request.URL.Path == "/login" || strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Next()
			return
		}

		// 从请求头或Cookie获取令牌
		var tokenString string

		// 先尝试从Authorization头获取
		auth := c.GetHeader("Authorization")
		if auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 如果请求头中没有令牌，尝试从Cookie获取
		if tokenString == "" {
			token, err := c.Cookie("token")
			if err == nil {
				tokenString = token
			}
		}

		// 如果还没有找到令牌，尝试从URL参数获取（用于WebSocket连接）
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		// 如果都没有找到令牌
		if tokenString == "" {
			// 如果是API请求返回JSON错误
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  "error",
					"message": "Authorization is required",
				})
				c.Abort()
				return
			}
			// 否则重定向到登录页面
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户ID存储在上下文中
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
