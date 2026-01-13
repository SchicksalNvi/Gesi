package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-cesi/internal/supervisor"
	"go-cesi/internal/errors"
)

type SupervisorAPI struct {
	service *supervisor.SupervisorService
}

func NewSupervisorAPI(service *supervisor.SupervisorService) *SupervisorAPI {
	return &SupervisorAPI{service: service}
}

func (api *SupervisorAPI) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/nodes", api.GetNodes)
	router.GET("/nodes/:nodeName/processes", api.GetNodeProcesses)
	router.POST("/nodes/:nodeName/processes/:processName/start", api.StartProcess)
	router.POST("/nodes/:nodeName/processes/:processName/stop", api.StopProcess)
	router.GET("/nodes/:nodeName/processes/:processName/logs", api.GetProcessLogs)
}

func (api *SupervisorAPI) GetNodes(c *gin.Context) {
	nodes := api.service.GetAllNodes()

	response := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		response[i] = node.Serialize()
	}

	c.JSON(http.StatusOK, response)
}

func (api *SupervisorAPI) GetNodeProcesses(c *gin.Context) {
	nodeName := c.Param("nodeName")

	processes, err := api.service.GetNodeProcesses(nodeName)
	if err != nil {
		if errors.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]map[string]interface{}, len(processes))
	for i, process := range processes {
		response[i] = map[string]interface{}{
			"name":       process.Name,
			"group":      process.Group,
			"state":      process.State,
			"pid":        process.PID,
			"start_time": process.StartTime,
			"stop_time":  process.StopTime,
			"uptime":     process.Uptime.Seconds(),
		}
	}

	c.JSON(http.StatusOK, response)
}

func (api *SupervisorAPI) StartProcess(c *gin.Context) {
	nodeName := c.Param("nodeName")
	processName := c.Param("processName")

	err := api.service.StartProcess(nodeName, processName)
	if err != nil {
		if errors.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node or process not found"})
			return
		}
		if errors.IsConnectionError(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Node not connected"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api *SupervisorAPI) StopProcess(c *gin.Context) {
	nodeName := c.Param("nodeName")
	processName := c.Param("processName")

	err := api.service.StopProcess(nodeName, processName)
	if err != nil {
		if errors.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node or process not found"})
			return
		}
		if errors.IsConnectionError(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Node not connected"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api *SupervisorAPI) GetProcessLogs(c *gin.Context) {
	nodeName := c.Param("nodeName")
	processName := c.Param("processName")

	logs, err := api.service.GetProcessLogs(nodeName, processName)
	if err != nil {
		if errors.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node or process not found"})
			return
		}
		if errors.IsConnectionError(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Node not connected"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}
