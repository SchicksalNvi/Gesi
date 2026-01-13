package config

import (
	"time"
	
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Database         string                   `mapstructure:"database"`
	ActivityLog      string                   `mapstructure:"activity_log"`
	Admin            AdminConfig              `mapstructure:"admin"`
	// 保留向后兼容的字段
	AdminUsername    string                   `mapstructure:"admin_username"`
	AdminPassword    string                   `mapstructure:"admin_password"`
	NodeNames        []string                 `mapstructure:"node_names"`
	NodeEnvironments []string                 `mapstructure:"node_environments"`
	Nodes            []NodeConfig             `mapstructure:"nodes"`
	DeveloperTools   DeveloperToolsConfig     `mapstructure:"developer_tools"`
	Performance      PerformanceConfig        `mapstructure:"performance"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
}

// DeveloperToolsConfig 开发者工具配置
type DeveloperToolsConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	LogPath           string `mapstructure:"log_path"`
	MaxLogLines       int    `mapstructure:"max_log_lines"`
	APIDocsEnabled    bool   `mapstructure:"api_docs_enabled"`
	SampleDataEnabled bool   `mapstructure:"sample_data_enabled"`
	MetricsEnabled    bool   `mapstructure:"metrics_enabled"`
	SystemMetrics     SystemMetricsConfig `mapstructure:"system_metrics"`
}

// SystemMetricsConfig 系统指标配置
type SystemMetricsConfig struct {
	CPUMonitorEnabled  bool `mapstructure:"cpu_monitor_enabled"`
	DiskMonitorEnabled bool `mapstructure:"disk_monitor_enabled"`
	UpdateInterval     int  `mapstructure:"update_interval_seconds"`
}

// PerformanceConfig 性能监控配置
type PerformanceConfig struct {
	MemoryMonitoringEnabled    bool          `toml:"memory_monitoring_enabled" json:"memory_monitoring_enabled"`
	MemoryUpdateInterval       time.Duration `toml:"memory_update_interval" json:"memory_update_interval"`
	MetricsCleanupEnabled      bool          `toml:"metrics_cleanup_enabled" json:"metrics_cleanup_enabled"`
	MetricsResetInterval       time.Duration `toml:"metrics_reset_interval" json:"metrics_reset_interval"`
	EndpointCleanupThreshold   time.Duration `toml:"endpoint_cleanup_threshold" json:"endpoint_cleanup_threshold"`
	
	// Scalability settings - simple and direct
	MaxConcurrentConnections   int           `toml:"max_concurrent_connections" json:"max_concurrent_connections"`
	MaxWebSocketConnections    int           `toml:"max_websocket_connections" json:"max_websocket_connections"`
}

type NodeConfig struct {
	Name        string `mapstructure:"name" toml:"name"`
	Environment string `mapstructure:"environment" toml:"environment"`
	Host        string `mapstructure:"host" toml:"host"`
	Port        int    `mapstructure:"port" toml:"port"`
	Username    string `mapstructure:"username" toml:"username"`
	Password    string `mapstructure:"password" toml:"password"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// 处理新旧配置格式的兼容性
	// 如果使用新格式（Admin 结构体），将其复制到旧字段
	if cfg.Admin.Username != "" {
		cfg.AdminUsername = cfg.Admin.Username
	}
	if cfg.Admin.Password != "" {
		cfg.AdminPassword = cfg.Admin.Password
	}

	// Set defaults
	if cfg.Database == "" {
		cfg.Database = "sqlite:///users.db"
	}
	if cfg.ActivityLog == "" {
		cfg.ActivityLog = "activity.log"
	}
	if cfg.AdminUsername == "" {
		cfg.AdminUsername = "admin"
	}
	if cfg.AdminPassword == "" {
		cfg.AdminPassword = "admin"
	}
	
	// Set DeveloperTools defaults
	if cfg.DeveloperTools.LogPath == "" {
		cfg.DeveloperTools.LogPath = "logs/app.log"
	}
	if cfg.DeveloperTools.MaxLogLines == 0 {
		cfg.DeveloperTools.MaxLogLines = 1000
	}
	if cfg.DeveloperTools.SystemMetrics.UpdateInterval == 0 {
		cfg.DeveloperTools.SystemMetrics.UpdateInterval = 30
	}

	// 设置性能监控默认值
	if cfg.Performance.MemoryUpdateInterval == 0 {
		cfg.Performance.MemoryUpdateInterval = 30 * time.Second
	}
	if cfg.Performance.MetricsResetInterval == 0 {
		cfg.Performance.MetricsResetInterval = 24 * time.Hour // 24小时重置一次
	}
	if cfg.Performance.EndpointCleanupThreshold == 0 {
		cfg.Performance.EndpointCleanupThreshold = 2 * time.Hour // 2小时未访问则清理
	}
	// Set scalability defaults - simple and practical
	if cfg.Performance.MaxConcurrentConnections == 0 {
		cfg.Performance.MaxConcurrentConnections = 100 // Support 100 concurrent connections
	}
	if cfg.Performance.MaxWebSocketConnections == 0 {
		cfg.Performance.MaxWebSocketConnections = 500 // Support 500 WebSocket connections
	}
	// 默认启用内存监控和指标清理
	if !cfg.Performance.MemoryMonitoringEnabled && !cfg.Performance.MetricsCleanupEnabled {
		cfg.Performance.MemoryMonitoringEnabled = true
		cfg.Performance.MetricsCleanupEnabled = true
	}

	return &cfg, nil
}

func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		zap.L().Fatal("Failed to load config", zap.Error(err))
	}
	return cfg
}
