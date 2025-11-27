package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go-cesi/internal/logger"
	"go-cesi/internal/models"
	"go-cesi/internal/supervisor"
	"go.uber.org/zap"
)

// WebSocketHub interface for broadcasting messages
type WebSocketHub interface {
	Broadcast(message []byte)
}

// AlertMonitor 监控节点和进程状态变化，自动创建和解决告警
type AlertMonitor struct {
	alertService      *AlertService
	supervisorService *supervisor.SupervisorService
	hub               WebSocketHub
	stopChan          chan struct{}
	wg                sync.WaitGroup
	mu                sync.RWMutex
	
	// 缓存上一次的状态，用于检测变化
	lastNodeStatus    map[string]bool // nodeName -> isConnected
	lastProcessStatus map[string]int  // "nodeName:processName" -> state
}

// NewAlertMonitor 创建 AlertMonitor 实例
func NewAlertMonitor(alertService *AlertService, supervisorService *supervisor.SupervisorService, hub WebSocketHub) *AlertMonitor {
	monitor := &AlertMonitor{
		alertService:      alertService,
		supervisorService: supervisorService,
		hub:               hub,
		stopChan:          make(chan struct{}),
		lastNodeStatus:    make(map[string]bool),
		lastProcessStatus: make(map[string]int),
	}
	
	// 确保系统默认规则存在
	monitor.ensureDefaultRules()
	
	return monitor
}

// ensureDefaultRules 确保系统默认告警规则存在
func (m *AlertMonitor) ensureDefaultRules() {
	// 检查规则是否存在（规则已通过数据库迁移或手动创建）
	var nodeRule models.AlertRule
	if err := m.alertService.db.Where("id = ?", 1).First(&nodeRule).Error; err != nil {
		logger.Warn("Default node offline rule (ID=1) not found, alerts may fail", zap.Error(err))
	} else {
		logger.Debug("Default node offline rule found", zap.Uint("rule_id", nodeRule.ID))
	}
	
	var processRule models.AlertRule
	if err := m.alertService.db.Where("id = ?", 2).First(&processRule).Error; err != nil {
		logger.Warn("Default process stopped rule (ID=2) not found, alerts may fail", zap.Error(err))
	} else {
		logger.Debug("Default process stopped rule found", zap.Uint("rule_id", processRule.ID))
	}
}

// Start 启动 Alert Monitor
func (m *AlertMonitor) Start() {
	logger.Info("Starting Alert Monitor")
	
	// 初始化当前状态
	m.initializeStatus()
	
	// 启动监控 goroutine
	m.wg.Add(1)
	go m.monitorLoop()
	
	logger.Info("Alert Monitor started")
}

// Stop 停止 Alert Monitor
func (m *AlertMonitor) Stop() {
	logger.Info("Stopping Alert Monitor")
	close(m.stopChan)
	m.wg.Wait()
	logger.Info("Alert Monitor stopped")
}

// initializeStatus 初始化当前节点和进程状态
func (m *AlertMonitor) initializeStatus() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	nodes := m.supervisorService.GetAllNodes()
	for _, node := range nodes {
		m.lastNodeStatus[node.Name] = node.IsConnected
		
		// 初始化进程状态
		if node.IsConnected {
			for _, process := range node.Processes {
				key := fmt.Sprintf("%s:%s", node.Name, process.Name)
				m.lastProcessStatus[key] = process.State
			}
		}
	}
	
	logger.Info("Alert Monitor status initialized",
		zap.Int("nodes", len(m.lastNodeStatus)),
		zap.Int("processes", len(m.lastProcessStatus)))
}

// monitorLoop 监控循环
func (m *AlertMonitor) monitorLoop() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.checkStatus()
		case <-m.stopChan:
			return
		}
	}
}

// checkStatus 检查所有节点和进程状态
func (m *AlertMonitor) checkStatus() {
	nodes := m.supervisorService.GetAllNodes()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 检查节点状态变化
	currentNodeStatus := make(map[string]bool)
	for _, node := range nodes {
		currentNodeStatus[node.Name] = node.IsConnected
		
		lastStatus, exists := m.lastNodeStatus[node.Name]
		
		// 节点状态变化
		if !exists || lastStatus != node.IsConnected {
			m.handleNodeStatusChange(node.Name, node.IsConnected)
		}
		
		// 检查进程状态变化（仅当节点在线时）
		if node.IsConnected {
			for _, process := range node.Processes {
				key := fmt.Sprintf("%s:%s", node.Name, process.Name)
				lastState, exists := m.lastProcessStatus[key]
				
				// 进程状态变化
				if !exists || lastState != process.State {
					m.handleProcessStatusChange(node.Name, process.Name, process.State)
					m.lastProcessStatus[key] = process.State
				}
			}
		}
	}
	
	// 更新节点状态缓存
	m.lastNodeStatus = currentNodeStatus
}

// handleNodeStatusChange 处理节点状态变化
func (m *AlertMonitor) handleNodeStatusChange(nodeName string, isConnected bool) {
	if isConnected {
		// 节点上线，解决告警
		logger.Info("Node came online, resolving alert",
			zap.String("node_name", nodeName))
		if err := m.alertService.ResolveNodeOfflineAlert(nodeName); err != nil {
			logger.Error("Failed to resolve node offline alert",
				zap.String("node_name", nodeName),
				zap.Error(err))
		} else {
			m.broadcastAlertEvent("alert_resolved", nodeName, "")
		}
	} else {
		// 节点离线，创建告警
		logger.Warn("Node went offline, creating alert",
			zap.String("node_name", nodeName))
		if err := m.alertService.CreateNodeOfflineAlert(nodeName); err != nil {
			logger.Error("Failed to create node offline alert",
				zap.String("node_name", nodeName),
				zap.Error(err))
		} else {
			m.broadcastAlertEvent("alert_created", nodeName, "")
		}
	}
}

// handleProcessStatusChange 处理进程状态变化
func (m *AlertMonitor) handleProcessStatusChange(nodeName, processName string, state int) {
	// Supervisor 状态码：
	// 0 = STOPPED
	// 10 = STARTING
	// 20 = RUNNING
	// 30 = BACKOFF
	// 40 = STOPPING
	// 100 = EXITED
	// 200 = FATAL
	// 1000 = UNKNOWN
	
	if state == 0 || state == 100 || state == 200 {
		// 进程停止/退出/致命错误，创建告警
		logger.Warn("Process stopped, creating alert",
			zap.String("node_name", nodeName),
			zap.String("process_name", processName),
			zap.Int("state", state))
		if err := m.alertService.CreateProcessStoppedAlert(nodeName, processName); err != nil {
			logger.Error("Failed to create process stopped alert",
				zap.String("node_name", nodeName),
				zap.String("process_name", processName),
				zap.Error(err))
		} else {
			m.broadcastAlertEvent("alert_created", nodeName, processName)
		}
	} else if state == 20 {
		// 进程运行中，解决告警
		logger.Info("Process started, resolving alert",
			zap.String("node_name", nodeName),
			zap.String("process_name", processName))
		if err := m.alertService.ResolveProcessStoppedAlert(nodeName, processName); err != nil {
			logger.Error("Failed to resolve process stopped alert",
				zap.String("node_name", nodeName),
				zap.String("process_name", processName),
				zap.Error(err))
		} else {
			m.broadcastAlertEvent("alert_resolved", nodeName, processName)
		}
	}
}

// CreateNodeOfflineAlert 创建节点离线告警
func (s *AlertService) CreateNodeOfflineAlert(nodeName string) error {
	alert := &models.Alert{
		RuleID:    1, // Node Offline Rule
		NodeName:  nodeName,
		Message:   fmt.Sprintf("Node '%s' is offline", nodeName),
		Severity:  models.AlertSeverityCritical,
		Status:    models.AlertStatusActive,
		StartTime: time.Now(),
	}
	
	// 依赖唯一索引防止重复，忽略冲突错误
	if err := s.db.Create(alert).Error; err != nil {
		// 唯一索引冲突 = 告警已存在，正常
		if isDuplicateError(err) {
			return nil
		}
		return err
	}
	
	logger.Info("Node offline alert created",
		zap.String("node_name", nodeName),
		zap.Uint("alert_id", alert.ID))
	return nil
}

// isDuplicateError 检查是否是重复键错误
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "UNIQUE constraint failed") || 
	       contains(errMsg, "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ResolveNodeOfflineAlert 解决节点离线告警
func (s *AlertService) ResolveNodeOfflineAlert(nodeName string) error {
	now := time.Now()
	result := s.db.Model(&models.Alert{}).
		Where("rule_id = ? AND node_name = ? AND status IN (?, ?)",
			1, nodeName, models.AlertStatusActive, models.AlertStatusAcknowledged).
		Updates(map[string]interface{}{
			"status":      models.AlertStatusResolved,
			"end_time":    now,
			"resolved_at": now,
		})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected > 0 {
		logger.Info("Node offline alert resolved",
			zap.String("node_name", nodeName),
			zap.Int64("count", result.RowsAffected))
	}
	return nil
}

// CreateProcessStoppedAlert 创建进程停止告警
func (s *AlertService) CreateProcessStoppedAlert(nodeName, processName string) error {
	alert := &models.Alert{
		RuleID:      2, // Process Stopped Rule
		NodeName:    nodeName,
		ProcessName: &processName,
		Message:     fmt.Sprintf("Process '%s' on node '%s' has stopped", processName, nodeName),
		Severity:    models.AlertSeverityHigh,
		Status:      models.AlertStatusActive,
		StartTime:   time.Now(),
	}
	
	if err := s.db.Create(alert).Error; err != nil {
		if isDuplicateError(err) {
			return nil
		}
		return err
	}
	
	logger.Info("Process stopped alert created",
		zap.String("node_name", nodeName),
		zap.String("process_name", processName),
		zap.Uint("alert_id", alert.ID))
	return nil
}

// ResolveProcessStoppedAlert 解决进程停止告警
func (s *AlertService) ResolveProcessStoppedAlert(nodeName, processName string) error {
	now := time.Now()
	result := s.db.Model(&models.Alert{}).
		Where("rule_id = ? AND node_name = ? AND process_name = ? AND status IN (?, ?)",
			2, nodeName, processName, models.AlertStatusActive, models.AlertStatusAcknowledged).
		Updates(map[string]interface{}{
			"status":      models.AlertStatusResolved,
			"end_time":    now,
			"resolved_at": now,
		})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected > 0 {
		logger.Info("Process stopped alert resolved",
			zap.String("node_name", nodeName),
			zap.String("process_name", processName),
			zap.Int64("count", result.RowsAffected))
	}
	return nil
}

// GetActiveAlerts 获取所有活跃和已确认的告警
func (s *AlertService) GetActiveAlerts() ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.db.Where("status IN (?, ?)",
		models.AlertStatusActive,
		models.AlertStatusAcknowledged).
		Order("created_at DESC").
		Find(&alerts).Error
	
	return alerts, err
}

// broadcastAlertEvent 广播告警事件到 WebSocket 客户端
func (m *AlertMonitor) broadcastAlertEvent(eventType string, nodeName string, processName string) {
	if m.hub == nil {
		return
	}
	
	event := map[string]interface{}{
		"type":         eventType,
		"node_name":    nodeName,
		"process_name": processName,
		"timestamp":    time.Now().Format(time.RFC3339),
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal alert event",
			zap.String("event_type", eventType),
			zap.Error(err))
		return
	}
	
	m.hub.Broadcast(data)
	logger.Debug("Alert event broadcasted",
		zap.String("event_type", eventType),
		zap.String("node_name", nodeName))
}
