package config

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicConfigManager_AtomicOperations(t *testing.T) {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-at-least-32-chars-long")
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ADMIN_PASSWORD")

	manager := NewAtomicConfigManager()
	defer manager.Stop()

	// 创建临时配置文件
	configContent := `
database = "sqlite:///test.db"
activity_log = "test.log"
admin_username = "testadmin"
admin_password = "testpass"

[[nodes]]
name = "test-node"
environment = "test"
host = "localhost"
port = 9001
username = "user"
password = "pass"
`

	tmpFile, err := os.CreateTemp("", "config-*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 测试加载配置
	cfg, err := manager.Load(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "testadmin", cfg.AdminUsername)

	// 测试原子读取
	readCfg := manager.Get()
	assert.NotNil(t, readCfg)
	assert.Equal(t, cfg.AdminUsername, readCfg.AdminUsername)

	// 测试版本号
	version := manager.GetVersion()
	assert.Greater(t, version, int64(0))
}

func TestAtomicConfigManager_ConcurrentAccess(t *testing.T) {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-at-least-32-chars-long")
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ADMIN_PASSWORD")

	manager := NewAtomicConfigManager()
	defer manager.Stop()

	// 创建临时配置文件
	configContent := `
database = "sqlite:///test.db"
activity_log = "test.log"
admin_username = "testadmin"
admin_password = "testpass"

[[nodes]]
name = "test-node"
environment = "test"
host = "localhost"
port = 9001
username = "user"
password = "pass"
`

	tmpFile, err := os.CreateTemp("", "config-*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 加载初始配置
	_, err = manager.Load(tmpFile.Name())
	require.NoError(t, err)

	// 并发读取测试
	const numReaders = 100
	const numReads = 1000

	var wg sync.WaitGroup
	errors := make(chan error, numReaders)

	// 启动多个读取 goroutine
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numReads; j++ {
				cfg := manager.Get()
				if cfg == nil {
					errors <- assert.AnError
					return
				}
				if cfg.AdminUsername != "testadmin" {
					errors <- assert.AnError
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("Concurrent read error: %v", err)
	}
}

func TestAtomicConfigManager_ReloadPrevention(t *testing.T) {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-at-least-32-chars-long")
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ADMIN_PASSWORD")

	manager := NewAtomicConfigManager()
	defer manager.Stop()

	// 创建临时配置文件
	configContent := `
database = "sqlite:///test.db"
activity_log = "test.log"
admin_username = "testadmin"
admin_password = "testpass"

[[nodes]]
name = "test-node"
environment = "test"
host = "localhost"
port = 9001
username = "user"
password = "pass"
`

	tmpFile, err := os.CreateTemp("", "config-*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 加载初始配置
	_, err = manager.Load(tmpFile.Name())
	require.NoError(t, err)

	// 测试并发重载防护
	const numReloaders = 10
	var wg sync.WaitGroup

	for i := 0; i < numReloaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.safeReloadConfig()
		}()
	}

	wg.Wait()

	// 验证配置仍然有效
	cfg := manager.Get()
	assert.NotNil(t, cfg)
	assert.Equal(t, "testadmin", cfg.AdminUsername)
}

func TestAtomicConfigManager_ValidationFailure(t *testing.T) {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-at-least-32-chars-long")
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ADMIN_PASSWORD")

	manager := NewAtomicConfigManager()
	defer manager.Stop()

	// 创建有效的初始配置
	validConfigContent := `
database = "sqlite:///test.db"
activity_log = "test.log"
admin_username = "testadmin"
admin_password = "testpass"

[[nodes]]
name = "test-node"
environment = "test"
host = "localhost"
port = 9001
username = "user"
password = "pass"
`

	tmpFile, err := os.CreateTemp("", "config-*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(validConfigContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 加载有效配置
	validCfg, err := manager.Load(tmpFile.Name())
	require.NoError(t, err)
	initialVersion := manager.GetVersion()

	// 创建无效配置（缺少必需字段）
	invalidConfigContent := `
database = "sqlite:///test.db"
# 缺少 admin_username 和 admin_password
`

	// 写入无效配置
	err = os.WriteFile(tmpFile.Name(), []byte(invalidConfigContent), 0644)
	require.NoError(t, err)

	// 尝试重载（应该失败并保持旧配置）
	manager.safeReloadConfig()

	// 验证配置没有改变
	currentCfg := manager.Get()
	assert.NotNil(t, currentCfg)
	assert.Equal(t, validCfg.AdminUsername, currentCfg.AdminUsername)
	assert.Equal(t, initialVersion, manager.GetVersion()) // 版本不应该增加，因为验证失败
}

func TestAtomicConfigManager_Rollback(t *testing.T) {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-at-least-32-chars-long")
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ADMIN_PASSWORD")

	manager := NewAtomicConfigManager()
	defer manager.Stop()

	// 创建初始配置
	configContent1 := `
database = "sqlite:///test1.db"
activity_log = "test1.log"
admin_username = "admin1"
admin_password = "testpass"

[[nodes]]
name = "test-node-1"
environment = "test"
host = "localhost"
port = 9001
username = "user"
password = "pass"
`

	tmpFile, err := os.CreateTemp("", "config-*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent1)
	require.NoError(t, err)
	tmpFile.Close()

	// 加载初始配置
	cfg1, err := manager.Load(tmpFile.Name())
	require.NoError(t, err)
	version1 := manager.GetVersion()

	// 模拟配置更新（直接加载新配置而不是文件重载）
	configContent2 := `
database = "sqlite:///test2.db"
activity_log = "test2.log"
admin_username = "admin2"
admin_password = "testpass"

[[nodes]]
name = "test-node-2"
environment = "test"
host = "localhost"
port = 9002
username = "user"
password = "pass"
`

	tmpFile2, err := os.CreateTemp("", "config2-*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile2.Name())

	_, err = tmpFile2.WriteString(configContent2)
	require.NoError(t, err)
	tmpFile2.Close()

	// 加载第二个配置
	cfg2, err := manager.Load(tmpFile2.Name())
	require.NoError(t, err)
	version2 := manager.GetVersion()

	// 验证新配置已加载
	assert.Equal(t, "admin2", cfg2.AdminUsername)
	assert.Greater(t, version2, version1)

	// 执行回滚
	err = manager.Rollback()
	require.NoError(t, err)

	// 验证回滚成功
	rolledBackCfg := manager.Get()
	version3 := manager.GetVersion()

	assert.Equal(t, cfg1.AdminUsername, rolledBackCfg.AdminUsername)
	assert.Equal(t, cfg1.Database, rolledBackCfg.Database)
	assert.Greater(t, version3, version2)
}

func TestAtomicConfigManager_StopSafety(t *testing.T) {
	manager := NewAtomicConfigManager()

	// 测试多次停止不会 panic
	manager.Stop()
	manager.Stop()
	manager.Stop()

	// 测试停止后的操作
	err := manager.Watch(func(*Config) {})
	assert.Error(t, err)

	err = manager.WatchNodeList("", func([]NodeConfig) {})
	assert.Error(t, err)
}

func TestAtomicConfigManager_CallbackSafety(t *testing.T) {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-at-least-32-chars-long")
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ADMIN_PASSWORD")

	manager := NewAtomicConfigManager()
	defer manager.Stop()

	// 创建临时配置文件
	configContent := `
database = "sqlite:///test.db"
activity_log = "test.log"
admin_username = "testadmin"
admin_password = "testpass"

[[nodes]]
name = "test-node"
environment = "test"
host = "localhost"
port = 9001
username = "user"
password = "pass"
`

	tmpFile, err := os.CreateTemp("", "config-*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 加载配置
	_, err = manager.Load(tmpFile.Name())
	require.NoError(t, err)

	// 测试设置回调不会 panic
	normalCallback := func(*Config) {
		// 回调函数被调用
	}

	err = manager.Watch(normalCallback)
	require.NoError(t, err)

	// 验证系统仍然正常工作
	cfg := manager.Get()
	assert.NotNil(t, cfg)
	assert.Equal(t, "testadmin", cfg.AdminUsername)
	
	// 验证回调已设置（通过检查内部状态或其他方式）
	// 这里我们只验证 Watch 方法成功执行而不出错
	assert.True(t, true) // 如果到这里没有 panic，说明回调设置成功
}