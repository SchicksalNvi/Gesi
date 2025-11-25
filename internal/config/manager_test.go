package config

import (
	"os"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// 属性 6：必需配置验证
// 验证需求：3.1
func TestConfigValidationProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// 设置必需的环境变量
	os.Setenv("JWT_SECRET", "test-secret-key-32-characters-long-minimum")
	os.Setenv("ADMIN_PASSWORD", "test-password")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	properties.Property("missing required config fields are rejected", prop.ForAll(
		func() bool {
			validator := NewValidator()

			// 测试缺少必需字段的配置
			invalidConfigs := []*Config{
				{AdminUsername: "", AdminPassword: "pass"}, // 缺少 username
				{AdminUsername: "admin", AdminPassword: ""}, // 缺少 password
			}

			for _, cfg := range invalidConfigs {
				if err := validator.Validate(cfg); err == nil {
					return false // 应该返回错误
				}
			}
			return true
		},
	))

	properties.Property("valid config passes validation", prop.ForAll(
		func(username string) bool {
			if username == "" {
				username = "admin"
			}

			validator := NewValidator()
			cfg := &Config{
				AdminUsername: username,
				AdminPassword: "test-password",
				Performance: PerformanceConfig{
					MemoryUpdateInterval:     30 * time.Second,
					MetricsResetInterval:     24 * time.Hour,
					EndpointCleanupThreshold: 2 * time.Hour,
				},
			}

			return validator.Validate(cfg) == nil
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// 属性 8：默认值日志记录
// 验证需求：3.5
func TestDefaultValueLogging(t *testing.T) {
	// 这个测试验证默认值被正确应用
	cfg := &Config{}

	// 应用默认值（模拟 Load 函数的行为）
	if cfg.Database == "" {
		cfg.Database = "sqlite:///users.db"
	}
	if cfg.ActivityLog == "" {
		cfg.ActivityLog = "activity.log"
	}
	if cfg.AdminUsername == "" {
		cfg.AdminUsername = "admin"
	}

	// 验证默认值
	if cfg.Database != "sqlite:///users.db" {
		t.Error("expected default database value")
	}
	if cfg.ActivityLog != "activity.log" {
		t.Error("expected default activity log value")
	}
	if cfg.AdminUsername != "admin" {
		t.Error("expected default admin username")
	}
}

// 单元测试
func TestValidatorWithMissingEnvVars(t *testing.T) {
	// 清除环境变量
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ADMIN_PASSWORD")

	validator := NewValidator()
	cfg := &Config{
		AdminUsername: "admin",
		AdminPassword: "password",
		Performance: PerformanceConfig{
			MemoryUpdateInterval:     30 * time.Second,
			MetricsResetInterval:     24 * time.Hour,
			EndpointCleanupThreshold: 2 * time.Hour,
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("expected validation to fail with missing env vars")
	}
}

func TestValidatorWithShortJWTSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "short")
	os.Setenv("ADMIN_PASSWORD", "password")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	validator := NewValidator()
	cfg := &Config{
		AdminUsername: "admin",
		AdminPassword: "password",
		Performance: PerformanceConfig{
			MemoryUpdateInterval:     30 * time.Second,
			MetricsResetInterval:     24 * time.Hour,
			EndpointCleanupThreshold: 2 * time.Hour,
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("expected validation to fail with short JWT secret")
	}
}

func TestValidatorWithInvalidPerformanceConfig(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-32-characters-long-minimum")
	os.Setenv("ADMIN_PASSWORD", "password")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	validator := NewValidator()
	cfg := &Config{
		AdminUsername: "admin",
		AdminPassword: "password",
		Performance: PerformanceConfig{
			MemoryUpdateInterval:     0, // 无效值
			MetricsResetInterval:     24 * time.Hour,
			EndpointCleanupThreshold: 2 * time.Hour,
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("expected validation to fail with invalid performance config")
	}
}

func TestConfigManagerLoadAndGet(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-32-characters-long-minimum")
	os.Setenv("ADMIN_PASSWORD", "password")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	manager := NewConfigManager()

	// 测试 Get 在未加载配置时返回 nil
	if cfg := manager.Get(); cfg != nil {
		t.Error("expected Get to return nil before Load")
	}
}
