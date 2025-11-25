package lifecycle

import (
	"context"
	"time"
)

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                 `json:"status"`    // "healthy", "degraded", "unhealthy"
	Timestamp time.Time              `json:"timestamp"` // 检查时间
	Details   map[string]interface{} `json:"details"`   // 详细信息
}

// Lifecycle 生命周期管理接口
type Lifecycle interface {
	// Start 启动组件
	Start(ctx context.Context) error

	// Stop 停止组件
	Stop(ctx context.Context) error

	// Health 健康检查
	Health() HealthStatus
}

// Component 组件信息
type Component struct {
	Name      string
	Lifecycle Lifecycle
	Priority  int // 优先级，数字越小越先启动，越后停止
}
