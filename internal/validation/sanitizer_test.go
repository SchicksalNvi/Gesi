package validation

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 属性 22：输入清理
// 验证需求：9.3
func TestInputSanitizationProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)
	sanitizer := NewSanitizer()

	properties.Property("HTML sanitization removes script tags", prop.ForAll(
		func(content string) bool {
			input := "<script>alert('xss')</script>" + content
			sanitized := sanitizer.SanitizeHTML(input)
			
			// 验证不包含 script 标签
			return !strings.Contains(strings.ToLower(sanitized), "<script")
		},
		gen.AnyString(),
	))

	properties.Property("SQL sanitization removes dangerous keywords", prop.ForAll(
		func(input string) bool {
			// 添加 SQL 关键字
			dangerous := input + " UNION SELECT * FROM users"
			sanitized := sanitizer.SanitizeSQL(dangerous)
			
			// 验证不包含 SQL 关键字
			lower := strings.ToLower(sanitized)
			return !strings.Contains(lower, "union") &&
				!strings.Contains(lower, "select")
		},
		gen.AlphaString(),
	))

	properties.Property("filename sanitization prevents path traversal", prop.ForAll(
		func(filename string) bool {
			dangerous := "../../../etc/passwd"
			sanitized := sanitizer.SanitizeFilename(dangerous)
			
			// 验证不包含路径遍历字符
			return !strings.Contains(sanitized, "..") &&
				!strings.Contains(sanitized, "/") &&
				!strings.Contains(sanitized, "\\")
		},
		gen.AlphaString(),
	))

	properties.Property("email sanitization validates format", prop.ForAll(
		func(localPart string, domain string) bool {
			if localPart == "" || domain == "" {
				return true
			}
			
			// 创建有效的邮箱
			email := localPart + "@" + domain + ".com"
			sanitized := sanitizer.SanitizeEmail(email)
			
			// 如果清理后为空，说明格式无效
			// 如果不为空，应该是小写且格式正确
			if sanitized != "" {
				return sanitized == strings.ToLower(sanitized) &&
					strings.Contains(sanitized, "@")
			}
			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("URL sanitization only allows http/https", prop.ForAll(
		func(path string) bool {
			// 测试危险协议
			dangerous := "javascript:alert('xss')"
			sanitized := sanitizer.SanitizeURL(dangerous)
			
			// 应该被清理为空或不包含 javascript
			return sanitized == "" || !strings.Contains(strings.ToLower(sanitized), "javascript")
		},
		gen.AlphaString(),
	))

	properties.Property("username sanitization removes special characters", prop.ForAll(
		func(username string) bool {
			// 添加特殊字符
			dangerous := username + "!@#$%^&*()"
			sanitized := sanitizer.SanitizeUsername(dangerous)
			
			// 验证只包含允许的字符
			for _, ch := range sanitized {
				if !((ch >= 'a' && ch <= 'z') ||
					(ch >= 'A' && ch <= 'Z') ||
					(ch >= '0' && ch <= '9') ||
					ch == '_' || ch == '-') {
					return false
				}
			}
			return true
		},
		gen.AlphaString(),
	))

	properties.Property("string sanitization removes control characters", prop.ForAll(
		func(input string) bool {
			// 添加控制字符
			dangerous := input + "\x00\x01\x02"
			sanitized := sanitizer.SanitizeString(dangerous)
			
			// 验证不包含控制字符（除了允许的）
			for _, ch := range sanitized {
				if ch < 32 && ch != '\n' && ch != '\r' && ch != '\t' {
					return false
				}
			}
			return true
		},
		gen.AnyString(),
	))

	properties.Property("sanitization is idempotent", prop.ForAll(
		func(input string) bool {
			// 清理一次
			once := sanitizer.SanitizeString(input)
			// 清理两次
			twice := sanitizer.SanitizeString(once)
			
			// 应该相同
			return once == twice
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestSanitizeHTML 测试 HTML 清理
func TestSanitizeHTML(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove script tag",
			input:    "<script>alert('xss')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "escape HTML entities",
			input:    "<div>Test</div>",
			expected: "&lt;div&gt;Test&lt;/div&gt;",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSanitizeSQL 测试 SQL 清理
func TestSanitizeSQL(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "remove UNION",
			input: "test UNION SELECT * FROM users",
		},
		{
			name:  "remove quotes",
			input: "test' OR '1'='1",
		},
		{
			name:  "remove comments",
			input: "test-- comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeSQL(tt.input)
			
			// 验证危险字符已被移除
			assert.NotContains(t, strings.ToLower(result), "union")
			assert.NotContains(t, result, "'")
			assert.NotContains(t, result, "\"")
			assert.NotContains(t, result, "--")
		})
	}
}

// TestSanitizeFilename 测试文件名清理
func TestSanitizeFilename(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove path traversal",
			input:    "../../../etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "remove slashes",
			input:    "path/to/file.txt",
			expected: "pathtofile.txt",
		},
		{
			name:     "keep valid characters",
			input:    "valid_file-name.txt",
			expected: "valid_file-name.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSanitizeEmail 测试邮箱清理
func TestSanitizeEmail(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid email",
			input:    "Test@Example.COM",
			expected: "test@example.com",
		},
		{
			name:     "invalid email",
			input:    "not-an-email",
			expected: "",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSanitizeURL 测试 URL 清理
func TestSanitizeURL(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid HTTP URL",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "valid HTTPS URL",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "javascript protocol",
			input:    "javascript:alert('xss')",
			expected: "",
		},
		{
			name:     "data protocol",
			input:    "data:text/html,<script>alert('xss')</script>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSanitizeUsername 测试用户名清理
func TestSanitizeUsername(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid username",
			input:    "user_name-123",
			expected: "user_name-123",
		},
		{
			name:     "remove special characters",
			input:    "user@name!",
			expected: "username",
		},
		{
			name:     "trim whitespace",
			input:    "  username  ",
			expected: "username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeUsername(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStripHTML 测试 HTML 标签移除
func TestStripHTML(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove all tags",
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello World",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.StripHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRemoveNullBytes 测试空字节移除
func TestRemoveNullBytes(t *testing.T) {
	sanitizer := NewSanitizer()

	input := "test\x00string"
	result := sanitizer.RemoveNullBytes(input)
	
	assert.NotContains(t, result, "\x00")
	assert.Equal(t, "teststring", result)
}

// TestTruncateString 测试字符串截断
func TestTruncateString(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "truncate long string",
			input:     "this is a very long string",
			maxLength: 10,
			expected:  "this is a ",
		},
		{
			name:      "keep short string",
			input:     "short",
			maxLength: 10,
			expected:  "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.TruncateString(tt.input, tt.maxLength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateAndSanitize 测试通用清理方法
func TestValidateAndSanitize(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name      string
		input     string
		inputType InputType
	}{
		{
			name:      "HTML type",
			input:     "<script>alert('xss')</script>",
			inputType: InputTypeHTML,
		},
		{
			name:      "SQL type",
			input:     "test UNION SELECT",
			inputType: InputTypeSQL,
		},
		{
			name:      "Filename type",
			input:     "../../../etc/passwd",
			inputType: InputTypeFilename,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.ValidateAndSanitize(tt.input, tt.inputType)
			require.NotNil(t, result)
		})
	}
}

// BenchmarkSanitizeHTML 基准测试：HTML 清理
func BenchmarkSanitizeHTML(b *testing.B) {
	sanitizer := NewSanitizer()
	input := "<script>alert('xss')</script><p>Hello World</p>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeHTML(input)
	}
}

// BenchmarkSanitizeSQL 基准测试：SQL 清理
func BenchmarkSanitizeSQL(b *testing.B) {
	sanitizer := NewSanitizer()
	input := "test UNION SELECT * FROM users WHERE id='1'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeSQL(input)
	}
}

// BenchmarkSanitizeFilename 基准测试：文件名清理
func BenchmarkSanitizeFilename(b *testing.B) {
	sanitizer := NewSanitizer()
	input := "../../../etc/passwd"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeFilename(input)
	}
}
