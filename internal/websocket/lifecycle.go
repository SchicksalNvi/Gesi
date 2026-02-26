package websocket

import (
	"context"
	"sync"
	"time"

	"superview/internal/lifecycle"
	"superview/internal/logger"
	"go.uber.org/zap"
)

// HubLifecycle 为 WebSocket Hub 实现 Lifecycle 接口
type HubLifecycle struct {
	hub     *Hub
	mu      sync.RWMutex
	started bool
	stopped bool
}

// NewHubLifecycle 创建 WebSocket Hub 的生命周期管理器
func NewHubLifecycle(hub *Hub) *HubLifecycle {
	return &HubLifecycle{
		hub: hub,
	}
}

// Start 启动 WebSocket Hub
func (hl *HubLifecycle) Start(ctx context.Context) error {
	hl.mu.Lock()
	defer hl.mu.Unlock()

	if hl.started {
		return nil // 已经启动
	}

	logger.Info("Starting WebSocket Hub", zap.String("component", "websocket"))

	// 启动 Hub 的运行循环
	go hl.hub.Run()

	hl.started = true
	logger.Info("WebSocket Hub started successfully", zap.String("component", "websocket"))
	return nil
}

// Stop 停止 WebSocket Hub
func (hl *HubLifecycle) Stop(ctx context.Context) error {
	hl.mu.Lock()
	defer hl.mu.Unlock()

	if hl.stopped {
		return nil // 已经停止
	}

	logger.Info("Stopping WebSocket Hub", zap.String("component", "websocket"))

	// 设置停止超时
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 关闭 Hub
	hl.hub.Close()

	// 等待所有消息发送完成
	if err := hl.waitForMessageCompletion(stopCtx); err != nil {
		logger.Warn("Timeout waiting for message completion", zap.Error(err))
	}

	// 关闭所有连接
	hl.closeAllConnections()

	hl.stopped = true
	logger.Info("WebSocket Hub stopped successfully", zap.String("component", "websocket"))
	return nil
}

// Health 检查 WebSocket Hub 健康状态
func (hl *HubLifecycle) Health() lifecycle.HealthStatus {
	hl.mu.RLock()
	defer hl.mu.RUnlock()

	if !hl.started || hl.stopped {
		return lifecycle.HealthStatus{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"reason":  "hub not running",
				"started": hl.started,
				"stopped": hl.stopped,
			},
		}
	}

	// 获取连接统计
	stats := hl.getConnectionStats()

	// 确定健康状态
	var status string
	if stats.ActiveConnections >= 0 {
		status = "healthy"
	} else {
		status = "degraded"
	}

	return lifecycle.HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"active_connections": stats.ActiveConnections,
			"max_connections":    hl.hub.config.MaxConnections,
			"broadcast_queue":    len(hl.hub.broadcast),
		},
	}
}

// ConnectionStats 连接统计信息
type ConnectionStats struct {
	ActiveConnections int
	MaxConnections    int
	BroadcastQueue    int
}

// getConnectionStats 获取连接统计信息
func (hl *HubLifecycle) getConnectionStats() ConnectionStats {
	return ConnectionStats{
		ActiveConnections: int(hl.hub.GetConnectionCount()),
		MaxConnections:    hl.hub.config.MaxConnections,
		BroadcastQueue:    len(hl.hub.broadcast),
	}
}

// waitForMessageCompletion 等待消息发送完成
func (hl *HubLifecycle) waitForMessageCompletion(ctx context.Context) error {
	// 等待广播队列清空
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// 检查广播队列是否为空
			if len(hl.hub.broadcast) == 0 {
				return nil
			}
		}
	}
}

// closeAllConnections 关闭所有连接
func (hl *HubLifecycle) closeAllConnections() {
	logger.Info("Closing all WebSocket connections")

	hl.hub.clientsMu.RLock()
	clients := make([]*Client, 0, len(hl.hub.clients))
	for client := range hl.hub.clients {
		clients = append(clients, client)
	}
	hl.hub.clientsMu.RUnlock()

	// 发送关闭消息并关闭连接
	shutdownMsg := []byte(`{"type":"server_shutdown","message":"Server is shutting down"}`)
	
	for _, client := range clients {
		client.mu.Lock()
		if !client.closed {
			// 尝试发送关闭消息
			select {
			case client.send <- shutdownMsg:
				// 等待一小段时间让消息发送
				time.Sleep(100 * time.Millisecond)
			default:
				// 如果发送队列满了，直接跳过
			}
			
			// 关闭连接
			client.closed = true
			client.conn.Close()
		}
		client.mu.Unlock()
	}

	logger.Info("All WebSocket connections closed", zap.Int("count", len(clients)))
}
