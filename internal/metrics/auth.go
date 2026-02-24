package metrics

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// BasicAuthMiddleware 提供 Basic Auth 认证中间件
type BasicAuthMiddleware struct {
	username string
	password string
	realm    string
}

// NewBasicAuthMiddleware 创建 Basic Auth 中间件
func NewBasicAuthMiddleware(username, password string) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		username: username,
		password: password,
		realm:    "Prometheus Metrics",
	}
}

// Wrap 包装 handler 添加 Basic Auth 认证
func (m *BasicAuthMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	// 如果没有配置认证信息，直接返回原 handler
	if m.username == "" && m.password == "" {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			m.unauthorized(w)
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			m.unauthorized(w)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			m.unauthorized(w)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			m.unauthorized(w)
			return
		}

		// 使用常量时间比较防止时序攻击
		usernameMatch := subtle.ConstantTimeCompare([]byte(pair[0]), []byte(m.username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(pair[1]), []byte(m.password)) == 1

		if !usernameMatch || !passwordMatch {
			m.unauthorized(w)
			return
		}

		next(w, r)
	}
}

func (m *BasicAuthMiddleware) unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+m.realm+`"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
