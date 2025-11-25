package logger

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SensitiveFilter 敏感信息过滤器
type SensitiveFilter struct {
	// 敏感字段名称（不区分大小写）
	sensitiveFields []string
	
	// 敏感字段的正则表达式
	sensitivePatterns []*regexp.Regexp
	
	// 替换文本
	replacement string
}

// NewSensitiveFilter 创建敏感信息过滤器
func NewSensitiveFilter() *SensitiveFilter {
	filter := &SensitiveFilter{
		sensitiveFields: []string{
			"password",
			"passwd",
			"pwd",
			"secret",
			"token",
			"api_key",
			"apikey",
			"access_token",
			"refresh_token",
			"auth_token",
			"authorization",
			"jwt",
			"bearer",
			"private_key",
			"privatekey",
			"credit_card",
			"creditcard",
			"ssn",
			"social_security",
		},
		replacement: "***REDACTED***",
	}

	// 编译正则表达式
	filter.sensitivePatterns = make([]*regexp.Regexp, 0)
	for _, field := range filter.sensitiveFields {
		// 匹配 JSON 格式: "field": "value"
		pattern := regexp.MustCompile(`(?i)"` + field + `"\s*:\s*"[^"]*"`)
		filter.sensitivePatterns = append(filter.sensitivePatterns, pattern)
		
		// 匹配 key=value 格式
		pattern2 := regexp.MustCompile(`(?i)` + field + `\s*=\s*[^\s&]+`)
		filter.sensitivePatterns = append(filter.sensitivePatterns, pattern2)
	}

	return filter
}

// FilterString 过滤字符串中的敏感信息
func (f *SensitiveFilter) FilterString(input string) string {
	if input == "" {
		return input
	}

	result := input

	// 应用所有正则表达式
	for i, pattern := range f.sensitivePatterns {
		field := f.sensitiveFields[i/2] // 每个字段有两个模式
		
		if i%2 == 0 {
			// JSON 格式
			result = pattern.ReplaceAllStringFunc(result, func(match string) string {
				return `"` + field + `": "` + f.replacement + `"`
			})
		} else {
			// key=value 格式
			result = pattern.ReplaceAllStringFunc(result, func(match string) string {
				return field + "=" + f.replacement
			})
		}
	}

	return result
}

// FilterMap 过滤 map 中的敏感信息
func (f *SensitiveFilter) FilterMap(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	filtered := make(map[string]interface{})
	
	for key, value := range data {
		lowerKey := strings.ToLower(key)
		
		// 检查是否是敏感字段
		isSensitive := false
		for _, field := range f.sensitiveFields {
			if strings.Contains(lowerKey, field) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			filtered[key] = f.replacement
		} else {
			// 递归处理嵌套的 map
			switch v := value.(type) {
			case map[string]interface{}:
				filtered[key] = f.FilterMap(v)
			case []interface{}:
				filtered[key] = f.filterSlice(v)
			default:
				filtered[key] = value
			}
		}
	}

	return filtered
}

// filterSlice 过滤切片中的敏感信息
func (f *SensitiveFilter) filterSlice(data []interface{}) []interface{} {
	if data == nil {
		return nil
	}

	filtered := make([]interface{}, len(data))
	
	for i, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			filtered[i] = f.FilterMap(v)
		case []interface{}:
			filtered[i] = f.filterSlice(v)
		default:
			filtered[i] = item
		}
	}

	return filtered
}

// FilterJSON 过滤 JSON 字符串中的敏感信息
func (f *SensitiveFilter) FilterJSON(jsonStr string) string {
	if jsonStr == "" {
		return jsonStr
	}

	// 尝试解析为 JSON
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// 如果不是有效的 JSON，使用字符串过滤
		return f.FilterString(jsonStr)
	}

	// 根据类型过滤
	var filtered interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		filtered = f.FilterMap(v)
	case []interface{}:
		filtered = f.filterSlice(v)
	default:
		return jsonStr
	}

	// 转换回 JSON
	result, err := json.Marshal(filtered)
	if err != nil {
		return jsonStr
	}

	return string(result)
}

// AddSensitiveField 添加自定义敏感字段
func (f *SensitiveFilter) AddSensitiveField(field string) {
	f.sensitiveFields = append(f.sensitiveFields, field)
	
	// 添加对应的正则表达式
	pattern := regexp.MustCompile(`(?i)"` + field + `"\s*:\s*"[^"]*"`)
	f.sensitivePatterns = append(f.sensitivePatterns, pattern)
	
	pattern2 := regexp.MustCompile(`(?i)` + field + `\s*=\s*[^\s&]+`)
	f.sensitivePatterns = append(f.sensitivePatterns, pattern2)
}

// SetReplacement 设置替换文本
func (f *SensitiveFilter) SetReplacement(replacement string) {
	f.replacement = replacement
}

// FilterLogMessage 过滤日志消息
func (f *SensitiveFilter) FilterLogMessage(message string, fields map[string]interface{}) (string, map[string]interface{}) {
	// 过滤消息
	filteredMessage := f.FilterString(message)
	
	// 过滤字段
	filteredFields := f.FilterMap(fields)
	
	return filteredMessage, filteredFields
}

// IsSensitiveField 检查字段名是否敏感
func (f *SensitiveFilter) IsSensitiveField(fieldName string) bool {
	lowerField := strings.ToLower(fieldName)
	
	for _, sensitive := range f.sensitiveFields {
		if strings.Contains(lowerField, sensitive) {
			return true
		}
	}
	
	return false
}

// FilterURL 过滤 URL 中的敏感信息（如查询参数）
func (f *SensitiveFilter) FilterURL(url string) string {
	if url == "" {
		return url
	}

	// 分离 URL 和查询参数
	parts := strings.SplitN(url, "?", 2)
	if len(parts) < 2 {
		return url
	}

	baseURL := parts[0]
	query := parts[1]

	// 过滤查询参数
	filteredQuery := f.FilterString(query)

	return baseURL + "?" + filteredQuery
}

// FilterHeaders 过滤 HTTP 头中的敏感信息
func (f *SensitiveFilter) FilterHeaders(headers map[string][]string) map[string][]string {
	if headers == nil {
		return nil
	}

	filtered := make(map[string][]string)
	
	for key, values := range headers {
		// 移除连字符并转换为小写进行比较
		normalizedKey := strings.ToLower(strings.ReplaceAll(key, "-", ""))
		
		// 检查是否是敏感头
		isSensitive := false
		for _, field := range f.sensitiveFields {
			normalizedField := strings.ReplaceAll(field, "_", "")
			if strings.Contains(normalizedKey, normalizedField) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			filtered[key] = []string{f.replacement}
		} else {
			filtered[key] = values
		}
	}

	return filtered
}

// MaskEmail 部分隐藏邮箱地址
func (f *SensitiveFilter) MaskEmail(email string) string {
	if email == "" {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domain := parts[1]

	// 只显示前2个字符
	if len(localPart) <= 2 {
		return "**@" + domain
	}

	masked := localPart[:2] + strings.Repeat("*", len(localPart)-2)
	return masked + "@" + domain
}

// MaskCreditCard 部分隐藏信用卡号
func (f *SensitiveFilter) MaskCreditCard(cardNumber string) string {
	if cardNumber == "" {
		return cardNumber
	}

	// 移除空格和连字符
	cleaned := strings.ReplaceAll(cardNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	if len(cleaned) < 4 {
		return strings.Repeat("*", len(cleaned))
	}

	// 只显示最后4位
	masked := strings.Repeat("*", len(cleaned)-4) + cleaned[len(cleaned)-4:]
	return masked
}

// MaskPhoneNumber 部分隐藏电话号码
func (f *SensitiveFilter) MaskPhoneNumber(phone string) string {
	if phone == "" {
		return phone
	}

	// 移除非数字字符
	cleaned := regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")

	if len(cleaned) < 4 {
		return strings.Repeat("*", len(cleaned))
	}

	// 只显示最后4位
	masked := strings.Repeat("*", len(cleaned)-4) + cleaned[len(cleaned)-4:]
	return masked
}

// 全局敏感信息过滤器实例
var GlobalSensitiveFilter = NewSensitiveFilter()
