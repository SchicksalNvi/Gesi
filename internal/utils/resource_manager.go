package utils

import (
	"context"
	"sync"
	"time"

	"go-cesi/internal/logger"
	"go.uber.org/zap"
)

// ResourceManager 资源管理器，用于统一管理应用程序资源的生命周期
type ResourceManager struct {
	mu        sync.RWMutex
	resources []Resource
	closed    bool
}

// Resource 资源接口，所有需要清理的资源都应该实现这个接口
type Resource interface {
	Close() error
	Name() string
}

// TickerResource 定时器资源包装器
type TickerResource struct {
	name   string
	ticker *time.Ticker
	stop   chan struct{}
}

func NewTickerResource(name string, ticker *time.Ticker, stop chan struct{}) *TickerResource {
	return &TickerResource{
		name:   name,
		ticker: ticker,
		stop:   stop,
	}
}

func (tr *TickerResource) Close() error {
	if tr.ticker != nil {
		tr.ticker.Stop()
	}
	if tr.stop != nil {
		close(tr.stop)
	}
	return nil
}

func (tr *TickerResource) Name() string {
	return tr.name
}

// ChannelResource 通道资源包装器
type ChannelResource struct {
	name string
	ch   chan struct{}
}

func NewChannelResource(name string, ch chan struct{}) *ChannelResource {
	return &ChannelResource{
		name: name,
		ch:   ch,
	}
}

func (cr *ChannelResource) Close() error {
	if cr.ch != nil {
		close(cr.ch)
	}
	return nil
}

func (cr *ChannelResource) Name() string {
	return cr.name
}

// GoroutineResource goroutine资源包装器
type GoroutineResource struct {
	name   string
	cancel context.CancelFunc
	done   <-chan struct{}
}

func NewGoroutineResource(name string, cancel context.CancelFunc, done <-chan struct{}) *GoroutineResource {
	return &GoroutineResource{
		name:   name,
		cancel: cancel,
		done:   done,
	}
}

func (gr *GoroutineResource) Close() error {
	if gr.cancel != nil {
		gr.cancel()
	}
	// 等待goroutine结束，最多等待5秒
	if gr.done != nil {
		select {
		case <-gr.done:
			return nil
		case <-time.After(5 * time.Second):
			logger.Warn("Goroutine did not exit within timeout", zap.String("name", gr.name))
			return nil
		}
	}
	return nil
}

func (gr *GoroutineResource) Name() string {
	return gr.name
}

// NewResourceManager 创建新的资源管理器
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make([]Resource, 0),
		closed:    false,
	}
}

// Register 注册资源
func (rm *ResourceManager) Register(resource Resource) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if rm.closed {
		logger.Warn("Attempting to register resource after manager is closed", 
			zap.String("resource", resource.Name()))
		return
	}
	
	rm.resources = append(rm.resources, resource)
	logger.Debug("Resource registered", zap.String("name", resource.Name()))
}

// Unregister 注销资源
func (rm *ResourceManager) Unregister(resourceName string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	for i, resource := range rm.resources {
		if resource.Name() == resourceName {
			// 移除资源
			rm.resources = append(rm.resources[:i], rm.resources[i+1:]...)
			logger.Debug("Resource unregistered", zap.String("name", resourceName))
			return
		}
	}
}

// CloseAll 关闭所有注册的资源
func (rm *ResourceManager) CloseAll() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if rm.closed {
		return
	}
	
	logger.Info("Closing all resources", zap.Int("count", len(rm.resources)))
	
	// 逆序关闭资源（后注册的先关闭）
	for i := len(rm.resources) - 1; i >= 0; i-- {
		resource := rm.resources[i]
		if err := resource.Close(); err != nil {
			logger.Error("Failed to close resource", 
				zap.String("name", resource.Name()),
				zap.Error(err))
		} else {
			logger.Debug("Resource closed successfully", 
				zap.String("name", resource.Name()))
		}
	}
	
	rm.resources = nil
	rm.closed = true
	logger.Info("All resources closed")
}

// IsClosed 检查资源管理器是否已关闭
func (rm *ResourceManager) IsClosed() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.closed
}

// GetResourceCount 获取当前注册的资源数量
func (rm *ResourceManager) GetResourceCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.resources)
}

// ListResources 列出所有注册的资源名称
func (rm *ResourceManager) ListResources() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	names := make([]string, len(rm.resources))
	for i, resource := range rm.resources {
		names[i] = resource.Name()
	}
	return names
}

// 全局资源管理器实例
var GlobalResourceManager = NewResourceManager()

// RegisterGlobalResource 注册全局资源
func RegisterGlobalResource(resource Resource) {
	GlobalResourceManager.Register(resource)
}

// UnregisterGlobalResource 注销全局资源
func UnregisterGlobalResource(resourceName string) {
	GlobalResourceManager.Unregister(resourceName)
}

// CloseAllGlobalResources 关闭所有全局资源
func CloseAllGlobalResources() {
	GlobalResourceManager.CloseAll()
}