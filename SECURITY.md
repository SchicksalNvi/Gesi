# 安全配置指南

## 已修复的安全问题

### 1. JWT密钥安全
- **问题**: JWT密钥硬编码在代码中
- **修复**: 改为从环境变量 `JWT_SECRET` 读取
- **配置**: 复制 `.env.example` 为 `.env` 并设置强密钥

```bash
cp .env.example .env
# 编辑 .env 文件，设置强JWT密钥
```

### 2. 密码安全
- **问题**: 配置文件中使用弱密码
- **修复**: 更新为强密码
- **建议**: 定期更换密码，使用密码管理器

### 3. 权限控制
- **问题**: 查看敏感信息缺少权限检查
- **修复**: 实现了权限验证机制
- **权限**: 
  - `config:view_secret` - 查看敏感配置
  - `env_var:view_secret` - 查看敏感环境变量

### 4. 日志安全
- **问题**: 明文密码记录到日志
- **修复**: 移除敏感信息的日志记录
- **建议**: 定期审查日志内容

### 5. WebSocket安全
- **问题**: 允许所有来源的WebSocket连接
- **修复**: 限制为特定来源列表
- **配置**: 通过环境变量 `WEBSOCKET_ALLOWED_ORIGINS` 配置

## 生产环境安全建议

### 必须配置的环境变量
```bash
# 强JWT密钥（至少32字符）
JWT_SECRET=your-super-secure-jwt-secret-key-change-this-in-production

# 限制WebSocket来源
WEBSOCKET_ALLOWED_ORIGINS=https://yourdomain.com

# 禁用调试模式
DEBUG=false
```

### 文件权限
```bash
# 限制配置文件权限
chmod 600 config.toml
chmod 600 .env

# 限制数据库文件权限
chmod 600 data/cesi.db
```

### 网络安全
- 使用HTTPS
- 配置防火墙
- 定期更新依赖
- 启用访问日志

### 用户管理
- 定期审查用户权限
- 强制使用强密码
- 启用双因素认证（如果支持）
- 定期清理无效用户

## 安全审计

定期检查以下项目：
1. 用户权限分配
2. 敏感信息访问日志
3. 异常登录活动
4. 系统漏洞更新
5. 配置文件安全性

## 报告安全问题

如果发现安全问题，请通过以下方式报告：
- 邮箱: security@yourcompany.com
- 创建私有issue
- 直接联系管理员

**请勿在公开渠道披露安全漏洞**