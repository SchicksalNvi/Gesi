package supervisor

import (
	"fmt"
	"go-cesi/internal/errors"
	"strings"
	"time"

	"go-cesi/internal/supervisor/xmlrpc"
)

// ErrNodeNotConnected 节点未连接错误
var ErrNodeNotConnected = errors.NewConnectionError("node is not connected")

type Node struct {
	Name         string
	Environment  string
	Host         string
	Port         int
	Username     string
	Password     string
	IsConnected  bool
	LastPing     time.Time
	Processes    []Process
	client       *xmlrpc.SupervisorClient
}

type Process struct {
	Name         string
	Group        string
	State        int
	StateString  string
	Description  string
	StartTime    time.Time
	StopTime     time.Time
	PID          int
	ExitStatus   int
	Logfile      string
	StdoutLogfile string
	StderrLogfile string
	Uptime       time.Duration
	Now          time.Time
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
	if err != nil {
		n.IsConnected = false
		return err
	}

	n.IsConnected = true
	n.LastPing = time.Now()
	return nil
}

func (n *Node) RefreshProcesses() error {
	if !n.IsConnected {
		return ErrNodeNotConnected
	}

	processInfos, err := n.client.GetAllProcessInfo()
	if err != nil {
		return err
	}

	n.Processes = make([]Process, len(processInfos))
	for i, info := range processInfos {
		// 计算运行时长
		var uptime time.Duration
		if info.State == 20 { // RUNNING状态
			uptime = time.Since(time.Unix(info.Start, 0))
		} else if info.Stop > 0 {
			uptime = time.Unix(info.Stop, 0).Sub(time.Unix(info.Start, 0))
		}

		// 状态字符串映射
		stateString := "UNKNOWN"
		switch info.State {
		case 0:  stateString = "STOPPED"
		case 10: stateString = "STARTING"
		case 20: stateString = "RUNNING"
		case 30: stateString = "BACKOFF"
		case 40: stateString = "STOPPING"
		case 100: stateString = "EXITED"
		case 200: stateString = "FATAL"
		case 1000: stateString = "UNKNOWN"
		}

		n.Processes[i] = Process{
			Name:         info.Name,
			Group:        info.Group,
			State:        info.State,
			StateString:  stateString,
			Description:  info.Description,
			StartTime:    time.Unix(info.Start, 0),
			StopTime:     time.Unix(info.Stop, 0),
			PID:          info.PID,
			ExitStatus:   info.ExitStatus,
			Uptime:       uptime,
			Now:          time.Now(),
		}
	}

	return nil
}

func (n *Node) StartProcess(name string) error {
	if !n.IsConnected {
		return ErrNodeNotConnected
	}

	return n.client.StartProcess(name)
}

func (n *Node) StopProcess(name string) error {
	if !n.IsConnected {
		return ErrNodeNotConnected
	}

	return n.client.StopProcess(name)
}

func (n *Node) RestartProcess(name string) error {
	if !n.IsConnected {
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
	if !n.IsConnected {
		return nil, ErrNodeNotConnected
	}

	stdout, err := n.client.TailProcessStdoutLog(name, 0, 500)
	if err != nil {
		return nil, err
	}

	stderr, err := n.client.TailProcessStderrLog(name, 0, 500)
	if err != nil {
		return nil, err
	}

	return map[string][]string{
		"stdout": splitLines(stdout),
		"stderr": splitLines(stderr),
	}, nil
}

func (n *Node) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"name":         n.Name,
		"environment":  n.Environment,
		"is_connected": n.IsConnected,
		"host":         n.Host,
		"port":         n.Port,
	}
}

func (n *Node) SerializeProcesses() []map[string]interface{} {
	var processes []map[string]interface{}
	for _, p := range n.Processes {
		processes = append(processes, map[string]interface{}{
			"name":           p.Name,
			"group":          p.Group,
			"state":          p.State,
			"state_string":   p.StateString,
			"description":    p.Description,
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

// 辅助函数：将日志字符串分割成行
func splitLines(s string) []string {
	lines := make([]string, 0)
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}