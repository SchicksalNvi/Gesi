package utils

import (
	"context"
	"sync"
	"time"

	"superview/internal/logger"
	"go.uber.org/zap"
)

// Task 任务接口
type Task interface {
	Execute(ctx context.Context) error
	ID() string
}

// WorkerPool 工作池
type WorkerPool struct {
	workers    int
	taskQueue  chan Task
	results    chan TaskResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	activeJobs int
	totalJobs  int64
	errors     int64
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID    string
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// WorkerPoolConfig 工作池配置
type WorkerPoolConfig struct {
	Workers       int           // 工作协程数量
	QueueSize     int           // 任务队列大小
	ResultBuffer  int           // 结果缓冲区大小
	TaskTimeout   time.Duration // 任务超时时间
}

// GetDefaultWorkerPoolConfig 获取默认工作池配置
func GetDefaultWorkerPoolConfig() *WorkerPoolConfig {
	return &WorkerPoolConfig{
		Workers:      10,
		QueueSize:    100,
		ResultBuffer: 100,
		TaskTimeout:  30 * time.Second,
	}
}

// NewWorkerPool 创建工作池
func NewWorkerPool(config *WorkerPoolConfig) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &WorkerPool{
		workers:   config.Workers,
		taskQueue: make(chan Task, config.QueueSize),
		results:   make(chan TaskResult, config.ResultBuffer),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 启动工作协程
	for i := 0; i < config.Workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i, config.TaskTimeout)
	}

	logger.Info("Worker pool started",
		zap.Int("workers", config.Workers),
		zap.Int("queue_size", config.QueueSize))

	return pool
}

// worker 工作协程
func (wp *WorkerPool) worker(id int, timeout time.Duration) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			logger.Debug("Worker stopping", zap.Int("worker_id", id))
			return

		case task, ok := <-wp.taskQueue:
			if !ok {
				return
			}

			wp.mu.Lock()
			wp.activeJobs++
			wp.totalJobs++
			wp.mu.Unlock()

			// 执行任务
			startTime := time.Now()
			taskCtx, cancel := context.WithTimeout(wp.ctx, timeout)
			
			err := task.Execute(taskCtx)
			cancel()
			
			duration := time.Since(startTime)

			// 记录结果
			result := TaskResult{
				TaskID:    task.ID(),
				Error:     err,
				Duration:  duration,
				Timestamp: time.Now(),
			}

			if err != nil {
				wp.mu.Lock()
				wp.errors++
				wp.mu.Unlock()
				
				logger.Error("Task execution failed",
					zap.String("task_id", task.ID()),
					zap.Error(err),
					zap.Duration("duration", duration))
			} else {
				logger.Debug("Task completed",
					zap.String("task_id", task.ID()),
					zap.Duration("duration", duration))
			}

			// 发送结果
			select {
			case wp.results <- result:
			default:
				logger.Warn("Result buffer full, dropping result",
					zap.String("task_id", task.ID()))
			}

			wp.mu.Lock()
			wp.activeJobs--
			wp.mu.Unlock()
		}
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task Task) error {
	// 检查工作池是否已停止
	select {
	case <-wp.ctx.Done():
		return context.Canceled
	default:
	}
	
	// 尝试提交任务，使用非阻塞发送避免 panic
	select {
	case <-wp.ctx.Done():
		return context.Canceled
	case wp.taskQueue <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

// SubmitWithTimeout 带超时的任务提交
func (wp *WorkerPool) SubmitWithTimeout(task Task, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-wp.ctx.Done():
		return context.Canceled
	case wp.taskQueue <- task:
		return nil
	}
}

// Results 获取结果通道
func (wp *WorkerPool) Results() <-chan TaskResult {
	return wp.results
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	logger.Info("Stopping worker pool")
	
	// 取消上下文
	wp.cancel()
	
	// 关闭任务队列
	close(wp.taskQueue)
	
	// 等待所有工作协程完成
	wp.wg.Wait()
	
	// 关闭结果通道
	close(wp.results)
	
	logger.Info("Worker pool stopped")
}

// StopWithTimeout 带超时的停止
func (wp *WorkerPool) StopWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})
	
	go func() {
		wp.Stop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return ErrStopTimeout
	}
}

// Stats 获取工作池统计信息
func (wp *WorkerPool) Stats() WorkerPoolStats {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return WorkerPoolStats{
		Workers:      wp.workers,
		ActiveJobs:   wp.activeJobs,
		QueuedJobs:   len(wp.taskQueue),
		TotalJobs:    wp.totalJobs,
		Errors:       wp.errors,
		QueueSize:    cap(wp.taskQueue),
		ResultBuffer: cap(wp.results),
	}
}

// WorkerPoolStats 工作池统计信息
type WorkerPoolStats struct {
	Workers      int   `json:"workers"`
	ActiveJobs   int   `json:"active_jobs"`
	QueuedJobs   int   `json:"queued_jobs"`
	TotalJobs    int64 `json:"total_jobs"`
	Errors       int64 `json:"errors"`
	QueueSize    int   `json:"queue_size"`
	ResultBuffer int   `json:"result_buffer"`
}

// 错误定义
var (
	ErrQueueFull   = &WorkerPoolError{Message: "task queue is full"}
	ErrStopTimeout = &WorkerPoolError{Message: "stop timeout exceeded"}
)

// WorkerPoolError 工作池错误
type WorkerPoolError struct {
	Message string
}

func (e *WorkerPoolError) Error() string {
	return e.Message
}
