package supervisor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
	"go-cesi/internal/config"
)

func TestSupervisorServiceScalability(t *testing.T) {
	// Test with configuration that supports hundreds of nodes
	perfConfig := &config.PerformanceConfig{
		MaxConcurrentConnections: 200,
		MaxWebSocketConnections:  1000,
	}
	
	service := NewSupervisorServiceWithConfig(perfConfig)
	
	// Verify service uses the configured limits
	if cap(service.connectionSemaphore) != 200 {
		t.Errorf("Expected connection semaphore capacity=200, got %d", 
			cap(service.connectionSemaphore))
	}
	
	if service.config.MaxConcurrentConnections != 200 {
		t.Errorf("Expected MaxConcurrentConnections=200, got %d", 
			service.config.MaxConcurrentConnections)
	}
}

func TestConcurrentNodeOperations(t *testing.T) {
	perfConfig := &config.PerformanceConfig{
		MaxConcurrentConnections: 100,
	}
	
	service := NewSupervisorServiceWithConfig(perfConfig)
	
	// Start the service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop(ctx)
	
	// Simulate adding many nodes concurrently
	const numNodes = 50
	var wg sync.WaitGroup
	errors := make(chan error, numNodes)
	
	for i := 0; i < numNodes; i++ {
		wg.Add(1)
		go func(nodeID int) {
			defer wg.Done()
			
			// Add node (will fail to connect but should not crash)
			err := service.AddNode(
				fmt.Sprintf("test-node-%d", nodeID),
				"test",
				"127.0.0.1",
				9001+nodeID,
				"test",
				"test",
			)
			
			// Connection failure is expected, but service should handle it gracefully
			if err != nil && !isConnectionError(err) {
				errors <- err
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for unexpected errors
	for err := range errors {
		t.Errorf("Unexpected error during concurrent node operations: %v", err)
	}
	
	// Verify all nodes were added (even if not connected)
	nodes := service.GetAllNodes()
	if len(nodes) != numNodes {
		t.Errorf("Expected %d nodes, got %d", numNodes, len(nodes))
	}
}

func isConnectionError(err error) bool {
	// Simple check for connection-related errors
	errStr := err.Error()
	return contains(errStr, "connection") || 
		   contains(errStr, "refused") || 
		   contains(errStr, "timeout")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || (len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		   containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}