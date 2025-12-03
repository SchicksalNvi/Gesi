# Design Document

## Overview

本设计实现节点配置与系统配置的分离，将节点列表从主配置文件 `config.toml` 迁移到独立的 `config/nodelist.toml` 文件。设计遵循以下原则：

1. **职责分离**：系统配置管理应用行为，节点配置管理数据源
2. **向后兼容**：现有配置格式继续有效，无需强制迁移
3. **独立热重载**：节点列表变更不影响系统配置
4. **零破坏性**：不改变现有 API 和数据结构

## Architecture

### 配置文件结构

```
config/
├── config.toml      # 系统配置（服务器、数据库、管理员、性能参数）
└── nodelist.toml    # 节点配置（节点列表）
```

### 配置加载流程

```
启动 → 加载 config/config.toml → 加载 config/nodelist.toml → 合并节点配置 → 验证 → 初始化服务
```

### 热重载机制

- **系统配置热重载**：监听 `config/config.toml`，触发完整配置回调
- **节点配置热重载**：监听 `config/nodelist.toml`，仅触发节点列表更新
- **独立监听**：两个文件使用独立的 fsnotify watcher

## Components and Interfaces

### 1. 配置结构（保持不变）

```go
type Config struct {
    Database         string
    ActivityLog      string
    AdminUsername    string
    AdminPassword    string
    Nodes            []NodeConfig  // 合并后的节点列表
    DeveloperTools   DeveloperToolsConfig
    Performance      PerformanceConfig
}

type NodeConfig struct {
    Name        string
    Environment string
    Host        string
    Port        int
    Username    string
    Password    string
}
```

### 2. 配置加载器

```go
// ConfigLoader 负责加载和合并配置
type ConfigLoader struct {
    mainConfigPath string
    nodeListPath   string
}

// Load 加载完整配置
func (l *ConfigLoader) Load() (*Config, error)

// LoadMainConfig 加载系统配置
func (l *ConfigLoader) LoadMainConfig() (*Config, error)

// LoadNodeList 加载节点列表
func (l *ConfigLoader) LoadNodeList() ([]NodeConfig, error)

// MergeNodes 合并节点配置（nodelist.toml 优先）
func (l *ConfigLoader) MergeNodes(mainNodes, nodeListNodes []NodeConfig) []NodeConfig
```

### 3. 配置管理器（扩展）

```go
type ConfigManager interface {
    Load(path string) (*Config, error)
    Validate(cfg *Config) error
    Watch(callback func(*Config)) error
    WatchNodeList(callback func([]NodeConfig)) error  // 新增
    Get() *Config
    Stop()
}
```

### 4. 节点服务集成

`SupervisorService` 无需修改，继续使用 `Config.Nodes` 字段。配置管理器负责在热重载时调用 `AddNode` 或移除节点。

## Data Models

### config/config.toml

```toml
[server]
port = 8081

[admin]
username = "admin"
password = "${ADMIN_PASSWORD}"
email = "${ADMIN_EMAIL}"

database = "data/cesi.db"

[performance]
memory_monitoring_enabled = true
memory_update_interval = "30s"

[developer_tools]
enabled = true

# 可选：向后兼容的节点配置
[[nodes]]
name = "legacy-node"
environment = "production"
host = "127.0.0.1"
port = 9001
username = "user"
password = "${NODE_PASSWORD}"
```

### config/nodelist.toml

```toml
[[nodes]]
name = "node-1"
environment = "production"
host = "192.168.1.10"
port = 9001
username = "supervisor"
password = "${NODE_PASSWORD}"

[[nodes]]
name = "node-2"
environment = "staging"
host = "192.168.1.11"
port = 9001
username = "supervisor"
password = "${NODE_PASSWORD}"

# ... 可扩展到上百个节点
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Node configuration parsing round-trip

*For any* valid node configuration structure, writing it to TOML format and parsing it back should produce an equivalent configuration.

**Validates: Requirements 1.2**

### Property 2: Node list merging with priority

*For any* two sets of node configurations (from config.toml and nodelist.toml), the merged result should contain all unique nodes, and for duplicate node names, the configuration from nodelist.toml should be used.

**Validates: Requirements 1.3, 1.4**

### Property 3: Environment variable expansion

*For any* configuration file containing environment variable references (e.g., `${VAR_NAME}`), after expansion, all references should be replaced with their corresponding environment variable values.

**Validates: Requirements 2.5**

### Property 4: Validation consistency across sources

*For any* node configuration, the validation result should be identical regardless of whether it comes from config.toml or nodelist.toml.

**Validates: Requirements 5.2**

### Property 5: Hot reload atomicity

*For any* valid node configuration update, either the entire new configuration is applied, or the system keeps the previous configuration unchanged - no partial updates should occur.

**Validates: Requirements 3.2, 3.3, 3.4**

### Property 6: Concurrent read safety

*For any* number of concurrent goroutines reading the configuration, no data races should occur and all reads should return consistent snapshots of the configuration.

**Validates: Requirements 5.5**

### Property 7: Error messages contain source information

*For any* configuration parsing error, the error message should contain the source file path and, when available, the line number where the error occurred.

**Validates: Requirements 5.3**

## Error Handling

### 配置加载错误

- **文件不存在**：`config/config.toml` 不存在时返回错误；`config/nodelist.toml` 不存在时使用空节点列表
- **解析错误**：返回包含文件路径和行号的详细错误信息
- **验证错误**：返回具体的验证失败原因（如缺少必填字段、端口范围错误）
- **环境变量未设置**：记录警告日志，使用空字符串作为默认值

### 热重载错误

- **文件监听失败**：记录错误日志，继续使用当前配置
- **重载验证失败**：保持当前配置不变，记录详细错误信息
- **回调执行失败**：记录错误但不影响配置更新

### 节点配置冲突

- **重复节点名**：使用 nodelist.toml 中的配置，记录警告日志
- **无效节点配置**：跳过该节点，记录错误，继续加载其他节点

## Testing Strategy

### Unit Testing

使用 Go 标准库 `testing` 包进行单元测试：

1. **配置加载测试**
   - 测试加载仅包含系统配置的 config.toml
   - 测试加载包含节点的 nodelist.toml
   - 测试 nodelist.toml 不存在时的向后兼容
   - 测试环境变量展开

2. **配置合并测试**
   - 测试两个文件都有节点时的合并
   - 测试节点名冲突时的优先级
   - 测试空节点列表的合并

3. **热重载测试**
   - 测试文件变更检测
   - 测试验证失败时保持原配置
   - 测试回调触发

4. **错误处理测试**
   - 测试无效 TOML 格式
   - 测试缺少必填字段
   - 测试无效端口号

### Property-Based Testing

使用 **gopter** 库进行属性测试（每个测试至少 100 次迭代）：

1. **Property 1: Node configuration parsing round-trip**
   - 生成随机节点配置
   - 序列化为 TOML
   - 反序列化并验证等价性

2. **Property 2: Node list merging with priority**
   - 生成两组随机节点配置
   - 执行合并
   - 验证所有节点存在且优先级正确

3. **Property 3: Environment variable expansion**
   - 生成随机配置和环境变量
   - 执行展开
   - 验证所有引用被替换

4. **Property 4: Validation consistency**
   - 生成随机节点配置
   - 从两个来源验证
   - 验证结果一致

5. **Property 5: Hot reload atomicity**
   - 生成随机配置更新
   - 模拟验证成功/失败
   - 验证配置状态的原子性

6. **Property 6: Concurrent read safety**
   - 启动多个 goroutine 并发读取
   - 使用 race detector 检测数据竞争
   - 验证读取的一致性

7. **Property 7: Error messages contain source information**
   - 生成随机的无效配置
   - 触发解析错误
   - 验证错误消息包含文件路径

每个属性测试必须：
- 使用注释标记对应的设计文档属性：`// Feature: node-config-separation, Property N: <property_text>`
- 配置至少 100 次迭代
- 使用智能生成器约束输入空间

### Integration Testing

1. **完整启动流程测试**
   - 测试从空目录启动
   - 测试从现有配置启动
   - 测试配置迁移场景

2. **SupervisorService 集成测试**
   - 测试节点服务正确接收配置
   - 测试热重载后节点列表更新

## Implementation Notes

### 文件路径约定

- 主配置文件：`config/config.toml`（向后兼容 `config.toml`）
- 节点列表文件：`config/nodelist.toml`
- 示例文件：`config/config.example.toml`, `config/nodelist.example.toml`

### 迁移策略

系统启动时检测配置格式：
1. 如果 `config.toml` 包含 `[[nodes]]` 且 `config/nodelist.toml` 不存在，记录迁移建议
2. 迁移建议包含：
   - 创建 `config/` 目录
   - 将节点配置移动到 `config/nodelist.toml`
   - 更新主配置路径为 `config/config.toml`

### 性能考虑

- 节点列表文件大小：支持上百个节点（预计 < 100KB）
- 热重载延迟：文件变更后 < 1 秒内完成重载
- 内存占用：节点配置结构轻量，100 个节点约 10KB 内存

### 向后兼容保证

- 现有 `config.toml` 格式完全兼容
- 不强制用户迁移配置
- API 和数据结构零变更

