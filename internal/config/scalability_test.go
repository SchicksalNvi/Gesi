package config

import (
	"testing"
)

func TestScalabilityDefaults(t *testing.T) {
	cfg := &Config{
		Performance: PerformanceConfig{},
	}
	
	// Apply defaults
	if cfg.Performance.MaxConcurrentConnections == 0 {
		cfg.Performance.MaxConcurrentConnections = 100
	}
	if cfg.Performance.MaxWebSocketConnections == 0 {
		cfg.Performance.MaxWebSocketConnections = 500
	}
	
	// Verify defaults support hundreds of nodes
	if cfg.Performance.MaxConcurrentConnections < 100 {
		t.Errorf("MaxConcurrentConnections too low: %d, need at least 100 for scalability", 
			cfg.Performance.MaxConcurrentConnections)
	}
	
	if cfg.Performance.MaxWebSocketConnections < 500 {
		t.Errorf("MaxWebSocketConnections too low: %d, need at least 500 for scalability", 
			cfg.Performance.MaxWebSocketConnections)
	}
}

func TestScalabilityConfiguration(t *testing.T) {
	cfg := &Config{
		Performance: PerformanceConfig{
			MaxConcurrentConnections: 200,
			MaxWebSocketConnections:  1000,
		},
	}
	
	// Should respect configured values
	if cfg.Performance.MaxConcurrentConnections != 200 {
		t.Errorf("Expected MaxConcurrentConnections=200, got %d", 
			cfg.Performance.MaxConcurrentConnections)
	}
	
	if cfg.Performance.MaxWebSocketConnections != 1000 {
		t.Errorf("Expected MaxWebSocketConnections=1000, got %d", 
			cfg.Performance.MaxWebSocketConnections)
	}
}