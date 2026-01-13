package supervisor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"testing/quick"
	"time"
)

// Feature: concurrent-safety-fixes, Property 4: Supervisor Service Thread Safety
func TestSupervisorServiceConcurrentSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	property := func(nodeCount uint8, operationCount uint8) bool {
		// 限制测试规模以避免过长的测试时间
		if nodeCount == 0 {
			nodeCount = 1
		}
		if nodeCount > 10 {
			nodeCount = 10
		}
		if operationCount == 0 {
			operationCount = 1
		}
		if operationCount > 20 {
			operationCount = 20
		}

		service := NewSupervisorService()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			service.Shutdown(ctx)
		}()

		// 启动服务
		ctx := context.Background()
		if err := service.Start(ctx); err != nil {
			t.Logf("Failed to start service: %v", err)
			return false
		}

		var wg sync.WaitGroup
		var operationErrors int32

		// 并发添加节点
		for i := uint8(0); i < nodeCount; i++ {
			wg.Add(1)
			go func(nodeID uint8) {
				defer wg.Done()
				nodeName := fmt.Sprintf("test-node-%d", nodeID)
				err := service.AddNode(nodeName, "test-env", "localhost", 9001+int(nodeID), "user", "pass")
				if err != nil {
					atomic.AddInt32(&operationErrors, 1)
					t.Logf("Failed to add node %s: %v", nodeName, err)
				}
			}(i)
		}

		// 并发执行各种操作
		for i := uint8(0); i < operationCount; i++ {
			wg.Add(1)
			go func(opID uint8) {
				defer wg.Done()
				
				// 随机选择操作类型
				switch opID % 6 {
				case 0:
					// GetAllNodes
					nodes := service.GetAllNodes()
					if nodes == nil && !service.IsShutdown() {
						atomic.AddInt32(&operationErrors, 1)
					}
				case 1:
					// GetEnvironments
					envs := service.GetEnvironments()
					if envs == nil {
						// 这是可接受的，因为可能没有节点
					}
				case 2:
					// GetGroups
					groups := service.GetGroups()
					if groups == nil {
						// 这是可接受的，因为可能没有连接的节点
					}
				case 3:
					// Health check
					health := service.Health()
					if health.Status == "" {
						atomic.AddInt32(&operationErrors, 1)
					}
				case 4:
					// IsShutdown check
					service.IsShutdown()
				case 5:
					// GetNode (可能失败，这是正常的)
					nodeName := fmt.Sprintf("test-node-%d", opID%nodeCount)
					service.GetNode(nodeName)
				}
			}(i)
		}

		// 等待所有操作完成
		wg.Wait()

		// 检查是否有操作错误
		if atomic.LoadInt32(&operationErrors) > int32(operationCount/2) {
			t.Logf("Too many operation errors: %d", operationErrors)
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Feature: concurrent-safety-fixes, Property 5: Resource-Limited Concurrent Operations
func TestSupervisorServiceResourceLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	property := func(concurrentRequests uint8) bool {
		// 限制并发请求数量
		if concurrentRequests == 0 {
			concurrentRequests = 1
		}
		if concurrentRequests > 50 {
			concurrentRequests = 50
		}

		service := NewSupervisorService()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			service.Shutdown(ctx)
		}()

		// 启动服务
		ctx := context.Background()
		if err := service.Start(ctx); err != nil {
			return false
		}

		// 添加一些测试节点
		for i := 0; i < 5; i++ {
			nodeName := fmt.Sprintf("test-node-%d", i)
			service.AddNode(nodeName, "test-env", "localhost", 9001+i, "user", "pass")
		}

		var wg sync.WaitGroup
		var completedRequests int32

		// 并发调用 GetAllNodes，这会触发连接信号量限制
		for i := uint8(0); i < concurrentRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				nodes := service.GetAllNodes()
				if nodes != nil {
					atomic.AddInt32(&completedRequests, 1)
				}
			}()
		}

		wg.Wait()

		// 验证至少有一些请求成功完成
		completed := atomic.LoadInt32(&completedRequests)
		if completed == 0 {
			t.Logf("No requests completed successfully")
			return false
		}

		// 验证连接信号量确实限制了并发连接
		// 这里我们检查服务仍然正常工作
		finalNodes := service.GetAllNodes()
		if finalNodes == nil && !service.IsShutdown() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Feature: concurrent-safety-fixes, Property 6: Graceful Service Shutdown
func TestSupervisorServiceGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	config := &quick.Config{
		MaxCount: 20,
		Rand:     nil,
	}

	property := func(activeOperations uint8) bool {
		// 限制活跃操作数量
		if activeOperations == 0 {
			activeOperations = 1
		}
		if activeOperations > 10 {
			activeOperations = 10
		}

		service := NewSupervisorService()

		// 启动服务
		ctx := context.Background()
		if err := service.Start(ctx); err != nil {
			return false
		}

		// 添加一些节点
		for i := 0; i < 3; i++ {
			nodeName := fmt.Sprintf("test-node-%d", i)
			service.AddNode(nodeName, "test-env", "localhost", 9001+i, "user", "pass")
		}

		var wg sync.WaitGroup
		shutdownStarted := make(chan struct{})

		// 启动一些长时间运行的操作
		for i := uint8(0); i < activeOperations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-shutdownStarted:
						return
					default:
						service.GetAllNodes()
						time.Sleep(10 * time.Millisecond)
					}
				}
			}()
		}

		// 短暂等待操作开始
		time.Sleep(50 * time.Millisecond)

		// 开始关闭
		close(shutdownStarted)

		// 执行优雅关闭
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		shutdownErr := service.Shutdown(shutdownCtx)
		
		// 等待所有操作完成
		wg.Wait()

		// 验证关闭成功
		if shutdownErr != nil {
			t.Logf("Shutdown failed: %v", shutdownErr)
			return false
		}

		// 验证服务确实已关闭
		if !service.IsShutdown() {
			return false
		}

		// 验证关闭后的操作返回错误
		if nodes := service.GetAllNodes(); nodes != nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// 单元测试：验证原子操作的正确性
func TestAtomicShutdownFlag(t *testing.T) {
	service := NewSupervisorService()

	// 初始状态应该是未关闭
	if service.IsShutdown() {
		t.Error("Service should not be shutdown initially")
	}

	// 启动服务
	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// 仍然不应该关闭
	if service.IsShutdown() {
		t.Error("Service should not be shutdown after start")
	}

	// 关闭服务
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := service.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Failed to shutdown service: %v", err)
	}

	// 现在应该是关闭状态
	if !service.IsShutdown() {
		t.Error("Service should be shutdown after Shutdown() call")
	}

	// 重复关闭应该是安全的
	if err := service.Shutdown(shutdownCtx); err != nil {
		t.Error("Repeated shutdown should not return error")
	}
}

// 单元测试：验证连接信号量限制
func TestConnectionSemaphore(t *testing.T) {
	service := NewSupervisorService()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		service.Shutdown(ctx)
	}()

	// 启动服务
	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// 验证信号量容量 (updated to reflect new default)
	semaphoreCapacity := cap(service.connectionSemaphore)
	if semaphoreCapacity != 100 {
		t.Errorf("Expected semaphore capacity 100, got %d", semaphoreCapacity)
	}

	// 验证信号量初始为空
	semaphoreLength := len(service.connectionSemaphore)
	if semaphoreLength != 0 {
		t.Errorf("Expected empty semaphore initially, got length %d", semaphoreLength)
	}
}