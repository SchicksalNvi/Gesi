package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoader_LoadMainConfig(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	mainConfigPath := filepath.Join(tmpDir, "config.toml")

	// 写入测试配置
	configContent := `
database = "test.db"
activity_log = "test.log"

[admin]
username = "testadmin"
password = "testpass"

[[nodes]]
name = "test-node"
environment = "test"
host = "localhost"
port = 9001
username = "user"
password = "pass"
`
	err := os.WriteFile(mainConfigPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 测试加载
	loader := NewConfigLoader(mainConfigPath, "")
	cfg, err := loader.LoadMainConfig()

	require.NoError(t, err)
	assert.Equal(t, "test.db", cfg.Database)
	assert.Equal(t, "testadmin", cfg.AdminUsername)
	assert.Len(t, cfg.Nodes, 1)
	assert.Equal(t, "test-node", cfg.Nodes[0].Name)
}

func TestConfigLoader_LoadNodeList(t *testing.T) {
	tmpDir := t.TempDir()
	nodeListPath := filepath.Join(tmpDir, "nodelist.toml")

	// 写入节点列表
	nodeListContent := `
[[nodes]]
name = "node-1"
environment = "production"
host = "192.168.1.10"
port = 9001
username = "supervisor"
password = "secret"

[[nodes]]
name = "node-2"
environment = "staging"
host = "192.168.1.11"
port = 9002
username = "supervisor"
password = "secret2"
`
	err := os.WriteFile(nodeListPath, []byte(nodeListContent), 0644)
	require.NoError(t, err)

	// 测试加载
	loader := NewConfigLoader("", nodeListPath)
	nodes, err := loader.LoadNodeList()

	require.NoError(t, err)
	assert.Len(t, nodes, 2)
	assert.Equal(t, "node-1", nodes[0].Name)
	assert.Equal(t, "192.168.1.10", nodes[0].Host)
	assert.Equal(t, 9001, nodes[0].Port)
	assert.Equal(t, "node-2", nodes[1].Name)
}

func TestConfigLoader_LoadNodeList_FileNotExists(t *testing.T) {
	// 测试文件不存在的情况
	loader := NewConfigLoader("", "/nonexistent/nodelist.toml")
	nodes, err := loader.LoadNodeList()

	require.NoError(t, err, "should not return error when file doesn't exist")
	assert.Empty(t, nodes, "should return empty slice")
}

func TestConfigLoader_MergeNodes(t *testing.T) {
	loader := NewConfigLoader("", "")

	mainNodes := []NodeConfig{
		{Name: "node-1", Host: "main-host-1", Port: 9001},
		{Name: "node-2", Host: "main-host-2", Port: 9002},
		{Name: "node-3", Host: "main-host-3", Port: 9003},
	}

	nodeListNodes := []NodeConfig{
		{Name: "node-1", Host: "nodelist-host-1", Port: 9101}, // 重复
		{Name: "node-4", Host: "nodelist-host-4", Port: 9104},
	}

	merged := loader.MergeNodes(mainNodes, nodeListNodes)

	// 应该有 4 个节点：nodelist 的 2 个 + main 的 2 个不重复的
	assert.Len(t, merged, 4)

	// 检查 nodelist 节点优先
	assert.Equal(t, "node-1", merged[0].Name)
	assert.Equal(t, "nodelist-host-1", merged[0].Host)
	assert.Equal(t, 9101, merged[0].Port)

	// 检查 node-4 来自 nodelist
	assert.Equal(t, "node-4", merged[1].Name)

	// 检查 node-2 和 node-3 来自 main（不重复）
	nodeNames := make(map[string]bool)
	for _, node := range merged {
		nodeNames[node.Name] = true
	}
	assert.True(t, nodeNames["node-2"])
	assert.True(t, nodeNames["node-3"])
}

func TestConfigLoader_Load_BothSources(t *testing.T) {
	tmpDir := t.TempDir()
	mainConfigPath := filepath.Join(tmpDir, "config.toml")
	nodeListPath := filepath.Join(tmpDir, "nodelist.toml")

	// 写入主配置
	mainConfigContent := `
database = "test.db"

[admin]
username = "admin"
password = "pass"

[[nodes]]
name = "main-node"
environment = "production"
host = "main-host"
port = 9001
username = "user"
password = "pass"
`
	err := os.WriteFile(mainConfigPath, []byte(mainConfigContent), 0644)
	require.NoError(t, err)

	// 写入节点列表
	nodeListContent := `
[[nodes]]
name = "nodelist-node"
environment = "staging"
host = "nodelist-host"
port = 9002
username = "user2"
password = "pass2"
`
	err = os.WriteFile(nodeListPath, []byte(nodeListContent), 0644)
	require.NoError(t, err)

	// 测试加载
	loader := NewConfigLoader(mainConfigPath, nodeListPath)
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, "test.db", cfg.Database)
	assert.Len(t, cfg.Nodes, 2)

	// 验证两个节点都存在
	nodeNames := make(map[string]bool)
	for _, node := range cfg.Nodes {
		nodeNames[node.Name] = true
	}
	assert.True(t, nodeNames["main-node"])
	assert.True(t, nodeNames["nodelist-node"])
}

func TestConfigLoader_Load_OnlyMainConfig(t *testing.T) {
	tmpDir := t.TempDir()
	mainConfigPath := filepath.Join(tmpDir, "config.toml")

	// 写入主配置（包含节点）
	mainConfigContent := `
database = "test.db"

[admin]
username = "admin"
password = "pass"

[[nodes]]
name = "legacy-node"
environment = "production"
host = "legacy-host"
port = 9001
username = "user"
password = "pass"
`
	err := os.WriteFile(mainConfigPath, []byte(mainConfigContent), 0644)
	require.NoError(t, err)

	// nodelist.toml 不存在
	nodeListPath := filepath.Join(tmpDir, "nodelist.toml")

	// 测试加载
	loader := NewConfigLoader(mainConfigPath, nodeListPath)
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.Len(t, cfg.Nodes, 1)
	assert.Equal(t, "legacy-node", cfg.Nodes[0].Name)
}

func TestConfigLoader_ExpandEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	mainConfigPath := filepath.Join(tmpDir, "config.toml")

	// 设置环境变量
	os.Setenv("TEST_PASSWORD", "secret123")
	os.Setenv("TEST_HOST", "test-host")
	defer os.Unsetenv("TEST_PASSWORD")
	defer os.Unsetenv("TEST_HOST")

	// 写入包含环境变量的配置
	mainConfigContent := `
database = "test.db"

[admin]
username = "admin"
password = "${TEST_PASSWORD}"

[[nodes]]
name = "test-node"
environment = "production"
host = "${TEST_HOST}"
port = 9001
username = "user"
password = "${TEST_PASSWORD}"
`
	err := os.WriteFile(mainConfigPath, []byte(mainConfigContent), 0644)
	require.NoError(t, err)

	// 测试加载
	loader := NewConfigLoader(mainConfigPath, "")
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, "secret123", cfg.AdminPassword)
	assert.Equal(t, "test-host", cfg.Nodes[0].Host)
	assert.Equal(t, "secret123", cfg.Nodes[0].Password)
}

func TestConfigLoader_MergeNodes_EmptyNodeList(t *testing.T) {
	loader := NewConfigLoader("", "")

	mainNodes := []NodeConfig{
		{Name: "node-1", Host: "host-1", Port: 9001},
	}
	nodeListNodes := []NodeConfig{}

	merged := loader.MergeNodes(mainNodes, nodeListNodes)

	assert.Len(t, merged, 1)
	assert.Equal(t, "node-1", merged[0].Name)
}

func TestConfigLoader_MergeNodes_EmptyMainNodes(t *testing.T) {
	loader := NewConfigLoader("", "")

	mainNodes := []NodeConfig{}
	nodeListNodes := []NodeConfig{
		{Name: "node-1", Host: "host-1", Port: 9001},
	}

	merged := loader.MergeNodes(mainNodes, nodeListNodes)

	assert.Len(t, merged, 1)
	assert.Equal(t, "node-1", merged[0].Name)
}
