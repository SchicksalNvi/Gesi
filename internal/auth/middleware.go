package auth

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// AuthMiddleware 实现基本的JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "未提供认证令牌"})
			return
		}

		// 从 Bearer Token 中提取令牌
		parts := strings.SplitN(token, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{"error": "无效的认证令牌格式"})
			return
		}
		token = parts[1]

		// 验证JWT令牌
		claims, err := ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "无效的认证令牌"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
