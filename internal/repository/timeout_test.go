package repository

import (
	"context"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// 属性 11：超时控制
// 验证需求：4.5
func TestTimeoutControlProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("all queries respect timeout", prop.ForAll(
		func(timeoutMs int) bool {
			if timeoutMs <= 0 {
				timeoutMs = 100
			}
			if timeoutMs > 5000 {
				timeoutMs = 5000
			}

			timeout := time.Duration(timeoutMs) * time.Millisecond
			ctx := context.Background()

			// 创建带超时的上下文
			timeoutCtx, cancel := WithTimeout(ctx, timeout)
			defer cancel()

			// 验证上下文有截止时间
			deadline, ok := timeoutCtx.Deadline()
			if !ok {
				return false
			}

			// 验证截止时间在合理范围内
			expectedDeadline := time.Now().Add(timeout)
			diff := deadline.Sub(expectedDeadline).Abs()

			// 允许 100ms 的误差
			return diff < 100*time.Millisecond
		},
		gen.IntRange(1, 5000),
	))

	properties.Property("zero timeout uses default", prop.ForAll(
		func() bool {
			ctx := context.Background()

			// 使用零超时
			timeoutCtx, cancel := WithTimeout(ctx, 0)
			defer cancel()

			// 验证使用了默认超时
			deadline, ok := timeoutCtx.Deadline()
			if !ok {
				return false
			}

			expectedDeadline := time.Now().Add(DefaultQueryTimeout)
			diff := deadline.Sub(expectedDeadline).Abs()

			// 允许 100ms 的误差
			return diff < 100*time.Millisecond
		},
	))

	properties.Property("negative timeout uses default", prop.ForAll(
		func(negativeTimeout int) bool {
			if negativeTimeout >= 0 {
				negativeTimeout = -1
			}

			ctx := context.Background()
			timeout := time.Duration(negativeTimeout) * time.Millisecond

			// 使用负超时
			timeoutCtx, cancel := WithTimeout(ctx, timeout)
			defer cancel()

			// 验证使用了默认超时
			deadline, ok := timeoutCtx.Deadline()
			if !ok {
				return false
			}

			expectedDeadline := time.Now().Add(DefaultQueryTimeout)
			diff := deadline.Sub(expectedDeadline).Abs()

			// 允许 100ms 的误差
			return diff < 100*time.Millisecond
		},
		gen.IntRange(-10000, -1),
	))

	properties.Property("context cancellation propagates", prop.ForAll(
		func(timeoutMs int) bool {
			if timeoutMs <= 0 {
				timeoutMs = 100
			}
			if timeoutMs > 1000 {
				timeoutMs = 1000
			}

			timeout := time.Duration(timeoutMs) * time.Millisecond
			ctx := context.Background()

			// 创建带超时的上下文
			timeoutCtx, cancel := WithTimeout(ctx, timeout)

			// 立即取消
			cancel()

			// 验证上下文已取消
			select {
			case <-timeoutCtx.Done():
				return timeoutCtx.Err() == context.Canceled
			case <-time.After(100 * time.Millisecond):
				return false
			}
		},
		gen.IntRange(1, 1000),
	))

	properties.Property("timeout triggers context cancellation", prop.ForAll(
		func() bool {
			ctx := context.Background()

			// 创建非常短的超时
			timeoutCtx, cancel := WithTimeout(ctx, 10*time.Millisecond)
			defer cancel()

			// 等待超时
			<-timeoutCtx.Done()

			// 验证是超时导致的取消
			return timeoutCtx.Err() == context.DeadlineExceeded
		},
	))

	properties.Property("default timeout is reasonable", prop.ForAll(
		func() bool {
			ctx := context.Background()

			// 使用默认超时
			timeoutCtx, cancel := WithDefaultTimeout(ctx)
			defer cancel()

			// 验证默认超时是 30 秒
			deadline, ok := timeoutCtx.Deadline()
			if !ok {
				return false
			}

			expectedDeadline := time.Now().Add(DefaultQueryTimeout)
			diff := deadline.Sub(expectedDeadline).Abs()

			// 允许 100ms 的误差
			return diff < 100*time.Millisecond
		},
	))

	properties.TestingRun(t)
}

// TestTimeoutScenarios 测试具体的超时场景
func TestTimeoutScenarios(t *testing.T) {
	// 测试正常超时
	t.Run("normal timeout", func(t *testing.T) {
		ctx := context.Background()
		timeoutCtx, cancel := WithTimeout(ctx, 1*time.Second)
		defer cancel()

		deadline, ok := timeoutCtx.Deadline()
		assert.True(t, ok)
		assert.True(t, time.Until(deadline) > 0)
		assert.True(t, time.Until(deadline) <= 1*time.Second)
	})

	// 测试默认超时
	t.Run("default timeout", func(t *testing.T) {
		ctx := context.Background()
		timeoutCtx, cancel := WithDefaultTimeout(ctx)
		defer cancel()

		deadline, ok := timeoutCtx.Deadline()
		assert.True(t, ok)
		assert.True(t, time.Until(deadline) > 0)
		assert.True(t, time.Until(deadline) <= DefaultQueryTimeout)
	})

	// 测试零超时使用默认值
	t.Run("zero timeout uses default", func(t *testing.T) {
		ctx := context.Background()
		timeoutCtx, cancel := WithTimeout(ctx, 0)
		defer cancel()

		deadline, ok := timeoutCtx.Deadline()
		assert.True(t, ok)

		expectedDeadline := time.Now().Add(DefaultQueryTimeout)
		diff := deadline.Sub(expectedDeadline).Abs()
		assert.Less(t, diff, 100*time.Millisecond)
	})

	// 测试负超时使用默认值
	t.Run("negative timeout uses default", func(t *testing.T) {
		ctx := context.Background()
		timeoutCtx, cancel := WithTimeout(ctx, -1*time.Second)
		defer cancel()

		deadline, ok := timeoutCtx.Deadline()
		assert.True(t, ok)

		expectedDeadline := time.Now().Add(DefaultQueryTimeout)
		diff := deadline.Sub(expectedDeadline).Abs()
		assert.Less(t, diff, 100*time.Millisecond)
	})

	// 测试超时触发
	t.Run("timeout triggers", func(t *testing.T) {
		ctx := context.Background()
		timeoutCtx, cancel := WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		// 等待超时
		<-timeoutCtx.Done()

		// 验证错误类型
		assert.Equal(t, context.DeadlineExceeded, timeoutCtx.Err())
	})

	// 测试手动取消
	t.Run("manual cancellation", func(t *testing.T) {
		ctx := context.Background()
		timeoutCtx, cancel := WithTimeout(ctx, 1*time.Second)

		// 立即取消
		cancel()

		// 验证上下文已取消
		select {
		case <-timeoutCtx.Done():
			assert.Equal(t, context.Canceled, timeoutCtx.Err())
		case <-time.After(100 * time.Millisecond):
			t.Error("context should be cancelled immediately")
		}
	})

	// 测试嵌套超时
	t.Run("nested timeout", func(t *testing.T) {
		ctx := context.Background()

		// 外层超时 1 秒
		outerCtx, outerCancel := WithTimeout(ctx, 1*time.Second)
		defer outerCancel()

		// 内层超时 100 毫秒
		innerCtx, innerCancel := WithTimeout(outerCtx, 100*time.Millisecond)
		defer innerCancel()

		// 等待内层超时
		<-innerCtx.Done()

		// 验证内层超时先触发
		assert.Equal(t, context.DeadlineExceeded, innerCtx.Err())

		// 外层上下文应该还没超时
		select {
		case <-outerCtx.Done():
			t.Error("outer context should not be done yet")
		default:
			// 正常
		}
	})

	// 测试并发超时
	t.Run("concurrent timeouts", func(t *testing.T) {
		ctx := context.Background()
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				timeoutCtx, cancel := WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()

				<-timeoutCtx.Done()
				assert.Equal(t, context.DeadlineExceeded, timeoutCtx.Err())
				done <- true
			}()
		}

		// 等待所有 goroutine 完成
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				// 正常
			case <-time.After(200 * time.Millisecond):
				t.Error("timeout should trigger within 200ms")
			}
		}
	})
}
