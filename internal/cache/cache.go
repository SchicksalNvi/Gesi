package cache

import (
	"context"
	"sync"
	"time"

	"superview/internal/logger"
	"go.uber.org/zap"
)

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(key string) (interface{}, bool)

	// Set 设置缓存值
	Set(key string, value interface{}, ttl time.Duration)

	// Delete 删除缓存值
	Delete(key string)

	// Clear 清空所有缓存
	Clear()

	// Size 获取缓存大小
	Size() int

	// Keys 获取所有键
	Keys() []string
}

// cacheItem 缓存项
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// isExpired 检查是否过期
func (item *cacheItem) isExpired() bool {
	if item.expiration.IsZero() {
		return false
	}
	return time.Now().After(item.expiration)
}

// memoryCache 内存缓存实现
type memoryCache struct {
	items map[string]*cacheItem
	mu    sync.RWMutex
	ctx   context.Context
	cancel context.CancelFunc
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache() Cache {
	ctx, cancel := context.WithCancel(context.Background())
	
	cache := &memoryCache{
		items:  make(map[string]*cacheItem),
		ctx:    ctx,
		cancel: cancel,
	}

	// 启动清理协程
	go cache.cleanupExpired()

	logger.Info("Memory cache initialized")
	return cache
}

// Get 获取缓存值
func (c *memoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if item.isExpired() {
		return nil, false
	}

	return item.value, true
}

// Set 设置缓存值
func (c *memoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.items[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}

	logger.Debug("Cache set",
		zap.String("key", key),
		zap.Duration("ttl", ttl))
}

// Delete 删除缓存值
func (c *memoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)

	logger.Debug("Cache deleted", zap.String("key", key))
}

// Clear 清空所有缓存
func (c *memoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)

	logger.Info("Cache cleared")
}

// Size 获取缓存大小
func (c *memoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Keys 获取所有键
func (c *memoryCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return keys
}

// cleanupExpired 清理过期项
func (c *memoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.removeExpired()
		}
	}
}

// removeExpired 移除过期项
func (c *memoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	for key, item := range c.items {
		if !item.expiration.IsZero() && now.After(item.expiration) {
			delete(c.items, key)
			removed++
		}
	}

	if removed > 0 {
		logger.Debug("Expired cache items removed", zap.Int("count", removed))
	}
}

// Stop 停止缓存（清理资源）
func (c *memoryCache) Stop() {
	c.cancel()
	c.Clear()
	logger.Info("Memory cache stopped")
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Size      int       `json:"size"`
	Keys      []string  `json:"keys"`
	Timestamp time.Time `json:"timestamp"`
}

// Stats 获取缓存统计信息
func (c *memoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return CacheStats{
		Size:      len(c.items),
		Keys:      keys,
		Timestamp: time.Now(),
	}
}
