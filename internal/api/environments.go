package api

import (
	"net/http"

	"superview/internal/supervisor"

	"github.com/gin-gonic/gin"
)

type EnvironmentsAPI struct {
	service *supervisor.SupervisorService
}

func NewEnvironmentsAPI(service *supervisor.SupervisorService) *EnvironmentsAPI {
	return &EnvironmentsAPI{service: service}
}

// GetEnvironments 获取所有环境列表
func (e *EnvironmentsAPI) GetEnvironments(c *gin.Context) {
	environments := e.service.GetEnvironments()

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"environments": environments,
	})
}

// GetEnvironmentDetails 获取特定环境的详细信息
func (e *EnvironmentsAPI) GetEnvironmentDetails(c *gin.Context) {
	environmentName := c.Param("environment_name")

	environment := e.service.GetEnvironmentDetails(environmentName)
	if environment == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Environment not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"environment": environment,
	})
}
