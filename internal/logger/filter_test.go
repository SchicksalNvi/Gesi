package logger

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 属性 23：敏感信息保护
// 验证需求：9.5
func TestSensitiveInformationProtectionProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)
	filter := NewSensitiveFilter()

	properties.Property("password fields are always filtered", prop.ForAll(
		func(password string) bool {
			input := map[string]interface{}{
				"username": "testuser",
				"password": password,
			}
			
			filtered := filter.FilterMap(input)
			
			// 验证密码被过滤
			return filtered["password"] == filter.replacement
		},
		gen.AnyString(),
	))

	properties.Property("token fields are always filtered", prop.ForAll(
		func(token string) bool {
			input := map[string]interface{}{
				"user_id": "123",
				"token":   token,
			}
			
			filtered := filter.FilterMap(input)
			
			// 验证令牌被过滤
			return filtered["token"] == filter.replacement
		},
		gen.AnyString(),
	))

	properties.Property("nested sensitive fields are filtered", prop.ForAll(
		func(secret string) bool {
			input := map[string]interface{}{
				"user": map[string]interface{}{
					"name":   "test",
					"secret": secret,
				},
			}
			
			filtered := filter.FilterMap(input)
			
			// 验证嵌套的敏感字段被过滤
			userMap, ok := filtered["user"].(map[string]interface{})
			if !ok {
				return false
			}
			
			return userMap["secret"] == filter.replacement
		},
		gen.AnyString(),
	))

	properties.Property("non-sensitive fields are preserved", prop.ForAll(
		func(username string, email string) bool {
			input := map[string]interface{}{
				"username": username,
				"email":    email,
				"password": "secret123",
			}
			
			filtered := filter.FilterMap(input)
			
			// 验证非敏感字段保持不变
			return filtered["username"] == username &&
				filtered["email"] == email &&
				filtered["password"] == filter.replacement
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("JSON string filtering works", prop.ForAll(
		func(password string) bool {
			jsonStr := `{"username":"test","password":"` + password + `"}`
			filtered := filter.FilterJSON(jsonStr)
			
			// 验证 JSON 中的密码被过滤
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(filtered), &result); err != nil {
				return false
			}
			
			return result["password"] == filter.replacement
		},
		gen.AlphaString(),
	))

	properties.Property("URL query parameters are filtered", prop.ForAll(
		func(token string) bool {
			if token == "" {
				token = "test"
			}
			url := "https://example.com/api?user=test&token=" + token
			filtered := filter.FilterURL(url)
			
			// 验证 URL 中的令牌被过滤
			return strings.Contains(filtered, "token="+filter.replacement)
		},
		gen.AlphaString(),
	))

	properties.Property("email masking preserves domain", prop.ForAll(
		func(localPart string, domain string) bool {
			if localPart == "" || domain == "" {
				return true
			}
			
			email := localPart + "@" + domain + ".com"
			masked := filter.MaskEmail(email)
			
			// 验证域名被保留
			return strings.Contains(masked, "@"+domain+".com")
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("filtering is idempotent", prop.ForAll(
		func(password string) bool {
			input := map[string]interface{}{
				"password": password,
			}
			
			// 过滤一次
			once := filter.FilterMap(input)
			// 过滤两次
			twice := filter.FilterMap(once)
			
			// 应该相同
			return once["password"] == twice["password"]
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestFilterString 测试字符串过滤
func TestFilterString(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "filter JSON password",
			input:    `{"username":"test","password":"secret123"}`,
			contains: filter.replacement,
		},
		{
			name:     "filter key=value password",
			input:    "username=test&password=secret123",
			contains: filter.replacement,
		},
		{
			name:     "filter token",
			input:    `{"token":"abc123"}`,
			contains: filter.replacement,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.FilterString(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

// TestFilterMap 测试 map 过滤
func TestFilterMap(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name     string
		input    map[string]interface{}
		checkKey string
		expected string
	}{
		{
			name: "filter password",
			input: map[string]interface{}{
				"username": "test",
				"password": "secret123",
			},
			checkKey: "password",
			expected: filter.replacement,
		},
		{
			name: "filter token",
			input: map[string]interface{}{
				"user_id": "123",
				"token":   "abc123",
			},
			checkKey: "token",
			expected: filter.replacement,
		},
		{
			name: "filter api_key",
			input: map[string]interface{}{
				"name":    "test",
				"api_key": "key123",
			},
			checkKey: "api_key",
			expected: filter.replacement,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.FilterMap(tt.input)
			assert.Equal(t, tt.expected, result[tt.checkKey])
		})
	}
}

// TestFilterNestedMap 测试嵌套 map 过滤
func TestFilterNestedMap(t *testing.T) {
	filter := NewSensitiveFilter()

	input := map[string]interface{}{
		"user": map[string]interface{}{
			"name":     "test",
			"password": "secret123",
			"profile": map[string]interface{}{
				"email":  "test@example.com",
				"secret": "mysecret",
			},
		},
	}

	result := filter.FilterMap(input)

	// 验证嵌套的敏感字段被过滤
	userMap := result["user"].(map[string]interface{})
	assert.Equal(t, filter.replacement, userMap["password"])

	profileMap := userMap["profile"].(map[string]interface{})
	assert.Equal(t, filter.replacement, profileMap["secret"])
	assert.Equal(t, "test@example.com", profileMap["email"])
}

// TestFilterJSON 测试 JSON 过滤
func TestFilterJSON(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "valid JSON with password",
			input: `{"username":"test","password":"secret123"}`,
		},
		{
			name:  "valid JSON with token",
			input: `{"user_id":"123","token":"abc123"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.FilterJSON(tt.input)
			
			// 验证结果是有效的 JSON
			var data map[string]interface{}
			err := json.Unmarshal([]byte(result), &data)
			require.NoError(t, err)
			
			// 验证敏感字段被过滤
			for key, value := range data {
				if filter.IsSensitiveField(key) {
					assert.Equal(t, filter.replacement, value)
				}
			}
		})
	}
}

// TestFilterURL 测试 URL 过滤
func TestFilterURL(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "filter token in query",
			input:    "https://example.com/api?user=test&token=abc123",
			contains: "token=" + filter.replacement,
		},
		{
			name:     "filter api_key in query",
			input:    "https://example.com/api?api_key=key123",
			contains: "api_key=" + filter.replacement,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.FilterURL(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

// TestFilterHeaders 测试 HTTP 头过滤
func TestFilterHeaders(t *testing.T) {
	filter := NewSensitiveFilter()

	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer token123"},
		"X-API-Key":     {"key123"},
	}

	result := filter.FilterHeaders(headers)

	assert.Equal(t, []string{"application/json"}, result["Content-Type"])
	assert.Equal(t, []string{filter.replacement}, result["Authorization"])
	assert.Equal(t, []string{filter.replacement}, result["X-API-Key"])
}

// TestMaskEmail 测试邮箱隐藏
func TestMaskEmail(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal email",
			input:    "testuser@example.com",
			expected: "te******@example.com",
		},
		{
			name:     "short email",
			input:    "ab@example.com",
			expected: "**@example.com",
		},
		{
			name:     "empty email",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.MaskEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskCreditCard 测试信用卡号隐藏
func TestMaskCreditCard(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "16 digit card",
			input:    "1234567890123456",
			expected: "************3456",
		},
		{
			name:     "card with spaces",
			input:    "1234 5678 9012 3456",
			expected: "************3456",
		},
		{
			name:     "card with dashes",
			input:    "1234-5678-9012-3456",
			expected: "************3456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.MaskCreditCard(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskPhoneNumber 测试电话号码隐藏
func TestMaskPhoneNumber(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "10 digit phone",
			input:    "1234567890",
			expected: "******7890",
		},
		{
			name:     "phone with formatting",
			input:    "(123) 456-7890",
			expected: "******7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.MaskPhoneNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAddSensitiveField 测试添加自定义敏感字段
func TestAddSensitiveField(t *testing.T) {
	filter := NewSensitiveFilter()
	
	// 添加自定义字段
	filter.AddSensitiveField("custom_secret")

	input := map[string]interface{}{
		"name":          "test",
		"custom_secret": "mysecret",
	}

	result := filter.FilterMap(input)

	assert.Equal(t, "test", result["name"])
	assert.Equal(t, filter.replacement, result["custom_secret"])
}

// TestIsSensitiveField 测试敏感字段检查
func TestIsSensitiveField(t *testing.T) {
	filter := NewSensitiveFilter()

	tests := []struct {
		name      string
		fieldName string
		expected  bool
	}{
		{
			name:      "password field",
			fieldName: "password",
			expected:  true,
		},
		{
			name:      "token field",
			fieldName: "access_token",
			expected:  true,
		},
		{
			name:      "normal field",
			fieldName: "username",
			expected:  false,
		},
		{
			name:      "case insensitive",
			fieldName: "PASSWORD",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.IsSensitiveField(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterLogMessage 测试日志消息过滤
func TestFilterLogMessage(t *testing.T) {
	filter := NewSensitiveFilter()

	message := "User login with password=secret123"
	fields := map[string]interface{}{
		"username": "test",
		"password": "secret123",
	}

	filteredMessage, filteredFields := filter.FilterLogMessage(message, fields)

	assert.Contains(t, filteredMessage, filter.replacement)
	assert.Equal(t, filter.replacement, filteredFields["password"])
	assert.Equal(t, "test", filteredFields["username"])
}

// TestSetReplacement 测试设置替换文本
func TestSetReplacement(t *testing.T) {
	filter := NewSensitiveFilter()
	customReplacement := "[HIDDEN]"
	
	filter.SetReplacement(customReplacement)

	input := map[string]interface{}{
		"password": "secret123",
	}

	result := filter.FilterMap(input)

	assert.Equal(t, customReplacement, result["password"])
}

// BenchmarkFilterMap 基准测试：map 过滤
func BenchmarkFilterMap(b *testing.B) {
	filter := NewSensitiveFilter()
	input := map[string]interface{}{
		"username": "test",
		"password": "secret123",
		"email":    "test@example.com",
		"token":    "abc123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.FilterMap(input)
	}
}

// BenchmarkFilterJSON 基准测试：JSON 过滤
func BenchmarkFilterJSON(b *testing.B) {
	filter := NewSensitiveFilter()
	jsonStr := `{"username":"test","password":"secret123","email":"test@example.com","token":"abc123"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.FilterJSON(jsonStr)
	}
}

// BenchmarkFilterString 基准测试：字符串过滤
func BenchmarkFilterString(b *testing.B) {
	filter := NewSensitiveFilter()
	input := "username=test&password=secret123&token=abc123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.FilterString(input)
	}
}
