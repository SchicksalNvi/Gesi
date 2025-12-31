package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"go.uber.org/zap"
)

// ConfigLoader 负责从多个源加载和合并配置
type ConfigLoader struct {
	mainConfigPath string
	nodeListPath   string
}

// NodeListConfig 节点列表配置文件结构
type NodeListConfig struct {
	Nodes []NodeConfig `toml:"nodes"`
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader(mainPath, nodeListPath string) *ConfigLoader {
	return &ConfigLoader{
		mainConfigPath: mainPath,
		nodeListPath:   nodeListPath,
	}
}

// Load 加载完整配置（系统配置 + 节点配置）
func (l *ConfigLoader) Load() (*Config, error) {
	// 加载主配置
	cfg, err := l.LoadMainConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load main config: %w", err)
	}

	// 加载节点列表
	nodeListNodes, err := l.LoadNodeList()
	if err != nil {
		return nil, fmt.Errorf("failed to load node list: %w", err)
	}

	// 合并节点配置
	cfg.Nodes = l.MergeNodes(cfg.Nodes, nodeListNodes)

	// 展开环境变量
	l.expandEnvVars(cfg)

	return cfg, nil
}

// LoadWithDefaults 加载完整配置并应用默认值（用于初始化）
func (l *ConfigLoader) LoadWithDefaults() (*Config, error) {
	// 使用原有的 Load 函数加载主配置（包含默认值）
	cfg, err := Load(l.mainConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load main config: %w", err)
	}

	// 加载节点列表
	nodeListNodes, err := l.LoadNodeList()
	if err != nil {
		return nil, fmt.Errorf("failed to load node list: %w", err)
	}

	// 合并节点配置
	cfg.Nodes = l.MergeNodes(cfg.Nodes, nodeListNodes)

	// 展开环境变量
	l.expandEnvVars(cfg)

	return cfg, nil
}

// LoadMainConfig 加载系统配置
func (l *ConfigLoader) LoadMainConfig() (*Config, error) {
	// 直接解析配置文件，不应用默认值（热重载时需要）
	data, err := os.ReadFile(l.mainConfigPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read file: %w", l.mainConfigPath, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%s: parse error: %w", l.mainConfigPath, err)
	}

	// 处理新旧配置格式的兼容性
	if cfg.Admin.Username != "" {
		cfg.AdminUsername = cfg.Admin.Username
	}
	if cfg.Admin.Password != "" {
		cfg.AdminPassword = cfg.Admin.Password
	}

	return &cfg, nil
}

// LoadNodeList 加载节点列表（如果文件不存在返回空切片）
func (l *ConfigLoader) LoadNodeList() ([]NodeConfig, error) {
	// 如果路径为空，返回空切片
	if l.nodeListPath == "" {
		return []NodeConfig{}, nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(l.nodeListPath); os.IsNotExist(err) {
		// 文件不存在，返回空切片（不是错误）
		return []NodeConfig{}, nil
	}

	// 读取文件
	data, err := os.ReadFile(l.nodeListPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read file: %w", l.nodeListPath, err)
	}

	// 解析 TOML
	var nodeListCfg NodeListConfig
	if err := toml.Unmarshal(data, &nodeListCfg); err != nil {
		// 包装错误以包含文件路径
		return nil, fmt.Errorf("%s: parse error: %w", l.nodeListPath, err)
	}

	return nodeListCfg.Nodes, nil
}

// MergeNodes 合并节点配置（nodelist.toml 优先）
func (l *ConfigLoader) MergeNodes(mainNodes, nodeListNodes []NodeConfig) []NodeConfig {
	// 创建 map 跟踪 nodelist 中的节点名
	nodeListMap := make(map[string]NodeConfig)
	for _, node := range nodeListNodes {
		nodeListMap[node.Name] = node
	}

	// 结果切片，先添加所有 nodelist 节点
	result := make([]NodeConfig, 0, len(mainNodes)+len(nodeListNodes))
	result = append(result, nodeListNodes...)

	// 添加 main config 中不重复的节点
	for _, node := range mainNodes {
		if _, exists := nodeListMap[node.Name]; exists {
			// 节点名重复，记录警告
			zap.L().Warn("Duplicate node name found, using configuration from nodelist.toml",
				zap.String("node_name", node.Name),
				zap.String("source", "config.toml"),
				zap.String("priority", "nodelist.toml"))
		} else {
			// 节点名不重复，添加到结果
			result = append(result, node)
		}
	}

	return result
}

// expandEnvVars 展开配置中的环境变量
func (l *ConfigLoader) expandEnvVars(cfg *Config) {
	// 展开管理员密码
	cfg.AdminPassword = os.ExpandEnv(cfg.AdminPassword)

	// 展开数据库路径
	cfg.Database = os.ExpandEnv(cfg.Database)

	// 展开活动日志路径
	cfg.ActivityLog = os.ExpandEnv(cfg.ActivityLog)

	// 展开节点配置中的环境变量
	for i := range cfg.Nodes {
		cfg.Nodes[i].Host = os.ExpandEnv(cfg.Nodes[i].Host)
		cfg.Nodes[i].Username = os.ExpandEnv(cfg.Nodes[i].Username)
		cfg.Nodes[i].Password = os.ExpandEnv(cfg.Nodes[i].Password)
	}

	// 展开开发者工具配置
	cfg.DeveloperTools.LogPath = os.ExpandEnv(cfg.DeveloperTools.LogPath)
}
