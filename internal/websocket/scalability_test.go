package websocket

import (
	"testing"
	"superview/internal/config"
	"superview/internal/supervisor"
)

func TestWebSocketScalability(t *testing.T) {
	// Test WebSocket config with performance settings
	perfConfig := &config.PerformanceConfig{
		MaxWebSocketConnections: 1000,
	}
	
	wsConfig := GetWebSocketConfigFromPerformance(perfConfig)
	
	// Verify WebSocket supports the configured limit
	if wsConfig.MaxConnections != 1000 {
		t.Errorf("Expected MaxConnections=1000, got %d", wsConfig.MaxConnections)
	}
	
	// Verify it can handle hundreds of nodes scenario
	if wsConfig.MaxConnections < 500 {
		t.Errorf("WebSocket MaxConnections too low for hundreds of nodes: %d", 
			wsConfig.MaxConnections)
	}
}

func TestWebSocketDefaultScalability(t *testing.T) {
	// Test default configuration supports reasonable scale
	wsConfig := GetDefaultWebSocketConfig()
	
	// Default should support at least 500 connections
	if wsConfig.MaxConnections < 500 {
		t.Errorf("Default WebSocket MaxConnections too low: %d, need at least 500", 
			wsConfig.MaxConnections)
	}
}

func TestHubCreationWithScalableConfig(t *testing.T) {
	// Create a mock supervisor service
	service := supervisor.NewSupervisorService()
	
	// Test Hub creation with scalable config
	perfConfig := &config.PerformanceConfig{
		MaxWebSocketConnections: 800,
	}
	
	wsConfig := GetWebSocketConfigFromPerformance(perfConfig)
	hub := NewHubWithConfig(service, wsConfig)
	
	if hub.config.MaxConnections != 800 {
		t.Errorf("Hub not using configured MaxConnections: expected 800, got %d", 
			hub.config.MaxConnections)
	}
	
	// Clean up
	hub.Close()
}