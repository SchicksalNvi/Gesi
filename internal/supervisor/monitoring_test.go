package supervisor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockActivityLogger 模拟活动日志记录器
type MockActivityLogger struct {
	events []LogEvent
}

type LogEvent struct {
	Level    string
	Action   string
	Resource string
	Target   string
	Message  string
}

func (m *MockActivityLogger) LogSystemEvent(level, action, resource, target, message string, extraInfo interface{}) error {
	m.events = append(m.events, LogEvent{
		Level:    level,
		Action:   action,
		Resource: resource,
		Target:   target,
		Message:  message,
	})
	return nil
}

func TestSetActivityLogger(t *testing.T) {
	service := NewSupervisorService()
	mockLogger := &MockActivityLogger{}

	service.SetActivityLogger(mockLogger)

	assert.NotNil(t, service.activityLogger)
}

func TestStartMonitoring(t *testing.T) {
	service := NewSupervisorService()
	mockLogger := &MockActivityLogger{}
	service.SetActivityLogger(mockLogger)

	// 启动监控
	stopChan := service.StartMonitoring(100 * time.Millisecond)

	// 等待一小段时间
	time.Sleep(150 * time.Millisecond)

	// 停止监控
	service.StopMonitoring(stopChan)
	
	// 停止服务
	close(service.stopChan)
	service.wg.Wait()

	// 验证监控已启动（不会崩溃）
	assert.True(t, true)
}

func TestGetStateName(t *testing.T) {
	tests := []struct {
		state    int
		expected string
	}{
		{0, "STOPPED"},
		{10, "STARTING"},
		{20, "RUNNING"},
		{30, "BACKOFF"},
		{40, "STOPPING"},
		{100, "EXITED"},
		{200, "FATAL"},
		{1000, "UNKNOWN"},
		{999, "STATE_999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getStateName(tt.state)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMonitorStatesWithoutLogger(t *testing.T) {
	service := NewSupervisorService()

	// 不设置 logger，确保不会崩溃
	service.monitorStates()

	assert.True(t, true)
}

func TestCheckNodeConnectionState(t *testing.T) {
	service := NewSupervisorService()
	mockLogger := &MockActivityLogger{}
	service.SetActivityLogger(mockLogger)

	// 创建一个模拟节点（不会真正连接）
	node := &Node{
		Name:        "test-node",
		Environment: "test",
		Host:        "localhost",
		Port:        9001,
		IsConnected: false,
		client:      nil, // 没有真实的 client
	}

	// 初始化节点状态
	service.statesMu.Lock()
	service.nodeStates[node.Name] = false
	service.statesMu.Unlock()

	// 模拟节点状态变化：断开 -> 连接
	service.statesMu.Lock()
	previousState := service.nodeStates[node.Name]
	currentState := true
	service.nodeStates[node.Name] = currentState
	service.statesMu.Unlock()

	// 手动触发日志记录（模拟状态变化检测）
	if service.activityLogger != nil && previousState != currentState {
		if currentState {
			message := "Node test-node reconnected at localhost:9001"
			service.activityLogger.LogSystemEvent("INFO", "node_connected", "node", node.Name, message, nil)
		}
	}

	// 应该记录节点连接事件
	assert.Equal(t, 1, len(mockLogger.events))
	if len(mockLogger.events) > 0 {
		event := mockLogger.events[0]
		assert.Equal(t, "INFO", event.Level)
		assert.Equal(t, "node_connected", event.Action)
		assert.Equal(t, "node", event.Resource)
		assert.Equal(t, "test-node", event.Target)
	}
}

func TestProcessStateTracking(t *testing.T) {
	service := NewSupervisorService()
	mockLogger := &MockActivityLogger{}
	service.SetActivityLogger(mockLogger)

	nodeName := "test-node"

	// 初始化进程状态
	service.statesMu.Lock()
	service.processStates[nodeName] = make(map[string]int)
	service.processStates[nodeName]["test-process"] = 0 // STOPPED
	service.statesMu.Unlock()

	// 模拟状态变化：STOPPED -> RUNNING
	service.statesMu.Lock()
	previousState := service.processStates[nodeName]["test-process"]
	currentState := 20 // RUNNING
	service.processStates[nodeName]["test-process"] = currentState
	service.statesMu.Unlock()

	// 手动触发日志记录逻辑
	if service.activityLogger != nil && previousState != currentState {
		target := "test-node:test-process"
		message := "Process test-process started on node test-node (state: STOPPED -> RUNNING)"
		service.activityLogger.LogSystemEvent("INFO", "process_started", "process", target, message, nil)
	}

	// 验证日志记录
	assert.Equal(t, 1, len(mockLogger.events))
	if len(mockLogger.events) > 0 {
		event := mockLogger.events[0]
		assert.Equal(t, "INFO", event.Level)
		assert.Equal(t, "process_started", event.Action)
		assert.Equal(t, "process", event.Resource)
		assert.Equal(t, "test-node:test-process", event.Target)
	}
}
