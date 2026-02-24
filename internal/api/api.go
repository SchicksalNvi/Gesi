package api

import (
	"go-cesi/internal/auth"
	"go-cesi/internal/middleware"
	"go-cesi/internal/repository"
	"go-cesi/internal/services"
	"go-cesi/internal/supervisor"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WebSocketHub interface for broadcasting
type WebSocketHub interface {
	Broadcast(message []byte)
	GetConnectionCount() int64
}

func SetupRoutes(r *gin.Engine, db *gorm.DB, service *supervisor.SupervisorService, hub WebSocketHub) {
	// 添加性能监控中间件
	r.Use(middleware.PerformanceMiddleware())

	authService := auth.NewAuthService(db)
	nodesAPI := NewNodesAPI(service)
	userAPI := NewUserAPI(db)
	environmentsAPI := NewEnvironmentsAPI(service)
	groupsAPI := NewGroupsAPI(service)
	activityLogService := services.NewActivityLogService(db)
	processesAPI := NewProcessesAPI(service, activityLogService)
	activityLogsAPI := NewActivityLogsAPI(activityLogService)
	healthAPI := NewHealthAPI(db, service)
	logManagementAPI := NewLogManagementAPI()

	roleHandler := NewRoleHandler(db)
	processEnhancedHandler := NewProcessEnhancedHandler(db)
	configurationHandler := NewConfigurationHandler(db)
	logAnalysisHandler := NewLogAnalysisHandler(db)

	// Discovery service and API
	// Requirements: 9.3, 9.4 - Authentication required for all discovery endpoints
	discoveryRepo := repository.NewDiscoveryRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	discoveryService := services.NewDiscoveryService(db, discoveryRepo, nodeRepo, hub)
	discoveryAPI := NewDiscoveryAPI(discoveryService, activityLogService)

	// Auth routes
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", authService.Login)
		authGroup.POST("/logout", authService.Logout)
		authGroup.GET("/user", authService.AuthMiddleware(), authService.GetCurrentUser)
	}

	// Protected API routes
	apiGroup := r.Group("/api", authService.AuthMiddleware())
	{
		// Health check endpoints
		healthGroup := apiGroup.Group("/health")
		{
			healthGroup.GET("", healthAPI.GetHealth)
			healthGroup.GET("/live", healthAPI.GetHealthLive)
			healthGroup.GET("/ready", healthAPI.GetHealthReady)
		}

		// Nodes routes
		nodesGroup := apiGroup.Group("/nodes")
		{
			nodesGroup.GET("", nodesAPI.GetNodes)
			nodesGroup.GET("/:node_name", nodesAPI.GetNode)
			nodesGroup.GET("/:node_name/processes", nodesAPI.GetNodeProcesses)
			nodesGroup.POST("/:node_name/processes/:process_name/start", nodesAPI.StartProcess)
			nodesGroup.POST("/:node_name/processes/:process_name/stop", nodesAPI.StopProcess)
			nodesGroup.POST("/:node_name/processes/:process_name/restart", nodesAPI.RestartProcess)
			nodesGroup.GET("/:node_name/processes/:process_name/logs", nodesAPI.GetProcessLogs)
			nodesGroup.GET("/:node_name/processes/:process_name/logs/stream", nodesAPI.GetProcessLogStream)
			// Batch operations
			nodesGroup.POST("/:node_name/processes/start-all", nodesAPI.StartAllProcesses)
			nodesGroup.POST("/:node_name/processes/stop-all", nodesAPI.StopAllProcesses)
			nodesGroup.POST("/:node_name/processes/restart-all", nodesAPI.RestartAllProcesses)
		}

		// User management API
		userHandler := NewUserHandler(db)
		userGroup := apiGroup.Group("/users")
		{
			userGroup.GET("", userHandler.GetUsers)
			userGroup.POST("", userHandler.CreateUser)
			userGroup.GET("/:id", userHandler.GetUserByID)
			userGroup.PUT("/:id", userHandler.UpdateUser)
			userGroup.DELETE("/:id", userHandler.DeleteUser)
			userGroup.PATCH("/:id/toggle", userHandler.ToggleUserStatus)
		}

		// Profile management API
		profileGroup := apiGroup.Group("/profile")
		{
			profileGroup.GET("", userAPI.GetProfile)
			profileGroup.PUT("", userAPI.UpdateProfile)
		}

		// Environments API
		environmentsGroup := apiGroup.Group("/environments")
		{
			environmentsGroup.GET("", environmentsAPI.GetEnvironments)
			environmentsGroup.GET("/:environment_name", environmentsAPI.GetEnvironmentDetails)
		}

		// Groups API
		groupsGroup := apiGroup.Group("/groups")
		{
			groupsGroup.GET("", groupsAPI.GetGroups)
			groupsGroup.GET("/:group_name", groupsAPI.GetGroupDetails)
			groupsGroup.POST("/:group_name/start", groupsAPI.StartGroupProcesses)
			groupsGroup.POST("/:group_name/stop", groupsAPI.StopGroupProcesses)
			groupsGroup.POST("/:group_name/restart", groupsAPI.RestartGroupProcesses)
		}

		// Processes Aggregation API
		processesGroup := apiGroup.Group("/processes")
		{
			processesGroup.GET("/aggregated", processesAPI.GetAggregatedProcesses)
			processesGroup.POST("/:process_name/start", processesAPI.BatchStartProcess)
			processesGroup.POST("/:process_name/stop", processesAPI.BatchStopProcess)
			processesGroup.POST("/:process_name/restart", processesAPI.BatchRestartProcess)
		}

		// Activity Logs API
		activityLogsGroup := apiGroup.Group("/activity-logs")
		{
			activityLogsGroup.GET("", activityLogsAPI.GetActivityLogs)
			activityLogsGroup.GET("/recent", activityLogsAPI.GetRecentLogs)
			activityLogsGroup.GET("/statistics", activityLogsAPI.GetLogStatistics)
			activityLogsGroup.GET("/export", activityLogsAPI.ExportLogs)
			activityLogsGroup.DELETE("/clean", activityLogsAPI.CleanOldLogs)
			activityLogsGroup.DELETE("", activityLogsAPI.DeleteLogs)
		}

		// Roles and Permissions API
		rolesGroup := apiGroup.Group("/roles")
		{
			rolesGroup.GET("", roleHandler.GetRoles)
			rolesGroup.POST("", roleHandler.CreateRole)
			rolesGroup.GET("/:id", roleHandler.GetRole)
			rolesGroup.PUT("/:id", roleHandler.UpdateRole)
			rolesGroup.DELETE("/:id", roleHandler.DeleteRole)
			rolesGroup.POST("/:id/permissions", roleHandler.AssignPermissions)
		}

		// Role-User assignment API (separate group to avoid conflicts)
		roleUsersGroup := apiGroup.Group("/role-users")
		{
			roleUsersGroup.POST("/:roleId/users/:userId", roleHandler.AssignRoleToUser)
			roleUsersGroup.DELETE("/:roleId/users/:userId", roleHandler.RemoveRoleFromUser)
		}

		// Permissions API
		permissionsGroup := apiGroup.Group("/permissions")
		{
			permissionsGroup.GET("", roleHandler.GetPermissions)
		}

		// Alerts API
		alertHandler := NewAlertHandler(db, hub)
		alertsGroup := apiGroup.Group("/alerts")
		{
			// Alert rules management
			alertsGroup.POST("/rules", alertHandler.CreateAlertRule)
			alertsGroup.GET("/rules", alertHandler.GetAlertRules)
			alertsGroup.GET("/rules/:id", alertHandler.GetAlertRule)
			alertsGroup.PUT("/rules/:id", alertHandler.UpdateAlertRule)
			alertsGroup.DELETE("/rules/:id", alertHandler.DeleteAlertRule)

			// Alert records management
			alertsGroup.GET("", alertHandler.GetAlerts)
			alertsGroup.GET("/:id", alertHandler.GetAlert)
			alertsGroup.POST("/:id/acknowledge", alertHandler.AcknowledgeAlert)
			alertsGroup.POST("/:id/resolve", alertHandler.ResolveAlert)

			// Notification channels management
			alertsGroup.POST("/channels", alertHandler.CreateNotificationChannel)
			alertsGroup.GET("/channels", alertHandler.GetNotificationChannels)
			alertsGroup.GET("/channels/:id", alertHandler.GetNotificationChannel)
			alertsGroup.PUT("/channels/:id", alertHandler.UpdateNotificationChannel)
			alertsGroup.DELETE("/channels/:id", alertHandler.DeleteNotificationChannel)
			alertsGroup.POST("/channels/:id/test", alertHandler.TestNotificationChannel)

			// System metrics and statistics
			alertsGroup.POST("/metrics", alertHandler.RecordSystemMetric)
			alertsGroup.GET("/metrics", alertHandler.GetSystemMetrics)
			alertsGroup.GET("/statistics", alertHandler.GetAlertStatistics)
		}

		// Process Enhanced API
		processEnhancedGroup := apiGroup.Group("/process-enhanced")
		{
			// Task scheduler management
			processEnhancedGroup.POST("/scheduler/start", processEnhancedHandler.StartScheduler)
			processEnhancedGroup.POST("/scheduler/stop", processEnhancedHandler.StopScheduler)

			// Process group management
			processEnhancedGroup.POST("/groups", processEnhancedHandler.CreateProcessGroup)
			processEnhancedGroup.GET("/groups", processEnhancedHandler.GetProcessGroups)
			processEnhancedGroup.GET("/groups/:id", processEnhancedHandler.GetProcessGroup)
			processEnhancedGroup.PUT("/groups/:id", processEnhancedHandler.UpdateProcessGroup)
			processEnhancedGroup.DELETE("/groups/:id", processEnhancedHandler.DeleteProcessGroup)
			processEnhancedGroup.POST("/groups/:id/processes", processEnhancedHandler.AddProcessToGroup)
			processEnhancedGroup.DELETE("/groups/:id/processes", processEnhancedHandler.RemoveProcessFromGroup)
			processEnhancedGroup.PUT("/groups/:id/reorder", processEnhancedHandler.ReorderProcessesInGroup)

			// Process dependency management
			processEnhancedGroup.POST("/dependencies", processEnhancedHandler.CreateProcessDependency)
			processEnhancedGroup.GET("/dependencies", processEnhancedHandler.GetProcessDependencies)
			processEnhancedGroup.GET("/dependent-processes", processEnhancedHandler.GetDependentProcesses)
			processEnhancedGroup.DELETE("/dependencies/:id", processEnhancedHandler.DeleteProcessDependency)
			processEnhancedGroup.POST("/startup-order", processEnhancedHandler.GetStartupOrder)

			// Scheduled task management
			processEnhancedGroup.POST("/scheduled-tasks", processEnhancedHandler.CreateScheduledTask)
			processEnhancedGroup.GET("/scheduled-tasks", processEnhancedHandler.GetScheduledTasks)
			processEnhancedGroup.GET("/scheduled-tasks/:id", processEnhancedHandler.GetScheduledTask)
			processEnhancedGroup.PUT("/scheduled-tasks/:id", processEnhancedHandler.UpdateScheduledTask)
			processEnhancedGroup.DELETE("/scheduled-tasks/:id", processEnhancedHandler.DeleteScheduledTask)
			processEnhancedGroup.GET("/scheduled-tasks/:id/executions", processEnhancedHandler.GetTaskExecutions)

			// Process template management
			processEnhancedGroup.POST("/templates", processEnhancedHandler.CreateProcessTemplate)
			processEnhancedGroup.GET("/templates", processEnhancedHandler.GetProcessTemplates)
			processEnhancedGroup.GET("/templates/:id", processEnhancedHandler.GetProcessTemplate)
			processEnhancedGroup.PUT("/templates/:id", processEnhancedHandler.UpdateProcessTemplate)
			processEnhancedGroup.DELETE("/templates/:id", processEnhancedHandler.DeleteProcessTemplate)
			processEnhancedGroup.POST("/templates/:id/use", processEnhancedHandler.UseTemplate)

			// Process configuration backup management
			processEnhancedGroup.POST("/backups", processEnhancedHandler.CreateProcessBackup)
			processEnhancedGroup.GET("/backups", processEnhancedHandler.GetProcessBackups)
			processEnhancedGroup.POST("/backups/:id/restore", processEnhancedHandler.RestoreProcessBackup)

			// Process performance metrics
			processEnhancedGroup.POST("/metrics", processEnhancedHandler.RecordProcessMetrics)
			processEnhancedGroup.GET("/metrics", processEnhancedHandler.GetProcessMetrics)
			processEnhancedGroup.GET("/metrics/statistics", processEnhancedHandler.GetProcessMetricsStatistics)

			// Data cleanup
			processEnhancedGroup.POST("/cleanup", processEnhancedHandler.CleanupOldData)
		}

		// Configuration API
		configurationGroup := apiGroup.Group("/configuration")
		{
			// 配置项管理
			configurationGroup.GET("", configurationHandler.GetConfigurations)
			configurationGroup.POST("", configurationHandler.CreateConfiguration)
			configurationGroup.GET("/:id", configurationHandler.GetConfiguration)
			configurationGroup.PUT("/:id", configurationHandler.UpdateConfiguration)
			configurationGroup.DELETE("/:id", configurationHandler.DeleteConfiguration)

			// 环境变量管理
			configurationGroup.GET("/env-vars", configurationHandler.GetEnvironmentVariables)
			configurationGroup.POST("/env-vars", configurationHandler.CreateEnvironmentVariable)
			configurationGroup.GET("/env-vars/:id", configurationHandler.GetEnvironmentVariable)
			configurationGroup.PUT("/env-vars/:id", configurationHandler.UpdateEnvironmentVariable)
			configurationGroup.DELETE("/env-vars/:id", configurationHandler.DeleteEnvironmentVariable)

			// 配置备份管理
			configurationGroup.GET("/backups", configurationHandler.GetBackups)
			configurationGroup.POST("/backups", configurationHandler.CreateBackup)
			configurationGroup.GET("/backups/:id", configurationHandler.GetBackup)
			configurationGroup.POST("/backups/:id/restore", configurationHandler.RestoreBackup)
			configurationGroup.DELETE("/backups/:id", configurationHandler.DeleteBackup)

			// 配置导入导出
			configurationGroup.GET("/export", configurationHandler.ExportConfigurations)
			configurationGroup.POST("/import", configurationHandler.ImportConfigurations)

			// 配置变更历史
			configurationGroup.GET("/history", configurationHandler.GetConfigurationHistory)

			// 审计日志
			configurationGroup.GET("/audit-logs", configurationHandler.GetAuditLogs)

			// 数据清理
			configurationGroup.POST("/cleanup", configurationHandler.CleanupOldData)
		}

		// Log Analysis API
		logAnalysisGroup := apiGroup.Group("/logs")
		{
			// 日志条目管理
			logAnalysisGroup.GET("", logAnalysisHandler.GetLogEntries)
			logAnalysisGroup.POST("", logAnalysisHandler.CreateLogEntry)
			logAnalysisGroup.GET("/:id", logAnalysisHandler.GetLogEntry)
			logAnalysisGroup.DELETE("/:id", logAnalysisHandler.DeleteLogEntry)

			// 分析规则管理
			logAnalysisGroup.GET("/rules", logAnalysisHandler.GetAnalysisRules)
			logAnalysisGroup.POST("/rules", logAnalysisHandler.CreateAnalysisRule)
			logAnalysisGroup.GET("/rules/:id", logAnalysisHandler.GetAnalysisRule)
			logAnalysisGroup.PUT("/rules/:id", logAnalysisHandler.UpdateAnalysisRule)
			logAnalysisGroup.DELETE("/rules/:id", logAnalysisHandler.DeleteAnalysisRule)

			// 日志统计
			logAnalysisGroup.GET("/statistics", logAnalysisHandler.GetLogStatistics)

			// 日志告警
			logAnalysisGroup.GET("/alerts", logAnalysisHandler.GetLogAlerts)
			logAnalysisGroup.POST("/alerts/:id/acknowledge", logAnalysisHandler.AcknowledgeAlert)
			logAnalysisGroup.POST("/alerts/:id/resolve", logAnalysisHandler.ResolveAlert)

			// 日志过滤器
			logAnalysisGroup.GET("/filters", logAnalysisHandler.GetLogFilters)
			logAnalysisGroup.POST("/filters", logAnalysisHandler.CreateLogFilter)
			logAnalysisGroup.PUT("/filters/:id", logAnalysisHandler.UpdateLogFilter)
			logAnalysisGroup.DELETE("/filters/:id", logAnalysisHandler.DeleteLogFilter)

			// 日志导出
			logAnalysisGroup.GET("/exports", logAnalysisHandler.GetLogExports)
			logAnalysisGroup.POST("/exports", logAnalysisHandler.CreateLogExport)
			logAnalysisGroup.GET("/exports/:id", logAnalysisHandler.GetLogExport)
			logAnalysisGroup.DELETE("/exports/:id", logAnalysisHandler.DeleteLogExport)

			// 保留策略
			logAnalysisGroup.GET("/retention-policies", logAnalysisHandler.GetRetentionPolicies)
			logAnalysisGroup.POST("/retention-policies", logAnalysisHandler.CreateRetentionPolicy)
			logAnalysisGroup.PUT("/retention-policies/:id", logAnalysisHandler.UpdateRetentionPolicy)
			logAnalysisGroup.DELETE("/retention-policies/:id", logAnalysisHandler.DeleteRetentionPolicy)
			logAnalysisGroup.POST("/retention-policies/execute", logAnalysisHandler.ExecuteRetentionPolicies)

			// 数据清理
			logAnalysisGroup.POST("/cleanup", logAnalysisHandler.CleanupOldLogs)
		}

		// Data Management API
		dataManagementHandler := NewDataManagementAPI()
		dataManagementGroup := apiGroup.Group("/data-management")
		{
			// 数据导出
			dataManagementGroup.POST("/export", dataManagementHandler.ExportData)
			dataManagementGroup.GET("/exports", dataManagementHandler.GetExportRecords)
			dataManagementGroup.GET("/exports/:id/download", dataManagementHandler.DownloadExportFile)
			dataManagementGroup.DELETE("/exports/:id", dataManagementHandler.DeleteExportRecord)

			// 数据备份
			dataManagementGroup.POST("/backup", dataManagementHandler.CreateBackup)
			dataManagementGroup.GET("/backups", dataManagementHandler.GetBackupRecords)
			dataManagementGroup.GET("/backups/:id/download", dataManagementHandler.DownloadBackupFile)
			dataManagementGroup.DELETE("/backups/:id", dataManagementHandler.DeleteBackupRecord)

			// 数据导入
			dataManagementGroup.POST("/import", dataManagementHandler.ImportData)
		}

		// System Settings API
		systemSettingsHandler := NewSystemSettingsAPI(db)
		systemSettingsGroup := apiGroup.Group("/system-settings")
		{
			// 系统设置管理
			systemSettingsGroup.GET("", systemSettingsHandler.GetSystemSettings)
			systemSettingsGroup.GET("/:key", systemSettingsHandler.GetSystemSetting)
			systemSettingsGroup.PUT("/:key", systemSettingsHandler.UpdateSystemSetting)
			systemSettingsGroup.PUT("/batch", systemSettingsHandler.UpdateMultipleSettings)
			systemSettingsGroup.DELETE("/:key", systemSettingsHandler.DeleteSystemSetting)
			systemSettingsGroup.POST("/reset", systemSettingsHandler.ResetToDefaults)

			// 用户偏好设置（当前用户）
			systemSettingsGroup.GET("/user-preferences", systemSettingsHandler.GetUserPreferences)
			systemSettingsGroup.PUT("/user-preferences", systemSettingsHandler.UpdateUserPreferences)

			// 管理员管理其他用户偏好
			systemSettingsGroup.GET("/users/:userId/preferences", systemSettingsHandler.GetUserPreferencesByAdmin)
			systemSettingsGroup.PUT("/users/:userId/preferences", systemSettingsHandler.UpdateUserPreferencesByAdmin)

			// 邮件配置测试
			systemSettingsGroup.POST("/test-email", systemSettingsHandler.TestEmailConfiguration)
		}

		// Developer Tools API
		developerToolsHandler := NewDeveloperToolsAPI(db, service, nil, hub)
		developerGroup := apiGroup.Group("/developer")
		{
			// API 文档
			developerGroup.GET("/api-docs", developerToolsHandler.GetApiEndpoints)
			developerGroup.POST("/test-api", developerToolsHandler.TestApiEndpoint)

			// 调试工具
			developerGroup.GET("/debug-logs", developerToolsHandler.GetDebugLogs)
			developerGroup.DELETE("/debug-logs", developerToolsHandler.ClearDebugLogs)
			developerGroup.PUT("/log-level", developerToolsHandler.SetLogLevel)

			// 性能监控
			developerGroup.GET("/performance", developerToolsHandler.GetPerformanceMetrics)
			developerGroup.POST("/performance/reset", developerToolsHandler.ResetPerformanceMetrics)
			developerGroup.GET("/performance/slow-endpoints", developerToolsHandler.GetTopSlowEndpoints)
			developerGroup.GET("/performance/error-rates", developerToolsHandler.GetErrorRateByEndpoint)
			developerGroup.GET("/system-metrics", developerToolsHandler.GetSystemMetrics)
			developerGroup.GET("/api-metrics", developerToolsHandler.GetApiMetrics)
			developerGroup.GET("/database-metrics", developerToolsHandler.GetDatabaseMetrics)
			developerGroup.GET("/websocket-metrics", developerToolsHandler.GetWebSocketMetrics)

			// 日志级别管理
			logGroup := developerGroup.Group("/logs")
			{
				logGroup.GET("/level", logManagementAPI.GetLogLevel)
				logGroup.PUT("/level", logManagementAPI.SetLogLevel)
				logGroup.POST("/level/temporary", logManagementAPI.SetTemporaryLogLevel)
				logGroup.POST("/level/reset", logManagementAPI.ResetLogLevel)
				logGroup.GET("/levels", logManagementAPI.GetAvailableLogLevels)
				logGroup.DELETE("/level/history", logManagementAPI.ClearLogLevelHistory)
			}
		}

		// Discovery API - Node discovery and network scanning
		// Requirements: 9.3, 9.4 - All discovery endpoints require authentication
		discoveryGroup := apiGroup.Group("/discovery")
		{
			discoveryGroup.POST("/tasks", discoveryAPI.StartDiscovery)
			discoveryGroup.GET("/tasks", discoveryAPI.ListTasks)
			discoveryGroup.GET("/tasks/:id", discoveryAPI.GetTask)
			discoveryGroup.POST("/tasks/:id/cancel", discoveryAPI.CancelTask)
			discoveryGroup.DELETE("/tasks/:id", discoveryAPI.DeleteTask)
			discoveryGroup.GET("/tasks/:id/progress", discoveryAPI.GetTaskProgress)
			discoveryGroup.POST("/validate-cidr", discoveryAPI.ValidateCIDR)
		}
	}
}
