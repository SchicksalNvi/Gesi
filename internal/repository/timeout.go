package repository

import (
	"context"
	"time"
)

// DefaultQueryTimeout 默认查询超时时间
const DefaultQueryTimeout = 30 * time.Second

// WithTimeout 为上下文添加超时控制
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultQueryTimeout
	}
	return context.WithTimeout(ctx, timeout)
}

// WithDefaultTimeout 使用默认超时时间
func WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithTimeout(ctx, DefaultQueryTimeout)
}
