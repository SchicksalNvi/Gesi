package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"superview/internal/supervisor"
)

// PrometheusMetrics 管理 Prometheus 指标收集
type PrometheusMetrics struct {
	supervisorService *supervisor.SupervisorService
	mu                sync.RWMutex
	lastCollectTime   time.Time
	cachedMetrics     string
	cacheDuration     time.Duration
}

// NewPrometheusMetrics 创建 PrometheusMetrics 实例
func NewPrometheusMetrics(svc *supervisor.SupervisorService) *PrometheusMetrics {
	return &PrometheusMetrics{
		supervisorService: svc,
		cacheDuration:     5 * time.Second, // 缓存5秒避免频繁采集
	}
}

// Handler 返回 Prometheus metrics HTTP handler
func (p *PrometheusMetrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := p.collectMetrics()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metrics))
	}
}

// collectMetrics 收集所有指标
func (p *PrometheusMetrics) collectMetrics() string {
	p.mu.RLock()
	if time.Since(p.lastCollectTime) < p.cacheDuration && p.cachedMetrics != "" {
		defer p.mu.RUnlock()
		return p.cachedMetrics
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查
	if time.Since(p.lastCollectTime) < p.cacheDuration && p.cachedMetrics != "" {
		return p.cachedMetrics
	}

	var sb strings.Builder

	// 写入 HELP 和 TYPE 注释
	p.writeMetricHelp(&sb)

	// 收集节点指标
	p.collectNodeMetrics(&sb)

	// 收集进程指标
	p.collectProcessMetrics(&sb)

	// 收集汇总指标
	p.collectSummaryMetrics(&sb)

	p.cachedMetrics = sb.String()
	p.lastCollectTime = time.Now()

	return p.cachedMetrics
}

// writeMetricHelp 写入指标帮助信息
func (p *PrometheusMetrics) writeMetricHelp(sb *strings.Builder) {
	// 节点指标
	sb.WriteString("# HELP superview_node_up Node connection status (1=connected, 0=disconnected)\n")
	sb.WriteString("# TYPE superview_node_up gauge\n")

	sb.WriteString("# HELP superview_node_last_ping_timestamp_seconds Last successful ping timestamp\n")
	sb.WriteString("# TYPE superview_node_last_ping_timestamp_seconds gauge\n")

	// 进程指标
	sb.WriteString("# HELP superview_process_state Process state (0=STOPPED, 10=STARTING, 20=RUNNING, 30=BACKOFF, 40=STOPPING, 100=EXITED, 200=FATAL, 1000=UNKNOWN)\n")
	sb.WriteString("# TYPE superview_process_state gauge\n")

	sb.WriteString("# HELP superview_process_up Process running status (1=running, 0=not running)\n")
	sb.WriteString("# TYPE superview_process_up gauge\n")

	sb.WriteString("# HELP superview_process_pid Process PID\n")
	sb.WriteString("# TYPE superview_process_pid gauge\n")

	sb.WriteString("# HELP superview_process_uptime_seconds Process uptime in seconds\n")
	sb.WriteString("# TYPE superview_process_uptime_seconds gauge\n")

	sb.WriteString("# HELP superview_process_start_timestamp_seconds Process start timestamp\n")
	sb.WriteString("# TYPE superview_process_start_timestamp_seconds gauge\n")

	sb.WriteString("# HELP superview_process_exit_status Process exit status code\n")
	sb.WriteString("# TYPE superview_process_exit_status gauge\n")

	// 汇总指标
	sb.WriteString("# HELP superview_nodes_total Total number of configured nodes\n")
	sb.WriteString("# TYPE superview_nodes_total gauge\n")

	sb.WriteString("# HELP superview_nodes_connected Number of connected nodes\n")
	sb.WriteString("# TYPE superview_nodes_connected gauge\n")

	sb.WriteString("# HELP superview_processes_total Total number of processes\n")
	sb.WriteString("# TYPE superview_processes_total gauge\n")

	sb.WriteString("# HELP superview_processes_running Number of running processes\n")
	sb.WriteString("# TYPE superview_processes_running gauge\n")

	sb.WriteString("# HELP superview_processes_stopped Number of stopped processes\n")
	sb.WriteString("# TYPE superview_processes_stopped gauge\n")

	sb.WriteString("# HELP superview_processes_failed Number of failed processes (FATAL or EXITED with error)\n")
	sb.WriteString("# TYPE superview_processes_failed gauge\n")

	sb.WriteString("# HELP superview_info Superview build information\n")
	sb.WriteString("# TYPE superview_info gauge\n")
}

// collectNodeMetrics 收集节点指标
func (p *PrometheusMetrics) collectNodeMetrics(sb *strings.Builder) {
	nodes := p.supervisorService.GetAllNodes()
	if nodes == nil {
		return
	}

	for _, node := range nodes {
		isConnected, lastPing := node.GetConnectionStatus()

		// superview_node_up
		upValue := 0
		if isConnected {
			upValue = 1
		}
		sb.WriteString(fmt.Sprintf("superview_node_up{node=\"%s\",environment=\"%s\",host=\"%s\",port=\"%d\"} %d\n",
			escapeLabel(node.Name),
			escapeLabel(node.Environment),
			escapeLabel(node.Host),
			node.Port,
			upValue))

		// superview_node_last_ping_timestamp_seconds
		if !lastPing.IsZero() {
			sb.WriteString(fmt.Sprintf("superview_node_last_ping_timestamp_seconds{node=\"%s\"} %d\n",
				escapeLabel(node.Name),
				lastPing.Unix()))
		}
	}
}

// collectProcessMetrics 收集进程指标
func (p *PrometheusMetrics) collectProcessMetrics(sb *strings.Builder) {
	nodes := p.supervisorService.GetAllNodes()
	if nodes == nil {
		return
	}

	for _, node := range nodes {
		isConnected, _ := node.GetConnectionStatus()
		if !isConnected {
			continue
		}

		processes := node.SerializeProcesses()
		for _, proc := range processes {
			name, _ := proc["name"].(string)
			group, _ := proc["group"].(string)
			state, _ := proc["state"].(int)
			pid, _ := proc["pid"].(int)
			uptime, _ := proc["uptime"].(float64)
			startTime, _ := proc["start"].(int64)
			exitStatus, _ := proc["exitstatus"].(int)

			labels := fmt.Sprintf("node=\"%s\",process=\"%s\",group=\"%s\"",
				escapeLabel(node.Name),
				escapeLabel(name),
				escapeLabel(group))

			// superview_process_state
			sb.WriteString(fmt.Sprintf("superview_process_state{%s} %d\n", labels, state))

			// superview_process_up (1 if RUNNING, 0 otherwise)
			upValue := 0
			if state == 20 { // RUNNING
				upValue = 1
			}
			sb.WriteString(fmt.Sprintf("superview_process_up{%s} %d\n", labels, upValue))

			// superview_process_pid
			sb.WriteString(fmt.Sprintf("superview_process_pid{%s} %d\n", labels, pid))

			// superview_process_uptime_seconds
			sb.WriteString(fmt.Sprintf("superview_process_uptime_seconds{%s} %.0f\n", labels, uptime))

			// superview_process_start_timestamp_seconds
			if startTime > 0 {
				sb.WriteString(fmt.Sprintf("superview_process_start_timestamp_seconds{%s} %d\n", labels, startTime))
			}

			// superview_process_exit_status
			sb.WriteString(fmt.Sprintf("superview_process_exit_status{%s} %d\n", labels, exitStatus))
		}
	}
}

// 构建信息 - 从环境变量或编译时注入获取版本
func (p *PrometheusMetrics) getVersion() string {
	// 可通过 -ldflags "-X superview/internal/metrics.Version=x.x.x" 注入
	if Version != "" {
		return Version
	}
	return "dev"
}

// Version 版本号，可通过编译时注入
var Version = ""

// collectSummaryMetrics 收集汇总指标
func (p *PrometheusMetrics) collectSummaryMetrics(sb *strings.Builder) {
	nodes := p.supervisorService.GetAllNodes()
	if nodes == nil {
		return
	}

	totalNodes := len(nodes)
	connectedNodes := 0
	totalProcesses := 0
	runningProcesses := 0
	stoppedProcesses := 0
	failedProcesses := 0

	// 按环境统计
	envStats := make(map[string]struct {
		total     int
		connected int
		processes int
		running   int
	})

	for _, node := range nodes {
		isConnected, _ := node.GetConnectionStatus()
		if isConnected {
			connectedNodes++
		}

		env := node.Environment
		stats := envStats[env]
		stats.total++
		if isConnected {
			stats.connected++
		}

		if isConnected {
			processes := node.SerializeProcesses()
			for _, proc := range processes {
				totalProcesses++
				stats.processes++
				state, _ := proc["state"].(int)
				switch state {
				case 20: // RUNNING
					runningProcesses++
					stats.running++
				case 0: // STOPPED
					stoppedProcesses++
				case 100, 200: // EXITED, FATAL
					failedProcesses++
				}
			}
		}
		envStats[env] = stats
	}

	// 总体指标
	sb.WriteString(fmt.Sprintf("superview_nodes_total %d\n", totalNodes))
	sb.WriteString(fmt.Sprintf("superview_nodes_connected %d\n", connectedNodes))
	sb.WriteString(fmt.Sprintf("superview_processes_total %d\n", totalProcesses))
	sb.WriteString(fmt.Sprintf("superview_processes_running %d\n", runningProcesses))
	sb.WriteString(fmt.Sprintf("superview_processes_stopped %d\n", stoppedProcesses))
	sb.WriteString(fmt.Sprintf("superview_processes_failed %d\n", failedProcesses))

	// 按环境统计
	envNames := make([]string, 0, len(envStats))
	for env := range envStats {
		envNames = append(envNames, env)
	}
	sort.Strings(envNames)

	for _, env := range envNames {
		stats := envStats[env]
		sb.WriteString(fmt.Sprintf("superview_nodes_total{environment=\"%s\"} %d\n", escapeLabel(env), stats.total))
		sb.WriteString(fmt.Sprintf("superview_nodes_connected{environment=\"%s\"} %d\n", escapeLabel(env), stats.connected))
		sb.WriteString(fmt.Sprintf("superview_processes_total{environment=\"%s\"} %d\n", escapeLabel(env), stats.processes))
		sb.WriteString(fmt.Sprintf("superview_processes_running{environment=\"%s\"} %d\n", escapeLabel(env), stats.running))
	}

	// 构建信息
	sb.WriteString(fmt.Sprintf("superview_info{version=\"%s\"} 1\n", p.getVersion()))
}

// escapeLabel 转义 Prometheus label 值中的特殊字符
func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
