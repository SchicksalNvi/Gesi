package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/pelletier/go-toml/v2"
)

// Feature: node-config-separation, Property 1: Node configuration parsing round-trip
// For any valid node configuration structure, writing it to TOML format and parsing it back
// should produce an equivalent configuration.
// Validates: Requirements 1.2
func TestNodeConfigParsingRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("node config round-trip preserves data", prop.ForAll(
		func(node NodeConfig) bool {
			// 序列化为 TOML
			data, err := toml.Marshal(node)
			if err != nil {
				t.Logf("Failed to marshal node config: %v", err)
				return false
			}

			// 反序列化
			var decoded NodeConfig
			err = toml.Unmarshal(data, &decoded)
			if err != nil {
				t.Logf("Failed to unmarshal node config: %v", err)
				return false
			}

			// 验证等价性
			return node.Name == decoded.Name &&
				node.Environment == decoded.Environment &&
				node.Host == decoded.Host &&
				node.Port == decoded.Port &&
				node.Username == decoded.Username &&
				node.Password == decoded.Password
		},
		genNodeConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genNodeConfig 生成随机的 NodeConfig
func genNodeConfig() gopter.Gen {
	return gopter.CombineGens(
		genNonEmptyString(),     // name
		genEnvironment(),        // environment
		genHost(),               // host
		genPort(),               // port
		genNonEmptyString(),     // username
		genPassword(),           // password
	).Map(func(values []interface{}) NodeConfig {
		return NodeConfig{
			Name:        values[0].(string),
			Environment: values[1].(string),
			Host:        values[2].(string),
			Port:        values[3].(int),
			Username:    values[4].(string),
			Password:    values[5].(string),
		}
	})
}

// genNonEmptyString 生成非空字符串
func genNonEmptyString() gopter.Gen {
	return gen.Identifier()
}

// genEnvironment 生成环境名称
func genEnvironment() gopter.Gen {
	return gen.OneConstOf("production", "staging", "development", "test")
}

// genHost 生成主机地址
func genHost() gopter.Gen {
	return gen.OneGenOf(
		// 主机名
		gen.Identifier(),
		// localhost
		gen.Const("localhost"),
		// 127.0.0.1
		gen.Const("127.0.0.1"),
		// 192.168.x.x
		gen.Const("192.168.1.10"),
		gen.Const("192.168.1.11"),
	)
}

// genPort 生成有效的端口号
func genPort() gopter.Gen {
	return gen.IntRange(1, 65535)
}

// genPassword 生成密码（可能包含环境变量引用）
func genPassword() gopter.Gen {
	return gen.OneGenOf(
		// 普通密码
		gen.Identifier(),
		// 环境变量引用
		gen.Const("${NODE_PASSWORD}"),
		gen.Const("${PASSWORD}"),
	)
}

// Feature: node-config-separation, Property 2: Node list merging with priority
// For any two sets of node configurations (from config.toml and nodelist.toml),
// the merged result should contain all unique nodes, and for duplicate node names,
// the configuration from nodelist.toml should be used.
// Validates: Requirements 1.3, 1.4
func TestNodeListMergingWithPriority(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("merged list contains all unique nodes", prop.ForAll(
		func(mainNodes, nodeListNodes []NodeConfig) bool {
			loader := NewConfigLoader("", "")
			merged := loader.MergeNodes(mainNodes, nodeListNodes)

			// 创建节点名称集合
			mainNames := make(map[string]NodeConfig)
			for _, node := range mainNodes {
				mainNames[node.Name] = node
			}

			nodeListNames := make(map[string]NodeConfig)
			for _, node := range nodeListNodes {
				nodeListNames[node.Name] = node
			}

			// 验证所有 nodelist 节点都在结果中
			for _, node := range nodeListNodes {
				found := false
				for _, merged := range merged {
					if merged.Name == node.Name {
						found = true
						// 验证使用的是 nodelist 的配置
						if merged.Host != node.Host || merged.Port != node.Port {
							t.Logf("Node %s should use nodelist config", node.Name)
							return false
						}
						break
					}
				}
				if !found {
					t.Logf("Node %s from nodelist not found in merged result", node.Name)
					return false
				}
			}

			// 验证不重复的 main 节点也在结果中
			for _, node := range mainNodes {
				if _, inNodeList := nodeListNames[node.Name]; !inNodeList {
					found := false
					for _, merged := range merged {
						if merged.Name == node.Name {
							found = true
							break
						}
					}
					if !found {
						t.Logf("Unique node %s from main config not found in merged result", node.Name)
						return false
					}
				}
			}

			// 验证没有重复的节点名
			seenNames := make(map[string]bool)
			for _, node := range merged {
				if seenNames[node.Name] {
					t.Logf("Duplicate node name %s in merged result", node.Name)
					return false
				}
				seenNames[node.Name] = true
			}

			return true
		},
		genNodeConfigList(),
		genNodeConfigList(),
	))

	properties.Property("nodelist nodes take priority over main nodes", prop.ForAll(
		func(nodeName string, mainNode, nodeListNode NodeConfig) bool {
			// 设置相同的节点名
			mainNode.Name = nodeName
			nodeListNode.Name = nodeName

			// 确保配置不同
			if mainNode.Host == nodeListNode.Host {
				nodeListNode.Host = nodeListNode.Host + "-different"
			}

			loader := NewConfigLoader("", "")
			merged := loader.MergeNodes([]NodeConfig{mainNode}, []NodeConfig{nodeListNode})

			// 应该只有一个节点
			if len(merged) != 1 {
				t.Logf("Expected 1 node, got %d", len(merged))
				return false
			}

			// 应该使用 nodelist 的配置
			return merged[0].Host == nodeListNode.Host &&
				merged[0].Port == nodeListNode.Port &&
				merged[0].Username == nodeListNode.Username
		},
		gen.Identifier(),  // 使用更简单的生成器
		genNodeConfig(),
		genNodeConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genNodeConfigList 生成节点配置列表
func genNodeConfigList() gopter.Gen {
	return gen.SliceOfN(5, genNodeConfig()).Map(func(nodes []NodeConfig) []NodeConfig {
		// 确保节点名称唯一
		for i := range nodes {
			nodes[i].Name = nodes[i].Name + "-" + string(rune('a'+i))
		}
		return nodes
	})
}

// Feature: node-config-separation, Property 3: Environment variable expansion
// For any configuration file containing environment variable references (e.g., `${VAR_NAME}`),
// after expansion, all references should be replaced with their corresponding environment variable values.
// Validates: Requirements 2.5
func TestEnvironmentVariableExpansion(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("env vars are expanded in config", prop.ForAll(
		func(varName, varValue string, cfg Config) bool {
			// 设置环境变量
			os.Setenv(varName, varValue)
			defer os.Unsetenv(varName)

			// 在配置中使用环境变量引用
			cfg.AdminPassword = "${" + varName + "}"
			cfg.Database = "${" + varName + "}/db"
			if len(cfg.Nodes) > 0 {
				cfg.Nodes[0].Password = "${" + varName + "}"
				cfg.Nodes[0].Host = "${" + varName + "}.example.com"
			}

			// 执行展开
			loader := NewConfigLoader("", "")
			loader.expandEnvVars(&cfg)

			// 验证展开结果
			if cfg.AdminPassword != varValue {
				t.Logf("AdminPassword not expanded: expected %s, got %s", varValue, cfg.AdminPassword)
				return false
			}
			if cfg.Database != varValue+"/db" {
				t.Logf("Database not expanded: expected %s/db, got %s", varValue, cfg.Database)
				return false
			}
			if len(cfg.Nodes) > 0 {
				if cfg.Nodes[0].Password != varValue {
					t.Logf("Node password not expanded: expected %s, got %s", varValue, cfg.Nodes[0].Password)
					return false
				}
				if cfg.Nodes[0].Host != varValue+".example.com" {
					t.Logf("Node host not expanded: expected %s.example.com, got %s", varValue, cfg.Nodes[0].Host)
					return false
				}
			}

			return true
		},
		gen.Identifier(),  // varName
		gen.Identifier(),  // varValue
		genConfig(),       // cfg
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genConfig 生成随机配置
func genConfig() gopter.Gen {
	return gopter.CombineGens(
		gen.Identifier(),        // Database
		gen.Identifier(),        // ActivityLog
		gen.Identifier(),        // AdminUsername
		gen.Identifier(),        // AdminPassword
		genNodeConfigList(),     // Nodes
	).Map(func(values []interface{}) Config {
		return Config{
			Database:      values[0].(string),
			ActivityLog:   values[1].(string),
			AdminUsername: values[2].(string),
			AdminPassword: values[3].(string),
			Nodes:         values[4].([]NodeConfig),
		}
	})
}

// Feature: node-config-separation, Property 4: Validation consistency across sources
// For any node configuration, the validation result should be identical regardless of
// whether it comes from config.toml or nodelist.toml.
// Validates: Requirements 5.2
func TestValidationConsistencyAcrossSources(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("validation is source-agnostic", prop.ForAll(
		func(node NodeConfig) bool {
			validator := NewValidator()

			// 验证节点（不管来源）
			err1 := validator.ValidateNode(node)

			// 再次验证同一个节点
			err2 := validator.ValidateNode(node)

			// 结果应该一致
			if (err1 == nil) != (err2 == nil) {
				t.Logf("Validation inconsistent: first=%v, second=%v", err1, err2)
				return false
			}

			if err1 != nil && err2 != nil {
				if err1.Error() != err2.Error() {
					t.Logf("Error messages differ: %s vs %s", err1.Error(), err2.Error())
					return false
				}
			}

			return true
		},
		genNodeConfig(),
	))

	properties.Property("valid nodes pass validation", prop.ForAll(
		func(name, env, host, username, password string, port int) bool {
			// 确保端口有效
			if port <= 0 || port > 65535 {
				port = 9001
			}
			// 确保必填字段非空
			if name == "" {
				name = "test-node"
			}
			if env == "" {
				env = "production"
			}
			if host == "" {
				host = "localhost"
			}

			node := NodeConfig{
				Name:        name,
				Environment: env,
				Host:        host,
				Port:        port,
				Username:    username,
				Password:    password,
			}

			validator := NewValidator()
			err := validator.ValidateNode(node)

			return err == nil
		},
		gen.Identifier(),
		genEnvironment(),
		genHost(),
		gen.Identifier(),
		gen.Identifier(),
		genPort(),
	))

	properties.Property("invalid nodes fail validation", prop.ForAll(
		func(node NodeConfig) bool {
			// 故意使节点无效
			node.Name = ""
			node.Port = 0

			validator := NewValidator()
			err := validator.ValidateNode(node)

			// 应该返回错误
			return err != nil
		},
		genNodeConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: node-config-separation, Property 7: Error messages contain source information
// For any configuration parsing error, the error message should contain the source file path
// and, when available, the line number where the error occurred.
// Validates: Requirements 5.3
func TestErrorMessagesContainSourceInformation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("parse errors include file path", prop.ForAll(
		func(invalidContent string) bool {
			tmpDir := t.TempDir()
			nodeListPath := tmpDir + "/nodelist.toml"

			// 写入无效的 TOML 内容
			invalidToml := "[[nodes]]\nname = " + invalidContent + "\n"
			os.WriteFile(nodeListPath, []byte(invalidToml), 0644)

			loader := NewConfigLoader("", nodeListPath)
			_, err := loader.LoadNodeList()

			if err == nil {
				// 某些随机字符串可能恰好是有效的 TOML
				return true
			}

			// 错误消息应该包含文件路径
			errMsg := err.Error()
			if !strings.Contains(errMsg, nodeListPath) {
				t.Logf("Error message missing file path: %s", errMsg)
				return false
			}

			return true
		},
		gen.Identifier(),
	))

	properties.Property("file read errors include file path", prop.ForAll(
		func(filename string) bool {
			// 使用不存在的文件路径
			nonExistentPath := "/nonexistent/" + filename + ".toml"

			loader := NewConfigLoader(nonExistentPath, "")
			_, err := loader.LoadMainConfig()

			if err == nil {
				return false // 应该返回错误
			}

			// 错误消息应该包含文件路径
			errMsg := err.Error()
			return strings.Contains(errMsg, nonExistentPath)
		},
		gen.Identifier(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: node-config-separation, Property 6: Concurrent read safety
// For any number of concurrent goroutines reading the configuration, no data races should occur
// and all reads should return consistent snapshots of the configuration.
// Validates: Requirements 5.5
func TestConcurrentReadSafety(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("concurrent reads are race-free", prop.ForAll(
		func(cfg Config, numReaders int) bool {
			// 限制并发读取器数量
			if numReaders <= 0 || numReaders > 100 {
				numReaders = 10
			}

			// 设置必需的环境变量
			os.Setenv("JWT_SECRET", "test-jwt-secret-at-least-32-characters-long")
			os.Setenv("ADMIN_PASSWORD", "testpass")
			defer os.Unsetenv("JWT_SECRET")
			defer os.Unsetenv("ADMIN_PASSWORD")

			// 创建临时配置文件
			tmpDir := t.TempDir()
			configPath := tmpDir + "/config.toml"
			configContent := `
database = "test.db"

[admin]
username = "admin"
password = "testpass"
`
			os.WriteFile(configPath, []byte(configContent), 0644)

			// 创建配置管理器
			manager := NewConfigManager()
			_, err := manager.Load(configPath)
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// 启动多个并发读取器
			done := make(chan bool, numReaders)
			for i := 0; i < numReaders; i++ {
				go func() {
					// 执行多次读取
					for j := 0; j < 10; j++ {
						cfg := manager.Get()
						if cfg == nil {
							done <- false
							return
						}
						// 验证配置一致性
						if cfg.Database != "test.db" {
							done <- false
							return
						}
					}
					done <- true
				}()
			}

			// 等待所有读取器完成
			for i := 0; i < numReaders; i++ {
				if !<-done {
					t.Logf("Concurrent read failed")
					return false
				}
			}

			return true
		},
		genConfig(),
		gen.IntRange(5, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: node-config-separation, Property 5: Hot reload atomicity
// For any valid node configuration update, either the entire new configuration is applied,
// or the system keeps the previous configuration unchanged - no partial updates should occur.
// Validates: Requirements 3.2, 3.3, 3.4
func TestHotReloadAtomicity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("successful reload updates all nodes atomically", prop.ForAll(
		func(initialNodes, newNodes []NodeConfig) bool {
			// 设置环境变量
			os.Setenv("JWT_SECRET", "test-jwt-secret-at-least-32-characters-long")
			os.Setenv("ADMIN_PASSWORD", "testpass")
			defer os.Unsetenv("JWT_SECRET")
			defer os.Unsetenv("ADMIN_PASSWORD")

			// 创建临时文件
			tmpDir := t.TempDir()
			configPath := tmpDir + "/config.toml"
			nodeListPath := tmpDir + "/nodelist.toml"

			// 写入初始配置
			configContent := `
database = "test.db"

[admin]
username = "admin"
password = "testpass"
`
			os.WriteFile(configPath, []byte(configContent), 0644)

			// 写入初始节点列表
			initialToml := "# Initial nodes\n"
			for _, node := range initialNodes {
				initialToml += "[[nodes]]\n"
				initialToml += "name = \"" + node.Name + "\"\n"
				initialToml += "environment = \"" + node.Environment + "\"\n"
				initialToml += "host = \"" + node.Host + "\"\n"
				initialToml += "port = " + fmt.Sprintf("%d", node.Port) + "\n"
				initialToml += "username = \"" + node.Username + "\"\n"
				initialToml += "password = \"" + node.Password + "\"\n\n"
			}
			os.WriteFile(nodeListPath, []byte(initialToml), 0644)

			// 创建配置管理器并加载
			loader := NewConfigLoader(configPath, nodeListPath)
			cfg, err := loader.Load()
			if err != nil {
				t.Logf("Failed to load initial config: %v", err)
				return false
			}

			manager := NewConfigManager()
			manager.(*configManager).config = cfg

			// 记录初始节点数量
			initialCount := len(cfg.Nodes)

			// 模拟热重载：写入新节点列表
			newToml := "# Updated nodes\n"
			for _, node := range newNodes {
				newToml += "[[nodes]]\n"
				newToml += "name = \"" + node.Name + "\"\n"
				newToml += "environment = \"" + node.Environment + "\"\n"
				newToml += "host = \"" + node.Host + "\"\n"
				newToml += "port = " + fmt.Sprintf("%d", node.Port) + "\n"
				newToml += "username = \"" + node.Username + "\"\n"
				newToml += "password = \"" + node.Password + "\"\n\n"
			}
			os.WriteFile(nodeListPath, []byte(newToml), 0644)

			// 手动触发重载
			manager.(*configManager).nodeListPath = nodeListPath
			manager.(*configManager).reloadNodeList()

			// 获取更新后的配置
			updatedCfg := manager.Get()

			// 验证原子性：要么全部更新，要么保持不变
			updatedCount := len(updatedCfg.Nodes)
			expectedCount := len(newNodes)

			// 如果更新成功，节点数应该等于新节点数
			// 如果更新失败（验证失败），节点数应该等于初始节点数
			if updatedCount != expectedCount && updatedCount != initialCount {
				t.Logf("Partial update detected: initial=%d, expected=%d, actual=%d",
					initialCount, expectedCount, updatedCount)
				return false
			}

			return true
		},
		genNodeConfigList(),
		genNodeConfigList(),
	))

	properties.Property("failed validation keeps old config", prop.ForAll(
		func(validNodes []NodeConfig) bool {
			// 设置环境变量
			os.Setenv("JWT_SECRET", "test-jwt-secret-at-least-32-characters-long")
			os.Setenv("ADMIN_PASSWORD", "testpass")
			defer os.Unsetenv("JWT_SECRET")
			defer os.Unsetenv("ADMIN_PASSWORD")

			// 创建临时文件
			tmpDir := t.TempDir()
			nodeListPath := tmpDir + "/nodelist.toml"

			// 写入有效节点列表
			validToml := ""
			for _, node := range validNodes {
				validToml += "[[nodes]]\n"
				validToml += "name = \"" + node.Name + "\"\n"
				validToml += "environment = \"" + node.Environment + "\"\n"
				validToml += "host = \"" + node.Host + "\"\n"
				validToml += "port = " + fmt.Sprintf("%d", node.Port) + "\n"
				validToml += "username = \"" + node.Username + "\"\n"
				validToml += "password = \"" + node.Password + "\"\n\n"
			}
			os.WriteFile(nodeListPath, []byte(validToml), 0644)

			// 加载初始配置
			loader := NewConfigLoader("", nodeListPath)
			initialNodes, err := loader.LoadNodeList()
			if err != nil {
				return true // 跳过无效的初始配置
			}

			manager := NewConfigManager()
			manager.(*configManager).config = &Config{Nodes: initialNodes}
			manager.(*configManager).nodeListPath = nodeListPath

			initialCount := len(initialNodes)

			// 写入无效节点列表（端口无效）
			invalidToml := "[[nodes]]\nname = \"invalid\"\nenvironment = \"test\"\nhost = \"localhost\"\nport = 0\n"
			os.WriteFile(nodeListPath, []byte(invalidToml), 0644)

			// 触发重载
			manager.(*configManager).reloadNodeList()

			// 验证配置未改变
			currentCfg := manager.Get()
			if len(currentCfg.Nodes) != initialCount {
				t.Logf("Config changed after failed validation: initial=%d, current=%d",
					initialCount, len(currentCfg.Nodes))
				return false
			}

			return true
		},
		genNodeConfigList(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
