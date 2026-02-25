package api

import (
	"net/http"

	"superview/internal/supervisor"

	"github.com/gin-gonic/gin"
)

type GroupsAPI struct {
	service *supervisor.SupervisorService
}

func NewGroupsAPI(service *supervisor.SupervisorService) *GroupsAPI {
	return &GroupsAPI{service: service}
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

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Group processes restarted successfully",
	})
}
