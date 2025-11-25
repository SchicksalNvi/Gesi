package validation

import (
	"html"
	"regexp"
	"strings"
)

// Sanitizer 输入清理器
type Sanitizer struct {
	// XSS 防护相关
	scriptTagRegex *regexp.Regexp
	htmlTagRegex   *regexp.Regexp
	
	// SQL 注入防护相关
	sqlKeywordRegex *regexp.Regexp
}

// NewSanitizer 创建新的清理器
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		scriptTagRegex:  regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		htmlTagRegex:    regexp.MustCompile(`<[^>]+>`),
		sqlKeywordRegex: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|onerror|onload)`),
	}
}

// SanitizeHTML 清理 HTML 输入，防止 XSS 攻击
func (s *Sanitizer) SanitizeHTML(input string) string {
	if input == "" {
		return ""
	}

	// 移除 script 标签
	cleaned := s.scriptTagRegex.ReplaceAllString(input, "")
	
	// HTML 转义
	cleaned = html.EscapeString(cleaned)
	
	return cleaned
}

// SanitizeString 清理普通字符串输入
func (s *Sanitizer) SanitizeString(input string) string {
	if input == "" {
		return ""
	}

	// 移除控制字符
	cleaned := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)

	// 移除前后空白
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// SanitizeSQL 清理 SQL 相关输入（注意：这不能替代参数化查询）
func (s *Sanitizer) SanitizeSQL(input string) string {
	if input == "" {
		return ""
	}

	// 移除潜在的 SQL 关键字
	cleaned := s.sqlKeywordRegex.ReplaceAllString(input, "")
	
	// 移除特殊字符
	cleaned = strings.ReplaceAll(cleaned, "'", "")
	cleaned = strings.ReplaceAll(cleaned, "\"", "")
	cleaned = strings.ReplaceAll(cleaned, ";", "")
	cleaned = strings.ReplaceAll(cleaned, "--", "")
	cleaned = strings.ReplaceAll(cleaned, "/*", "")
	cleaned = strings.ReplaceAll(cleaned, "*/", "")
	
	return cleaned
}

// SanitizeFilename 清理文件名
func (s *Sanitizer) SanitizeFilename(filename string) string {
	if filename == "" {
		return ""
	}

	// 移除路径遍历字符
	cleaned := strings.ReplaceAll(filename, "..", "")
	cleaned = strings.ReplaceAll(cleaned, "/", "")
	cleaned = strings.ReplaceAll(cleaned, "\\", "")
	
	// 只保留字母、数字、点、下划线和连字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	cleaned = reg.ReplaceAllString(cleaned, "_")
	
	// 限制长度
	if len(cleaned) > 255 {
		cleaned = cleaned[:255]
	}
	
	return cleaned
}

// SanitizeEmail 清理邮箱地址
func (s *Sanitizer) SanitizeEmail(email string) string {
	if email == "" {
		return ""
	}

	// 转换为小写
	cleaned := strings.ToLower(email)
	
	// 移除空白
	cleaned = strings.TrimSpace(cleaned)
	
	// 基本验证格式
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(cleaned) {
		return ""
	}
	
	return cleaned
}

// SanitizeURL 清理 URL
func (s *Sanitizer) SanitizeURL(url string) string {
	if url == "" {
		return ""
	}

	// 移除空白
	cleaned := strings.TrimSpace(url)
	
	// 只允许 http 和 https 协议
	if !strings.HasPrefix(cleaned, "http://") && !strings.HasPrefix(cleaned, "https://") {
		return ""
	}
	
	// 移除潜在的 JavaScript 协议
	cleaned = strings.ReplaceAll(cleaned, "javascript:", "")
	cleaned = strings.ReplaceAll(cleaned, "data:", "")
	cleaned = strings.ReplaceAll(cleaned, "vbscript:", "")
	
	return cleaned
}

// SanitizeUsername 清理用户名
func (s *Sanitizer) SanitizeUsername(username string) string {
	if username == "" {
		return ""
	}

	// 移除空白
	cleaned := strings.TrimSpace(username)
	
	// 只保留字母、数字、下划线和连字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	cleaned = reg.ReplaceAllString(cleaned, "")
	
	// 限制长度
	if len(cleaned) > 50 {
		cleaned = cleaned[:50]
	}
	
	return cleaned
}

// SanitizeJSONString 清理 JSON 字符串
func (s *Sanitizer) SanitizeJSONString(input string) string {
	if input == "" {
		return ""
	}

	// 转义特殊字符
	cleaned := strings.ReplaceAll(input, "\\", "\\\\")
	cleaned = strings.ReplaceAll(cleaned, "\"", "\\\"")
	cleaned = strings.ReplaceAll(cleaned, "\n", "\\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\\r")
	cleaned = strings.ReplaceAll(cleaned, "\t", "\\t")
	
	return cleaned
}

// RemoveNullBytes 移除空字节
func (s *Sanitizer) RemoveNullBytes(input string) string {
	return strings.ReplaceAll(input, "\x00", "")
}

// TruncateString 截断字符串到指定长度
func (s *Sanitizer) TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength]
}

// StripHTML 完全移除 HTML 标签
func (s *Sanitizer) StripHTML(input string) string {
	if input == "" {
		return ""
	}

	// 移除所有 HTML 标签
	cleaned := s.htmlTagRegex.ReplaceAllString(input, "")
	
	// HTML 反转义
	cleaned = html.UnescapeString(cleaned)
	
	return cleaned
}

// ValidateAndSanitizeInput 验证并清理输入的通用方法
type InputType int

const (
	InputTypeHTML InputType = iota
	InputTypeString
	InputTypeSQL
	InputTypeFilename
	InputTypeEmail
	InputTypeURL
	InputTypeUsername
)

// ValidateAndSanitize 根据类型验证并清理输入
func (s *Sanitizer) ValidateAndSanitize(input string, inputType InputType) string {
	switch inputType {
	case InputTypeHTML:
		return s.SanitizeHTML(input)
	case InputTypeString:
		return s.SanitizeString(input)
	case InputTypeSQL:
		return s.SanitizeSQL(input)
	case InputTypeFilename:
		return s.SanitizeFilename(input)
	case InputTypeEmail:
		return s.SanitizeEmail(input)
	case InputTypeURL:
		return s.SanitizeURL(input)
	case InputTypeUsername:
		return s.SanitizeUsername(input)
	default:
		return s.SanitizeString(input)
	}
}

// 全局清理器实例
var GlobalSanitizer = NewSanitizer()
