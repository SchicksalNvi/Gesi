package supervisor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"superview/internal/logger"
	"go.uber.org/zap"
)

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	// 单个操作超时
	SingleOperationTimeout time.Duration `json:"single_operation_timeout"`
	// 批量操作超时
	BatchOperationTimeout time.Duration `json:"batch_operation_timeout"`
	// 连接超时
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	// 健康检查超时
	HealthCheckTimeout time.Duration `json:"health_check_timeout"`
	// 重试间隔
	RetryInterval time.Duration `json:"retry_interval"`
	// 最大重试次数
	MaxRetries int `json:"max_retries"`
}

// GetDefaultTimeoutConfig 获取默认超时配置
func GetDefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		SingleOperationTimeout: 30 * time.Second,
		BatchOperationTimeout:  5 * time.Minute,
		ConnectionTimeout:      10 * time.Second,
		HealthCheckTimeout:     5 * time.Second,
		RetryInterval:          1 * time.Second,
		MaxRetries:             3,
	}
}

// CircuitBreaker 熔断器实现
type CircuitBreaker struct {
	mu           sync.RWMutex
	name         string
	maxFailures  int
	resetTimeout time.Duration
	
	failures     int64
	lastFailTime time.Time
	state        CircuitState
}

// CircuitState 熔断器状态
type CircuitState int

const (
	CircuitClosed CircuitState = iota // 关闭状态 - 正常工作
	CircuitOpen                       // 开启状态 - 熔断中
	CircuitHalfOpen                   // 半开状态 - 尝试恢复
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute 执行操作，带熔断保护
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker %s is OPEN", cb.name)
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// canExecute 检查是否可以执行操作
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// 检查是否可以转为半开状态
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == CircuitOpen && time.Since(cb.lastFailTime) > cb.resetTimeout {
				cb.state = CircuitHalfOpen
				logger.Info("Circuit breaker transitioning to HALF_OPEN",
					zap.String("name", cb.name))
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return cb.state == CircuitHalfOpen
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult 记录操作结果
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.state == CircuitHalfOpen {
			// 半开状态下失败，直接转为开启状态
			cb.state = CircuitOpen
			logger.Warn("Circuit breaker opened due to failure in HALF_OPEN state",
				zap.String("name", cb.name),
				zap.Error(err))
		} else if cb.failures >= int64(cb.maxFailures) {
			// 失败次数达到阈值，转为开启状态
			cb.state = CircuitOpen
			logger.Warn("Circuit breaker opened due to max failures",
				zap.String("name", cb.name),
				zap.Int64("failures", cb.failures),
				zap.Int("max_failures", cb.maxFailures))
		}
	} else {
		// 成功执行
		if cb.state == CircuitHalfOpen {
			// 半开状态下成功，转为关闭状态
			cb.state = CircuitClosed
			cb.failures = 0
			logger.Info("Circuit breaker closed after successful execution",
				zap.String("name", cb.name))
		}
	}
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures 获取失败次数
func (cb *CircuitBreaker) GetFailures() int64 {
	return atomic.LoadInt64(&cb.failures)
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failures = 0
	cb.lastFailTime = time.Time{}
	logger.Info("Circuit breaker reset", zap.String("name", cb.name))
}

// TimeoutManager 超时管理器
type TimeoutManager struct {
	config          *TimeoutConfig
	circuitBreakers map[string]*CircuitBreaker
	mu              sync.RWMutex
}

// NewTimeoutManager 创建超时管理器
func NewTimeoutManager(config *TimeoutConfig) *TimeoutManager {
	if config == nil {
		config = GetDefaultTimeoutConfig()
	}
	
	return &TimeoutManager{
		config:          config,
		circuitBreakers: make(map[string]*CircuitBreaker),
	}
}

// GetCircuitBreaker 获取或创建熔断器
func (tm *TimeoutManager) GetCircuitBreaker(name string) *CircuitBreaker {
	tm.mu.RLock()
	cb, exists := tm.circuitBreakers[name]
	tm.mu.RUnlock()

	if exists {
		return cb
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 双重检查
	if cb, exists := tm.circuitBreakers[name]; exists {
		return cb
	}

	// 创建新的熔断器
	cb = NewCircuitBreaker(name, tm.config.MaxRetries, tm.config.RetryInterval*time.Duration(tm.config.MaxRetries))
	tm.circuitBreakers[name] = cb
	
	logger.Info("Created circuit breaker",
		zap.String("name", name),
		zap.Int("max_failures", tm.config.MaxRetries))
	
	return cb
}

// ExecuteWithTimeout 执行带超时的操作
func (tm *TimeoutManager) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	if timeout <= 0 {
		timeout = tm.config.SingleOperationTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("operation panicked: %v", r)
			}
		}()
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v: %w", timeout, ctx.Err())
	}
}

// ExecuteWithRetry 执行带重试的操作
func (tm *TimeoutManager) ExecuteWithRetry(ctx context.Context, name string, fn func(context.Context) error) error {
	cb := tm.GetCircuitBreaker(name)
	
	return cb.Execute(func() error {
		var lastErr error
		for attempt := 0; attempt <= tm.config.MaxRetries; attempt++ {
			if attempt > 0 {
				// 等待重试间隔
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(tm.config.RetryInterval):
				}
				
				logger.Debug("Retrying operation",
					zap.String("name", name),
					zap.Int("attempt", attempt),
					zap.Int("max_retries", tm.config.MaxRetries))
			}

			err := tm.ExecuteWithTimeout(ctx, tm.config.SingleOperationTimeout, fn)
			if err == nil {
				if attempt > 0 {
					logger.Info("Operation succeeded after retry",
						zap.String("name", name),
						zap.Int("attempts", attempt+1))
				}
				return nil
			}

			lastErr = err
			
			// 检查是否是上下文取消错误，如果是则不重试
			if ctx.Err() != nil {
				break
			}
		}

		logger.Error("Operation failed after all retries",
			zap.String("name", name),
			zap.Int("max_retries", tm.config.MaxRetries),
			zap.Error(lastErr))
		
		return fmt.Errorf("operation %s failed after %d retries: %w", name, tm.config.MaxRetries, lastErr)
	})
}

// ExecuteBatchWithTimeout 执行批量操作，每个操作都有独立超时
func (tm *TimeoutManager) ExecuteBatchWithTimeout(ctx context.Context, operations []BatchOperation) []BatchResult {
	// 为批量操作创建带超时的上下文
	batchCtx, cancel := context.WithTimeout(ctx, tm.config.BatchOperationTimeout)
	defer cancel()

	results := make([]BatchResult, len(operations))
	var wg sync.WaitGroup
	
	// 限制并发数
	semaphore := make(chan struct{}, 10) // 最多10个并发操作

	for i, op := range operations {
		wg.Add(1)
		go func(index int, operation BatchOperation) {
			defer wg.Done()
			
			// 获取信号量
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-batchCtx.Done():
				results[index] = BatchResult{
					Index: index,
					Error: batchCtx.Err(),
				}
				return
			}

			// 为单个操作创建超时上下文
			opCtx, opCancel := context.WithTimeout(batchCtx, tm.config.SingleOperationTimeout)
			defer opCancel()

			// 执行操作
			err := operation.Execute(opCtx)
			results[index] = BatchResult{
				Index: index,
				Error: err,
			}

			if err != nil {
				logger.Debug("Batch operation failed",
					zap.Int("index", index),
					zap.String("name", operation.Name),
					zap.Error(err))
			}
		}(i, op)
	}

	wg.Wait()
	return results
}

// BatchOperation 批量操作接口
type BatchOperation struct {
	Name    string
	Execute func(context.Context) error
}

// BatchResult 批量操作结果
type BatchResult struct {
	Index int
	Error error
}

// GetCircuitBreakerStats 获取所有熔断器统计信息
func (tm *TimeoutManager) GetCircuitBreakerStats() map[string]CircuitBreakerStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for name, cb := range tm.circuitBreakers {
		stats[name] = CircuitBreakerStats{
			Name:     name,
			State:    cb.GetState().String(),
			Failures: cb.GetFailures(),
		}
	}
	return stats
}

// CircuitBreakerStats 熔断器统计信息
type CircuitBreakerStats struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Failures int64  `json:"failures"`
}

// Cleanup 清理资源
func (tm *TimeoutManager) Cleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	for name := range tm.circuitBreakers {
		delete(tm.circuitBreakers, name)
	}
	
	logger.Info("Timeout manager cleaned up")
}