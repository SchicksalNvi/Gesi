# 设计文档

## 概述

本设计文档描述了 Go-CESI 项目的优化方案。Go-CESI 是一个用 Go 语言实现的 Supervisor 集中管理界面，提供 Web UI 来便捷控制和监控多个 Supervisor 实例。

优化的核心目标是：
- 提高代码质量和可维护性
- 增强系统的可靠性和稳定性
- 优化性能和资源使用
- 加强安全性
- 完善测试覆盖

本优化遵循 Go 语言最佳实践和 Linus Torvalds 的工程哲学：简洁、清晰、实用。

## 架构

### 当前架构

Go-CESI 采用经典的三层架构：

```
┌─────────────────────────────────────┐
│         Web Frontend (React)        │
│  - Dashboard, Nodes, Users, Logs    │
└─────────────────────────────────────┘
                 │ HTTP/WebSocket
┌─────────────────────────────────────┐
│         API Layer (Gin)             │
│  - REST API, WebSocket Hub          │
│  - Authentication, Authorization    │
└─────────────────────────────────────┘
                 │
┌─────────────────────────────────────┐
│       Business Logic Layer          │
│  - SupervisorService                │
│  - AuthService, ActivityLogService  │
└─────────────────────────────────────┘
                 │
┌─────────────────────────────────────┐
│         Data Layer                  │
│  - GORM (SQLite)                    │
│  - Supervisor XML-RPC Client        │
└─────────────────────────────────────┘
```

### 优化后的架构改进

1. **统一的错误处理层**：引入标准化的错误类型和处理机制
2. **配置管理中心化**：统一配置加载、验证和热重载
3. **资源生命周期管理**：明确的启动、运行、关闭流程
4. **可观测性增强**：结构化日志、指标收集、健康检查

## 组件和接口

### 1. 错误处理模块 (internal/errors)

**职责**：提供统一的错误类型和错误处理机制

**接口设计**：

```go
// AppError 应用错误接口
type AppError interface {
    error
    Code() string          // 错误代码
    Message() string       // 用户友好的错误消息
    Details() interface{}  // 详细错误信息
    Cause() error         // 原始错误
    WithContext(key string, value interface{}) AppError
}

// 错误类型构造函数
func NewValidationError(field string, message string) AppError
func NewNotFoundError(resource string, id string) AppError
func NewConflictError(resource string, message string) AppError
func NewInternalError(message string, cause error) AppError
func NewUnauthorizedError(message string) AppError
```

**设计决策**：
- 使用接口而非具体类型，便于扩展
- 支持错误链，保留原始错误信息
- 支持上下文信息附加，便于调试
- 错误代码标准化，便于前端处理

### 2. 配置管理模块 (internal/config)

**职责**：统一管理应用配置，支持验证和热重载

**接口设计**：

```go
// ConfigManager 配置管理器接口
type ConfigManager interface {
    Load(path string) (*Config, error)
    Validate(cfg *Config) error
    Watch(callback func(*Config)) error
    Get() *Config
}

// Config 应用配置
type Config struct {
    Server      ServerConfig
    Database    DatabaseConfig
    Auth        AuthConfig
    Nodes       []NodeConfig
    Performance PerformanceConfig
    Logging     LoggingConfig
}

// Validator 配置验证器
type Validator interface {
    Validate(cfg *Config) []ValidationError
}
```

**设计决策**：
- 配置结构化，使用强类型
- 敏感信息从环境变量读取
- 支持配置验证，启动前检查
- 支持热重载，无需重启服务

### 3. 资源管理模块 (internal/lifecycle)

**职责**：管理应用生命周期和资源清理

**接口设计**：

```go
// Lifecycle 生命周期管理接口
type Lifecycle interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() HealthStatus
}

// Manager 资源管理器
type Manager struct {
    components []Lifecycle
}

func (m *Manager) Register(component Lifecycle)
func (m *Manager) StartAll(ctx context.Context) error
func (m *Manager) StopAll(ctx context.Context) error
func (m *Manager) HealthCheck() map[string]HealthStatus
```

**设计决策**：
- 统一的生命周期接口
- 支持优雅关闭，带超时控制
- 组件注册机制，自动管理依赖
- 健康检查集成

### 4. 数据库访问层优化 (internal/database)

**当前问题**：
- 缺少查询超时控制
- 没有连接池监控
- 事务管理不统一

**优化方案**：

```go
// Repository 仓储接口
type Repository interface {
    WithContext(ctx context.Context) Repository
    WithTransaction(fn func(Repository) error) error
}

// QueryBuilder 查询构建器
type QueryBuilder interface {
    Where(query interface{}, args ...interface{}) QueryBuilder
    Limit(limit int) QueryBuilder
    Offset(offset int) QueryBuilder
    Order(value interface{}) QueryBuilder
    Find(dest interface{}) error
}

// 分页支持
type Pagination struct {
    Page     int
    PageSize int
    Total    int64
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB
```

**设计决策**：
- 所有查询使用 context，支持超时和取消
- 统一的事务管理接口
- 分页查询标准化
- 连接池监控和健康检查

### 5. API 响应标准化 (internal/api)

**当前问题**：
- 响应格式不统一
- 错误处理分散

**优化方案**：

```go
// Response 标准响应格式
type Response struct {
    Status  string      `json:"status"`  // "success" or "error"
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// 响应辅助函数
func Success(c *gin.Context, data interface{})
func SuccessWithMessage(c *gin.Context, message string, data interface{})
func Error(c *gin.Context, err error)
func ValidationError(c *gin.Context, errors []ValidationError)
```

**设计决策**：
- 统一的响应格式，便于前端处理
- 错误信息结构化
- HTTP 状态码与业务错误分离
- 支持国际化（预留）

### 6. 日志系统增强 (internal/logger)

**当前问题**：
- 日志格式不统一
- 缺少上下文信息
- 日志级别固定

**优化方案**：

```go
// Logger 日志接口
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    With(fields ...Field) Logger
    WithContext(ctx context.Context) Logger
}

// Field 日志字段
type Field struct {
    Key   string
    Value interface{}
}

// 上下文日志
func FromContext(ctx context.Context) Logger
func WithRequestID(ctx context.Context, requestID string) context.Context
```

**设计决策**：
- 使用 zap 作为底层日志库
- 结构化日志，便于查询和分析
- 支持上下文传递，关联请求日志
- 动态日志级别调整

## 数据模型

### 错误模型

```go
type appError struct {
    code    string
    message string
    details interface{}
    cause   error
    context map[string]interface{}
}
```

### 配置模型

```go
type Config struct {
    Server struct {
        Port            int
        ReadTimeout     time.Duration
        WriteTimeout    time.Duration
        ShutdownTimeout time.Duration
    }
    
    Database struct {
        Path            string
        MaxOpenConns    int
        MaxIdleConns    int
        ConnMaxLifetime time.Duration
    }
    
    Auth struct {
        JWTSecret       string
        TokenExpiration time.Duration
    }
    
    Nodes []struct {
        Name        string
        Environment string
        Host        string
        Port        int
        Username    string
        Password    string
    }
}
```

### 健康状态模型

```go
type HealthStatus struct {
    Status    string                 // "healthy", "degraded", "unhealthy"
    Timestamp time.Time
    Details   map[string]interface{}
}
```

## 正确性属性

*属性是一个特征或行为，应该在系统的所有有效执行中保持为真。属性是人类可读规范和机器可验证正确性保证之间的桥梁。*


### 错误处理属性

**属性 1：错误上下文完整性**
*对于任何*系统错误，返回的错误对象应包含错误代码、消息和原始错误（如果存在）
**验证：需求 2.1**

**属性 2：可恢复错误不终止**
*对于任何*可恢复的错误（如网络超时、临时数据库错误），系统应记录错误并继续执行，而不是调用 Fatal 终止程序
**验证：需求 2.2**

**属性 3：验证错误详细性**
*对于任何*输入验证失败，返回的错误应包含失败的字段名称和具体原因
**验证：需求 2.3**

**属性 4：外部调用重试**
*对于任何*外部服务调用失败，系统应在配置的重试次数内重试，并在超时后返回错误
**验证：需求 2.4**

**属性 5：Panic 恢复**
*对于任何*可能发生 panic 的操作，系统应使用 recover 捕获 panic 并转换为错误返回
**验证：需求 2.5**

### 配置管理属性

**属性 6：必需配置验证**
*对于任何*配置加载操作，如果缺少必需的配置项（如 JWT_SECRET），系统应返回明确的验证错误
**验证：需求 3.1**

**属性 7：配置格式错误提示**
*对于任何*格式错误的配置文件（如 TOML 语法错误），系统应返回包含错误位置和原因的错误消息
**验证：需求 3.3**

**属性 8：默认值日志记录**
*对于任何*使用默认值的配置项，系统应在日志中记录该配置项的名称和默认值
**验证：需求 3.5**

### 资源管理属性

**属性 9：Goroutine 生命周期**
*对于任何*启动的 goroutine，应有对应的退出机制（通过 context 或 channel），并在系统关闭时正确退出
**验证：需求 4.1**

**属性 10：并发限制**
*对于任何*并发请求处理，系统应使用工作池或信号量限制并发数量，不超过配置的最大值
**验证：需求 4.3**

**属性 11：超时控制**
*对于任何*长时间运行的操作（如数据库查询、HTTP 请求），应使用 context.WithTimeout 设置超时，并在超时后取消操作
**验证：需求 4.5**

### 数据库操作属性

**属性 12：预编译语句使用**
*对于任何*数据库查询操作，系统应使用 GORM 的参数化查询或预编译语句，而不是字符串拼接
**验证：需求 5.1**

**属性 13：事务原子性**
*对于任何*批量数据库操作，如果任一操作失败，整个事务应回滚，数据库状态应保持一致
**验证：需求 5.2**

**属性 14：并发更新一致性**
*对于任何*并发数据更新操作，使用乐观锁或悲观锁后，最终数据状态应反映所有成功的更新，不应出现丢失更新
**验证：需求 5.4**

### API 设计属性

**属性 15：响应格式统一**
*对于任何*API 端点的成功响应，应包含 status、message 和 data 字段，格式符合标准响应结构
**验证：需求 6.2**

**属性 16：错误响应标准化**
*对于任何*API 错误响应，应包含 status、error 对象（含 code 和 message），HTTP 状态码应与错误类型匹配
**验证：需求 6.3**

**属性 17：输入验证完整性**
*对于任何*API 请求，所有必需参数应被验证，无效参数应返回 400 错误和详细的验证信息
**验证：需求 6.4**

**属性 18：权限验证**
*对于任何*需要权限的 API 操作，未授权用户应收到 403 错误，操作不应被执行
**验证：需求 6.5**

### 日志系统属性

**属性 19：结构化日志格式**
*对于任何*日志记录，应包含时间戳、级别、消息和结构化字段（如 request_id、user_id）
**验证：需求 7.1**

**属性 20：错误日志堆栈**
*对于任何*错误级别的日志，如果包含错误对象，应记录错误的堆栈跟踪信息
**验证：需求 7.2**

**属性 21：操作日志完整性**
*对于任何*用户操作日志，应包含用户 ID、操作类型、操作对象和操作时间
**验证：需求 7.3**

### 安全性属性

**属性 22：输入清理**
*对于任何*用户输入，在处理前应进行验证和清理，防止 XSS、SQL 注入等攻击
**验证：需求 9.3**

**属性 23：敏感信息保护**
*对于任何*日志记录，不应包含密码、JWT 令牌等敏感信息的明文
**验证：需求 9.5**

### 性能属性

**属性 24：响应时间**
*对于任何*不涉及外部调用的 HTTP 请求，响应时间应在 100ms 以内（P95）
**验证：需求 10.1**

**属性 25：缓存一致性**
*对于任何*使用缓存的数据，缓存失效后应从数据源重新加载，确保数据一致性
**验证：需求 10.3**

**属性 26：WebSocket 消息分发效率**
*对于任何*WebSocket 广播消息，应使用并发分发机制，单个客户端的慢速不应阻塞其他客户端
**验证：需求 10.4**

**属性 27：资源释放**
*对于任何*长时间空闲的资源（如数据库连接、HTTP 连接），应在配置的超时后自动释放
**验证：需求 10.5**

## 错误处理

### 错误分类

1. **验证错误** (ValidationError)
   - HTTP 状态码：400
   - 场景：输入参数不合法
   - 处理：返回详细的验证错误信息

2. **未找到错误** (NotFoundError)
   - HTTP 状态码：404
   - 场景：请求的资源不存在
   - 处理：返回资源类型和 ID

3. **冲突错误** (ConflictError)
   - HTTP 状态码：409
   - 场景：资源已存在或状态冲突
   - 处理：返回冲突原因

4. **未授权错误** (UnauthorizedError)
   - HTTP 状态码：401
   - 场景：未登录或令牌无效
   - 处理：清除令牌，重定向到登录页

5. **禁止错误** (ForbiddenError)
   - HTTP 状态码：403
   - 场景：权限不足
   - 处理：返回权限要求

6. **内部错误** (InternalError)
   - HTTP 状态码：500
   - 场景：系统内部错误
   - 处理：记录详细错误日志，返回通用错误消息

### 错误处理流程

```
请求 → 参数验证 → 业务逻辑 → 数据访问
  ↓         ↓          ↓          ↓
验证错误   业务错误   数据错误   系统错误
  ↓         ↓          ↓          ↓
错误转换 → 日志记录 → 响应构建 → 返回客户端
```

### 错误恢复策略

1. **重试机制**
   - 适用：临时性错误（网络超时、数据库连接失败）
   - 策略：指数退避重试，最多 3 次
   - 超时：每次重试超时递增

2. **降级处理**
   - 适用：非关键功能失败
   - 策略：返回缓存数据或默认值
   - 告警：记录降级事件

3. **熔断机制**
   - 适用：外部服务持续失败
   - 策略：快速失败，避免级联故障
   - 恢复：定期探测服务恢复

## 测试策略

### 单元测试

**目标**：覆盖核心业务逻辑和工具函数

**重点模块**：
- 错误处理模块：测试各种错误类型的创建和转换
- 配置管理模块：测试配置加载、验证和默认值
- 数据库仓储：测试 CRUD 操作和事务
- API 响应构建：测试响应格式和错误转换

**测试框架**：
- 使用 Go 标准库 `testing`
- 使用 `testify` 进行断言
- 使用表驱动测试模式

**示例**：

```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            config: &Config{
                Server: ServerConfig{Port: 8080},
                Auth:   AuthConfig{JWTSecret: "secret-key-32-characters-long"},
            },
            wantErr: false,
        },
        {
            name: "missing JWT secret",
            config: &Config{
                Server: ServerConfig{Port: 8080},
            },
            wantErr: true,
            errMsg:  "JWT_SECRET is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(tt.config)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 属性测试

**目标**：验证系统的通用属性和不变量

**测试库**：使用 `gopter` 进行属性测试

**重点属性**：
- 错误处理：所有错误都包含上下文信息
- 配置验证：无效配置总是被拒绝
- 事务原子性：失败的事务总是回滚
- API 响应：所有响应符合标准格式

**示例**：

```go
func TestErrorContextProperty(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("all errors contain context", prop.ForAll(
        func(msg string, code string) bool {
            err := NewInternalError(msg, nil).
                WithContext("code", code)
            
            appErr, ok := err.(AppError)
            if !ok {
                return false
            }
            
            return appErr.Code() != "" &&
                   appErr.Message() != "" &&
                   appErr.Details() != nil
        },
        gen.AnyString(),
        gen.AnyString(),
    ))
    
    properties.TestingRun(t)
}
```

### 集成测试

**目标**：测试组件间的交互和端到端流程

**测试场景**：
- API 端点测试：测试完整的请求-响应流程
- 数据库集成：测试数据持久化和查询
- WebSocket 通信：测试实时消息推送
- 认证授权：测试登录和权限验证

**测试环境**：
- 使用内存 SQLite 数据库
- 使用 `httptest` 模拟 HTTP 服务器
- 使用 `gorilla/websocket` 测试工具

**示例**：

```go
func TestLoginAPI(t *testing.T) {
    // 设置测试环境
    db := setupTestDB(t)
    defer db.Close()
    
    router := setupTestRouter(db)
    
    // 创建测试用户
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
    }
    user.SetPassword("password123")
    db.Create(user)
    
    // 测试登录
    w := httptest.NewRecorder()
    body := `{"username":"testuser","password":"password123"}`
    req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Equal(t, "success", response["status"])
    assert.NotEmpty(t, response["data"].(map[string]interface{})["token"])
}
```

### 性能测试

**目标**：验证系统性能指标

**测试工具**：
- 使用 `testing.B` 进行基准测试
- 使用 `pprof` 进行性能分析
- 使用 `vegeta` 进行负载测试

**测试指标**：
- API 响应时间（P50, P95, P99）
- 数据库查询性能
- WebSocket 消息吞吐量
- 内存使用和 GC 压力

**示例**：

```go
func BenchmarkAPIEndpoint(b *testing.B) {
    router := setupTestRouter(nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        req, _ := http.NewRequest("GET", "/api/nodes", nil)
        router.ServeHTTP(w, req)
    }
}
```

### 测试覆盖率目标

- 核心业务逻辑：>= 80%
- API 处理器：>= 70%
- 工具函数：>= 90%
- 整体覆盖率：>= 75%

### 持续集成

- 每次提交自动运行单元测试
- 每日运行完整测试套件（包括集成测试）
- 每周运行性能测试和负载测试
- 测试失败阻止合并

## 实施计划

### 阶段 1：基础设施优化（优先级：高）

1. **错误处理标准化**
   - 创建统一的错误类型
   - 重构现有错误处理代码
   - 添加错误处理测试

2. **配置管理改进**
   - 统一配置加载逻辑
   - 添加配置验证
   - 实现配置热重载

3. **日志系统增强**
   - 标准化日志格式
   - 添加上下文传递
   - 实现动态日志级别

### 阶段 2：核心功能优化（优先级：高）

1. **数据库访问层改进**
   - 添加查询超时控制
   - 统一事务管理
   - 实现连接池监控

2. **API 响应标准化**
   - 统一响应格式
   - 改进错误响应
   - 添加 API 测试

3. **资源生命周期管理**
   - 实现优雅关闭
   - 添加健康检查
   - 改进 goroutine 管理

### 阶段 3：性能和安全优化（优先级：中）

1. **性能优化**
   - 添加缓存机制
   - 优化数据库查询
   - 改进 WebSocket 性能

2. **安全增强**
   - 加强输入验证
   - 改进密码存储
   - 添加安全审计

3. **测试完善**
   - 添加单元测试
   - 实现属性测试
   - 完善集成测试

### 阶段 4：可观测性和运维（优先级：中）

1. **监控和指标**
   - 添加性能指标收集
   - 实现健康检查端点
   - 集成监控系统

2. **文档完善**
   - API 文档
   - 部署文档
   - 运维手册

## 技术选型

### 核心依赖

- **Web 框架**：Gin v1.10+
- **ORM**：GORM v1.30+
- **日志**：zap v1.27+
- **配置**：viper v1.20+
- **JWT**：golang-jwt/jwt v4.5+
- **WebSocket**：gorilla/websocket v1.5+

### 测试依赖

- **断言**：testify v1.9+
- **属性测试**：gopter v0.2+
- **Mock**：gomock v1.6+
- **HTTP 测试**：httptest (标准库)

### 工具

- **代码检查**：golangci-lint
- **代码格式化**：gofmt, goimports
- **性能分析**：pprof
- **负载测试**：vegeta

## 迁移策略

### 向后兼容性

- API 端点保持不变
- 数据库模式兼容
- 配置文件向后兼容（支持旧格式）

### 渐进式迁移

1. 新功能使用新的错误处理和响应格式
2. 逐步重构现有代码
3. 保持测试覆盖，确保不破坏现有功能
4. 使用特性开关控制新功能启用

### 回滚计划

- 保留旧代码分支
- 数据库迁移可回滚
- 配置文件兼容旧版本
- 监控关键指标，异常时快速回滚

## 前端优化

### 当前问题

1. **启动速度慢**
   - 开发服务器启动时间过长
   - 大量依赖导致编译缓慢
   - ESLint 检查拖慢启动

2. **构建性能差**
   - 生产构建时间过长
   - 未优化的 webpack 配置
   - 未使用代码分割

3. **错误处理混乱**
   - console.error/warn 散落各处
   - 缺少统一的错误处理
   - 错误信息不友好

4. **冗余代码**
   - 旧的静态模板文件（web/templates/）
   - 旧的静态 JS 文件（web/static/js/）
   - 未使用的依赖包

### 优化方案

#### 1. 构建性能优化

**Webpack 配置优化**：

```javascript
// craco.config.js
module.exports = {
  webpack: {
    configure: (webpackConfig, { env }) => {
      if (env === 'development') {
        // 使用更快的 source map
        webpackConfig.devtool = 'eval-cheap-module-source-map';
        
        // 减少文件监听开销
        webpackConfig.watchOptions = {
          ignored: /node_modules/,
          aggregateTimeout: 300,
          poll: false
        };
        
        // 优化模块解析
        webpackConfig.resolve.modules = ['node_modules'];
        webpackConfig.resolve.extensions = ['.js', '.jsx'];
      }
      
      if (env === 'production') {
        // 代码分割
        webpackConfig.optimization = {
          ...webpackConfig.optimization,
          splitChunks: {
            chunks: 'all',
            cacheGroups: {
              vendor: {
                test: /[\\/]node_modules[\\/]/,
                name: 'vendors',
                priority: 10
              },
              common: {
                minChunks: 2,
                priority: 5,
                reuseExistingChunk: true
              }
            }
          }
        };
      }
      
      return webpackConfig;
    }
  },
  devServer: {
    compress: false, // 开发环境不压缩
    hot: true,
    client: {
      overlay: {
        errors: true,
        warnings: false // 不显示警告
      }
    }
  },
  eslint: {
    enable: false // 开发时禁用 ESLint
  }
};
```

**依赖优化**：

- 移除未使用的依赖
- 使用更轻量的替代品
- 延迟加载非关键依赖

#### 2. 错误处理标准化

**创建统一的错误处理服务**：

```javascript
// src/services/errorHandler.js
class ErrorHandler {
  constructor() {
    this.logger = console; // 可替换为专业日志服务
  }
  
  handleError(error, context = {}) {
    const errorInfo = {
      message: error.message || 'Unknown error',
      stack: error.stack,
      context,
      timestamp: new Date().toISOString()
    };
    
    // 开发环境输出详细信息
    if (process.env.NODE_ENV === 'development') {
      this.logger.error('Error occurred:', errorInfo);
    }
    
    // 生产环境发送到监控服务
    if (process.env.NODE_ENV === 'production') {
      this.sendToMonitoring(errorInfo);
    }
    
    return this.getUserFriendlyMessage(error);
  }
  
  getUserFriendlyMessage(error) {
    if (error.response) {
      // API 错误
      return error.response.data?.message || 'Server error occurred';
    }
    if (error.request) {
      // 网络错误
      return 'Network error. Please check your connection.';
    }
    return error.message || 'An unexpected error occurred';
  }
  
  sendToMonitoring(errorInfo) {
    // 发送到监控服务（如 Sentry）
    // 这里只是占位符
  }
}

export const errorHandler = new ErrorHandler();
```

**在 API 服务中使用**：

```javascript
// src/services/api.js
import { errorHandler } from './errorHandler';

export const fetchNodes = async () => {
  try {
    const response = await axios.get('/api/nodes');
    return response.data;
  } catch (error) {
    const message = errorHandler.handleError(error, {
      action: 'fetchNodes',
      endpoint: '/api/nodes'
    });
    throw new Error(message);
  }
};
```

#### 3. 代码清理

**移除旧文件**：

- `web/templates/` - 旧的 HTML 模板
- `web/static/js/` - 旧的 JavaScript 文件
- `web/static/css/style.css` - 旧的样式文件（如果未使用）

**清理路由配置**：

```go
// cmd/main.go
// 移除旧模板路由
// protectedGroup := router.Group("/legacy", authService.AuthMiddleware())
// { ... }

// 只保留 React 应用路由
router.NoRoute(func(c *gin.Context) {
    if strings.HasPrefix(c.Request.URL.Path, "/api") {
        c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
        return
    }
    indexPath := filepath.Join(projectRoot, "web", "react-app", "build", "index.html")
    c.File(indexPath)
})
```

**清理未使用的依赖**：

```bash
# 分析未使用的依赖
npx depcheck

# 移除未使用的包
npm uninstall <unused-package>
```

#### 4. 性能监控

**添加性能指标收集**：

```javascript
// src/utils/performance.js
export const measurePageLoad = () => {
  if (window.performance && window.performance.timing) {
    const timing = window.performance.timing;
    const loadTime = timing.loadEventEnd - timing.navigationStart;
    const domReady = timing.domContentLoadedEventEnd - timing.navigationStart;
    
    console.log('Page Load Time:', loadTime, 'ms');
    console.log('DOM Ready Time:', domReady, 'ms');
    
    // 发送到监控服务
    sendMetrics({
      pageLoadTime: loadTime,
      domReadyTime: domReady
    });
  }
};
```

### 前端架构改进

**组件结构优化**：

```
src/
├── components/
│   ├── common/          # 通用组件
│   │   ├── Button/
│   │   ├── Card/
│   │   └── Table/
│   ├── layout/          # 布局组件
│   │   ├── Header/
│   │   ├── Sidebar/
│   │   └── Footer/
│   └── features/        # 功能组件
│       ├── NodeList/
│       ├── ProcessControl/
│       └── UserManagement/
├── pages/               # 页面组件
├── services/            # API 服务
│   ├── api.js
│   ├── errorHandler.js
│   └── websocket.js
├── hooks/               # 自定义 Hooks
│   ├── useNodes.js
│   ├── useAuth.js
│   └── useWebSocket.js
├── utils/               # 工具函数
│   ├── format.js
│   ├── validation.js
│   └── performance.js
└── contexts/            # React Context
    ├── AuthContext.js
    └── WebSocketContext.js
```

## 风险和缓解

### 风险 1：重构引入新 Bug

**缓解措施**：
- 完善的测试覆盖
- 代码审查
- 渐进式迁移
- 充分的测试环境验证

### 风险 2：性能回退

**缓解措施**：
- 性能基准测试
- 负载测试对比
- 性能监控
- 性能优化迭代

### 风险 3：学习曲线

**缓解措施**：
- 详细的文档
- 代码示例
- 团队培训
- 逐步推广

### 风险 4：前端构建失败

**缓解措施**：
- 保留旧构建配置作为备份
- 渐进式更新依赖
- 充分测试构建流程
- 准备回滚方案

## 成功指标

### 后端指标
- 代码覆盖率达到 75%+
- API 响应时间 P95 < 100ms
- 系统稳定性 99.9%+
- 错误处理标准化率 100%
- 配置验证覆盖率 100%
- 资源泄漏事件 = 0

### 前端指标
- 开发服务器启动时间 < 30 秒
- 生产构建时间 < 2 分钟
- 首次内容绘制 (FCP) < 1.5 秒
- 最大内容绘制 (LCP) < 2.5 秒
- 首次输入延迟 (FID) < 100ms
- 累积布局偏移 (CLS) < 0.1
- 打包体积减少 30%+
- 移除所有旧代码文件
