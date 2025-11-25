package logger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

// TestSetLogLevel 测试设置日志级别
func TestSetLogLevel(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	tests := []struct {
		name      string
		level     string
		changedBy string
		reason    string
		wantErr   bool
	}{
		{
			name:      "Set to debug level",
			level:     "debug",
			changedBy: "test",
			reason:    "testing",
			wantErr:   false,
		},
		{
			name:      "Set to info level",
			level:     "info",
			changedBy: "test",
			reason:    "testing",
			wantErr:   false,
		},
		{
			name:      "Set to warn level",
			level:     "warn",
			changedBy: "test",
			reason:    "testing",
			wantErr:   false,
		},
		{
			name:      "Set to error level",
			level:     "error",
			changedBy: "test",
			reason:    "testing",
			wantErr:   false,
		},
		{
			name:      "Invalid level",
			level:     "invalid",
			changedBy: "test",
			reason:    "testing",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetLogLevel(tt.level, tt.changedBy, tt.reason)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.level, GetLogLevel())
			}
		})
	}
}

// TestGetLogLevel 测试获取日志级别
func TestGetLogLevel(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	// 设置日志级别
	err := SetLogLevel("debug", "test", "testing")
	assert.NoError(t, err)

	// 获取日志级别
	level := GetLogLevel()
	assert.Equal(t, "debug", level)
}

// TestGetLogLevelInfo 测试获取日志级别信息
func TestGetLogLevelInfo(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	// 清空历史记录
	ClearLevelHistory("test")

	// 设置几次日志级别
	SetLogLevel("debug", "test", "test1")
	SetLogLevel("info", "test", "test2")
	SetLogLevel("warn", "test", "test3")

	// 获取日志级别信息
	info := GetLogLevelInfo()

	assert.Equal(t, "warn", info.CurrentLevel)
	assert.Equal(t, []string{"debug", "info", "warn", "error", "fatal"}, info.AvailableLevels)
	assert.Equal(t, 3, len(info.History))
	assert.NotNil(t, info.LastChanged)
}

// TestResetLogLevel 测试重置日志级别
func TestResetLogLevel(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	// 设置为 debug 级别
	err := SetLogLevel("debug", "test", "testing")
	assert.NoError(t, err)
	assert.Equal(t, "debug", GetLogLevel())

	// 重置到默认级别
	err = ResetLogLevel("test")
	assert.NoError(t, err)
	assert.Equal(t, "info", GetLogLevel())
}

// TestSetTemporaryLogLevel 测试临时日志级别
func TestSetTemporaryLogLevel(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	// 设置初始级别为 info
	SetLogLevel("info", "test", "initial")
	assert.Equal(t, "info", GetLogLevel())

	// 设置临时级别为 debug，持续 100ms
	err := SetTemporaryLogLevel("debug", 100*time.Millisecond, "test", "temporary debug")
	assert.NoError(t, err)
	assert.Equal(t, "debug", GetLogLevel())

	// 等待自动恢复
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, "info", GetLogLevel())
}

// TestGetAvailableLogLevels 测试获取可用日志级别
func TestGetAvailableLogLevels(t *testing.T) {
	levels := GetAvailableLogLevels()
	expected := []string{"debug", "info", "warn", "error", "fatal"}
	assert.Equal(t, expected, levels)
}

// TestClearLevelHistory 测试清空历史记录
func TestClearLevelHistory(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	// 设置几次日志级别
	SetLogLevel("debug", "test", "test1")
	SetLogLevel("info", "test", "test2")

	// 验证历史记录存在
	info := GetLogLevelInfo()
	assert.Greater(t, len(info.History), 0)

	// 清空历史记录
	ClearLevelHistory("test")

	// 验证历史记录已清空
	info = GetLogLevelInfo()
	assert.Equal(t, 0, len(info.History))
}

// TestLevelChangeHistory 测试级别变更历史记录
func TestLevelChangeHistory(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	// 清空历史记录
	ClearLevelHistory("test")

	// 设置日志级别
	SetLogLevel("debug", "user1", "reason1")
	SetLogLevel("info", "user2", "reason2")

	// 获取历史记录
	info := GetLogLevelInfo()
	assert.Equal(t, 2, len(info.History))

	// 验证第一条记录
	assert.Equal(t, zapcore.InfoLevel, info.History[0].OldLevel)
	assert.Equal(t, zapcore.DebugLevel, info.History[0].NewLevel)
	assert.Equal(t, "user1", info.History[0].ChangedBy)
	assert.Equal(t, "reason1", info.History[0].Reason)

	// 验证第二条记录
	assert.Equal(t, zapcore.DebugLevel, info.History[1].OldLevel)
	assert.Equal(t, zapcore.InfoLevel, info.History[1].NewLevel)
	assert.Equal(t, "user2", info.History[1].ChangedBy)
	assert.Equal(t, "reason2", info.History[1].Reason)
}

// TestConcurrentLevelChanges 测试并发级别变更
func TestConcurrentLevelChanges(t *testing.T) {
	// 初始化动态日志系统
	if err := InitDynamicLogger(); err != nil {
		t.Fatalf("Failed to initialize dynamic logger: %v", err)
	}

	// 清空历史记录
	ClearLevelHistory("test")

	// 并发设置日志级别
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			level := "info"
			if index%2 == 0 {
				level = "debug"
			}
			SetLogLevel(level, "test", "concurrent test")
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证最终状态一致
	level := GetLogLevel()
	assert.Contains(t, []string{"debug", "info"}, level)

	// 验证历史记录数量
	info := GetLogLevelInfo()
	assert.Equal(t, 10, len(info.History))
}
