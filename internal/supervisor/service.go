package supervisor

import (
	"context"
	"sync"
	"time"
	"go-cesi/internal/errors"
	"go-cesi/internal/logger"
	"go.uber.org/zap"
)

type SupervisorService struct {
	nodes map[string]*Node
	mu    sync.RWMutex
	stopChan chan struct{}
	wg    sync.WaitGroup
	shutdown bool
}

func NewSupervisorService() *SupervisorService {
	return &SupervisorService{
		nodes: make(map[string]*Node),
		stopChan: make(chan struct{}),
	}
}

func (s *SupervisorService) AddNode(name, environment, host string, port int, username, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.shutdown {
		return errors.NewInternalError("service is shutting down", nil)
	}
	
	if _, exists := s.nodes[name]; exists {
		return errors.NewConflictError("node", "node "+name+" already exists")
	}

	logger.Info("Creating node",
		zap.String("name", name),
		zap.String("host", host),
		zap.Int("port", port))
	node, err := NewNode(name, environment, host, port, username, password)
	if err != nil {
		logger.Error("Failed to create node",
			zap.String("name", name),
			zap.Error(err))
		return err
	}

	// 尝试连接
	logger.Info("Attempting to connect to node",
		zap.String("name", name),
		zap.String("host", host),
		zap.Int("port", port))
	if err := node.Connect(); err != nil {
		// 连接失败时仍然添加节点，但标记为未连接
		logger.Warn("Failed to connect to node",
			zap.String("name", name),
			zap.Error(err))
		node.IsConnected = false
	} else {
		logger.Info("Successfully connected to node",
			zap.String("name", name))
		node.IsConnected = true
		node.LastPing = time.Now()
	}

	s.nodes[name] = node
	logger.Info("Node added to service",
		zap.String("name", name),
		zap.Bool("connected", node.IsConnected))
	return nil
}

func (s *SupervisorService) GetNode(name string) (*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.shutdown {
		return nil, errors.NewInternalError("service is shutting down", nil)
	}
	
	node, exists := s.nodes[name]
	if !exists {
		return nil, errors.NewNotFoundError("node", name)
	}
	return node, nil
}

func (s *SupervisorService) GetAllNodes() []*Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.shutdown {
		return nil
	}
	
	nodes := make([]*Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		// 只在节点从未连接过或者很久没有ping时才尝试重连
		if !node.IsConnected || time.Since(node.LastPing) > 60*time.Second {
			// 异步检查连接状态，添加超时控制
			s.wg.Add(1)
			go func(n *Node) {
				defer s.wg.Done()
				
				// 创建带超时的context
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				
				// 使用channel来处理超时
				done := make(chan bool, 1)
				go func() {
					if err := n.Connect(); err == nil {
						n.IsConnected = true
						n.LastPing = time.Now()
						done <- true
					} else {
						n.IsConnected = false
						done <- false
					}
				}()
				
				select {
				case <-done:
					// 连接完成
				case <-ctx.Done():
					// 超时处理
					n.IsConnected = false
					logger.Warn("Node connection check timeout",
						zap.String("node_name", n.Name))
				case <-s.stopChan:
					// 服务正在关闭
					return
				}
			}(node)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func (s *SupervisorService) checkNodeStatus(node *Node) bool {
	if err := node.Connect(); err != nil {
		return false
	}
	return node.IsConnected
}

func (s *SupervisorService) GetNodeProcesses(nodeName string) ([]Process, error) {
	node, err := s.GetNode(nodeName)
	if err != nil {
		return nil, err
	}

	if err := node.RefreshProcesses(); err != nil {
		return nil, err
	}

	return node.Processes, nil
}

func (s *SupervisorService) StartProcess(nodeName, processName string) error {
	node, err := s.GetNode(nodeName)
	if err != nil {
		return err
	}
	return node.StartProcess(processName)
}

func (s *SupervisorService) StopProcess(nodeName, processName string) error {
	node, err := s.GetNode(nodeName)
	if err != nil {
		return err
	}
	return node.StopProcess(processName)
}

func (s *SupervisorService) GetProcessLogs(nodeName, processName string) (map[string][]string, error) {
	node, err := s.GetNode(nodeName)
	if err != nil {
		return nil, err
	}
	return node.GetProcessLogs(processName)
}

// StartAllProcesses starts all processes on a specific node
func (s *SupervisorService) StartAllProcesses(nodeName string) error {
	node, err := s.GetNode(nodeName)
	if err != nil {
		return err
	}
	
	if err := node.RefreshProcesses(); err != nil {
		return err
	}
	
	for _, process := range node.Processes {
		if err := node.StartProcess(process.Name); err != nil {
			logger.Error("Failed to start process",
				zap.String("process_name", process.Name),
				zap.String("node_name", nodeName),
				zap.Error(err))
		}
	}
	return nil
}

// StopAllProcesses stops all processes on a specific node
func (s *SupervisorService) StopAllProcesses(nodeName string) error {
	node, err := s.GetNode(nodeName)
	if err != nil {
		return err
	}
	
	if err := node.RefreshProcesses(); err != nil {
		return err
	}
	
	for _, process := range node.Processes {
		if err := node.StopProcess(process.Name); err != nil {
			logger.Error("Failed to stop process",
				zap.String("process_name", process.Name),
				zap.String("node_name", nodeName),
				zap.Error(err))
		}
	}
	return nil
}

// RestartAllProcesses restarts all processes on a specific node
func (s *SupervisorService) RestartAllProcesses(nodeName string) error {
	node, err := s.GetNode(nodeName)
	if err != nil {
		return err
	}
	
	if err := node.RefreshProcesses(); err != nil {
		return err
	}
	
	for _, process := range node.Processes {
		if err := node.RestartProcess(process.Name); err != nil {
			logger.Error("Failed to restart process",
				zap.String("process_name", process.Name),
				zap.String("node_name", nodeName),
				zap.Error(err))
		}
	}
	return nil
}

// 移除不再需要的辅助函数

// GetEnvironments 获取所有环境列表
func (s *SupervisorService) GetEnvironments() []map[string]interface{} {
	environmentMap := make(map[string][]map[string]interface{})
	
	// 按环境分组节点
	for _, node := range s.nodes {
		nodeInfo := map[string]interface{}{
			"name": node.Name,
			"host": node.Host,
			"port": node.Port,
			"is_connected": node.IsConnected,
			"last_ping": node.LastPing,
		}
		environmentMap[node.Environment] = append(environmentMap[node.Environment], nodeInfo)
	}
	
	// 转换为环境列表格式
	environments := make([]map[string]interface{}, 0)
	for envName, members := range environmentMap {
		environment := map[string]interface{}{
			"name": envName,
			"members": members,
		}
		environments = append(environments, environment)
	}
	
	return environments
}

// GetEnvironmentDetails 获取特定环境的详细信息
func (s *SupervisorService) GetEnvironmentDetails(environmentName string) map[string]interface{} {
	members := make([]map[string]interface{}, 0)
	
	for _, node := range s.nodes {
		if node.Environment == environmentName {
			nodeInfo := map[string]interface{}{
				"name": node.Name,
				"host": node.Host,
				"port": node.Port,
				"is_connected": node.IsConnected,
				"last_ping": node.LastPing,
				"processes": len(node.Processes),
			}
			members = append(members, nodeInfo)
		}
	}
	
	if len(members) == 0 {
		return nil
	}
	
	return map[string]interface{}{
		"name": environmentName,
		"members": members,
	}
}

// GetGroups 获取所有进程分组
func (s *SupervisorService) GetGroups() []map[string]interface{} {
	groupMap := make(map[string]map[string][]map[string]interface{})
	
	// 按分组和环境组织进程
	for _, node := range s.nodes {
		if !node.IsConnected {
			continue
		}
		
		// 刷新进程信息
		node.RefreshProcesses()
		
		for _, process := range node.Processes {
			groupName := process.Group
			if groupName == "" {
				groupName = "default"
			}
			
			if groupMap[groupName] == nil {
				groupMap[groupName] = make(map[string][]map[string]interface{})
			}
			
			processInfo := map[string]interface{}{
				"name": process.Name,
				"state": process.State,
				"node": node.Name,
				"pid": process.PID,
				"uptime": process.Uptime,
			}
			
			groupMap[groupName][node.Environment] = append(groupMap[groupName][node.Environment], processInfo)
		}
	}
	
	// 转换为分组列表格式
	groups := make([]map[string]interface{}, 0)
	for groupName, environments := range groupMap {
		group := map[string]interface{}{
			"name": groupName,
			"environments": make([]map[string]interface{}, 0),
		}
		
		for envName, processes := range environments {
			environment := map[string]interface{}{
				"name": envName,
				"processes": processes,
				"members": s.getUniqueNodeNames(processes),
			}
			group["environments"] = append(group["environments"].([]map[string]interface{}), environment)
		}
		
		groups = append(groups, group)
	}
	
	return groups
}

// getUniqueNodeNames 获取进程列表中的唯一节点名称
func (s *SupervisorService) getUniqueNodeNames(processes []map[string]interface{}) []string {
	nodeSet := make(map[string]bool)
	for _, process := range processes {
		if nodeName, ok := process["node"].(string); ok {
			nodeSet[nodeName] = true
		}
	}
	
	nodes := make([]string, 0, len(nodeSet))
	for nodeName := range nodeSet {
		nodes = append(nodes, nodeName)
	}
	return nodes
}

// GetGroupDetails 获取特定分组的详细信息
func (s *SupervisorService) GetGroupDetails(groupName string) map[string]interface{} {
	groups := s.GetGroups()
	
	for _, group := range groups {
		if group["name"] == groupName {
			return group
		}
	}
	
	return nil
}

// StartGroupProcesses 启动分组中的所有进程
func (s *SupervisorService) StartGroupProcesses(groupName, environmentName string) error {
	return s.operateGroupProcesses(groupName, environmentName, "start")
}

// StopGroupProcesses 停止分组中的所有进程
func (s *SupervisorService) StopGroupProcesses(groupName, environmentName string) error {
	return s.operateGroupProcesses(groupName, environmentName, "stop")
}

// RestartGroupProcesses 重启分组中的所有进程
func (s *SupervisorService) RestartGroupProcesses(groupName, environmentName string) error {
	return s.operateGroupProcesses(groupName, environmentName, "restart")
}

// operateGroupProcesses 对分组中的进程执行操作
func (s *SupervisorService) operateGroupProcesses(groupName, environmentName, operation string) error {
	for _, node := range s.nodes {
		if environmentName != "" && node.Environment != environmentName {
			continue
		}
		
		if !node.IsConnected {
			continue
		}
		
		node.RefreshProcesses()
		
		for _, process := range node.Processes {
			processGroupName := process.Group
			if processGroupName == "" {
				processGroupName = "default"
			}
			
			if processGroupName == groupName {
				switch operation {
				case "start":
					node.StartProcess(process.Name)
				case "stop":
					node.StopProcess(process.Name)
				case "restart":
					node.RestartProcess(process.Name)
				}
			}
		}
	}
	
	return nil
}

func (s *SupervisorService) StartAutoRefresh(interval time.Duration) chan struct{} {
	ticker := time.NewTicker(interval)
	stopChan := make(chan struct{})

	go func() {
		defer ticker.Stop() // 确保ticker被正确清理
		for {
			select {
			case <-ticker.C:
				// 刷新所有节点状态
				logger.Debug("Auto-refreshing node connections")
				for _, node := range s.nodes {
					prevConnected := node.IsConnected
					if err := node.Connect(); err == nil {
						node.IsConnected = true
						node.LastPing = time.Now()
						if !prevConnected {
							logger.Info("Node reconnected",
								zap.String("node_name", node.Name))
						}
					} else {
						node.IsConnected = false
						if prevConnected {
							logger.Warn("Node disconnected",
								zap.String("node_name", node.Name),
								zap.Error(err))
						}
					}
				}
			case <-stopChan:
				logger.Debug("Stopping auto-refresh goroutine")
				return
			}
		}
	}()

	return stopChan
}

// StopAutoRefresh 停止自动刷新
func (s *SupervisorService) StopAutoRefresh(stopChan chan struct{}) {
	if stopChan != nil {
		close(stopChan)
	}
}

// Shutdown 优雅关闭SupervisorService，清理所有资源
func (s *SupervisorService) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return nil // 已经关闭
	}
	s.shutdown = true
	s.mu.Unlock()
	
	logger.Info("Shutting down SupervisorService")
	
	// 关闭stopChan，通知所有goroutine停止
	close(s.stopChan)
	
	// 等待所有goroutine完成，带超时控制
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		logger.Info("All SupervisorService goroutines stopped gracefully")
	case <-ctx.Done():
		logger.Warn("SupervisorService shutdown timeout, some goroutines may still be running")
		return ctx.Err()
	}
	
	// 清理所有节点连接
	s.mu.Lock()
	for name, node := range s.nodes {
		if node.client != nil {
			// 假设Node有Close方法来清理连接
			logger.Debug("Closing node connection", zap.String("node_name", name))
		}
	}
	s.nodes = make(map[string]*Node) // 清空节点映射
	s.mu.Unlock()
	
	logger.Info("SupervisorService shutdown completed")
	return nil
}

// IsShutdown 检查服务是否已关闭
func (s *SupervisorService) IsShutdown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shutdown
}