package supervisor

import (
	"fmt"
	"go-cesi/internal/errors"
	"go-cesi/internal/logger"
	"strings"
	"sync"
	"time"

	"go-cesi/internal/supervisor/xmlrpc"
	"go.uber.org/zap"
)

// ErrNodeNotConnected 节点未连接错误
var ErrNodeNotConnected = errors.NewConnectionError("node", nil)

type Node struct {
	Name         string
	Environment  string
	Host         string
	Port         int
	Username     string
	Password     string
	
	// 需要同步保护的字段
	mu           sync.RWMutex
	IsConnected  bool
	LastPing     time.Time
	Processes    []Process
	
	client       *xmlrpc.SupervisorClient
}

// Process 表示一个进程 - 符合 Supervisor API 规范
type Process struct {
	Name           string        `json:"name"`           // 进程名称
	Group          string        `json:"group"`          // 进程组名称
	State          int           `json:"state"`          // 状态码
	StateString    string        `json:"state_string"`   // 状态名称
	StartTime      time.Time     `json:"start_time"`     // 启动时间
	StopTime       time.Time     `json:"stop_time"`      // 停止时间
	PID            int           `json:"pid"`            // 进程 PID
	ExitStatus     int           `json:"exit_status"`    // 退出状态码
	Logfile        string        `json:"logfile"`        // 日志文件 (已弃用)
	StdoutLogfile  string        `json:"stdout_logfile"` // stdout 日志文件
	StderrLogfile  string        `json:"stderr_logfile"` // stderr 日志文件
	Uptime         time.Duration `json:"uptime"`         // 运行时间
	UptimeHuman    string        `json:"uptime_human"`   // 人类可读的运行时间
	SpawnErr       string        `json:"spawn_err"`      // 启动错误
	Now            time.Time     `json:"now"`            // 当前时间
}

type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Source      string    `json:"source"`
	ProcessName string    `json:"process_name"`
	NodeName    string    `json:"node_name"`
}

// LogStream 表示日志流 - 符合 Supervisor API 规范
type LogStream struct {
	ProcessName string     `json:"process_name"`
	NodeName    string     `json:"node_name"`
	LogType     string     `json:"log_type"`     // "stdout" 或 "stderr"
	Entries     []LogEntry `json:"entries"`
	LastOffset  int        `json:"last_offset"`  // 下一次读取的偏移量
	Overflow    bool       `json:"overflow"`     // 是否有日志溢出
}

func NewNode(name, environment, host string, port int, username, password string) (*Node, error) {
	client, err := xmlrpc.NewSupervisorClient(host, port, username, password)
	if err != nil {
		return nil, err
	}

	return &Node{
		Name:        name,
		Environment: environment,
		Host:        host,
		Port:        port,
		Username:    username,
		Password:    password,
		client:      client,
		Processes:   make([]Process, 0),
	}, nil
}

func (n *Node) Connect() error {
	// 尝试获取进程信息来测试连接
	_, err := n.client.GetAllProcessInfo()
	
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if err != nil {
		n.IsConnected = false
		return err
	}

	n.IsConnected = true
	n.LastPing = time.Now()
	return nil
}

func (n *Node) RefreshProcesses() error {
	n.mu.RLock()
	connected := n.IsConnected
	n.mu.RUnlock()
	
	if !connected {
		return ErrNodeNotConnected
	}

	logger.Debug("Refreshing processes for node",
		zap.String("node", n.Name),
		zap.String("host", n.Host),
		zap.Int("port", n.Port))

	processInfos, err := n.client.GetAllProcessInfo()
	if err != nil {
		logger.Error("Failed to get process info from node",
			zap.String("node", n.Name),
			zap.String("host", n.Host),
			zap.Int("port", n.Port),
			zap.Error(err))
		return err
	}

	logger.Debug("Retrieved process info from node",
		zap.String("node", n.Name),
		zap.Int("process_count", len(processInfos)))

	processes := make([]Process, len(processInfos))
	for i, info := range processInfos {
		// 转换时间戳
		var startTime, stopTime, nowTime time.Time
		if info.Start > 0 {
			startTime = time.Unix(info.Start, 0)
		}
		if info.Stop > 0 {
			stopTime = time.Unix(info.Stop, 0)
		}
		if info.Now > 0 {
			nowTime = time.Unix(info.Now, 0)
		}

		// 计算运行时长
		var uptime time.Duration
		var uptimeHuman string
		
		if info.State == 20 && info.Start > 0 && info.Now > 0 { // RUNNING状态
			// 使用当前时间和启动时间计算实际运行时间
			actualUptime := info.Now - info.Start
			uptime = time.Duration(actualUptime) * time.Second
			uptimeHuman = formatDuration(uptime)
		} else {
			// 对于非运行状态的进程，uptime为0
			uptime = 0
			uptimeHuman = "0s"
		}

		// 状态字符串映射 - 符合 Supervisor 规范
		stateString := getStateNameFromCode(info.State)

		processes[i] = Process{
			Name:          info.Name,
			Group:         info.Group,
			State:         info.State,
			StateString:   stateString,
			StartTime:     startTime,
			StopTime:      stopTime,
			PID:           getPIDForState(info.State, info.PID),
			ExitStatus:    info.ExitStatus,
			StdoutLogfile: info.StdoutLogfile,
			StderrLogfile: info.StderrLogfile,
			SpawnErr:      info.SpawnErr,
			Uptime:        uptime,
			UptimeHuman:   uptimeHuman,
			Now:           nowTime,
		}
	}

	n.mu.Lock()
	n.Processes = processes
	n.mu.Unlock()

	logger.Info("Successfully refreshed processes for node",
		zap.String("node", n.Name),
		zap.Int("process_count", len(processes)))

	return nil
}

// getStateNameFromCode 根据状态码获取状态名称 - 符合 Supervisor 规范
func getStateNameFromCode(state int) string {
	switch state {
	case 0:
		return "STOPPED"
	case 10:
		return "STARTING"
	case 20:
		return "RUNNING"
	case 30:
		return "BACKOFF"
	case 40:
		return "STOPPING"
	case 100:
		return "EXITED"
	case 200:
		return "FATAL"
	case 1000:
		return "UNKNOWN"
	default:
		return fmt.Sprintf("STATE_%d", state)
	}
}

// getPIDForState 根据进程状态返回正确的PID
// 只有运行状态的进程才有有效的PID
func getPIDForState(state int, originalPID int) int {
	switch state {
	case 20: // RUNNING
		return originalPID
	case 10: // STARTING - 可能有PID
		return originalPID
	default:
		// STOPPED, EXITED, FATAL等状态下PID应该为0
		return 0
	}
}

func (n *Node) StartProcess(name string) error {
	n.mu.RLock()
	connected := n.IsConnected
	n.mu.RUnlock()
	
	if !connected {
		return ErrNodeNotConnected
	}

	return n.client.StartProcess(name)
}

func (n *Node) StopProcess(name string) error {
	n.mu.RLock()
	connected := n.IsConnected
	n.mu.RUnlock()
	
	if !connected {
		return ErrNodeNotConnected
	}

	return n.client.StopProcess(name)
}

func (n *Node) RestartProcess(name string) error {
	n.mu.RLock()
	connected := n.IsConnected
	n.mu.RUnlock()
	
	if !connected {
		return ErrNodeNotConnected
	}

	// 先停止进程
	err := n.client.StopProcess(name)
	if err != nil {
		return err
	}
	
	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)
	
	// 再启动进程
	return n.client.StartProcess(name)
}

func (n *Node) GetProcessLogs(name string) (map[string][]string, error) {
	n.mu.RLock()
	connected := n.IsConnected
	n.mu.RUnlock()
	
	if !connected {
		return nil, ErrNodeNotConnected
	}

	// 使用新的 API 签名 - 返回 (content, offset, overflow, error)
	stdout, _, _, err := n.client.TailProcessStdoutLog(name, 0, 500)
	if err != nil {
		return nil, err
	}

	stderr, _, _, err := n.client.TailProcessStderrLog(name, 0, 500)
	if err != nil {
		return nil, err
	}

	// 分割日志行，过滤空行
	stdoutLines := make([]string, 0)
	for _, line := range strings.Split(stdout, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			stdoutLines = append(stdoutLines, line)
		}
	}

	stderrLines := make([]string, 0)
	for _, line := range strings.Split(stderr, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			stderrLines = append(stderrLines, line)
		}
	}

	return map[string][]string{
		"stdout": stdoutLines,
		"stderr": stderrLines,
	}, nil
}

// GetProcessLogStream 获取结构化的日志流 - 使用正确的 Supervisor API
func (n *Node) GetProcessLogStream(name string, offset int, maxLines int) (*LogStream, error) {
	n.mu.RLock()
	connected := n.IsConnected
	n.mu.RUnlock()
	
	if !connected {
		return nil, ErrNodeNotConnected
	}

	// 使用 tailProcessStdoutLog API，它返回 [bytes, nextOffset, overflow]
	bytesToRead := maxLines * 100 // 估算每行 100 字节
	stdout, nextOffset, overflow, err := n.client.TailProcessStdoutLog(name, offset, bytesToRead)
	if err != nil {
		return nil, err
	}

	// 解析日志为结构化条目
	entries := n.parseLogEntries(stdout, "stdout", name)

	// 限制返回的条目数
	if len(entries) > maxLines {
		entries = entries[len(entries)-maxLines:]
	}

	return &LogStream{
		ProcessName: name,
		NodeName:    n.Name,
		LogType:     "stdout",
		Entries:     entries,
		LastOffset:  nextOffset,  // 使用 API 返回的正确偏移量
		Overflow:    overflow,    // 标记是否有日志溢出
	}, nil
}

// parseLogEntries 解析日志文本为结构化条目
func (n *Node) parseLogEntries(logText, logType, processName string) []LogEntry {
	var entries []LogEntry
	
	lines := strings.Split(logText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		entry := LogEntry{
			Timestamp:   time.Now(), // 默认当前时间，后续可以解析日志中的时间戳
			Level:       extractLogLevel(line),
			Message:     line,
			Source:      logType,
			ProcessName: processName,
			NodeName:    n.Name,
		}
		
		// 尝试解析时间戳
		if timestamp := extractTimestamp(line); !timestamp.IsZero() {
			entry.Timestamp = timestamp
		}
		
		entries = append(entries, entry)
	}
	
	return entries
}

// extractLogLevel 从日志行中提取日志级别
func extractLogLevel(line string) string {
	line = strings.ToUpper(line)
	
	levels := []string{"ERROR", "WARN", "WARNING", "INFO", "DEBUG", "TRACE", "FATAL"}
	for _, level := range levels {
		if strings.Contains(line, level) {
			return level
		}
	}
	return "INFO" // 默认级别
}

// extractTimestamp 从日志行中提取时间戳
func extractTimestamp(line string) time.Time {
	// 常见的时间戳格式
	formats := []string{
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"Jan 02 15:04:05",
	}
	
	for _, format := range formats {
		// 尝试从行首提取时间戳
		if len(line) >= len(format) {
			if t, err := time.Parse(format, line[:len(format)]); err == nil {
				return t
			}
		}
		
		// 尝试查找时间戳模式
		for i := 0; i <= len(line)-len(format); i++ {
			if t, err := time.Parse(format, line[i:i+len(format)]); err == nil {
				return t
			}
		}
	}
	
	return time.Time{} // 返回零值表示未找到
}

func (n *Node) Serialize() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	return map[string]interface{}{
		"name":         n.Name,
		"environment":  n.Environment,
		"is_connected": n.IsConnected,
		"host":         n.Host,
		"port":         n.Port,
	}
}

func (n *Node) SerializeProcesses() []map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	var processes []map[string]interface{}
	for _, p := range n.Processes {
		processes = append(processes, map[string]interface{}{
			"name":           p.Name,
			"group":          p.Group,
			"state":          p.State,
			"state_string":   p.StateString,
			"start_time":     p.StartTime,
			"stop_time":      p.StopTime,
			"pid":           p.PID,
			"exit_status":   p.ExitStatus,
			"logfile":       p.Logfile,
			"stdout_logfile": p.StdoutLogfile,
			"stderr_logfile": p.StderrLogfile,
			"uptime":        p.Uptime.Seconds(),
			"uptime_human":  formatDuration(p.Uptime),
			"now":           p.Now,
		})
	}
	return processes
}

// GetConnectionStatus 安全地获取连接状态
func (n *Node) GetConnectionStatus() (bool, time.Time) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.IsConnected, n.LastPing
}

// GetProcessCount 安全地获取进程数量
func (n *Node) GetProcessCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.Processes)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}