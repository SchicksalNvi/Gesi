package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"superview/internal/config"
)

// LogEntry 日志条目结构
type LogEntry struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
	Details   string    `json:"details"`
}

// LogReader 日志读取器
type LogReader struct {
	config *config.DeveloperToolsConfig
}

// NewLogReader 创建新的日志读取器
func NewLogReader(cfg *config.DeveloperToolsConfig) *LogReader {
	return &LogReader{
		config: cfg,
	}
}

// ReadLogs 从日志文件读取日志
func (lr *LogReader) ReadLogs(level, component string, limit int) ([]LogEntry, error) {
	if !lr.config.Enabled {
		return []LogEntry{}, nil
	}

	// 检查日志文件是否存在
	logPath := lr.config.LogPath
	if !filepath.IsAbs(logPath) {
		// 如果是相对路径，转换为绝对路径
		wd, _ := os.Getwd()
		logPath = filepath.Join(wd, logPath)
	}

	file, err := os.Open(logPath)
	if err != nil {
		// 如果文件不存在，返回空结果而不是错误
		if os.IsNotExist(err) {
			return []LogEntry{}, nil
		}
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	id := 1

	// 正则表达式匹配常见的日志格式
	// 支持多种日志格式：JSON、标准格式等
	jsonLogRegex := regexp.MustCompile(`^\{.*\}$`)
	standardLogRegex := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?)\s+\[(\w+)\]\s+(?:\[(\w+)\]\s+)?(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry LogEntry
		entry.ID = id
		id++

		// 尝试解析JSON格式日志
		if jsonLogRegex.MatchString(line) {
			var jsonLog map[string]interface{}
			if err := json.Unmarshal([]byte(line), &jsonLog); err == nil {
				entry = lr.parseJSONLog(jsonLog)
				entry.ID = id - 1
			}
		} else if matches := standardLogRegex.FindStringSubmatch(line); len(matches) >= 4 {
			// 解析标准格式日志
			entry = lr.parseStandardLog(matches)
			entry.ID = id - 1
		} else {
			// 简单文本日志
			entry = LogEntry{
				ID:        id - 1,
				Timestamp: time.Now(),
				Level:     "INFO",
				Component: "Unknown",
				Message:   line,
				Details:   "",
			}
		}

		// 应用过滤器
		if level != "" && !strings.EqualFold(entry.Level, level) {
			continue
		}
		if component != "" && !strings.EqualFold(entry.Component, component) {
			continue
		}

		entries = append(entries, entry)

		// 限制读取的行数以避免内存问题
		if len(entries) >= lr.config.MaxLogLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %v", err)
	}

	// 按时间戳倒序排列（最新的在前）
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// 应用限制
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// parseJSONLog 解析JSON格式的日志
func (lr *LogReader) parseJSONLog(jsonLog map[string]interface{}) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: "Unknown",
	}

	// 解析时间戳
	if ts, ok := jsonLog["timestamp"]; ok {
		if tsStr, ok := ts.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, tsStr); err == nil {
				entry.Timestamp = parsedTime
			}
		}
	} else if ts, ok := jsonLog["time"]; ok {
		if tsStr, ok := ts.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, tsStr); err == nil {
				entry.Timestamp = parsedTime
			}
		}
	}

	// 解析日志级别
	if level, ok := jsonLog["level"]; ok {
		if levelStr, ok := level.(string); ok {
			entry.Level = strings.ToUpper(levelStr)
		}
	}

	// 解析组件
	if component, ok := jsonLog["component"]; ok {
		if compStr, ok := component.(string); ok {
			entry.Component = compStr
		}
	} else if logger, ok := jsonLog["logger"]; ok {
		if loggerStr, ok := logger.(string); ok {
			entry.Component = loggerStr
		}
	}

	// 解析消息
	if msg, ok := jsonLog["message"]; ok {
		if msgStr, ok := msg.(string); ok {
			entry.Message = msgStr
		}
	} else if msg, ok := jsonLog["msg"]; ok {
		if msgStr, ok := msg.(string); ok {
			entry.Message = msgStr
		}
	}

	// 解析详细信息
	if details, ok := jsonLog["details"]; ok {
		if detailsStr, ok := details.(string); ok {
			entry.Details = detailsStr
		}
	} else if stack, ok := jsonLog["stack"]; ok {
		if stackStr, ok := stack.(string); ok {
			entry.Details = stackStr
		}
	}

	return entry
}

// parseStandardLog 解析标准格式的日志
func (lr *LogReader) parseStandardLog(matches []string) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: "Unknown",
	}

	// 解析时间戳
	if len(matches) > 1 {
		if parsedTime, err := time.Parse("2006-01-02T15:04:05.000Z", matches[1]); err == nil {
			entry.Timestamp = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02 15:04:05", matches[1]); err == nil {
			entry.Timestamp = parsedTime
		}
	}

	// 解析日志级别
	if len(matches) > 2 {
		entry.Level = strings.ToUpper(matches[2])
	}

	// 解析组件（可选）
	if len(matches) > 4 && matches[3] != "" {
		entry.Component = matches[3]
		entry.Message = matches[4]
	} else if len(matches) > 3 {
		entry.Message = matches[3]
	}

	return entry
}

// GetLogStats 获取日志统计信息
func (lr *LogReader) GetLogStats() (map[string]interface{}, error) {
	if !lr.config.Enabled {
		return map[string]interface{}{"enabled": false}, nil
	}

	logPath := lr.config.LogPath
	if !filepath.IsAbs(logPath) {
		wd, _ := os.Getwd()
		logPath = filepath.Join(wd, logPath)
	}

	fileInfo, err := os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"enabled":    true,
				"file_exists": false,
				"path":       logPath,
			}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"enabled":     true,
		"file_exists": true,
		"path":        logPath,
		"size":        fileInfo.Size(),
		"modified":    fileInfo.ModTime(),
	}, nil
}