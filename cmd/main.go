package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"go-cesi/internal/api"
	"go-cesi/internal/auth"
	"go-cesi/internal/config"
	"go-cesi/internal/database"
	"go-cesi/internal/logger"
	"go-cesi/internal/loggers"
	"go-cesi/internal/middleware"
	"go-cesi/internal/models"
	"go-cesi/internal/services"
	"go-cesi/internal/supervisor"
	"go-cesi/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	}
	Admin struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Email    string `mapstructure:"email"`
	} `mapstructure:"admin"`
	Nodes []struct {
		Name        string `mapstructure:"name"`
		Environment string `mapstructure:"environment"`
		Host        string `mapstructure:"host"`
		Port        int    `mapstructure:"port"`
		Username    string `mapstructure:"username"`
		Password    string `mapstructure:"password"`
	} `mapstructure:"nodes"`
}

// 创建管理员用户
func createAdminUser(db *gorm.DB, config *Config) error {
	// 检查是否已存在管理员用户
	var count int64
	if err := db.Model(&models.User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check admin user: %v", err)
	}

	// 如果没有管理员用户，创建一个默认的管理员用户
	if count == 0 {
		adminUser := models.User{
			Username: config.Admin.Username,
			Email:    config.Admin.Email,
			IsAdmin:  true,
		}
		logger.Info("Creating admin user", zap.String("username", config.Admin.Username))
		if err := adminUser.SetPassword(config.Admin.Password); err != nil {
			return fmt.Errorf("failed to set admin password: %v", err)
		}
		logger.Info("Initial admin account created",
			zap.String("username", config.Admin.Username),
			zap.String("email", config.Admin.Email))

		if err := db.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}
		logger.Info("Default admin user created successfully")
	}

	return nil
}

func createAdminCommand(username, password, email string) {
	if err := database.InitDB(); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	db := database.DB

	adminUser := models.User{
		Username: username,
		Email:    email,
		IsAdmin:  true,
	}
	if err := adminUser.SetPassword(password); err != nil {
		logger.Fatal("Failed to set password", zap.Error(err))
	}

	if err := db.Create(&adminUser).Error; err != nil {
		logger.Fatal("Failed to create admin user", zap.Error(err))
	}
	logger.Info("Admin user created successfully",
		zap.String("username", username),
		zap.String("email", email))
}

// validateEnvironmentVariables 验证必要的环境变量
func validateEnvironmentVariables() error {
	requiredEnvVars := []string{
		"JWT_SECRET",
		"ADMIN_PASSWORD",
		"NODE_PASSWORD",
	}

	var missingVars []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	// 验证JWT_SECRET长度
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}

	return nil
}

func main() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		// .env文件不存在或加载失败时，继续执行（可能使用系统环境变量）
		fmt.Printf("Warning: .env file not found or failed to load: %v\n", err)
	}

	// 初始化动态日志系统
	if err := logger.InitDynamicLogger(); err != nil {
		// 由于logger初始化失败，使用标准错误输出
		fmt.Fprintf(os.Stderr, "Failed to initialize dynamic logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// 记录日志系统启动信息
	logger.Info("Dynamic logging system initialized successfully")

	// 验证环境变量
	if err := validateEnvironmentVariables(); err != nil {
		logger.Fatal("Environment validation failed", zap.Error(err))
	}

	// 加载应用配置
	appConfig, err := config.Load("config.toml")
	if err != nil {
		logger.Fatal("Failed to load application config", zap.Error(err))
	}
	logger.Info("Application config loaded successfully")

	if len(os.Args) > 1 && os.Args[1] == "create-admin" {
		if len(os.Args) < 6 {
			logger.Fatal("Usage: go run cmd/main.go create-admin --username <username> --password <password> --email <email>")
		}
		// 解析命令行参数
		var username, password, email string
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--username":
				username = os.Args[i+1]
				i++
			case "--password":
				password = os.Args[i+1]
				i++
			case "--email":
				email = os.Args[i+1]
				i++
			}
		}
		if username == "" || password == "" || email == "" {
			logger.Fatal("Missing required arguments")
		}
		createAdminCommand(username, password, email)
		return
	}

	// 加载节点配置
	nodeConfig, err := loadConfig()
	if err != nil {
		logger.Fatal("Failed to load node config", zap.Error(err))
	}
	logger.Info("Node config loaded",
		zap.Int("server_port", nodeConfig.Server.Port),
		zap.String("admin_username", nodeConfig.Admin.Username),
		zap.Int("nodes_count", len(nodeConfig.Nodes)))

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	db := database.DB

	// 创建管理员用户
	if err := createAdminUser(db, nodeConfig); err != nil {
		logger.Fatal("Failed to create admin user", zap.Error(err))
	}

	// 启动性能监控
	if appConfig.Performance.MemoryMonitoringEnabled {
		middleware.StartMemoryMonitoring(appConfig.Performance.MemoryUpdateInterval)
		logger.Info("Memory monitoring started",
			zap.Duration("update_interval", appConfig.Performance.MemoryUpdateInterval))
	}

	if appConfig.Performance.MetricsCleanupEnabled {
		middleware.StartPerformanceCleanup(
			appConfig.Performance.MetricsResetInterval,
			appConfig.Performance.EndpointCleanupThreshold,
		)
		logger.Info("Performance metrics cleanup started",
			zap.Duration("reset_interval", appConfig.Performance.MetricsResetInterval),
			zap.Duration("cleanup_threshold", appConfig.Performance.EndpointCleanupThreshold))
	}

	// 初始化活动日志服务
	loggers.InitActivityLogService(db)

	// 初始化Supervisor服务
	supervisorService := supervisor.NewSupervisorService()

	// 初始化WebSocket Hub
	hub := websocket.NewHub(supervisorService)
	go hub.Run()

	// 初始化Alert服务和监控
	alertService := services.NewAlertService(db)
	alertMonitor := services.NewAlertMonitor(alertService, supervisorService, hub)
	alertMonitor.Start()
	logger.Info("Alert Monitor started")

	// 添加配置的节点
	logger.Info("Loading nodes from config", zap.Int("nodes_count", len(nodeConfig.Nodes)))
	for _, node := range nodeConfig.Nodes {
		logger.Info("Adding node",
			zap.String("name", node.Name),
			zap.String("environment", node.Environment),
			zap.String("host", node.Host),
			zap.Int("port", node.Port),
			zap.String("username", node.Username))
		err := supervisorService.AddNode(
			node.Name,
			node.Environment,
			node.Host,
			node.Port,
			node.Username,
			node.Password,
		)
		if err != nil {
			logger.Error("Failed to add node",
				zap.String("node_name", node.Name),
				zap.Error(err))
		} else {
			logger.Info("Successfully added node", zap.String("node_name", node.Name))
		}
	}

	// 启动自动刷新
	stopRefresh := supervisorService.StartAutoRefresh(30 * time.Second)

	// 设置Gin路由
	router := gin.Default()

	// 配置CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// 获取当前工作目录作为项目根目录
	projectRoot, err := os.Getwd()
	if err != nil {
		logger.Fatal("Failed to get working directory", zap.Error(err))
	}

	// 设置API路由
	api.SetupRoutes(router, db, supervisorService, hub)

	// extractToken extracts JWT token from query parameter or Authorization header
	extractToken := func(c *gin.Context) string {
		// Query parameter takes precedence
		if token := c.Query("token"); token != "" {
			return token
		}
		// Fallback to Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			return ""
		}
		// Strip "Bearer " prefix if present
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 设置WebSocket路由
	router.GET("/ws", func(c *gin.Context) {
		// Extract token from query parameter or header
		token := extractToken(c)
		if token == "" {
			logger.Warn("WebSocket authentication failed: missing token",
				zap.String("remote_addr", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed: missing token"})
			return
		}

		// Validate token
		claims, err := auth.ParseToken(token)
		if err != nil {
			logger.Warn("WebSocket authentication failed: invalid token",
				zap.String("remote_addr", c.ClientIP()),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed: invalid token"})
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		logger.Debug("WebSocket authentication successful",
			zap.String("user_id", claims.UserID),
			zap.String("remote_addr", c.ClientIP()))

		// Upgrade to WebSocket
		hub.HandleWebSocket(c)
	})

	// 服务 React 构建产物
	// 静态资源（JS/CSS/图片等）
	staticPath := filepath.Join(projectRoot, "web", "react-app", "dist", "assets")
	router.Static("/assets", staticPath)

	// 前端应用路由 - 服务于所有前端路由
	router.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api") || strings.HasPrefix(c.Request.URL.Path, "/ws") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		// 否则服务 React 应用的 index.html
		// 设置缓存控制头：不缓存 HTML，让浏览器每次都获取最新版本
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		indexPath := filepath.Join(projectRoot, "web", "react-app", "dist", "index.html")
		c.File(indexPath)
	})

	// 设置优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	// 启动配置热重载协程
	go func() {
		for range sigChan {
			logger.Info("Received SIGHUP, reloading configuration")
			newConfig, err := loadConfig()
			if err != nil {
				logger.Error("Failed to reload config", zap.Error(err))
				continue
			}

			// 更新节点配置
			for _, node := range newConfig.Nodes {
				if _, err := supervisorService.GetNode(node.Name); err != nil {
					// 新增节点
					if err := supervisorService.AddNode(
						node.Name,
						node.Environment,
						node.Host,
						node.Port,
						node.Username,
						node.Password,
					); err != nil {
						logger.Error("Failed to add new node",
							zap.String("node_name", node.Name),
							zap.Error(err))
					}
				}
			}

			// 更新admin配置
			if newConfig.Admin.Username != nodeConfig.Admin.Username ||
				newConfig.Admin.Password != nodeConfig.Admin.Password {
				if err := updateAdminUser(db, newConfig); err != nil {
					logger.Error("Failed to update admin user", zap.Error(err))
				}
			}

			nodeConfig = newConfig
			logger.Info("Configuration reloaded successfully")
		}
	}()

	// 启动HTTP服务器
	serverAddr := fmt.Sprintf("0.0.0.0:%d", nodeConfig.Server.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 在goroutine中启动服务器
	go func() {
		logger.Info("Server starting", zap.String("address", serverAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	<-quit
	logger.Info("Shutting down server...")

	// 优雅关闭资源
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 关闭WebSocket Hub
	hub.Close()

	// 停止Alert Monitor
	alertMonitor.Stop()
	logger.Info("Alert Monitor stopped")

	// 停止自动刷新
	supervisorService.StopAutoRefresh(stopRefresh)

	// 停止性能监控
	if appConfig.Performance.MemoryMonitoringEnabled {
		middleware.StopMemoryMonitoring()
		logger.Info("Memory monitoring stopped")
	}

	if appConfig.Performance.MetricsCleanupEnabled {
		middleware.StopPerformanceCleanup()
		logger.Info("Performance metrics cleanup stopped")
	}

	// 关闭数据库连接
	database.Close()
	logger.Info("Database connection closed")

	logger.Info("Server exited")
}

func updateAdminUser(db *gorm.DB, config *Config) error {
	adminUser := models.User{
		Username: config.Admin.Username,
		Email:    config.Admin.Email,
		IsAdmin:  true,
	}
	if err := adminUser.SetPassword(config.Admin.Password); err != nil {
		return fmt.Errorf("failed to set admin password: %v", err)
	}

	// 更新或创建admin用户
	var existingUser models.User
	if err := db.Where("is_admin = ?", true).First(&existingUser).Error; err == nil {
		// 更新现有admin
		if err := db.Model(&existingUser).Updates(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to update admin user: %v", err)
		}
	} else {
		// 创建新admin
		if err := db.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}
	}
	return nil
}

func loadConfig() (*Config, error) {
	// 获取当前工作目录作为项目根目录
	projectRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(projectRoot)
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("admin.username", "admin")
	viper.SetDefault("admin.email", "admin@example.com")

	// 启用环境变量替换
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	logger.Info("Config file loaded",
		zap.String("config_file", viper.ConfigFileUsed()),
		zap.Any("nodes_config", viper.Get("nodes")))

	// 解析配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// 从环境变量获取敏感信息
	config.Admin.Password = os.Getenv("ADMIN_PASSWORD")
	for i := range config.Nodes {
		config.Nodes[i].Password = os.Getenv("NODE_PASSWORD")
	}

	return &config, nil
}
