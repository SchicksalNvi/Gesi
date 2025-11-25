package config

import (
	"fmt"
	"os"
	"strings"
)

// Validator 配置验证器接口
type Validator interface {
	Validate(cfg *Config) error
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
		if node.Name == "" {
			errors = append(errors, fmt.Sprintf("node[%d].name is required", i))
		}
		if node.Environment == "" {
			errors = append(errors, fmt.Sprintf("node[%d].environment is required", i))
		}
		if node.Address == "" {
			errors = append(errors, fmt.Sprintf("node[%d].address is required", i))
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
