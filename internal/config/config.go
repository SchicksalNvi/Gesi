package config

import (
	"time"
	
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Database         string
	ActivityLog      string
	AdminUsername    string
	AdminPassword    string
	NodeNames        []string
	NodeEnvironments []string
	Nodes            []NodeConfig
	DeveloperTools   DeveloperToolsConfig
	Performance      PerformanceConfig    `toml:"performance" json:"performance"`
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
}

type NodeConfig struct {
	Name        string
	Environment string
	Address     string
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
