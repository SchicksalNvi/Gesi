package config

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"go-cesi/internal/logger"
	"go.uber.org/zap"
)

// AtomicConfigManager 线程安全的配置管理器
type AtomicConfigManager struct {
	// 使用 atomic.Value 确保配置读取的原子性
	config atomic.Value // *Config
	
	// 配置更新序列化
	updateMu sync.Mutex
	
	// 文件监听
	watcher          *fsnotify.Watcher
	nodeListWatcher  *fsnotify.Watcher
	
	// 控制通道
	stopChan         chan struct{}
	nodeListStopChan chan struct{}
	
	// 回调函数
	callback         func(*Config)
	nodeListCallback func([]NodeConfig)
	
	// 配置路径
	mainConfigPath string
	nodeListPath   string
	
	// 状态管理
	stopped    int32 // atomic flag
	reloading  int32 // atomic flag - 防止并发重载
	
	// 配置版本和历史
	version        int64 // atomic counter
	lastValidConfig atomic.Value // *Config - 用于回滚
}

// NewAtomicConfigManager 创建原子配置管理器
func NewAtomicConfigManager() *AtomicConfigManager {
	return &AtomicConfigManager{
		stopChan:         make(chan struct{}),
		nodeListStopChan: make(chan struct{}),
	}
}

// Load 加载配置
func (m *AtomicConfigManager) Load(path string) (*Config, error) {
	loader := NewConfigLoader(path, "")
	cfg, err := loader.LoadWithDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 验证配置
	if err := m.validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 原子更新配置
	m.updateMu.Lock()
	// 保存当前配置作为上一个有效配置（如果存在）
	if currentConfig := m.config.Load(); currentConfig != nil {
		m.lastValidConfig.Store(currentConfig)
	}
	m.config.Store(cfg)
	m.mainConfigPath = path
	atomic.AddInt64(&m.version, 1)
	m.updateMu.Unlock()

	m.logDefaultValues(cfg)
	logger.Info("Config loaded successfully", 
		zap.String("path", path),
		zap.Int64("version", atomic.LoadInt64(&m.version)))

	return cfg, nil
}

// LoadWithNodeList 加载配置（支持节点列表分离）
func (m *AtomicConfigManager) LoadWithNodeList(mainPath, nodeListPath string) (*Config, error) {
	loader := NewConfigLoader(mainPath, nodeListPath)
	cfg, err := loader.LoadWithDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 验证配置
	if err := m.validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 原子更新配置
	m.updateMu.Lock()
	// 保存当前配置作为上一个有效配置（如果存在）
	if currentConfig := m.config.Load(); currentConfig != nil {
		m.lastValidConfig.Store(currentConfig)
	}
	m.config.Store(cfg)
	m.mainConfigPath = mainPath
	m.nodeListPath = nodeListPath
	atomic.AddInt64(&m.version, 1)
	m.updateMu.Unlock()

	m.logDefaultValues(cfg)
	logger.Info("Config loaded successfully", 
		zap.String("main_path", mainPath),
		zap.String("node_list_path", nodeListPath),
		zap.Int64("version", atomic.LoadInt64(&m.version)))

	return cfg, nil
}

// Get 获取当前配置（原子操作）
func (m *AtomicConfigManager) Get() *Config {
	if cfg := m.config.Load(); cfg != nil {
		return cfg.(*Config)
	}
	return nil
}

// GetVersion 获取配置版本
func (m *AtomicConfigManager) GetVersion() int64 {
	return atomic.LoadInt64(&m.version)
}

// Watch 监听配置文件变化
func (m *AtomicConfigManager) Watch(callback func(*Config)) error {
	if atomic.LoadInt32(&m.stopped) != 0 {
		return fmt.Errorf("manager is stopped")
	}

	m.callback = callback

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	m.watcher = watcher

	configFile := m.mainConfigPath
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

// WatchNodeList 监听节点列表文件变化
func (m *AtomicConfigManager) WatchNodeList(nodeListPath string, callback func([]NodeConfig)) error {
	if atomic.LoadInt32(&m.stopped) != 0 {
		return fmt.Errorf("manager is stopped")
	}

	m.nodeListPath = nodeListPath
	m.nodeListCallback = callback

	if nodeListPath == "" {
		logger.Info("Node list path is empty, skipping watch")
		return nil
	}

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

// watchLoop 主配置文件监听循环
func (m *AtomicConfigManager) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				logger.Info("Config file changed, reloading", zap.String("file", event.Name))
				m.safeReloadConfig()
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

// watchNodeListLoop 节点列表文件监听循环
func (m *AtomicConfigManager) watchNodeListLoop() {
	for {
		select {
		case event, ok := <-m.nodeListWatcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				logger.Info("Node list file changed, reloading", zap.String("file", event.Name))
				m.safeReloadNodeList()
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

// safeReloadConfig 安全重载配置（防止并发）
func (m *AtomicConfigManager) safeReloadConfig() {
	// 使用原子操作防止并发重载
	if !atomic.CompareAndSwapInt32(&m.reloading, 0, 1) {
		logger.Debug("Config reload already in progress, skipping")
		return
	}
	defer atomic.StoreInt32(&m.reloading, 0)

	if atomic.LoadInt32(&m.stopped) != 0 {
		logger.Debug("Manager is stopped, skipping config reload")
		return
	}

	m.reloadConfig()
}

// safeReloadNodeList 安全重载节点列表（防止并发）
func (m *AtomicConfigManager) safeReloadNodeList() {
	// 使用原子操作防止并发重载
	if !atomic.CompareAndSwapInt32(&m.reloading, 0, 1) {
		logger.Debug("Node list reload already in progress, skipping")
		return
	}
	defer atomic.StoreInt32(&m.reloading, 0)

	if atomic.LoadInt32(&m.stopped) != 0 {
		logger.Debug("Manager is stopped, skipping node list reload")
		return
	}

	m.reloadNodeList()
}

// reloadConfig 重新加载配置
func (m *AtomicConfigManager) reloadConfig() {
	oldVersion := atomic.LoadInt64(&m.version)
	
	// 序列化配置更新
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	mainPath := m.mainConfigPath
	nodeListPath := m.nodeListPath
	
	if mainPath == "" {
		logger.Error("No config file path saved, cannot reload")
		return
	}

	// 加载新配置
	loader := NewConfigLoader(mainPath, nodeListPath)
	newConfig, err := loader.Load()
	if err != nil {
		logger.Error("Failed to reload config", 
			zap.Error(err),
			zap.Int64("version", oldVersion))
		return
	}

	// 验证新配置
	if err := m.validateConfig(newConfig); err != nil {
		logger.Error("New config validation failed, keeping old config", 
			zap.Error(err),
			zap.Int64("version", oldVersion))
		return
	}

	// 保存旧配置用于回滚
	if oldConfig := m.config.Load(); oldConfig != nil {
		m.lastValidConfig.Store(oldConfig)
	}

	// 原子更新配置
	m.config.Store(newConfig)
	newVersion := atomic.AddInt64(&m.version, 1)

	logger.Info("Config reloaded successfully", 
		zap.Int64("old_version", oldVersion),
		zap.Int64("new_version", newVersion))

	// 调用回调函数
	if m.callback != nil {
		// 在单独的 goroutine 中调用回调，避免阻塞
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Config callback panic", zap.Any("panic", r))
				}
			}()
			m.callback(newConfig)
		}()
	}
}

// reloadNodeList 重新加载节点列表
func (m *AtomicConfigManager) reloadNodeList() {
	oldVersion := atomic.LoadInt64(&m.version)
	
	// 序列化配置更新
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	mainPath := m.mainConfigPath
	nodeListPath := m.nodeListPath

	// 使用 ConfigLoader 重新加载完整配置
	loader := NewConfigLoader(mainPath, nodeListPath)
	
	mainConfig, err := loader.LoadMainConfig()
	if err != nil {
		logger.Error("Failed to reload main config during node list reload", 
			zap.Error(err),
			zap.Int64("version", oldVersion))
		return
	}
	
	nodeListNodes, err := loader.LoadNodeList()
	if err != nil {
		logger.Error("Failed to reload node list", 
			zap.Error(err),
			zap.Int64("version", oldVersion))
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
				zap.Error(err),
				zap.Int64("version", oldVersion))
			return
		}
	}

	// 获取当前配置并创建新版本
	currentConfig := m.Get()
	if currentConfig == nil {
		logger.Error("No current config available for node list update")
		return
	}

	// 保存旧配置用于回滚
	m.lastValidConfig.Store(currentConfig)

	// 创建新配置（只更新节点列表）
	newConfig := *currentConfig // 浅拷贝
	newConfig.Nodes = mergedNodes

	// 原子更新配置
	m.config.Store(&newConfig)
	newVersion := atomic.AddInt64(&m.version, 1)

	logger.Info("Node list reloaded successfully", 
		zap.Int("node_count", len(mergedNodes)),
		zap.Int64("old_version", oldVersion),
		zap.Int64("new_version", newVersion))

	// 调用回调函数
	if m.nodeListCallback != nil {
		// 在单独的 goroutine 中调用回调，避免阻塞
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Node list callback panic", zap.Any("panic", r))
				}
			}()
			m.nodeListCallback(mergedNodes)
		}()
	}
}

// Rollback 回滚到上一个有效配置
func (m *AtomicConfigManager) Rollback() error {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	lastValid := m.lastValidConfig.Load()
	if lastValid == nil {
		return fmt.Errorf("no valid config available for rollback")
	}

	oldVersion := atomic.LoadInt64(&m.version)
	
	// 原子更新配置
	m.config.Store(lastValid)
	newVersion := atomic.AddInt64(&m.version, 1)

	logger.Warn("Config rolled back to previous version", 
		zap.Int64("old_version", oldVersion),
		zap.Int64("new_version", newVersion))

	return nil
}

// Stop 停止配置管理器
func (m *AtomicConfigManager) Stop() {
	// 使用原子操作设置停止标志
	if !atomic.CompareAndSwapInt32(&m.stopped, 0, 1) {
		return // 已经停止
	}

	logger.Info("Stopping atomic config manager")

	// 关闭通道
	close(m.stopChan)
	close(m.nodeListStopChan)

	// 关闭文件监听器
	if m.watcher != nil {
		m.watcher.Close()
	}
	if m.nodeListWatcher != nil {
		m.nodeListWatcher.Close()
	}

	logger.Info("Atomic config manager stopped")
}

// validateConfig 验证配置
func (m *AtomicConfigManager) validateConfig(cfg *Config) error {
	validator := NewValidator()
	return validator.Validate(cfg)
}

// logDefaultValues 记录使用的默认值
func (m *AtomicConfigManager) logDefaultValues(cfg *Config) {
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
}

// Validate 实现 ConfigManager 接口
func (m *AtomicConfigManager) Validate(cfg *Config) error {
	return m.validateConfig(cfg)
}