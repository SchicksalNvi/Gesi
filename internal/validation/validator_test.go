package validation

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// 属性 17：输入验证完整性
// 验证需求：6.4
func TestInputValidationProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("required validation rejects empty strings", prop.ForAll(
		func() bool {
			validator := NewValidator()
			validator.ValidateRequired("field", "")
			return validator.HasErrors()
		},
	))

	properties.Property("required validation accepts non-empty strings", prop.ForAll(
		func(value string) bool {
			if value == "" {
				return true // Skip empty strings
			}

			validator := NewValidator()
			validator.ValidateRequired("field", value)
			return !validator.HasErrors()
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	))

	properties.Property("length validation enforces min length", prop.ForAll(
		func(minLen int) bool {
			if minLen < 1 || minLen > 100 {
				minLen = 5
			}

			validator := NewValidator()
			shortString := strings.Repeat("a", minLen-1)
			validator.ValidateLength("field", shortString, minLen, 100)
			return validator.HasErrors()
		},
		gen.IntRange(1, 100),
	))

	properties.Property("length validation enforces max length", prop.ForAll(
		func(maxLen int) bool {
			if maxLen < 1 || maxLen > 100 {
				maxLen = 10
			}

			validator := NewValidator()
			longString := strings.Repeat("a", maxLen+1)
			validator.ValidateLength("field", longString, 1, maxLen)
			return validator.HasErrors()
		},
		gen.IntRange(1, 100),
	))

	properties.Property("length validation accepts valid length", prop.ForAll(
		func(length int) bool {
			if length < 5 || length > 50 {
				length = 10
			}

			validator := NewValidator()
			validString := strings.Repeat("a", length)
			validator.ValidateLength("field", validString, 5, 50)
			return !validator.HasErrors()
		},
		gen.IntRange(5, 50),
	))

	properties.Property("alphanumeric validation accepts valid strings", prop.ForAll(
		func(value string) bool {
			if value == "" {
				return true
			}

			// 只包含字母和数字
			isValid := true
			for _, ch := range value {
				if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
					isValid = false
					break
				}
			}

			if !isValid {
				return true // Skip invalid strings
			}

			validator := NewValidator()
			validator.ValidateAlphanumeric("field", value)
			return !validator.HasErrors()
		},
		gen.AlphaString(),
	))

	properties.Property("email validation rejects invalid emails", prop.ForAll(
		func(value string) bool {
			if value == "" || strings.Contains(value, "@") {
				return true // Skip empty or potentially valid emails
			}

			validator := NewValidator()
			validator.ValidateEmail("field", value)
			return validator.HasErrors()
		},
		gen.AlphaString(),
	))

	properties.Property("SQL injection patterns are detected", prop.ForAll(
		func(value string) bool {
			// 测试包含 SQL 关键字的字符串
			sqlKeywords := []string{"DROP", "DELETE", "INSERT", "UPDATE", "SELECT", "UNION"}
			
			containsSQL := false
			lowerValue := strings.ToLower(value)
			for _, keyword := range sqlKeywords {
				if strings.Contains(lowerValue, strings.ToLower(keyword)) {
					containsSQL = true
					break
				}
			}

			if !containsSQL {
				return true // Skip strings without SQL keywords
			}

			v := NewValidator()
			v.ValidateNoSQLInjection("field", value)
			return v.HasErrors()
		},
		gen.AnyString(),
	))

	properties.Property("malicious content detection works", prop.ForAll(
		func() bool {
			maliciousContent := []string{
				"<script>alert('xss')</script>",
				"<img src=x onerror=alert('xss')>",
				"javascript:alert('xss')",
				"<iframe src='javascript:alert(1)'>",
			}

			for _, content := range maliciousContent {
				v := NewValidator()
				v.ValidateNoMaliciousContent("field", content)
				if !v.HasErrors() {
					return false
				}
			}
			return true
		},
	))

	properties.Property("pagination validation normalizes values", prop.ForAll(
		func(page int, limit int) bool {
			validator := NewValidator()
			pageStr := ""
			limitStr := ""

			if page > 0 {
				pageStr = string(rune(page))
			}
			if limit > 0 {
				limitStr = string(rune(limit))
			}

			resultPage, resultLimit := validator.ValidatePagination(pageStr, limitStr)

			// 验证结果在合理范围内
			return resultPage >= 1 && resultPage <= 10000 &&
				resultLimit >= 1 && resultLimit <= 100
		},
		gen.IntRange(-100, 10000),
		gen.IntRange(-100, 200),
	))

	properties.Property("sanitization removes dangerous characters", prop.ForAll(
		func(value string) bool {
			sanitized := SanitizeInput(value)

			// 验证不包含危险字符
			dangerous := []string{"<", ">", "script", "javascript:", "onerror"}
			for _, d := range dangerous {
				if strings.Contains(strings.ToLower(sanitized), d) {
					return false
				}
			}
			return true
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestValidatorScenarios 测试具体的验证场景
func TestValidatorScenarios(t *testing.T) {
	// 测试必需字段验证
	t.Run("required field validation", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateRequired("username", "")
		assert.True(t, validator.HasErrors())
		errors := validator.Errors()
		assert.Len(t, errors, 1)
		assert.Equal(t, "username", errors[0].Field)

		validator2 := NewValidator()
		validator2.ValidateRequired("username", "john")
		assert.False(t, validator2.HasErrors())
	})

	// 测试长度验证
	t.Run("length validation", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateLength("password", "123", 8, 20)
		assert.True(t, validator.HasErrors())

		validator2 := NewValidator()
		validator2.ValidateLength("password", "12345678", 8, 20)
		assert.False(t, validator2.HasErrors())

		validator3 := NewValidator()
		validator3.ValidateLength("password", strings.Repeat("a", 25), 8, 20)
		assert.True(t, validator3.HasErrors())
	})

	// 测试字母数字验证
	t.Run("alphanumeric validation", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateAlphanumeric("username", "john@doe")
		assert.True(t, validator.HasErrors())

		validator2 := NewValidator()
		validator2.ValidateAlphanumeric("username", "johndoe123")
		assert.False(t, validator2.HasErrors())

		// 下划线和连字符是允许的
		validator3 := NewValidator()
		validator3.ValidateAlphanumeric("username", "john_doe-123")
		assert.False(t, validator3.HasErrors())
	})

	// 测试邮箱验证
	t.Run("email validation", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateEmail("email", "invalid-email")
		assert.True(t, validator.HasErrors())

		validator2 := NewValidator()
		validator2.ValidateEmail("email", "user@example.com")
		assert.False(t, validator2.HasErrors())
	})

	// 测试 SQL 注入检测
	t.Run("SQL injection detection", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateNoSQLInjection("input", "'; DROP TABLE users--")
		assert.True(t, validator.HasErrors())

		validator2 := NewValidator()
		validator2.ValidateNoSQLInjection("input", "normal input")
		assert.False(t, validator2.HasErrors())
	})

	// 测试恶意内容检测
	t.Run("malicious content detection", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateNoMaliciousContent("input", "<script>alert('xss')</script>")
		assert.True(t, validator.HasErrors())

		validator2 := NewValidator()
		validator2.ValidateNoMaliciousContent("input", "normal text")
		assert.False(t, validator2.HasErrors())
	})

	// 测试分页验证
	t.Run("pagination validation", func(t *testing.T) {
		validator := NewValidator()
		page, limit := validator.ValidatePagination("", "")
		assert.Equal(t, 1, page)
		assert.Equal(t, 20, limit)
		assert.False(t, validator.HasErrors())

		validator2 := NewValidator()
		page2, limit2 := validator2.ValidatePagination("2", "50")
		assert.Equal(t, 2, page2)
		assert.Equal(t, 50, limit2)
		assert.False(t, validator2.HasErrors())

		validator3 := NewValidator()
		page3, limit3 := validator3.ValidatePagination("-1", "200")
		// 无效值会被规范化，但会添加错误
		assert.True(t, validator3.HasErrors())
		// 返回值会被规范化为有效值
		assert.Equal(t, 1, page3)
		assert.LessOrEqual(t, limit3, 100) // 限制最大值
	})

	// 测试输入清理
	t.Run("input sanitization", func(t *testing.T) {
		sanitized := SanitizeInput("<script>alert('xss')</script>")
		assert.NotContains(t, sanitized, "<script>")
		assert.NotContains(t, sanitized, "</script>")

		sanitized2 := SanitizeInput("normal text")
		assert.Equal(t, "normal text", sanitized2)
	})

	// 测试多个错误
	t.Run("multiple errors", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateRequired("username", "")
		validator.ValidateRequired("email", "")
		validator.ValidateRequired("password", "")

		assert.True(t, validator.HasErrors())
		errors := validator.Errors()
		assert.Len(t, errors, 3)
		
		// 检查每个字段都有错误
		fields := make(map[string]bool)
		for _, err := range errors {
			fields[err.Field] = true
		}
		assert.True(t, fields["username"])
		assert.True(t, fields["email"])
		assert.True(t, fields["password"])
	})

	// 测试错误累积
	t.Run("error accumulation", func(t *testing.T) {
		validator := NewValidator()
		validator.ValidateRequired("field1", "")
		assert.True(t, validator.HasErrors())

		validator.ValidateLength("field2", "ab", 5, 10)
		assert.True(t, validator.HasErrors())

		errors := validator.Errors()
		assert.Len(t, errors, 2)
	})
}
