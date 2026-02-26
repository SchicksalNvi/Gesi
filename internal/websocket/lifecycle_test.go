package websocket

import (
	"context"
	"testing"
	"time"

	"superview/internal/supervisor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHubLifecycleStart 测试启动 WebSocket Hub
func TestHubLifecycleStart(t *testing.T) {
	// 创建 Hub
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	// 启动
	ctx := context.Background()
	err := lifecycle.Start(ctx)
	require.NoError(t, err)

	// 验证已启动
	assert.True(t, lifecycle.started)
	assert.False(t, lifecycle.stopped)

	// 清理
	lifecycle.Stop(ctx)
}

// TestHubLifecycleStop 测试停止 WebSocket Hub
func TestHubLifecycleStop(t *testing.T) {
	// 创建并启动 Hub
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()
	lifecycle.Start(ctx)

	// 停止
	err := lifecycle.Stop(ctx)
	require.NoError(t, err)

	// 验证已停止
	assert.True(t, lifecycle.stopped)
}

// TestHubLifecycleDoubleStart 测试重复启动
func TestHubLifecycleDoubleStart(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()

	// 第一次启动
	err := lifecycle.Start(ctx)
	require.NoError(t, err)

	// 第二次启动应该是安全的
	err = lifecycle.Start(ctx)
	require.NoError(t, err)

	// 清理
	lifecycle.Stop(ctx)
}

// TestHubLifecycleDoubleStop 测试重复停止
func TestHubLifecycleDoubleStop(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()
	lifecycle.Start(ctx)

	// 第一次停止
	err := lifecycle.Stop(ctx)
	require.NoError(t, err)

	// 第二次停止应该是安全的
	err = lifecycle.Stop(ctx)
	require.NoError(t, err)
}

// TestHubLifecycleHealth 测试健康检查
func TestHubLifecycleHealth(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()

	// 未启动时应该是 unhealthy
	health := lifecycle.Health()
	assert.Equal(t, "unhealthy", health.Status)

	// 启动后应该是 healthy
	lifecycle.Start(ctx)
	time.Sleep(100 * time.Millisecond) // 等待启动完成

	health = lifecycle.Health()
	assert.Equal(t, "healthy", health.Status)
	assert.NotNil(t, health.Details)
	assert.Contains(t, health.Details, "active_connections")

	// 停止后应该是 unhealthy
	lifecycle.Stop(ctx)
	health = lifecycle.Health()
	assert.Equal(t, "unhealthy", health.Status)
}

// TestHubLifecycleConnectionStats 测试连接统计
func TestHubLifecycleConnectionStats(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()
	lifecycle.Start(ctx)
	defer lifecycle.Stop(ctx)

	// 获取统计信息
	stats := lifecycle.getConnectionStats()
	assert.Equal(t, 0, stats.ActiveConnections)
	assert.Equal(t, hub.config.MaxConnections, stats.MaxConnections)
	assert.GreaterOrEqual(t, stats.BroadcastQueue, 0)
}

// TestHubLifecycleGracefulShutdown 测试优雅关闭
func TestHubLifecycleGracefulShutdown(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()
	lifecycle.Start(ctx)

	// 等待一段时间确保 Hub 正在运行
	time.Sleep(100 * time.Millisecond)

	// 优雅关闭
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := lifecycle.Stop(stopCtx)
	require.NoError(t, err)

	// 验证已停止
	assert.True(t, lifecycle.stopped)
}

// TestHubLifecycleStopTimeout 测试停止超时
func TestHubLifecycleStopTimeout(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()
	lifecycle.Start(ctx)

	// 使用非常短的超时
	stopCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// 停止应该成功，即使超时
	err := lifecycle.Stop(stopCtx)
	require.NoError(t, err)
}

// TestHubLifecycleHealthDetails 测试健康检查详细信息
func TestHubLifecycleHealthDetails(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	lifecycle := NewHubLifecycle(hub)

	ctx := context.Background()
	lifecycle.Start(ctx)
	defer lifecycle.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	health := lifecycle.Health()
	assert.Equal(t, "healthy", health.Status)
	assert.NotNil(t, health.Details)

	// 验证详细信息包含必要字段
	details := health.Details
	assert.Contains(t, details, "active_connections")
	assert.Contains(t, details, "max_connections")
	assert.Contains(t, details, "broadcast_queue")

	// 验证值的类型
	assert.IsType(t, 0, details["active_connections"])
	assert.IsType(t, 0, details["max_connections"])
	assert.IsType(t, 0, details["broadcast_queue"])
}
