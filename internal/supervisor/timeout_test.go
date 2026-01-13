package supervisor

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Feature: concurrent-safety-fixes, Property 13: Hierarchical Operation Timeouts
// For any batch operation, individual operations should have their own timeouts in addition to batch-level timeouts
func TestHierarchicalOperationTimeouts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping hierarchical timeout test in short mode")
	}

	config := &TimeoutConfig{
		SingleOperationTimeout: 100 * time.Millisecond,
		BatchOperationTimeout:  500 * time.Millisecond,
		ConnectionTimeout:      50 * time.Millisecond,
		HealthCheckTimeout:     30 * time.Millisecond,
		RetryInterval:          10 * time.Millisecond,
		MaxRetries:             2,
	}

	tm := NewTimeoutManager(config)
	defer tm.Cleanup()

	t.Run("SingleOperationTimeout", func(t *testing.T) {
		testSingleOperationTimeout(t, tm, config)
	})

	t.Run("BatchOperationTimeout", func(t *testing.T) {
		testBatchOperationTimeout(t, tm, config)
	})

	t.Run("HierarchicalTimeouts", func(t *testing.T) {
		testHierarchicalTimeouts(t, tm, config)
	})

	t.Run("TimeoutResourceCleanup", func(t *testing.T) {
		testTimeoutResourceCleanup(t, tm, config)
	})
}

// Feature: concurrent-safety-fixes, Property 14: Circuit Breaker Pattern
// For any node that fails repeatedly, the system should implement circuit breaker pattern to prevent cascading failures
func TestCircuitBreakerPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping circuit breaker test in short mode")
	}

	config := &TimeoutConfig{
		SingleOperationTimeout: 50 * time.Millisecond,
		BatchOperationTimeout:  200 * time.Millisecond,
		RetryInterval:          10 * time.Millisecond,
		MaxRetries:             3,
	}

	tm := NewTimeoutManager(config)
	defer tm.Cleanup()

	t.Run("CircuitBreakerStates", func(t *testing.T) {
		testCircuitBreakerStates(t, tm)
	})

	t.Run("CircuitBreakerRecovery", func(t *testing.T) {
		testCircuitBreakerRecovery(t, tm)
	})

	t.Run("CascadingFailurePrevention", func(t *testing.T) {
		testCascadingFailurePrevention(t, tm)
	})
}

// testSingleOperationTimeout 测试单个操作超时
func testSingleOperationTimeout(t *testing.T, tm *TimeoutManager, config *TimeoutConfig) {
	ctx := context.Background()
	
	// 测试正常操作（不超时）
	err := tm.ExecuteWithTimeout(ctx, config.SingleOperationTimeout, func(ctx context.Context) error {
		time.Sleep(config.SingleOperationTimeout / 2) // 一半时间
		return nil
	})
	if err != nil {
		t.Errorf("Normal operation should not timeout: %v", err)
	}

	// 测试超时操作
	start := time.Now()
	err = tm.ExecuteWithTimeout(ctx, config.SingleOperationTimeout, func(ctx context.Context) error {
		time.Sleep(config.SingleOperationTimeout * 2) // 两倍时间
		return nil
	})
	duration := time.Since(start)

	if err == nil {
		t.Error("Operation should have timed out")
	}

	// 验证超时时间合理
	expectedTimeout := config.SingleOperationTimeout
	if duration < expectedTimeout || duration > expectedTimeout+50*time.Millisecond {
		t.Errorf("Timeout duration unexpected: got %v, expected ~%v", duration, expectedTimeout)
	}
}

// testBatchOperationTimeout 测试批量操作超时
func testBatchOperationTimeout(t *testing.T, tm *TimeoutManager, config *TimeoutConfig) {
	ctx := context.Background()

	// 创建多个操作，其中一些会超时
	operations := []BatchOperation{
		{
			Name: "fast_op_1",
			Execute: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		},
		{
			Name: "slow_op_1",
			Execute: func(ctx context.Context) error {
				// 检查上下文取消
				select {
				case <-time.After(config.SingleOperationTimeout * 2): // 超过单个操作超时
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		},
		{
			Name: "fast_op_2",
			Execute: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		},
	}

	start := time.Now()
	results := tm.ExecuteBatchWithTimeout(ctx, operations)
	duration := time.Since(start)

	// 验证结果
	if len(results) != len(operations) {
		t.Errorf("Expected %d results, got %d", len(operations), len(results))
	}

	// 快速操作应该成功
	if results[0].Error != nil {
		t.Errorf("Fast operation should succeed: %v", results[0].Error)
	}
	if results[2].Error != nil {
		t.Errorf("Fast operation should succeed: %v", results[2].Error)
	}

	// 慢操作应该超时
	if results[1].Error == nil {
		t.Error("Slow operation should timeout")
	} else {
		t.Logf("Slow operation correctly timed out: %v", results[1].Error)
	}

	// 批量操作应该在合理时间内完成（不等待所有慢操作）
	if duration > config.BatchOperationTimeout+100*time.Millisecond {
		t.Errorf("Batch operation took too long: %v", duration)
	}
}

// testHierarchicalTimeouts 测试分层超时
func testHierarchicalTimeouts(t *testing.T, tm *TimeoutManager, config *TimeoutConfig) {
	ctx := context.Background()

	// 创建一个批量操作，包含不同速度的操作
	numOperations := 5
	operations := make([]BatchOperation, numOperations)
	
	for i := 0; i < numOperations; i++ {
		operations[i] = BatchOperation{
			Name: fmt.Sprintf("hierarchical_op_%d", i),
			Execute: func(ctx context.Context) error {
				// 随机延迟，有些会超过单个操作超时，但不会超过批量超时
				delay := time.Duration(rand.Intn(int(config.SingleOperationTimeout.Nanoseconds()*3))) * time.Nanosecond
				
				select {
				case <-time.After(delay):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		}
	}

	start := time.Now()
	results := tm.ExecuteBatchWithTimeout(ctx, operations)
	duration := time.Since(start)

	// 验证层次化超时行为
	successCount := 0
	timeoutCount := 0
	
	for i, result := range results {
		if result.Error == nil {
			successCount++
		} else if errors.Is(result.Error, context.DeadlineExceeded) {
			timeoutCount++
		} else {
			t.Errorf("Operation %d failed with unexpected error: %v", i, result.Error)
		}
	}

	t.Logf("Hierarchical timeout results: %d success, %d timeout, duration: %v", 
		successCount, timeoutCount, duration)

	// 应该有一些操作成功，一些超时
	if successCount == 0 {
		t.Error("At least some operations should succeed")
	}

	// 批量操作不应该超过批量超时
	if duration > config.BatchOperationTimeout+100*time.Millisecond {
		t.Errorf("Batch operation exceeded batch timeout: %v > %v", duration, config.BatchOperationTimeout)
	}
}

// testTimeoutResourceCleanup 测试超时后的资源清理
func testTimeoutResourceCleanup(t *testing.T, tm *TimeoutManager, config *TimeoutConfig) {
	ctx := context.Background()
	
	// 跟踪资源使用
	var activeOperations int64
	var maxActiveOperations int64

	// 创建多个会超时的操作
	numOperations := 10
	var wg sync.WaitGroup
	
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(opIndex int) {
			defer wg.Done()
			
			err := tm.ExecuteWithTimeout(ctx, config.SingleOperationTimeout, func(ctx context.Context) error {
				// 增加活跃操作计数
				current := atomic.AddInt64(&activeOperations, 1)
				defer atomic.AddInt64(&activeOperations, -1)
				
				// 更新最大活跃操作数
				for {
					max := atomic.LoadInt64(&maxActiveOperations)
					if current <= max || atomic.CompareAndSwapInt64(&maxActiveOperations, max, current) {
						break
					}
				}
				
				// 模拟长时间操作，但要检查上下文取消
				select {
				case <-time.After(config.SingleOperationTimeout * 2):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
			
			// 所有操作都应该超时
			if err == nil {
				t.Errorf("Operation %d should have timed out", opIndex)
			}
		}(i)
	}

	wg.Wait()
	
	// 给一点时间让goroutine完全清理
	time.Sleep(50 * time.Millisecond)

	// 验证资源清理
	finalActiveOperations := atomic.LoadInt64(&activeOperations)
	if finalActiveOperations != 0 {
		t.Errorf("Resource leak detected: %d operations still active", finalActiveOperations)
	}

	// 验证并发控制
	if maxActiveOperations > int64(numOperations) {
		t.Errorf("Too many concurrent operations: %d > %d", maxActiveOperations, numOperations)
	}

	t.Logf("Resource cleanup test: max concurrent operations: %d", maxActiveOperations)
}

// testCircuitBreakerStates 测试熔断器状态转换
func testCircuitBreakerStates(t *testing.T, tm *TimeoutManager) {
	cb := tm.GetCircuitBreaker("test_circuit")
	
	// 初始状态应该是关闭
	if cb.GetState() != CircuitClosed {
		t.Errorf("Initial state should be CLOSED, got %v", cb.GetState())
	}

	// 模拟失败操作直到熔断器开启
	maxFailures := tm.config.MaxRetries
	for i := 0; i < maxFailures; i++ {
		err := cb.Execute(func() error {
			return fmt.Errorf("simulated failure %d", i)
		})
		if err == nil {
			t.Errorf("Operation %d should have failed", i)
		}
	}

	// 现在熔断器应该是开启状态
	if cb.GetState() != CircuitOpen {
		t.Errorf("Circuit should be OPEN after %d failures, got %v", maxFailures, cb.GetState())
	}

	// 在开启状态下，操作应该立即失败
	err := cb.Execute(func() error {
		return nil // 这个函数不会被调用
	})
	if err == nil {
		t.Error("Operation should fail immediately when circuit is OPEN")
	}

	// 等待重置时间后，应该转为半开状态
	resetTimeout := tm.config.RetryInterval * time.Duration(tm.config.MaxRetries)
	time.Sleep(resetTimeout + 10*time.Millisecond)

	// 尝试执行成功操作
	err = cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Operation should succeed in HALF_OPEN state: %v", err)
	}

	// 成功后应该转为关闭状态
	if cb.GetState() != CircuitClosed {
		t.Errorf("Circuit should be CLOSED after successful operation, got %v", cb.GetState())
	}
}

// testCircuitBreakerRecovery 测试熔断器恢复
func testCircuitBreakerRecovery(t *testing.T, tm *TimeoutManager) {
	cb := tm.GetCircuitBreaker("recovery_test")
	
	// 触发熔断
	for i := 0; i < tm.config.MaxRetries; i++ {
		cb.Execute(func() error {
			return fmt.Errorf("failure %d", i)
		})
	}

	if cb.GetState() != CircuitOpen {
		t.Error("Circuit should be OPEN")
	}

	// 记录失败次数
	initialFailures := cb.GetFailures()

	// 等待重置时间
	resetTimeout := tm.config.RetryInterval * time.Duration(tm.config.MaxRetries)
	time.Sleep(resetTimeout + 10*time.Millisecond)

	// 执行成功操作进行恢复
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Recovery operation should succeed: %v", err)
	}

	// 验证恢复
	if cb.GetState() != CircuitClosed {
		t.Errorf("Circuit should be CLOSED after recovery, got %v", cb.GetState())
	}

	// 失败计数应该被重置
	if cb.GetFailures() >= initialFailures {
		t.Errorf("Failure count should be reset after recovery: %d >= %d", cb.GetFailures(), initialFailures)
	}
}

// testCascadingFailurePrevention 测试级联失败预防
func testCascadingFailurePrevention(t *testing.T, tm *TimeoutManager) {
	// 创建新的超时管理器以避免之前测试的影响
	cleanTM := NewTimeoutManager(tm.config)
	defer cleanTM.Cleanup()
	
	// 创建多个相关的熔断器
	services := []string{"service_a", "service_b", "service_c"}
	
	// 模拟服务A失败
	cbA := cleanTM.GetCircuitBreaker("service_a")
	for i := 0; i < cleanTM.config.MaxRetries; i++ {
		cbA.Execute(func() error {
			return fmt.Errorf("service A failure")
		})
	}

	if cbA.GetState() != CircuitOpen {
		t.Error("Service A circuit should be OPEN")
	}

	// 其他服务的熔断器应该仍然正常工作
	cbB := cleanTM.GetCircuitBreaker("service_b")
	cbC := cleanTM.GetCircuitBreaker("service_c")

	if cbB.GetState() != CircuitClosed {
		t.Error("Service B circuit should remain CLOSED")
	}
	if cbC.GetState() != CircuitClosed {
		t.Error("Service C circuit should remain CLOSED")
	}

	// 验证其他服务仍然可以正常工作
	err := cbB.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Service B should still work: %v", err)
	}

	err = cbC.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Service C should still work: %v", err)
	}

	// 验证服务A被隔离
	err = cbA.Execute(func() error {
		return nil // 这不会被执行
	})
	if err == nil {
		t.Error("Service A should be isolated and fail fast")
	}

	// 获取统计信息
	stats := cleanTM.GetCircuitBreakerStats()
	if len(stats) != len(services) {
		t.Errorf("Expected %d circuit breakers, got %d", len(services), len(stats))
	}

	// 验证统计信息
	for _, service := range services {
		stat, exists := stats[service]
		if !exists {
			t.Errorf("Missing stats for service %s", service)
			continue
		}

		if service == "service_a" {
			if stat.State != "OPEN" {
				t.Errorf("Service A should be OPEN, got %s", stat.State)
			}
			if stat.Failures == 0 {
				t.Error("Service A should have failure count > 0")
			}
		} else {
			if stat.State != "CLOSED" {
				t.Errorf("Service %s should be CLOSED, got %s", service, stat.State)
			}
		}
	}

	t.Logf("Circuit breaker stats: %+v", stats)
}