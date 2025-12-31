package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go-cesi/internal/logger"
	"go.uber.org/zap"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	Load(path string) (*Config, error)
	LoadWithNodeList(mainPath, nodeListPath string) (*Config, error)
	Validate(cfg *Config) error
	Watch(callback func(*Config)) error
	WatchNodeList(nodeListPath string, callback func([]NodeConfig)) error
	Get() *Config
	Stop()
}

// configManager 配置管理器实现
type configManager struct {
	config           *Config
	mu               sync.RWMutex
	watcher          *fsnotify.Watcher
	nodeListWatcher  *fsnotify.Watcher
	stopChan         chan struct{}
	nodeListStopChan chan struct{}
	callback         func(*Config)
	nodeListCallback func([]NodeConfig)
	mainConfigPath   string  // 添加主配置路径
	nodeListPath     string
	stopped          bool
}

// NewConfigManager 创建配置管理器
func NewConfigManager() ConfigManager {
	return &configManager{
		stopChan:         make(chan struct{}),
		nodeListStopChan: make(chan struct{}),
	}
}

// Load 加载配置
func (m *configManager) Load(path string) (*Config, error) {
	// 使用 ConfigLoader 加载配置（包含默认值）
	loader := NewConfigLoader(path, "")
	cfg, err := loader.LoadWithDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 验证配置
	if err := m.Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	m.mu.Lock()
	m.config = cfg
	m.mainConfigPath = path  // 保存主配置路径
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

	// 使用保存的主配置路径
	m.mu.RLock()
	configFile := m.mainConfigPath
	m.mu.RUnlock()
	
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
	// 使用保存的配置路径
	m.mu.RLock()
	mainPath := m.mainConfigPath
	nodeListPath := m.nodeListPath
	m.mu.RUnlock()
	
	if mainPath == "" {
		logger.Error("No config file path saved, cannot reload")
		return
	}

	// 使用 ConfigLoader 重新加载配置
	loader := NewConfigLoader(mainPath, nodeListPath)
	newConfig, err := loader.Load()
	if err != nil {
		logger.Error("Failed to reload config using ConfigLoader", zap.Error(err))
		return
	}

	// 验证新配置
	if err := m.Validate(newConfig); err != nil {
		logger.Error("New config validation failed, keeping old config", zap.Error(err))
		return
	}

	// 更新配置
	m.mu.Lock()
	m.config = newConfig
	m.mu.Unlock()

	logger.Info("Config reloaded successfully")

	// 调用回调函数
	if m.callback != nil {
		m.callback(newConfig)
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
	close(m.nodeListStopChan)
	if m.watcher != nil {
		m.watcher.Close()
	}
	if m.nodeListWatcher != nil {
		m.nodeListWatcher.Close()
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

// WatchNodeList 监听节点列表文件变化
func (m *configManager) WatchNodeList(nodeListPath string, callback func([]NodeConfig)) error {
	m.nodeListPath = nodeListPath
	m.nodeListCallback = callback

	// 如果路径为空，跳过监听
	if nodeListPath == "" {
		logger.Info("Node list path is empty, skipping watch")
		return nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(nodeListPath); os.IsNotExist(err) {
		logger.Info("Node list file does not exist, skipping watch", zap.String("file", nodeListPath))
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create node list watcher: %w", err)
	}
	m.nodeListWatcher = watcher

	if err := watcher.Add(nodeListPath); err != nil {
		return fmt.Errorf("failed to watch node list file: %w", err)
	}

	go m.watchNodeListLoop()

	logger.Info("Node list file watcher started", zap.String("file", nodeListPath))
	return nil
}

// watchNodeListLoop 监听节点列表文件循环
func (m *configManager) watchNodeListLoop() {
	for {
		select {
		case event, ok := <-m.nodeListWatcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				logger.Info("Node list file changed, reloading", zap.String("file", event.Name))
				m.reloadNodeList()
			}
		case err, ok := <-m.nodeListWatcher.Errors:
			if !ok {
				return
			}
			logger.Error("Node list watcher error", zap.Error(err))
		case <-m.nodeListStopChan:
			return
		}
	}
}

// reloadNodeList 重新加载节点列表
func (m *configManager) reloadNodeList() {
	// 获取当前配置路径
	m.mu.RLock()
	mainPath := m.mainConfigPath
	nodeListPath := m.nodeListPath
	m.mu.RUnlock()

	// 使用 ConfigLoader 重新加载完整配置（包括节点合并）
	loader := NewConfigLoader(mainPath, nodeListPath)
	
	// 加载主配置中的节点
	mainConfig, err := loader.LoadMainConfig()
	if err != nil {
		logger.Error("Failed to reload main config during node list reload", zap.Error(err))
		return
	}
	
	// 加载节点列表
	nodeListNodes, err := loader.LoadNodeList()
	if err != nil {
		logger.Error("Failed to reload node list", zap.Error(err))
		return
	}

	// 合并节点配置
	mergedNodes := loader.MergeNodes(mainConfig.Nodes, nodeListNodes)

	// 验证新节点配置
	validator := NewValidator()
	for i, node := range mergedNodes {
		if err := validator.ValidateNode(node); err != nil {
			logger.Error("Node list validation failed, keeping old config",
				zap.Int("node_index", i),
				zap.String("node_name", node.Name),
				zap.Error(err))
			return
		}
	}

	// 原子更新配置中的节点列表
	m.mu.Lock()
	if m.config != nil {
		m.config.Nodes = mergedNodes
	}
	m.mu.Unlock()

	logger.Info("Node list reloaded successfully", zap.Int("node_count", len(mergedNodes)))

	// 调用回调函数
	if m.nodeListCallback != nil {
		m.nodeListCallback(mergedNodes)
	}
}
// LoadWithNodeList 使用 ConfigLoader 加载配置（支持节点列表分离）
func (m *configManager) LoadWithNodeList(mainPath, nodeListPath string) (*Config, error) {
	// 使用 ConfigLoader 加载配置（包含默认值）
	loader := NewConfigLoader(mainPath, nodeListPath)
	cfg, err := loader.LoadWithDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 验证配置
	if err := m.Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	m.mu.Lock()
	m.config = cfg
	m.mainConfigPath = mainPath    // 保存主配置路径
	m.nodeListPath = nodeListPath  // 保存节点列表路径用于热重载
	m.mu.Unlock()

	// 记录使用的默认值
	m.logDefaultValues(cfg)

	return cfg, nil
}