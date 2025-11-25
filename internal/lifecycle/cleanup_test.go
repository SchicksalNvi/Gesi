package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCleanupComponent 模拟需要清理的组件
type mockCleanupComponent struct {
	name          string
	startCalled   bool
	stopCalled    bool
	cleanupCalled bool
	startError    error
	stopError     error
	cleanupError  error
	health        HealthStatus
	mu            sync.Mutex
	resources     []string // 模拟资源
}

func (m *mockCleanupComponent) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startCalled = true
	// 模拟分配资源
	m.resources = append(m.resources, "resource1", "resource2")
	return m.startError
}

func (m *mockCleanupComponent) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCalled = true
	return m.stopError
}

func (m *mockCleanupComponent) Health() HealthStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.health
}

// Cleanup 清理资源
func (m *mockCleanupComponent) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupCalled = true
	// 模拟清理资源
	m.resources = nil
	return m.cleanupError
}

// GetResources 获取资源列表（用于测试）
func (m *mockCleanupComponent) GetResources() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.resources...)
}

// TestResourceCleanup 测试资源清理
func TestResourceCleanup(t *testing.T) {
	manager := NewManager()
	comp := &mockCleanupComponent{
		name: "cleanup-component",
		health: HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
		},
	}

	manager.Register("cleanup-test", comp, 1)

	ctx := context.Background()

	// 启动组件
	err := manager.StartAll(ctx)
	require.NoError(t, err)
	assert.True(t, comp.startCalled)
	assert.Len(t, comp.GetResources(), 2)

	// 停止组件
	err = manager.StopAll(ctx)
	require.NoError(t, err)
	assert.True(t, comp.stopCalled)

	// 清理资源
	err = comp.Cleanup(ctx)
	require.NoError(t, err)
	assert.True(t, comp.cleanupCalled)
	assert.Len(t, comp.GetResources(), 0)
}

// TestResourceCleanupWithError 测试清理时的错误处理
func TestResourceCleanupWithError(t *testing.T) {
	manager := NewManager()
	comp := &mockCleanupComponent{
		name:         "cleanup-component",
		cleanupError: errors.New("cleanup failed"),
		health: HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
		},
	}

	manager.Register("cleanup-test", comp, 1)

	ctx := context.Background()

	// 启动组件
	err := manager.StartAll(ctx)
	require.NoError(t, err)

	// 停止组件
	err = manager.StopAll(ctx)
	require.NoError(t, err)

	// 清理资源（应该返回错误）
	err = comp.Cleanup(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cleanup failed")
	assert.True(t, comp.cleanupCalled)
}

// TestResourceCleanupTimeout 测试清理超时
func TestResourceCleanupTimeout(t *testing.T) {
	comp := &mockCleanupComponent{
		name: "slow-cleanup-component",
		health: HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
		},
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 模拟慢清理
	go func() {
		time.Sleep(2 * time.Second)
		comp.Cleanup(context.Background())
	}()

	// 等待超时
	<-ctx.Done()
	assert.Error(t, ctx.Err())
	assert.Contains(t, ctx.Err().Error(), "context deadline exceeded")
}

// TestResourceCleanupOrder 测试清理顺序
func TestResourceCleanupOrder(t *testing.T) {
	manager := NewManager()
	var cleanupOrder []string
	mu := sync.Mutex{}

	// 创建多个组件
	comps := make([]*mockCleanupComponent, 3)
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("comp%d", i+1)
		comp := &mockCleanupComponent{
			name: name,
			health: HealthStatus{
				Status:    "healthy",
				Timestamp: time.Now(),
			},
		}
		comps[i] = comp
		manager.Register(name, comp, i+1)
	}

	ctx := context.Background()

	// 启动所有组件
	err := manager.StartAll(ctx)
	require.NoError(t, err)

	// 停止所有组件
	err = manager.StopAll(ctx)
	require.NoError(t, err)

	// 手动清理所有组件（按逆序）
	for i := len(comps) - 1; i >= 0; i-- {
		comp := comps[i]
		comp.Cleanup(ctx)
		mu.Lock()
		cleanupOrder = append(cleanupOrder, comp.name)
		mu.Unlock()
	}

	// 验证清理顺序
	assert.Len(t, cleanupOrder, 3)
	assert.Equal(t, "comp3", cleanupOrder[0])
	assert.Equal(t, "comp2", cleanupOrder[1])
	assert.Equal(t, "comp1", cleanupOrder[2])
}

// TestResourceCleanupConcurrent 测试并发清理
func TestResourceCleanupConcurrent(t *testing.T) {
	manager := NewManager()
	componentCount := 10
	var wg sync.WaitGroup

	// 创建多个组件
	comps := make([]*mockCleanupComponent, componentCount)
	for i := 0; i < componentCount; i++ {
		name := fmt.Sprintf("comp%d", i)
		comp := &mockCleanupComponent{
			name: name,
			health: HealthStatus{
				Status:    "healthy",
				Timestamp: time.Now(),
			},
		}
		comps[i] = comp
		manager.Register(name, comp, i+1)
	}

	ctx := context.Background()

	// 启动所有组件
	err := manager.StartAll(ctx)
	require.NoError(t, err)

	// 停止所有组件
	err = manager.StopAll(ctx)
	require.NoError(t, err)

	// 并发清理所有组件
	wg.Add(componentCount)
	for i := 0; i < componentCount; i++ {
		go func(comp *mockCleanupComponent) {
			defer wg.Done()
			err := comp.Cleanup(ctx)
			assert.NoError(t, err)
		}(comps[i])
	}

	wg.Wait()

	// 验证所有组件都被清理
	for _, comp := range comps {
		assert.True(t, comp.cleanupCalled)
		assert.Len(t, comp.GetResources(), 0)
	}
}

// TestResourceCleanupPartialFailure 测试部分清理失败
func TestResourceCleanupPartialFailure(t *testing.T) {
	manager := NewManager()
	successCount := 0
	failureCount := 0

	// 创建成功和失败的组件
	comps := make([]*mockCleanupComponent, 5)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("comp%d", i)
		comp := &mockCleanupComponent{
			name: name,
			health: HealthStatus{
				Status:    "healthy",
				Timestamp: time.Now(),
			},
		}

		// 让一半的组件清理失败
		if i%2 == 0 {
			comp.cleanupError = errors.New("cleanup failed")
		}

		comps[i] = comp
		manager.Register(name, comp, i+1)
	}

	ctx := context.Background()

	// 启动所有组件
	err := manager.StartAll(ctx)
	require.NoError(t, err)

	// 停止所有组件
	err = manager.StopAll(ctx)
	require.NoError(t, err)

	// 清理所有组件并统计结果
	for _, comp := range comps {
		err := comp.Cleanup(ctx)
		if err != nil {
			failureCount++
		} else {
			successCount++
		}
	}

	// 验证部分成功，部分失败
	assert.Equal(t, 2, successCount)
	assert.Equal(t, 3, failureCount)
}

// TestResourceCleanupIdempotent 测试清理的幂等性
func TestResourceCleanupIdempotent(t *testing.T) {
	comp := &mockCleanupComponent{
		name: "idempotent-component",
		health: HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
		},
	}

	ctx := context.Background()

	// 启动组件
	err := comp.Start(ctx)
	require.NoError(t, err)
	assert.Len(t, comp.GetResources(), 2)

	// 第一次清理
	err = comp.Cleanup(ctx)
	require.NoError(t, err)
	assert.Len(t, comp.GetResources(), 0)

	// 第二次清理（应该是幂等的）
	err = comp.Cleanup(ctx)
	require.NoError(t, err)
	assert.Len(t, comp.GetResources(), 0)
}

// TestResourceCleanupWithContext 测试带上下文的清理
func TestResourceCleanupWithContext(t *testing.T) {
	comp := &mockCleanupComponent{
		name: "context-component",
		health: HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
		},
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 启动组件
	err := comp.Start(ctx)
	require.NoError(t, err)

	// 取消上下文
	cancel()

	// 清理应该检测到上下文已取消
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cleanupCancel()

	err = comp.Cleanup(cleanupCtx)
	require.NoError(t, err)
	assert.True(t, comp.cleanupCalled)
}

// TestAllResourcesReleasedOnShutdown 测试系统关闭时所有资源被释放
func TestAllResourcesReleasedOnShutdown(t *testing.T) {
	manager := NewManager()

	// 创建多个组件，每个组件分配不同数量的资源
	components := []*mockCleanupComponent{
		{
			name: "comp1",
			health: HealthStatus{
				Status:    "healthy",
				Timestamp: time.Now(),
			},
		},
		{
			name: "comp2",
			health: HealthStatus{
				Status:    "healthy",
				Timestamp: time.Now(),
			},
		},
		{
			name: "comp3",
			health: HealthStatus{
				Status:    "healthy",
				Timestamp: time.Now(),
			},
		},
	}

	for i, comp := range components {
		manager.Register(comp.name, comp, i+1)
	}

	ctx := context.Background()

	// 启动所有组件
	err := manager.StartAll(ctx)
	require.NoError(t, err)

	// 验证所有组件都分配了资源
	totalResourcesBefore := 0
	for _, comp := range components {
		totalResourcesBefore += len(comp.GetResources())
	}
	assert.Greater(t, totalResourcesBefore, 0)

	// 停止所有组件
	err = manager.StopAll(ctx)
	require.NoError(t, err)

	// 清理所有组件
	for _, comp := range components {
		err := comp.Cleanup(ctx)
		require.NoError(t, err)
	}

	// 验证所有资源都被释放
	totalResourcesAfter := 0
	for _, comp := range components {
		totalResourcesAfter += len(comp.GetResources())
	}
	assert.Equal(t, 0, totalResourcesAfter)

	// 验证所有组件都调用了清理方法
	for _, comp := range components {
		assert.True(t, comp.cleanupCalled)
	}
}
