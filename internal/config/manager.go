package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go-cesi/internal/logger"
	"go.uber.org/zap"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	Load(path string) (*Config, error)
	Validate(cfg *Config) error
	Watch(callback func(*Config)) error
	Get() *Config
	Stop()
}

// configManager 配置管理器实现
type configManager struct {
	config   *Config
	mu       sync.RWMutex
	watcher  *fsnotify.Watcher
	stopChan chan struct{}
	callback func(*Config)
	stopped  bool
}

// NewConfigManager 创建配置管理器
func NewConfigManager() ConfigManager {
	return &configManager{
		stopChan: make(chan struct{}),
	}
}

// Load 加载配置
func (m *configManager) Load(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 验证配置
	if err := m.Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	m.mu.Lock()
	m.config = cfg
	m.mu.Unlock()

	// 记录使用的默认值
	m.logDefaultValues(cfg)

	return cfg, nil
}

// Validate 验证配置
func (m *configManager) Validate(cfg *Config) error {
	validator := NewValidator()
	return validator.Validate(cfg)
}

// Watch 监听配置文件变化
func (m *configManager) Watch(callback func(*Config)) error {
	m.callback = callback

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	m.watcher = watcher

	// 监听配置文件
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return fmt.Errorf("no config file loaded")
	}

	if err := watcher.Add(configFile); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	go m.watchLoop()

	logger.Info("Config file watcher started", zap.String("file", configFile))
	return nil
}

// watchLoop 监听循环
func (m *configManager) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				logger.Info("Config file changed, reloading", zap.String("file", event.Name))
				m.reloadConfig()
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("Config watcher error", zap.Error(err))
		case <-m.stopChan:
			return
		}
	}
}

// reloadConfig 重新加载配置
func (m *configManager) reloadConfig() {
	// 重新读取配置
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Failed to reload config", zap.Error(err))
		return
	}

	var newConfig Config
	if err := viper.Unmarshal(&newConfig); err != nil {
		logger.Error("Failed to unmarshal config", zap.Error(err))
		return
	}

	// 从环境变量获取敏感信息
	newConfig.AdminPassword = os.Getenv("ADMIN_PASSWORD")

	// 验证新配置
	if err := m.Validate(&newConfig); err != nil {
		logger.Error("New config validation failed, keeping old config", zap.Error(err))
		return
	}

	// 更新配置
	m.mu.Lock()
	m.config = &newConfig
	m.mu.Unlock()

	logger.Info("Config reloaded successfully")

	// 调用回调函数
	if m.callback != nil {
		m.callback(&newConfig)
	}
}

// Get 获取当前配置
func (m *configManager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Stop 停止配置管理器
func (m *configManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.stopped {
		return // 已经停止，避免重复关闭
	}
	
	m.stopped = true
	close(m.stopChan)
	if m.watcher != nil {
		m.watcher.Close()
	}
}

// logDefaultValues 记录使用的默认值
func (m *configManager) logDefaultValues(cfg *Config) {
	if cfg.Database == "sqlite:///users.db" {
		logger.Info("Using default value for database", zap.String("value", cfg.Database))
	}
	if cfg.ActivityLog == "activity.log" {
		logger.Info("Using default value for activity_log", zap.String("value", cfg.ActivityLog))
	}
	if cfg.AdminUsername == "admin" {
		logger.Info("Using default value for admin_username", zap.String("value", cfg.AdminUsername))
	}
	if cfg.Performance.MemoryUpdateInterval == 30*time.Second {
		logger.Info("Using default value for memory_update_interval", zap.Duration("value", cfg.Performance.MemoryUpdateInterval))
	}
	if cfg.Performance.MetricsResetInterval == 24*time.Hour {
		logger.Info("Using default value for metrics_reset_interval", zap.Duration("value", cfg.Performance.MetricsResetInterval))
	}
	if cfg.Performance.EndpointCleanupThreshold == 2*time.Hour {
		logger.Info("Using default value for endpoint_cleanup_threshold", zap.Duration("value", cfg.Performance.EndpointCleanupThreshold))
	}
}
