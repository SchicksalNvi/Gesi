package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors 多个验证错误
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator 验证器
type Validator struct {
	errors ValidationErrors
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// AddError 添加验证错误
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors 检查是否有错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors 获取所有错误
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// ValidateRequired 验证必填字段
func (v *Validator) ValidateRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
	}
}

// ValidateLength 验证字符串长度
func (v *Validator) ValidateLength(field, value string, min, max int) {
	length := len(strings.TrimSpace(value))
	if length < min {
		v.AddError(field, fmt.Sprintf("must be at least %d characters", min))
	}
	if max > 0 && length > max {
		v.AddError(field, fmt.Sprintf("must be at most %d characters", max))
	}
}

// ValidateEmail 验证邮箱格式
func (v *Validator) ValidateEmail(field, email string) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		v.AddError(field, "must be a valid email address")
	}
}

// ValidateAlphanumeric 验证字母数字字符
func (v *Validator) ValidateAlphanumeric(field, value string) {
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			v.AddError(field, "must contain only letters, numbers, underscores, and hyphens")
			return
		}
	}
}

// ValidateRange 验证数值范围
func (v *Validator) ValidateRange(field string, value, min, max int) {
	if value < min {
		v.AddError(field, fmt.Sprintf("must be at least %d", min))
	}
	if value > max {
		v.AddError(field, fmt.Sprintf("must be at most %d", max))
	}
}

// ValidatePositive 验证正数
func (v *Validator) ValidatePositive(field string, value int) {
	if value <= 0 {
		v.AddError(field, "must be a positive number")
	}
}

// ValidateID 验证ID格式（正整数）
func (v *Validator) ValidateID(field, value string) int {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
		return 0
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, "must be a valid number")
		return 0
	}

	if id <= 0 {
		v.AddError(field, "must be a positive number")
		return 0
	}

	return id
}

// ValidateNodeName 验证节点名称
func (v *Validator) ValidateNodeName(field, name string) {
	v.ValidateRequired(field, name)
	v.ValidateLength(field, name, 1, 100)
	v.ValidateAlphanumeric(field, name)
}

// ValidateProcessName 验证进程名称
func (v *Validator) ValidateProcessName(field, name string) {
	v.ValidateRequired(field, name)
	v.ValidateLength(field, name, 1, 200)
	// 进程名称允许更多字符，包括路径分隔符
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && 
		   r != '_' && r != '-' && r != '.' && r != '/' && r != '\\' && r != ':' {
			v.AddError(field, "contains invalid characters")
			return
		}
	}
}

// ValidateCommand 验证命令
func (v *Validator) ValidateCommand(field, command string) {
	v.ValidateRequired(field, command)
	v.ValidateLength(field, command, 1, 1000)
	
	// 限制命令长度，防止缓冲区溢出
	if len(command) > 2000 {
		v.AddError(field, "command too long (max 2000 characters)")
		return
	}
	
	// 检查危险命令和模式
	dangerousCommands := []string{
		// 文件系统破坏
		"rm -rf", "rm -fr", "rm -r", "rmdir /s", "del /f", "del /s",
		"format", "mkfs", "fdisk", "parted", "gparted",
		// 系统控制
		"shutdown", "reboot", "halt", "poweroff", "init 0", "init 6",
		"systemctl poweroff", "systemctl reboot", "systemctl halt",
		// 网络攻击
		"nc -l", "netcat -l", "ncat -l", "socat", "telnet",
		"wget", "curl", "lynx", "w3m",
		// 进程操作
		"kill -9", "killall", "pkill", "fuser -k",
		// 权限提升
		"sudo", "su -", "su root", "passwd", "chown", "chmod 777",
		"usermod", "useradd", "userdel", "groupadd", "groupdel",
		// 恶意代码
		":(){ :|:& };:", // fork bomb
		"/dev/tcp", "/dev/udp", "exec", "eval",
		"base64 -d", "xxd -r", "uudecode",
		// 数据泄露
		"dd if=", "cat /etc/passwd", "cat /etc/shadow",
		"find / -name", "locate", "which", "whereis",
		// 环境变量操作
		"export", "unset", "env", "printenv",
		// 包管理
		"apt install", "yum install", "dnf install", "pacman -S",
		"pip install", "npm install", "gem install",
		// 编译执行
		"gcc", "g++", "make", "cmake", "python -c", "perl -e",
		"ruby -e", "node -e", "php -r",
	}
	
	lowerCommand := strings.ToLower(command)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(lowerCommand, dangerous) {
			v.AddError(field, "contains potentially dangerous command")
			return
		}
	}
	
	// 检查危险字符模式
	dangerousPatterns := []string{
		`[;&|]`, // 命令分隔符
		"`", // 命令替换
		"$(", // 命令替换
		">", "<", "<<", ">>", // 重定向
		"*", "?", "[", // 通配符（可能导致意外文件操作）
	}
	
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, command); matched {
			v.AddError(field, "contains potentially dangerous characters or patterns")
			return
		}
	}
	
	// 检查是否尝试访问敏感路径
	sensitivePaths := []string{
		"/etc/", "/root/", "/home/", "/var/", "/usr/", "/bin/", "/sbin/",
		"/proc/", "/sys/", "/dev/", "/tmp/", "/boot/",
		"C:\\", "D:\\", "E:\\", "F:\\", // Windows路径
		"%SYSTEMROOT%", "%PROGRAMFILES%", "%USERPROFILE%",
	}
	
	for _, path := range sensitivePaths {
		if strings.Contains(lowerCommand, strings.ToLower(path)) {
			v.AddError(field, "attempts to access sensitive system paths")
			return
		}
	}
}

// SanitizeInput 清理输入，防止SQL注入和XSS攻击
func SanitizeInput(input string) string {
	// 限制输入长度，防止DoS攻击
	if len(input) > 10000 {
		input = input[:10000]
	}
	
	// 移除或转义危险字符
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "\\", "\\\\")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "--", "")
	input = strings.ReplaceAll(input, "/*", "")
	input = strings.ReplaceAll(input, "*/", "")
	input = strings.ReplaceAll(input, "#", "")
	input = strings.ReplaceAll(input, "@", "")
	
	// XSS防护 - 移除HTML/JavaScript标签
	xssPatterns := []string{
		"<script", "</script>", "<iframe", "</iframe>",
		"javascript:", "vbscript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=",
	}
	
	for _, pattern := range xssPatterns {
		input = regexp.MustCompile(`(?i)`+regexp.QuoteMeta(pattern)).ReplaceAllString(input, "")
	}
	
	// 增强的SQL关键字过滤
	sqlKeywords := []string{
		"DROP", "DELETE", "INSERT", "UPDATE", "ALTER", "CREATE",
		"EXEC", "EXECUTE", "UNION", "SELECT", "SCRIPT", "TRUNCATE",
		"GRANT", "REVOKE", "BACKUP", "RESTORE", "SHUTDOWN",
		"WAITFOR", "DELAY", "BENCHMARK", "SLEEP", "LOAD_FILE",
		"INTO OUTFILE", "INTO DUMPFILE", "INFORMATION_SCHEMA",
	}
	
	for _, keyword := range sqlKeywords {
		// 使用单词边界确保精确匹配
		pattern := `(?i)\b` + regexp.QuoteMeta(keyword) + `\b`
		input = regexp.MustCompile(pattern).ReplaceAllString(input, "")
	}
	
	// 移除SQL注入常见模式
	sqlInjectionPatterns := []string{
		`(?i)\b(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)\b(or|and)\s+['"]\w+['"]\s*=\s*['"]\w+['"]`,
		`(?i)['"];\s*(drop|delete|insert|update)`,
		`(?i)\bunion\s+select`,
		`(?i)\bhaving\s+\d+\s*=\s*\d+`,
		`(?i)\bgroup\s+by\s+\d+`,
		`(?i)\border\s+by\s+\d+`,
		`(?i)0x[0-9a-f]+`, // 十六进制编码
		`(?i)char\(\d+\)`, // CHAR函数
		`(?i)concat\s*\(`, // CONCAT函数
		`(?i)substring\s*\(`, // SUBSTRING函数
	}
	
	for _, pattern := range sqlInjectionPatterns {
		input = regexp.MustCompile(pattern).ReplaceAllString(input, "")
	}
	
	// 移除多余的空白字符
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	
	return strings.TrimSpace(input)
}

// SanitizeString 清理字符串输入
func SanitizeString(input string) string {
	return SanitizeInput(input)
}

// ValidateUsername 验证用户名
func ValidateUsername(username string) error {
	if len(strings.TrimSpace(username)) < 3 {
		return ValidationError{Field: "username", Message: "must be at least 3 characters"}
	}
	if len(username) > 50 {
		return ValidationError{Field: "username", Message: "must be at most 50 characters"}
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return ValidationError{Field: "username", Message: "must contain only letters, numbers, underscores, and hyphens"}
		}
	}
	return nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ValidationError{Field: "password", Message: "must be at least 8 characters"}
	}
	if len(password) > 128 {
		return ValidationError{Field: "password", Message: "must be at most 128 characters"}
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	
	if !hasUpper || !hasLower || !hasDigit {
		return ValidationError{Field: "password", Message: "must contain at least one uppercase letter, one lowercase letter, and one digit"}
	}
	
	return nil
}

// ValidateEmail 验证邮箱格式（独立函数）
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ValidationError{Field: "email", Message: "must be a valid email address"}
	}
	return nil
}

// ValidateNoSQLInjection 验证是否包含SQL注入
func (v *Validator) ValidateNoSQLInjection(field, value string) {
	sqlPatterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)\s`,
		`(?i)(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)(or|and)\s+['"]\w+['"]\s*=\s*['"]\w+['"]`,
		`['"];\s*(drop|delete|insert|update)`,
		`/\*.*\*/`,
		`--`,
		`(?i)\bwaitfor\s+delay`,
		`(?i)\bbenchmark\s*\(`,
		`(?i)\bsleep\s*\(`,
		`(?i)\bload_file\s*\(`,
		`(?i)\binto\s+(outfile|dumpfile)`,
		`(?i)0x[0-9a-f]+`,
		`(?i)\bchar\s*\(`,
		`(?i)\bconcat\s*\(`,
		`(?i)\bsubstring\s*\(`,
		`(?i)\bhex\s*\(`,
		`(?i)\bunhex\s*\(`,
	}
	
	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			v.AddError(field, "contains potentially malicious SQL content")
			return
		}
	}
}

// ValidateNoMaliciousContent 验证是否包含恶意内容（综合检查）
func (v *Validator) ValidateNoMaliciousContent(field, value string) {
	// 检查长度限制
	if len(value) > 50000 {
		v.AddError(field, "content too large (max 50KB)")
		return
	}
	
	// XSS攻击模式
	xssPatterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)<iframe[^>]*>.*?</iframe>`,
		`(?i)javascript:\s*`,
		`(?i)vbscript:\s*`,
		`(?i)on\w+\s*=`,
		`(?i)<\s*img[^>]+src\s*=\s*['"]*javascript:`,
		`(?i)<\s*link[^>]+href\s*=\s*['"]*javascript:`,
		`(?i)<\s*object[^>]*>`,
		`(?i)<\s*embed[^>]*>`,
		`(?i)<\s*applet[^>]*>`,
		`(?i)<\s*meta[^>]+http-equiv`,
	}
	
	for _, pattern := range xssPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			v.AddError(field, "contains potentially malicious XSS content")
			return
		}
	}
	
	// 路径遍历攻击
	pathTraversalPatterns := []string{
		`\.\.\/`,
		`\.\.\\`,
		`%2e%2e%2f`,
		`%2e%2e%5c`,
		`\.\%2f`,
		`\.\%5c`,
	}
	
	for _, pattern := range pathTraversalPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			v.AddError(field, "contains path traversal attack patterns")
			return
		}
	}
	
	// 命令注入模式
	cmdInjectionPatterns := []string{
		`[;&|]\s*(rm|del|format|shutdown|reboot)`,
		`\$\([^)]*\)`,
		"`[^`]*`",
		`\|\s*(nc|netcat|telnet|wget|curl)`,
		`>\s*\/dev\/`,
		`<\s*\/dev\/`,
	}
	
	for _, pattern := range cmdInjectionPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			v.AddError(field, "contains command injection patterns")
			return
		}
	}
	
	// LDAP注入模式
	ldapPatterns := []string{
		`\*\)\(`,
		`\)\(\|`,
		`\)\(&`,
		`\*\)\(&`,
	}
	
	for _, pattern := range ldapPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			v.AddError(field, "contains LDAP injection patterns")
			return
		}
	}
	
	// 检查是否包含过多的特殊字符（可能的混淆攻击）
	specialCharCount := 0
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			specialCharCount++
		}
	}
	
	if len(value) > 0 && float64(specialCharCount)/float64(len(value)) > 0.3 {
		v.AddError(field, "contains too many special characters (possible obfuscation attack)")
		return
	}
}

// ValidatePagination 验证分页参数
func (v *Validator) ValidatePagination(page, limit string) (int, int) {
	pageNum := 1
	limitNum := 20
	
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 && p <= 10000 {
			pageNum = p
		} else {
			v.AddError("page", "must be a positive number between 1 and 10000")
		}
	}
	
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			limitNum = l
		} else {
			v.AddError("limit", "must be a positive number between 1 and 100")
		}
	}
	
	return pageNum, limitNum
}

// ValidateLogLevel 验证日志级别
func (v *Validator) ValidateLogLevel(field, level string) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	lowerLevel := strings.ToLower(level)
	
	for _, validLevel := range validLevels {
		if lowerLevel == validLevel {
			return
		}
	}
	
	v.AddError(field, "must be one of: debug, info, warn, error, fatal")
}

// ValidateRetentionDays 验证日志保留天数
func (v *Validator) ValidateRetentionDays(field string, days int) {
	if days < 1 {
		v.AddError(field, "must be at least 1 day")
	}
	if days > 3650 { // 最多10年
		v.AddError(field, "must be at most 3650 days (10 years)")
	}
}

// ValidateFileSize 验证文件大小（字节）
func (v *Validator) ValidateFileSize(field string, size int64) {
	const maxSize = 100 * 1024 * 1024 // 100MB
	if size < 0 {
		v.AddError(field, "file size cannot be negative")
	}
	if size > maxSize {
		v.AddError(field, "file size cannot exceed 100MB")
	}
}

// ValidateTimeout 验证超时时间（秒）
func (v *Validator) ValidateTimeout(field string, timeout int) {
	if timeout < 1 {
		v.AddError(field, "timeout must be at least 1 second")
	}
	if timeout > 3600 { // 最多1小时
		v.AddError(field, "timeout cannot exceed 3600 seconds (1 hour)")
	}
}

// ValidatePort 验证端口号
func (v *Validator) ValidatePort(field string, port int) {
	if port < 1 || port > 65535 {
		v.AddError(field, "port must be between 1 and 65535")
	}
}

// ValidateIPAddress 验证IP地址格式
func (v *Validator) ValidateIPAddress(field, ip string) {
	ipRegex := regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	if !ipRegex.MatchString(ip) {
		v.AddError(field, "must be a valid IPv4 address")
	}
}

// ValidateURL 验证URL格式
func (v *Validator) ValidateURL(field, url string) {
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(:[0-9]+)?(/.*)?$`)
	if !urlRegex.MatchString(url) {
		v.AddError(field, "must be a valid HTTP or HTTPS URL")
	}
}

// ValidateJSONString 验证JSON字符串格式
func (v *Validator) ValidateJSONString(field, jsonStr string) {
	if len(jsonStr) > 10240 { // 最大10KB
		v.AddError(field, "JSON string too large (max 10KB)")
		return
	}
	
	// 简单的JSON格式检查
	trimmed := strings.TrimSpace(jsonStr)
	if !((strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))) {
		v.AddError(field, "must be a valid JSON object or array")
	}
}

// ValidateSearchQuery 验证搜索查询，防止过于复杂的查询
func (v *Validator) ValidateSearchQuery(field, query string) {
	if len(query) > 500 {
		v.AddError(field, "search query too long (max 500 characters)")
	}
	
	// 检查是否包含过多的通配符
	wildcardCount := strings.Count(query, "*") + strings.Count(query, "?")
	if wildcardCount > 10 {
		v.AddError(field, "too many wildcards in search query (max 10)")
	}
}

// ValidatePassword 验证密码强度
func (v *Validator) ValidatePassword(field, password string) {
	v.ValidateRequired(field, password)
	v.ValidateLength(field, password, 8, 128)
	
	// 检查密码复杂性
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		v.AddError(field, "must contain at least one uppercase letter")
	}
	if !hasLower {
		v.AddError(field, "must contain at least one lowercase letter")
	}
	if !hasDigit {
		v.AddError(field, "must contain at least one digit")
	}
	if !hasSpecial {
		v.AddError(field, "must contain at least one special character")
	}
}

// ValidateHost 验证主机地址
func (v *Validator) ValidateHost(field, host string) {
	v.ValidateRequired(field, host)
	
	// 检查是否为IP地址
	ipRegex := regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	if ipRegex.MatchString(host) {
		return
	}
	
	// 检查是否为域名
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?))*$`)
	if !hostnameRegex.MatchString(host) {
		v.AddError(field, "must be a valid IP address or hostname")
	}
}

// ValidateMaxLength 验证最大长度
func (v *Validator) ValidateMaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("must be at most %d characters", max))
	}
}