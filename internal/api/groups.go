package api

import (
	"fmt"
	"net/http"

	"superview/internal/services"
	"superview/internal/supervisor"

	"github.com/gin-gonic/gin"
)

type GroupsAPI struct {
	service            *supervisor.SupervisorService
	activityLogService *services.ActivityLogService
}

func NewGroupsAPI(service *supervisor.SupervisorService, activityLogService ...*services.ActivityLogService) *GroupsAPI {
	api := &GroupsAPI{service: service}
	if len(activityLogService) > 0 {
		api.activityLogService = activityLogService[0]
	}
	return api
}

// GetGroups 获取所有进程分组
func (g *GroupsAPI) GetGroups(c *gin.Context) {
	groups := g.service.GetGroups()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"groups": groups,
	})
}

// GetGroupDetails 获取特定分组的详细信息
func (g *GroupsAPI) GetGroupDetails(c *gin.Context) {
	groupName := c.Param("group_name")

	group := g.service.GetGroupDetails(groupName)
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Group not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"group":  group,
	})
}

// StartGroupProcesses 启动分组中的所有进程
func (g *GroupsAPI) StartGroupProcesses(c *gin.Context) {
	groupName := c.Param("group_name")
	environmentName := c.Query("environment")

	err := g.service.StartGroupProcesses(groupName, environmentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if g.activityLogService != nil {
		msg := fmt.Sprintf("Started all processes in group %s", groupName)
		g.activityLogService.LogWithContext(c, "INFO", "start_group", "group", groupName, msg, nil)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Group processes started successfully",
	})
}

// StopGroupProcesses 停止分组中的所有进程
func (g *GroupsAPI) StopGroupProcesses(c *gin.Context) {
	groupName := c.Param("group_name")
	environmentName := c.Query("environment")

	err := g.service.StopGroupProcesses(groupName, environmentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if g.activityLogService != nil {
		msg := fmt.Sprintf("Stopped all processes in group %s", groupName)
		g.activityLogService.LogWithContext(c, "INFO", "stop_group", "group", groupName, msg, nil)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Group processes stopped successfully",
	})
}

// RestartGroupProcesses 重启分组中的所有进程
func (g *GroupsAPI) RestartGroupProcesses(c *gin.Context) {
	groupName := c.Param("group_name")
	environmentName := c.Query("environment")

	err := g.service.RestartGroupProcesses(groupName, environmentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if g.activityLogService != nil {
		msg := fmt.Sprintf("Restarted all processes in group %s", groupName)
		g.activityLogService.LogWithContext(c, "INFO", "restart_group", "group", groupName, msg, nil)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Group processes restarted successfully",
	})
}
