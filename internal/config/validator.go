package config

import (
	"fmt"
	"os"
	"strings"
)

// Validator 配置验证器接口
type Validator interface {
	Validate(cfg *Config) error
	ValidateNode(node NodeConfig) error
}

// validator 配置验证器实现
type validator struct{}

// NewValidator 创建配置验证器
func NewValidator() Validator {
	return &validator{}
}

// Validate 验证配置
func (v *validator) Validate(cfg *Config) error {
	var errors []string

	// 验证必需的环境变量
	if err := v.validateRequiredEnvVars(); err != nil {
		errors = append(errors, err.Error())
	}

	// 验证 JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		errors = append(errors, "JWT_SECRET must be at least 32 characters long")
	}

	// 验证管理员配置
	if cfg.AdminUsername == "" {
		errors = append(errors, "admin username is required")
	}
	if cfg.AdminPassword == "" {
		errors = append(errors, "admin password is required (set ADMIN_PASSWORD env var)")
	}

	// 验证性能配置
	if cfg.Performance.MemoryUpdateInterval <= 0 {
		errors = append(errors, "memory_update_interval must be positive")
	}
	if cfg.Performance.MetricsResetInterval <= 0 {
		errors = append(errors, "metrics_reset_interval must be positive")
	}
	if cfg.Performance.EndpointCleanupThreshold <= 0 {
		errors = append(errors, "endpoint_cleanup_threshold must be positive")
	}

	// 验证节点配置
	for i, node := range cfg.Nodes {
		if err := v.ValidateNode(node); err != nil {
			errors = append(errors, fmt.Sprintf("node[%d]: %s", i, err.Error()))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// validateRequiredEnvVars 验证必需的环境变量
func (v *validator) validateRequiredEnvVars() error {
	requiredVars := []string{
		"JWT_SECRET",
		"ADMIN_PASSWORD",
	}

	var missing []string
	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidateNode 验证单个节点配置（源无关）
func (v *validator) ValidateNode(node NodeConfig) error {
	var errors []string

	// 验证必填字段
	if node.Name == "" {
		errors = append(errors, "name is required")
	}
	if node.Environment == "" {
		errors = append(errors, "environment is required")
	}
	if node.Host == "" {
		errors = append(errors, "host is required")
	}

	// 验证端口范围
	if node.Port <= 0 || node.Port > 65535 {
		errors = append(errors, "port must be between 1 and 65535")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, ", "))
	}

	return nil
}
