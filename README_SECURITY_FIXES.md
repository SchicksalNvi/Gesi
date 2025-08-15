# 安全修复说明

本次修复解决了Go-CESI项目中的多个安全问题。以下是详细的修复内容和使用说明。

## 修复的安全问题

### 1. JWT密钥硬编码 ✅
**问题**: JWT密钥直接写在代码中，存在安全风险
**修复**: 改为从环境变量读取
**影响文件**: `internal/middleware/auth.go`

### 2. 弱密码配置 ✅
**问题**: 配置文件中使用简单密码
**修复**: 更新为强密码
**影响文件**: `config.toml`

### 3. 权限检查缺失 ✅
**问题**: 查看敏感信息时缺少权限验证
**修复**: 实现完整的权限检查机制
**影响文件**: 
- `internal/models/role.go` (新增权限常量)
- `internal/api/configuration.go` (实现权限检查)

### 4. 密码明文日志 ✅
**问题**: 密码相关操作记录明文到日志
**修复**: 移除敏感信息的日志记录
**影响文件**: `internal/models/user.go`

### 5. WebSocket跨域安全 ✅
**问题**: 允许所有来源的WebSocket连接
**修复**: 限制为特定来源列表
**影响文件**: `internal/websocket/hub.go`

## 快速开始

### 1. 环境配置
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，设置强JWT密钥
vim .env
```

### 2. 重置管理员密码（可选）
```bash
# 如果需要重置管理员密码
go run scripts/reset_admin_password.go "your_new_strong_password"
```

### 3. 编译和运行
```bash
# 安装依赖
go mod tidy

# 编译
go build -o go-cesi cmd/main.go

# 运行
./go-cesi
```

## 环境变量说明

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `JWT_SECRET` | 是 | 无 | JWT签名密钥，至少32字符 |
| `DB_PATH` | 否 | `data/cesi.db` | 数据库文件路径 |
| `SERVER_PORT` | 否 | `8081` | 服务器端口 |
| `WEBSOCKET_ALLOWED_ORIGINS` | 否 | 见配置 | WebSocket允许的来源 |
| `LOG_LEVEL` | 否 | `info` | 日志级别 |
| `DEBUG` | 否 | `false` | 是否启用调试模式 |

## 新增权限说明

### 权限常量
- `PermissionConfigViewSecret`: 查看敏感配置的权限
- `PermissionEnvVarViewSecret`: 查看敏感环境变量的权限

### API权限检查
- `GET /api/configurations/:id?showSecret=true`: 需要 `config:view_secret` 权限
- `GET /api/environment-variables/:id?showSecret=true`: 需要 `env_var:view_secret` 权限

## 安全最佳实践

### 生产环境部署
1. **强JWT密钥**: 使用至少32字符的随机字符串
2. **文件权限**: 限制配置文件和数据库文件的访问权限
3. **HTTPS**: 在生产环境中启用HTTPS
4. **防火墙**: 配置适当的防火墙规则
5. **定期更新**: 保持依赖库的最新版本

### 用户管理
1. **强密码策略**: 要求用户使用强密码
2. **权限最小化**: 只授予必要的权限
3. **定期审查**: 定期检查用户权限和访问日志
4. **账户清理**: 及时删除不再使用的账户

## 测试验证

### 1. 权限测试
```bash
# 测试无权限用户访问敏感信息
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8081/api/configurations/1?showSecret=true"
# 应该返回 403 Forbidden
```

### 2. JWT密钥测试
```bash
# 确保JWT密钥从环境变量读取
unset JWT_SECRET
./go-cesi
# 应该显示错误信息要求设置JWT_SECRET
```

### 3. WebSocket测试
```javascript
// 测试WebSocket连接限制
const ws = new WebSocket('ws://localhost:8081/ws');
// 只有允许的来源才能连接成功
```

## 故障排除

### 常见问题

1. **JWT_SECRET未设置**
   ```
   错误: JWT_SECRET environment variable is required
   解决: 在.env文件中设置JWT_SECRET
   ```

2. **权限不足**
   ```
   错误: 403 Forbidden
   解决: 确保用户具有相应的权限
   ```

3. **WebSocket连接失败**
   ```
   错误: WebSocket connection failed
   解决: 检查WEBSOCKET_ALLOWED_ORIGINS配置
   ```

## 更新日志

- **v1.0.1**: 修复JWT密钥硬编码问题
- **v1.0.2**: 实现权限检查机制
- **v1.0.3**: 修复密码日志泄露
- **v1.0.4**: 加强WebSocket安全
- **v1.0.5**: 更新配置文件密码

## 联系支持

如果在使用过程中遇到问题，请：
1. 查看 `SECURITY.md` 文件
2. 检查日志文件
3. 创建Issue报告问题
4. 联系技术支持团队

---

**重要提醒**: 在生产环境部署前，请务必完成所有安全配置，特别是JWT密钥和用户权限设置。