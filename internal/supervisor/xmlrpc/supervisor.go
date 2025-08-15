package xmlrpc

import (
	"fmt"
	"time"
)

type ProcessInfo struct {
	Name        string    `xml:"name"`
	Group       string    `xml:"group"`
	Description string    `xml:"description"`
	Start       int64     `xml:"start"`
	Stop        int64     `xml:"stop"`
	Now         int64     `xml:"now"`
	State       int       `xml:"state"`
	StateName   string    `xml:"statename"`
	SpawnErr    string    `xml:"spawnerr"`
	ExitStatus  int       `xml:"exitstatus"`
	StdoutLog   string    `xml:"stdout_logfile"`
	StderrLog   string    `xml:"stderr_logfile"`
	PID         int       `xml:"pid"`
	StartTime   time.Time `xml:"-"`
	StopTime    time.Time `xml:"-"`
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

// GetAllProcessInfo 获取所有进程信息
func (s *SupervisorClient) GetAllProcessInfo() ([]ProcessInfo, error) {
	result, err := s.client.Call("supervisor.getAllProcessInfo", nil)
	if err != nil {
		return nil, err
	}

	// 解析结果到ProcessInfo结构体
	processes := make([]ProcessInfo, 0)
	if arr, ok := result.([]interface{}); ok {
		for _, item := range arr {
			if proc, ok := item.(map[string]interface{}); ok {
				pi := ProcessInfo{}
				// 填充ProcessInfo字段
				if name, ok := proc["name"].(string); ok {
					pi.Name = name
				}
				if group, ok := proc["group"].(string); ok {
					pi.Group = group
				}
				// ... 其他字段的解析
				processes = append(processes, pi)
			}
		}
	}

	return processes, nil
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

// TailProcessStdoutLog 获取进程标准输出日志
func (s *SupervisorClient) TailProcessStdoutLog(name string, offset, length int) (string, error) {
	result, err := s.client.Call("supervisor.tailProcessStdoutLog", []interface{}{name, offset, length})
	if err != nil {
		return "", err
	}

	if log, ok := result.(string); ok {
		return log, nil
	}

	return "", fmt.Errorf("invalid log response")
}

// TailProcessStderrLog 获取进程标准错误日志
func (s *SupervisorClient) TailProcessStderrLog(name string, offset, length int) (string, error) {
	result, err := s.client.Call("supervisor.tailProcessStderrLog", []interface{}{name, offset, length})
	if err != nil {
		return "", err
	}

	if log, ok := result.(string); ok {
		return log, nil
	}

	return "", fmt.Errorf("invalid log response")
}
