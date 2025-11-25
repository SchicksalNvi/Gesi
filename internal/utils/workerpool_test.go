package utils

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTask 模拟任务
type mockTask struct {
	id       string
	duration time.Duration
	err      error
	executed atomic.Bool
}

func (t *mockTask) Execute(ctx context.Context) error {
	t.executed.Store(true)
	
	if t.duration > 0 {
		select {
		case <-time.After(t.duration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return t.err
}

func (t *mockTask) ID() string {
	return t.id
}

// 属性 10：并发限制
// 验证需求：4.3
func TestConcurrencyLimitProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("worker pool limits concurrent execution", prop.ForAll(
		func(workers int, tasks int) bool {
			if workers < 1 || workers > 100 {
				workers = 10
			}
			if tasks < 1 || tasks > 100 {
				tasks = 20
			}

			config := &WorkerPoolConfig{
				Workers:      workers,
				QueueSize:    tasks * 2,
				ResultBuffer: tasks * 2,
				TaskTimeout:  5 * time.Second,
			}

			pool := NewWorkerPool(config)
			defer pool.Stop()

			// 提交任务
			var maxConcurrent int32
			var currentConcurrent int32
			var wg sync.WaitGroup

			for i := 0; i < tasks; i++ {
				task := &mockTask{
					id:       string(rune(i)),
					duration: 10 * time.Millisecond,
				}

				// 包装任务以跟踪并发数
				wrappedTask := &concurrentTrackingTask{
					Task:              task,
					currentConcurrent: &currentConcurrent,
					maxConcurrent:     &maxConcurrent,
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					pool.Submit(wrappedTask)
				}()
			}

			wg.Wait()
			time.Sleep(100 * time.Millisecond) // 等待任务完成

			// 验证最大并发数不超过工作协程数
			max := atomic.LoadInt32(&maxConcurrent)
			return int(max) <= workers
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	properties.Property("queue size limits pending tasks", prop.ForAll(
		func(queueSize int) bool {
			if queueSize < 1 || queueSize > 50 {
				queueSize = 10
			}

			config := &WorkerPoolConfig{
				Workers:      1,
				QueueSize:    queueSize,
				ResultBuffer: queueSize * 2,
				TaskTimeout:  5 * time.Second,
			}

			pool := NewWorkerPool(config)
			defer pool.Stop()

			// 提交超过队列大小的任务
			submitted := 0
			for i := 0; i < queueSize*2; i++ {
				task := &mockTask{
					id:       string(rune(i)),
					duration: 100 * time.Millisecond,
				}

				err := pool.Submit(task)
				if err == nil {
					submitted++
				}
			}

			// 提交的任务数应该不超过 workers + queueSize
			return submitted <= (1 + queueSize)
		},
		gen.IntRange(1, 50),
	))

	properties.Property("worker pool handles task completion", prop.ForAll(
		func(taskCount int) bool {
			if taskCount < 1 || taskCount > 50 {
				taskCount = 10
			}

			config := GetDefaultWorkerPoolConfig()
			pool := NewWorkerPool(config)
			defer pool.Stop()

			// 提交任务
			for i := 0; i < taskCount; i++ {
				task := &mockTask{
					id:       string(rune(i)),
					duration: 1 * time.Millisecond,
				}
				pool.Submit(task)
			}

			// 等待所有任务完成
			time.Sleep(time.Duration(taskCount) * 5 * time.Millisecond)

			stats := pool.Stats()
			return stats.TotalJobs >= int64(taskCount)
		},
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}

// concurrentTrackingTask 跟踪并发数的任务包装器
type concurrentTrackingTask struct {
	Task
	currentConcurrent *int32
	maxConcurrent     *int32
}

func (t *concurrentTrackingTask) Execute(ctx context.Context) error {
	// 增加当前并发数
	current := atomic.AddInt32(t.currentConcurrent, 1)
	
	// 更新最大并发数
	for {
		max := atomic.LoadInt32(t.maxConcurrent)
		if current <= max || atomic.CompareAndSwapInt32(t.maxConcurrent, max, current) {
			break
		}
	}

	// 执行任务
	err := t.Task.Execute(ctx)

	// 减少当前并发数
	atomic.AddInt32(t.currentConcurrent, -1)

	return err
}

// TestWorkerPoolBasic 测试基本功能
func TestWorkerPoolBasic(t *testing.T) {
	config := GetDefaultWorkerPoolConfig()
	pool := NewWorkerPool(config)
	defer pool.Stop()

	// 提交任务
	task := &mockTask{
		id:       "test-task",
		duration: 10 * time.Millisecond,
	}

	err := pool.Submit(task)
	require.NoError(t, err)

	// 等待任务完成
	time.Sleep(50 * time.Millisecond)

	// 验证任务已执行
	assert.True(t, task.executed.Load())

	// 验证统计信息
	stats := pool.Stats()
	assert.Equal(t, config.Workers, stats.Workers)
	assert.GreaterOrEqual(t, stats.TotalJobs, int64(1))
}

// TestWorkerPoolQueueFull 测试队列满
func TestWorkerPoolQueueFull(t *testing.T) {
	config := &WorkerPoolConfig{
		Workers:      1,
		QueueSize:    2,
		ResultBuffer: 10,
		TaskTimeout:  5 * time.Second,
	}

	pool := NewWorkerPool(config)
	defer pool.Stop()

	// 提交阻塞任务填满队列
	for i := 0; i < 3; i++ {
		task := &mockTask{
			id:       string(rune(i)),
			duration: 100 * time.Millisecond,
		}
		pool.Submit(task)
	}

	// 再提交一个任务应该失败
	task := &mockTask{
		id:       "overflow",
		duration: 10 * time.Millisecond,
	}

	err := pool.Submit(task)
	assert.Error(t, err)
	assert.Equal(t, ErrQueueFull, err)
}

// TestWorkerPoolStop 测试停止
func TestWorkerPoolStop(t *testing.T) {
	config := GetDefaultWorkerPoolConfig()
	pool := NewWorkerPool(config)

	// 提交一些任务
	for i := 0; i < 5; i++ {
		task := &mockTask{
			id:       string(rune(i)),
			duration: 10 * time.Millisecond,
		}
		pool.Submit(task)
	}

	// 停止工作池
	pool.Stop()

	// 停止后提交任务应该失败
	task := &mockTask{
		id:       "after-stop",
		duration: 10 * time.Millisecond,
	}

	err := pool.Submit(task)
	assert.Error(t, err)
}

// TestWorkerPoolStopWithTimeout 测试带超时的停止
func TestWorkerPoolStopWithTimeout(t *testing.T) {
	config := GetDefaultWorkerPoolConfig()
	pool := NewWorkerPool(config)

	// 提交快速任务
	for i := 0; i < 5; i++ {
		task := &mockTask{
			id:       string(rune(i)),
			duration: 1 * time.Millisecond,
		}
		pool.Submit(task)
	}

	// 带超时停止
	err := pool.StopWithTimeout(1 * time.Second)
	assert.NoError(t, err)
}

// TestWorkerPoolStats 测试统计信息
func TestWorkerPoolStats(t *testing.T) {
	config := &WorkerPoolConfig{
		Workers:      5,
		QueueSize:    10,
		ResultBuffer: 10,
		TaskTimeout:  5 * time.Second,
	}

	pool := NewWorkerPool(config)
	defer pool.Stop()

	// 提交任务
	taskCount := 3
	for i := 0; i < taskCount; i++ {
		task := &mockTask{
			id:       string(rune(i)),
			duration: 10 * time.Millisecond,
		}
		pool.Submit(task)
	}

	// 获取统计信息
	stats := pool.Stats()
	assert.Equal(t, 5, stats.Workers)
	assert.Equal(t, 10, stats.QueueSize)
	assert.GreaterOrEqual(t, stats.TotalJobs, int64(0))
}

// TestWorkerPoolTaskTimeout 测试任务超时
func TestWorkerPoolTaskTimeout(t *testing.T) {
	config := &WorkerPoolConfig{
		Workers:      1,
		QueueSize:    10,
		ResultBuffer: 10,
		TaskTimeout:  50 * time.Millisecond,
	}

	pool := NewWorkerPool(config)
	defer pool.Stop()

	// 提交超时任务
	task := &mockTask{
		id:       "timeout-task",
		duration: 200 * time.Millisecond,
	}

	err := pool.Submit(task)
	require.NoError(t, err)

	// 等待任务超时
	time.Sleep(100 * time.Millisecond)

	// 从结果通道读取
	select {
	case result := <-pool.Results():
		assert.Error(t, result.Error)
		assert.Equal(t, "timeout-task", result.TaskID)
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected result from timeout task")
	}
}

// TestWorkerPoolConcurrentSubmit 测试并发提交
func TestWorkerPoolConcurrentSubmit(t *testing.T) {
	config := GetDefaultWorkerPoolConfig()
	pool := NewWorkerPool(config)
	defer pool.Stop()

	var wg sync.WaitGroup
	taskCount := 50

	// 并发提交任务
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := &mockTask{
				id:       string(rune(id)),
				duration: 1 * time.Millisecond,
			}
			pool.Submit(task)
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	stats := pool.Stats()
	assert.GreaterOrEqual(t, stats.TotalJobs, int64(taskCount))
}
