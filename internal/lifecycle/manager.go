package lifecycle

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go-cesi/internal/logger"
	"go.uber.org/zap"
)

// Manager 资源管理器
type Manager struct {
	components []Component
	mu         sync.RWMutex
	started    bool
}

// NewManager 创建资源管理器
func NewManager() *Manager {
	return &Manager{
		components: []Component{},
	}
}

// Register 注册组件
func (m *Manager) Register(name string, lifecycle Lifecycle, priority int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.components = append(m.components, Component{
		Name:      name,
		Lifecycle: lifecycle,
		Priority:  priority,
	})

	logger.Info("Component registered",
		zap.String("name", name),
		zap.Int("priority", priority))
}

// StartAll 启动所有组件
func (m *Manager) StartAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("components already started")
	}

	// 按优先级排序（优先级小的先启动）
	sort.Slice(m.components, func(i, j int) bool {
		return m.components[i].Priority < m.components[j].Priority
	})

	logger.Info("Starting all components", zap.Int("count", len(m.components)))

	// 启动所有组件
	for _, comp := range m.components {
		logger.Info("Starting component", zap.String("name", comp.Name))

		if err := comp.Lifecycle.Start(ctx); err != nil {
			logger.Error("Failed to start component",
				zap.String("name", comp.Name),
				zap.Error(err))
			// 启动失败时，停止已启动的组件
			m.stopStartedComponents(ctx)
			return fmt.Errorf("failed to start component %s: %w", comp.Name, err)
		}

		logger.Info("Component started successfully", zap.String("name", comp.Name))
	}

	m.started = true
	logger.Info("All components started successfully")
	return nil
}

// StopAll 停止所有组件
func (m *Manager) StopAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	logger.Info("Stopping all components", zap.Int("count", len(m.components)))

	// 按优先级逆序停止（优先级大的先停止）
	var errors []error
	for i := len(m.components) - 1; i >= 0; i-- {
		comp := m.components[i]
		logger.Info("Stopping component", zap.String("name", comp.Name))

		if err := comp.Lifecycle.Stop(ctx); err != nil {
			logger.Error("Failed to stop component",
				zap.String("name", comp.Name),
				zap.Error(err))
			errors = append(errors, fmt.Errorf("failed to stop component %s: %w", comp.Name, err))
		} else {
			logger.Info("Component stopped successfully", zap.String("name", comp.Name))
		}
	}

	m.started = false

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while stopping components: %v", errors)
	}

	logger.Info("All components stopped successfully")
	return nil
}

// stopStartedComponents 停止已启动的组件（用于启动失败时的清理）
func (m *Manager) stopStartedComponents(ctx context.Context) {
	logger.Warn("Stopping started components due to startup failure")

	// 逆序停止已启动的组件
	for i := len(m.components) - 1; i >= 0; i-- {
		comp := m.components[i]
		if err := comp.Lifecycle.Stop(ctx); err != nil {
			logger.Error("Failed to stop component during cleanup",
				zap.String("name", comp.Name),
				zap.Error(err))
		}
	}
}

// HealthCheck 健康检查所有组件
func (m *Manager) HealthCheck() map[string]HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]HealthStatus)

	for _, comp := range m.components {
		status := comp.Lifecycle.Health()
		results[comp.Name] = status
	}

	return results
}

// GetComponentHealth 获取特定组件的健康状态
func (m *Manager) GetComponentHealth(name string) (HealthStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, comp := range m.components {
		if comp.Name == name {
			return comp.Lifecycle.Health(), nil
		}
	}

	return HealthStatus{}, fmt.Errorf("component not found: %s", name)
}

// IsHealthy 检查所有组件是否健康
func (m *Manager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, comp := range m.components {
		status := comp.Lifecycle.Health()
		if status.Status != "healthy" {
			return false
		}
	}

	return true
}

// GetOverallHealth 获取整体健康状态
func (m *Manager) GetOverallHealth() HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	componentStatuses := make(map[string]string)
	allHealthy := true
	anyDegraded := false

	for _, comp := range m.components {
		status := comp.Lifecycle.Health()
		componentStatuses[comp.Name] = status.Status

		if status.Status == "unhealthy" {
			allHealthy = false
		} else if status.Status == "degraded" {
			anyDegraded = true
		}
	}

	overallStatus := "healthy"
	if !allHealthy {
		overallStatus = "unhealthy"
	} else if anyDegraded {
		overallStatus = "degraded"
	}

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"components": componentStatuses,
			"total":      len(m.components),
		},
	}
}
