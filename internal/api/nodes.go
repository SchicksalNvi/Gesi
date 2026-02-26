package api

import (
	"fmt"
	"net/http"
	"strconv"

	"superview/internal/services"
	"superview/internal/supervisor"
	"superview/internal/validation"

	"github.com/gin-gonic/gin"
)

type NodesAPI struct {
	service            *supervisor.SupervisorService
	activityLogService *services.ActivityLogService
}

func NewNodesAPI(service *supervisor.SupervisorService, activityLogService ...*services.ActivityLogService) *NodesAPI {
	api := &NodesAPI{service: service}
	if len(activityLogService) > 0 {
		api.activityLogService = activityLogService[0]
	}
	return api
}

func (api *NodesAPI) GetNodes(c *gin.Context) {
	nodes := api.service.GetAllNodes()
	response := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		response[i] = node.Serialize()
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"nodes":  response,
	})
}

func (api *NodesAPI) GetNode(c *gin.Context) {
	nodeName := c.Param("node_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateNoSQLInjection("node_name", nodeName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}

	node, err := api.service.GetNode(nodeName)
	if err != nil {
		handleNotFound(c, "node", nodeName)
		return
	}
	c.JSON(http.StatusOK, node.Serialize())
}

func (api *NodesAPI) GetNodeProcesses(c *gin.Context) {
	nodeName := c.Param("node_name")
	
	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateNoSQLInjection("node_name", nodeName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}

	
	node, err := api.service.GetNode(nodeName)
	if err != nil {
		handleAppError(c, err)
		return
	}
	
	if err := node.RefreshProcesses(); err != nil {
		handleAppError(c, err)
		return
	}
	
	// 使用SerializeProcesses方法返回格式化的数据
	processes := node.SerializeProcesses()
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"processes": processes,
	})
}

func (api *NodesAPI) StartProcess(c *gin.Context) {
	nodeName := c.Param("node_name")
	processName := c.Param("process_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateProcessName("process_name", processName)
	validator.ValidateNoSQLInjection("node_name", nodeName)
	validator.ValidateNoSQLInjection("process_name", processName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}


	if err := api.service.StartProcess(nodeName, processName); err != nil {
		handleAppError(c, err)
		return
	}

	if api.activityLogService != nil {
		msg := fmt.Sprintf("Started process %s on node %s", processName, nodeName)
		api.activityLogService.LogWithContext(c, "INFO", "start_process", "process", processName, msg, nil)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api *NodesAPI) StopProcess(c *gin.Context) {
	nodeName := c.Param("node_name")
	processName := c.Param("process_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateProcessName("process_name", processName)
	validator.ValidateNoSQLInjection("node_name", nodeName)
	validator.ValidateNoSQLInjection("process_name", processName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}


	if err := api.service.StopProcess(nodeName, processName); err != nil {
		handleAppError(c, err)
		return
	}

	if api.activityLogService != nil {
		msg := fmt.Sprintf("Stopped process %s on node %s", processName, nodeName)
		api.activityLogService.LogWithContext(c, "INFO", "stop_process", "process", processName, msg, nil)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api *NodesAPI) RestartProcess(c *gin.Context) {
	nodeName := c.Param("node_name")
	processName := c.Param("process_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateProcessName("process_name", processName)
	validator.ValidateNoSQLInjection("node_name", nodeName)
	validator.ValidateNoSQLInjection("process_name", processName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}


	if err := api.service.StopProcess(nodeName, processName); err != nil {
		handleInternalError(c, err)
		return
	}
	if err := api.service.StartProcess(nodeName, processName); err != nil {
		handleInternalError(c, err)
		return
	}

	if api.activityLogService != nil {
		msg := fmt.Sprintf("Restarted process %s on node %s", processName, nodeName)
		api.activityLogService.LogWithContext(c, "INFO", "restart_process", "process", processName, msg, nil)
	}

	handleSuccess(c, "Process restarted successfully", nil)
}

func (api *NodesAPI) GetProcessLogs(c *gin.Context) {
	nodeName := c.Param("node_name")
	processName := c.Param("process_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateProcessName("process_name", processName)
	validator.ValidateNoSQLInjection("node_name", nodeName)
	validator.ValidateNoSQLInjection("process_name", processName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}


	logs, err := api.service.GetProcessLogs(nodeName, processName)
	if err != nil {
		handleAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, logs)
}

// GetProcessLogStream 获取结构化的日志流
func (api *NodesAPI) GetProcessLogStream(c *gin.Context) {
	nodeName := c.Param("node_name")
	processName := c.Param("process_name")
	
	// 获取查询参数
	offset := -1 // -1 表示从文件末尾读取
	maxLines := 100
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}
	
	if maxLinesStr := c.Query("max_lines"); maxLinesStr != "" {
		if m, err := strconv.Atoi(maxLinesStr); err == nil && m > 0 && m <= 1000 {
			maxLines = m
		}
	}
	
	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateProcessName("process_name", processName)
	validator.ValidateNoSQLInjection("node_name", nodeName)
	validator.ValidateNoSQLInjection("process_name", processName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}

	
	node, err := api.service.GetNode(nodeName)
	if err != nil {
		handleAppError(c, err)
		return
	}
	
	// 如果 offset < 0，从文件末尾读取最新日志
	if offset < 0 {
		logStream, err := node.GetProcessLogStreamTail(processName, maxLines)
		if err != nil {
			handleAppError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   logStream,
		})
		return
	}
	
	// 从指定偏移量读取
	logStream, err := node.GetProcessLogStream(processName, offset, maxLines)
	if err != nil {
		handleAppError(c, err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   logStream,
	})
}

// StartAllProcesses starts all processes on a specific node
func (api *NodesAPI) StartAllProcesses(c *gin.Context) {
	nodeName := c.Param("node_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateNoSQLInjection("node_name", nodeName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}


	if err := api.service.StartAllProcesses(nodeName); err != nil {
		handleAppError(c, err)
		return
	}

	if api.activityLogService != nil {
		msg := fmt.Sprintf("Started all processes on node %s", nodeName)
		api.activityLogService.LogWithContext(c, "INFO", "start_process", "node", nodeName, msg, nil)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "All processes started"})
}

// StopAllProcesses stops all processes on a specific node
func (api *NodesAPI) StopAllProcesses(c *gin.Context) {
	nodeName := c.Param("node_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateNoSQLInjection("node_name", nodeName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}


	if err := api.service.StopAllProcesses(nodeName); err != nil {
		handleAppError(c, err)
		return
	}

	if api.activityLogService != nil {
		msg := fmt.Sprintf("Stopped all processes on node %s", nodeName)
		api.activityLogService.LogWithContext(c, "INFO", "stop_process", "node", nodeName, msg, nil)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "All processes stopped"})
}

// RestartAllProcesses restarts all processes on a specific node
func (api *NodesAPI) RestartAllProcesses(c *gin.Context) {
	nodeName := c.Param("node_name")

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateNodeName("node_name", nodeName)
	validator.ValidateNoSQLInjection("node_name", nodeName)

	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}


	if err := api.service.RestartAllProcesses(nodeName); err != nil {
		handleInternalError(c, err)
		return
	}

	if api.activityLogService != nil {
		msg := fmt.Sprintf("Restarted all processes on node %s", nodeName)
		api.activityLogService.LogWithContext(c, "INFO", "restart_process", "node", nodeName, msg, nil)
	}

	handleSuccess(c, "All processes restarted", nil)
}
