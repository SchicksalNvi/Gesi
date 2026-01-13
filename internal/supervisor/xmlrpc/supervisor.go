package xmlrpc

import (
	"fmt"
	"strconv"
	"strings"
)

// ProcessInfo 符合 Supervisor XML-RPC API 规范的进程信息
type ProcessInfo struct {
	Name           string `json:"name"`           // 进程名称
	Group          string `json:"group"`          // 进程组名称  
	Start          int64  `json:"start"`          // UNIX 启动时间戳
	Stop           int64  `json:"stop"`           // UNIX 停止时间戳 (0 表示从未停止)
	Now            int64  `json:"now"`            // 当前 UNIX 时间戳
	State          int    `json:"state"`          // 状态码 (见 Supervisor 文档)
	StateName      string `json:"statename"`      // 状态名称
	SpawnErr       string `json:"spawnerr"`       // 启动错误描述
	ExitStatus     int    `json:"exitstatus"`     // 退出状态码
	StdoutLogfile  string `json:"stdout_logfile"` // stdout 日志文件路径
	StderrLogfile  string `json:"stderr_logfile"` // stderr 日志文件路径
	PID            int    `json:"pid"`            // 进程 PID (0 表示未运行)
	
	// 计算字段 (非 API 返回)
	Uptime         int64  `json:"uptime"`         // 运行时间 (秒)
	UptimeHuman    string `json:"uptime_human"`   // 人类可读的运行时间
}

type SupervisorClient struct {
	client *Client
}

func NewSupervisorClient(host string, port int, username, password string) (*SupervisorClient, error) {
	client, err := NewClient(host, port, username, password)
	if err != nil {
		return nil, err
	}
	return &SupervisorClient{client: client}, nil
}

// GetAllProcessInfo 获取所有进程信息 - 符合官方 API 规范
func (s *SupervisorClient) GetAllProcessInfo() ([]ProcessInfo, error) {
	result, err := s.client.Call("supervisor.getAllProcessInfo", nil)
	if err != nil {
		return nil, fmt.Errorf("XML-RPC call failed: %v", err)
	}

	xmlResponse, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	// 检查是否有fault
	if strings.Contains(xmlResponse, "<fault>") {
		return nil, fmt.Errorf("XML-RPC fault in response")
	}

	// 使用正确的 XML 解析
	processes, err := parseProcessInfoXML(xmlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %v", err)
	}
	
	return processes, nil
}

// parseProcessInfoXML 使用正确的方法解析 XML 响应
func parseProcessInfoXML(xmlResponse string) ([]ProcessInfo, error) {
	var processes []ProcessInfo
	
	// 简单但有效的方法：逐个查找每个字段
	// 查找所有 <value><struct> 块
	structs := extractStructBlocks(xmlResponse)
	
	for _, structXML := range structs {
		process := ProcessInfo{}
		
		// 提取字符串字段
		process.Name = extractFieldValue(structXML, "name")
		process.Group = extractFieldValue(structXML, "group")
		process.StateName = extractFieldValue(structXML, "statename")
		process.SpawnErr = extractFieldValue(structXML, "spawnerr")
		process.StdoutLogfile = extractFieldValue(structXML, "stdout_logfile")
		process.StderrLogfile = extractFieldValue(structXML, "stderr_logfile")
		
		// 提取整数字段
		if stateStr := extractFieldValue(structXML, "state"); stateStr != "" {
			if state, err := strconv.Atoi(stateStr); err == nil {
				process.State = state
			}
		}
		
		if pidStr := extractFieldValue(structXML, "pid"); pidStr != "" {
			if pid, err := strconv.Atoi(pidStr); err == nil {
				process.PID = pid
			}
		}
		
		if exitStr := extractFieldValue(structXML, "exitstatus"); exitStr != "" {
			if exit, err := strconv.Atoi(exitStr); err == nil {
				process.ExitStatus = exit
			}
		}
		
		// 提取时间戳字段
		if startStr := extractFieldValue(structXML, "start"); startStr != "" {
			if start, err := strconv.ParseInt(startStr, 10, 64); err == nil {
				process.Start = start
			}
		}
		
		if stopStr := extractFieldValue(structXML, "stop"); stopStr != "" {
			if stop, err := strconv.ParseInt(stopStr, 10, 64); err == nil {
				process.Stop = stop
			}
		}
		
		if nowStr := extractFieldValue(structXML, "now"); nowStr != "" {
			if now, err := strconv.ParseInt(nowStr, 10, 64); err == nil {
				process.Now = now
			}
		}
		
		// 计算运行时间 - 直接使用API提供的时间戳
		if process.State == 20 && process.Start > 0 && process.Now > 0 { // RUNNING state
			process.Uptime = process.Now - process.Start
			process.UptimeHuman = formatUptime(process.Uptime)
		}
		
		if process.Name != "" {
			processes = append(processes, process)
		}
	}
	
	return processes, nil
}

// extractStructBlocks 提取所有 struct 块
func extractStructBlocks(xmlResponse string) []string {
	var blocks []string
	start := 0
	
	for {
		structStart := strings.Index(xmlResponse[start:], "<value><struct>")
		if structStart == -1 {
			break
		}
		structStart += start + 15 // len("<value><struct>")
		
		structEnd := strings.Index(xmlResponse[structStart:], "</struct></value>")
		if structEnd == -1 {
			break
		}
		
		block := xmlResponse[structStart : structStart+structEnd]
		blocks = append(blocks, block)
		start = structStart + structEnd + 16 // len("</struct></value>")
	}
	
	return blocks
}

// extractFieldValue 提取指定字段的值 - 更精确的匹配
func extractFieldValue(structXML, fieldName string) string {
	// 构建精确的模式：<member><name>fieldName</name><value>...
	memberStart := 0
	for {
		// 查找下一个 member 块
		memberPos := strings.Index(structXML[memberStart:], "<member>")
		if memberPos == -1 {
			break
		}
		memberPos += memberStart
		
		// 查找这个 member 的结束
		memberEnd := strings.Index(structXML[memberPos:], "</member>")
		if memberEnd == -1 {
			break
		}
		memberEnd += memberPos
		
		memberContent := structXML[memberPos:memberEnd]
		
		// 检查这个 member 是否包含我们要找的字段名
		namePattern := fmt.Sprintf("<name>%s</name>", fieldName)
		if strings.Contains(memberContent, namePattern) {
			// 提取值
			valueStart := strings.Index(memberContent, "<value>")
			if valueStart == -1 {
				break
			}
			valueStart += 7 // len("<value>")
			
			valueEnd := strings.Index(memberContent[valueStart:], "</value>")
			if valueEnd == -1 {
				break
			}
			
			valueContent := memberContent[valueStart : valueStart+valueEnd]
			
			// 提取具体类型的值
			for _, valueType := range []string{"string", "int", "boolean"} {
				startTag := fmt.Sprintf("<%s>", valueType)
				endTag := fmt.Sprintf("</%s>", valueType)
				
				if start := strings.Index(valueContent, startTag); start != -1 {
					start += len(startTag)
					if end := strings.Index(valueContent[start:], endTag); end != -1 {
						return valueContent[start : start+end]
					}
				}
			}
		}
		
		memberStart = memberEnd
	}
	
	return ""
}



// formatUptime 格式化运行时间为人类可读格式
func formatUptime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, seconds%60)
	}
	
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, minutes%60)
	}
	
	days := hours / 24
	return fmt.Sprintf("%dd %dh", days, hours%24)
}

// StartProcess 启动进程
func (s *SupervisorClient) StartProcess(name string) error {
	result, err := s.client.Call("supervisor.startProcess", []interface{}{name})
	if err != nil {
		return err
	}

	if success, ok := result.(bool); !ok || !success {
		return fmt.Errorf("failed to start process %s", name)
	}

	return nil
}

// StopProcess 停止进程
func (s *SupervisorClient) StopProcess(name string) error {
	result, err := s.client.Call("supervisor.stopProcess", []interface{}{name})
	if err != nil {
		return err
	}

	if success, ok := result.(bool); !ok || !success {
		return fmt.Errorf("failed to stop process %s", name)
	}

	return nil
}

// GetProcessInfo 获取单个进程信息
func (s *SupervisorClient) GetProcessInfo(name string) (*ProcessInfo, error) {
	result, err := s.client.Call("supervisor.getProcessInfo", []interface{}{name})
	if err != nil {
		return nil, err
	}

	if info, ok := result.(map[string]interface{}); ok {
		pi := &ProcessInfo{}
		// 填充ProcessInfo字段
		if name, ok := info["name"].(string); ok {
			pi.Name = name
		}
		// ... 其他字段的解析
		return pi, nil
	}

	return nil, fmt.Errorf("invalid process info response")
}

// TailProcessStdoutLog 获取进程 stdout 日志尾部 - 符合官方 API 规范
// 返回: [string bytes, int offset, bool overflow]
func (s *SupervisorClient) TailProcessStdoutLog(name string, offset, length int) (string, int, bool, error) {
	result, err := s.client.Call("supervisor.tailProcessStdoutLog", []interface{}{name, offset, length})
	if err != nil {
		return "", 0, false, fmt.Errorf("failed to tail stdout log: %w", err)
	}

	xmlResponse, ok := result.(string)
	if !ok {
		return "", 0, false, fmt.Errorf("unexpected response type: %T", result)
	}

	// 检查是否有fault
	if strings.Contains(xmlResponse, "<fault>") {
		return "", 0, false, fmt.Errorf("XML-RPC fault in response")
	}

	// 解析 tailProcessStdoutLog 返回的数组: [bytes, offset, overflow]
	logData, nextOffset, overflow := parseTailLogResponse(xmlResponse)
	formattedLog := formatLogContent(logData)
	return formattedLog, nextOffset, overflow, nil
}

// TailProcessStderrLog 获取进程 stderr 日志尾部 - 符合官方 API 规范  
// 返回: [string bytes, int offset, bool overflow]
func (s *SupervisorClient) TailProcessStderrLog(name string, offset, length int) (string, int, bool, error) {
	result, err := s.client.Call("supervisor.tailProcessStderrLog", []interface{}{name, offset, length})
	if err != nil {
		return "", 0, false, fmt.Errorf("failed to tail stderr log: %w", err)
	}

	xmlResponse, ok := result.(string)
	if !ok {
		return "", 0, false, fmt.Errorf("unexpected response type: %T", result)
	}

	// 检查是否有fault
	if strings.Contains(xmlResponse, "<fault>") {
		return "", 0, false, fmt.Errorf("XML-RPC fault in response")
	}

	// 解析 tailProcessStderrLog 返回的数组: [bytes, offset, overflow]
	logData, nextOffset, overflow := parseTailLogResponse(xmlResponse)
	formattedLog := formatLogContent(logData)
	return formattedLog, nextOffset, overflow, nil
}

// parseTailLogResponse 解析 tailProcessLog 的响应
// Supervisor API 返回格式: [string bytes, int offset, bool overflow]
func parseTailLogResponse(xmlResponse string) (string, int, bool) {
	// 查找数组结构 <data><value>...</value><value>...</value><value>...</value></data>
	logContent := ""
	nextOffset := 0
	overflow := false
	
	// 提取第一个值 (日志内容)
	if start := strings.Index(xmlResponse, "<value><string>"); start != -1 {
		start += 15 // len("<value><string>")
		if end := strings.Index(xmlResponse[start:], "</string></value>"); end != -1 {
			logContent = xmlResponse[start : start+end]
		}
	}
	
	// 提取第二个值 (偏移量)
	// 查找第二个 <value><int>
	firstIntEnd := strings.Index(xmlResponse, "</int></value>")
	if firstIntEnd != -1 {
		searchStart := firstIntEnd + 14 // len("</int></value>")
		if start := strings.Index(xmlResponse[searchStart:], "<value><int>"); start != -1 {
			start += searchStart + 12 // len("<value><int>")
			if end := strings.Index(xmlResponse[start:], "</int></value>"); end != -1 {
				if offset, err := strconv.Atoi(xmlResponse[start : start+end]); err == nil {
					nextOffset = offset
				}
			}
		}
	}
	
	// 提取第三个值 (溢出标志)
	if strings.Contains(xmlResponse, "<value><boolean>1</boolean></value>") {
		overflow = true
	}
	
	return logContent, nextOffset, overflow
}

// formatLogContent 格式化日志内容
func formatLogContent(content string) string {
	// 移除XML转义字符
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")
	
	// 按逗号分割并格式化为多行
	lines := strings.Split(content, ",")
	var formattedLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			formattedLines = append(formattedLines, line)
		}
	}
	
	return strings.Join(formattedLines, "\n")
}
