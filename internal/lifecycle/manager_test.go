package lifecycle

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Mock component for testing
type mockComponent struct {
	name          string
	startCalled   bool
	stopCalled    bool
	startError    error
	stopError     error
	healthStatus  string
	startDuration time.Duration
	stopDuration  time.Duration
}

func (m *mockComponent) Start(ctx context.Context) error {
	if m.startDuration > 0 {
		time.Sleep(m.startDuration)
	}
	m.startCalled = true
	return m.startError
}

func (m *mockComponent) Stop(ctx context.Context) error {
	if m.stopDuration > 0 {
		time.Sleep(m.stopDuration)
	}
	m.stopCalled = true
	return m.stopError
}

func (m *mockComponent) Health() HealthStatus {
	return HealthStatus{
		Status:    m.healthStatus,
		Timestamp: time.Now(),
		Details:   map[string]interface{}{"name": m.name},
	}
}

// 属性 9：Goroutine 生命周期
// 验证需求：4.1
func TestLifecycleProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("all registered components are started", prop.ForAll(
		func(count int) bool {
			if count < 1 || count > 10 {
				return true // 跳过无效输入
			}

			manager := NewManager()
			components := make([]*mockComponent, count)

			// 注册组件
			for i := 0; i < count; i++ {
				comp := &mockComponent{
					name:         string(rune('A' + i)),
					healthStatus: "healthy",
				}
				components[i] = comp
				manager.Register(comp.name, comp, i)
			}

			// 启动所有组件
			ctx := context.Background()
			if err := manager.StartAll(ctx); err != nil {
				return false
			}

			// 验证所有组件都被启动
			for _, comp := range components {
				if !comp.startCalled {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.Property("all started components are stopped", prop.ForAll(
		func(count int) bool {
			if count < 1 || count > 10 {
				return true
			}

			manager := NewManager()
			components := make([]*mockComponent, count)

			for i := 0; i < count; i++ {
				comp := &mockComponent{
					name:         string(rune('A' + i)),
					healthStatus: "healthy",
				}
				components[i] = comp
				manager.Register(comp.name, comp, i)
			}

			ctx := context.Background()
			manager.StartAll(ctx)
			manager.StopAll(ctx)

			// 验证所有组件都被停止
			for _, comp := range components {
				if !comp.stopCalled {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// 单元测试
func TestManagerRegister(t *testing.T) {
	manager := NewManager()
	comp := &mockComponent{name: "test", healthStatus: "healthy"}

	manager.Register("test", comp, 1)

	if len(manager.components) != 1 {
		t.Errorf("expected 1 component, got %d", len(manager.components))
	}
}

func TestManagerStartAll(t *testing.T) {
	manager := NewManager()
	comp1 := &mockComponent{name: "comp1", healthStatus: "healthy"}
	comp2 := &mockComponent{name: "comp2", healthStatus: "healthy"}

	manager.Register("comp1", comp1, 1)
	manager.Register("comp2", comp2, 2)

	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	if !comp1.startCalled {
		t.Error("comp1 was not started")
	}
	if !comp2.startCalled {
		t.Error("comp2 was not started")
	}
}

func TestManagerStartAllWithError(t *testing.T) {
	manager := NewManager()
	comp1 := &mockComponent{name: "comp1", healthStatus: "healthy"}
	comp2 := &mockComponent{
		name:         "comp2",
		healthStatus: "healthy",
		startError:   errors.New("start failed"),
	}

	manager.Register("comp1", comp1, 1)
	manager.Register("comp2", comp2, 2)

	ctx := context.Background()
	if err := manager.StartAll(ctx); err == nil {
		t.Error("expected StartAll to fail")
	}

	// comp1 应该被启动然后停止
	if !comp1.startCalled {
		t.Error("comp1 was not started")
	}
	if !comp1.stopCalled {
		t.Error("comp1 was not stopped during cleanup")
	}
}

func TestManagerStopAll(t *testing.T) {
	manager := NewManager()
	comp1 := &mockComponent{name: "comp1", healthStatus: "healthy"}
	comp2 := &mockComponent{name: "comp2", healthStatus: "healthy"}

	manager.Register("comp1", comp1, 1)
	manager.Register("comp2", comp2, 2)

	ctx := context.Background()
	manager.StartAll(ctx)
	if err := manager.StopAll(ctx); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if !comp1.stopCalled {
		t.Error("comp1 was not stopped")
	}
	if !comp2.stopCalled {
		t.Error("comp2 was not stopped")
	}
}

func TestManagerHealthCheck(t *testing.T) {
	manager := NewManager()
	comp1 := &mockComponent{name: "comp1", healthStatus: "healthy"}
	comp2 := &mockComponent{name: "comp2", healthStatus: "degraded"}

	manager.Register("comp1", comp1, 1)
	manager.Register("comp2", comp2, 2)

	results := manager.HealthCheck()

	if len(results) != 2 {
		t.Errorf("expected 2 health results, got %d", len(results))
	}

	if results["comp1"].Status != "healthy" {
		t.Errorf("expected comp1 to be healthy, got %s", results["comp1"].Status)
	}

	if results["comp2"].Status != "degraded" {
		t.Errorf("expected comp2 to be degraded, got %s", results["comp2"].Status)
	}
}

func TestManagerIsHealthy(t *testing.T) {
	manager := NewManager()
	comp1 := &mockComponent{name: "comp1", healthStatus: "healthy"}
	comp2 := &mockComponent{name: "comp2", healthStatus: "healthy"}

	manager.Register("comp1", comp1, 1)
	manager.Register("comp2", comp2, 2)

	if !manager.IsHealthy() {
		t.Error("expected manager to be healthy")
	}

	// 改变一个组件的健康状态
	comp2.healthStatus = "unhealthy"

	if manager.IsHealthy() {
		t.Error("expected manager to be unhealthy")
	}
}

func TestManagerGetOverallHealth(t *testing.T) {
	manager := NewManager()
	comp1 := &mockComponent{name: "comp1", healthStatus: "healthy"}
	comp2 := &mockComponent{name: "comp2", healthStatus: "healthy"}

	manager.Register("comp1", comp1, 1)
	manager.Register("comp2", comp2, 2)

	status := manager.GetOverallHealth()
	if status.Status != "healthy" {
		t.Errorf("expected overall status to be healthy, got %s", status.Status)
	}

	// 测试降级状态
	comp2.healthStatus = "degraded"
	status = manager.GetOverallHealth()
	if status.Status != "degraded" {
		t.Errorf("expected overall status to be degraded, got %s", status.Status)
	}

	// 测试不健康状态
	comp2.healthStatus = "unhealthy"
	status = manager.GetOverallHealth()
	if status.Status != "unhealthy" {
		t.Errorf("expected overall status to be unhealthy, got %s", status.Status)
	}
}

func TestManagerPriorityOrder(t *testing.T) {
	manager := NewManager()

	// 创建组件，优先级分别为 3, 2, 1
	compA := &mockComponent{name: "A", healthStatus: "healthy"}
	compB := &mockComponent{name: "B", healthStatus: "healthy"}
	compC := &mockComponent{name: "C", healthStatus: "healthy"}

	manager.Register("A", compA, 3)
	manager.Register("B", compB, 2)
	manager.Register("C", compC, 1)

	ctx := context.Background()
	manager.StartAll(ctx)

	// 验证所有组件都被启动
	if !compA.startCalled || !compB.startCalled || !compC.startCalled {
		t.Error("not all components were started")
	}

	// 验证组件按优先级排序（通过检查 manager.components 的顺序）
	if len(manager.components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(manager.components))
	}

	// 优先级小的应该在前面
	if manager.components[0].Name != "C" {
		t.Errorf("expected first component to be C, got %s", manager.components[0].Name)
	}
	if manager.components[1].Name != "B" {
		t.Errorf("expected second component to be B, got %s", manager.components[1].Name)
	}
	if manager.components[2].Name != "A" {
		t.Errorf("expected third component to be A, got %s", manager.components[2].Name)
	}
}
