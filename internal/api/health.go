package api

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"go-cesi/internal/logger"
	"go-cesi/internal/supervisor"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HealthAPI struct {
	db            *gorm.DB
	supervisorSvc *supervisor.SupervisorService
	startTime     time.Time
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceInfo `json:"services"`
	System    SystemInfo             `json:"system"`
	Database  DatabaseInfo           `json:"database"`
	Nodes     NodesInfo              `json:"nodes"`
}

type ServiceInfo struct {
	Status       string    `json:"status"`
	LastCheck    time.Time `json:"last_check"`
	Message      string    `json:"message,omitempty"`
	ResponseTime string    `json:"response_time,omitempty"`
}

type SystemInfo struct {
	GoVersion   string  `json:"go_version"`
	Goroutines  int     `json:"goroutines"`
	MemoryUsage string  `json:"memory_usage"`
	CPUCount    int     `json:"cpu_count"`
	LoadAverage float64 `json:"load_average,omitempty"`
}

type DatabaseInfo struct {
	Status       string `json:"status"`
	Connections  int    `json:"connections"`
	ResponseTime string `json:"response_time"`
	Message      string `json:"message,omitempty"`
}

type NodesInfo struct {
	Total   int `json:"total"`
	Online  int `json:"online"`
	Offline int `json:"offline"`
	Unknown int `json:"unknown"`
}

func NewHealthAPI(db *gorm.DB, supervisorSvc *supervisor.SupervisorService) *HealthAPI {
	return &HealthAPI{
		db:            db,
		supervisorSvc: supervisorSvc,
		startTime:     time.Now(),
	}
}

// GetHealth 获取系统健康状态
func (h *HealthAPI) GetHealth(c *gin.Context) {
	startTime := time.Now()

	// 检查数据库状态
	dbInfo := h.checkDatabaseHealth()

	// 检查节点状态
	nodesInfo := h.checkNodesHealth()

	// 检查系统信息
	systemInfo := h.getSystemInfo()

	// 检查各个服务状态
	services := map[string]ServiceInfo{
		"database": {
			Status:       dbInfo.Status,
			LastCheck:    time.Now(),
			Message:      dbInfo.Message,
			ResponseTime: dbInfo.ResponseTime,
		},
		"supervisor": {
			Status:       h.getSupervisorStatus(),
			LastCheck:    time.Now(),
			ResponseTime: time.Since(startTime).String(),
		},
		"logger": {
			Status:    "healthy",
			LastCheck: time.Now(),
			Message:   "Logger service is operational",
		},
	}

	// 确定整体状态
	overallStatus := "healthy"
	if dbInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	} else if nodesInfo.Offline > 0 {
		overallStatus = "degraded"
	}

	health := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
		Version:   "1.0.0", // 可以从配置或构建信息获取
		Services:  services,
		System:    systemInfo,
		Database:  dbInfo,
		Nodes:     nodesInfo,
	}

	// 根据状态返回相应的HTTP状态码
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, health)
}

// GetHealthLive 简单的存活检查
func (h *HealthAPI) GetHealthLive(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// GetHealthReady 就绪检查
func (h *HealthAPI) GetHealthReady(c *gin.Context) {
	// 检查关键依赖是否就绪
	dbReady := h.isDatabaseReady()

	if !dbReady {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"message": "Database is not ready",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now(),
	})
}

func (h *HealthAPI) checkDatabaseHealth() DatabaseInfo {
	startTime := time.Now()

	// 执行简单的数据库查询来检查连接
	var result int
	err := h.db.Raw("SELECT 1").Scan(&result).Error

	responseTime := time.Since(startTime).String()

	if err != nil {
		logger.Error("Database health check failed", zap.Error(err))
		return DatabaseInfo{
			Status:       "unhealthy",
			Connections:  0,
			ResponseTime: responseTime,
			Message:      err.Error(),
		}
	}

	// 获取数据库连接池信息
	sqlDB, err := h.db.DB()
	connections := 0
	if err == nil {
		stats := sqlDB.Stats()
		connections = stats.OpenConnections
	}

	return DatabaseInfo{
		Status:       "healthy",
		Connections:  connections,
		ResponseTime: responseTime,
		Message:      "Database connection is healthy",
	}
}

func (h *HealthAPI) checkNodesHealth() NodesInfo {
	if h.supervisorSvc == nil {
		return NodesInfo{
			Total:   0,
			Online:  0,
			Offline: 0,
			Unknown: 0,
		}
	}

	nodes := h.supervisorSvc.GetAllNodes()
	nodesInfo := NodesInfo{
		Total: len(nodes),
	}

	for _, node := range nodes {
		if node.IsConnected {
			nodesInfo.Online++
		} else {
			nodesInfo.Offline++
		}
	}

	return nodesInfo
}

func (h *HealthAPI) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		MemoryUsage: h.formatBytes(m.Alloc),
		CPUCount:    runtime.NumCPU(),
	}
}

// formatBytes formats bytes to human readable format
func (h *HealthAPI) formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (h *HealthAPI) getSupervisorStatus() string {
	if h.supervisorSvc == nil {
		return "unavailable"
	}
	return "healthy"
}

func (h *HealthAPI) isDatabaseReady() bool {
	var result int
	err := h.db.Raw("SELECT 1").Scan(&result).Error
	return err == nil
}
