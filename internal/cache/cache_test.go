package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// 属性 25：缓存一致性
// 验证需求：10.3
func TestCacheConsistencyProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("set then get returns same value", prop.ForAll(
		func(key string, value string) bool {
			if key == "" {
				key = "test"
			}

			cache := NewMemoryCache()
			defer cache.(*memoryCache).Stop()

			cache.Set(key, value, 1*time.Minute)
			retrieved, exists := cache.Get(key)

			return exists && retrieved == value
		},
		gen.AlphaString(),
		gen.AnyString(),
	))

	properties.Property("expired items are not retrievable", prop.ForAll(
		func(key string, value string) bool {
			if key == "" {
				key = "test"
			}

			cache := NewMemoryCache()
			defer cache.(*memoryCache).Stop()

			// 设置非常短的 TTL
			cache.Set(key, value, 1*time.Millisecond)

			// 等待过期
			time.Sleep(10 * time.Millisecond)

			_, exists := cache.Get(key)
			return !exists
		},
		gen.AlphaString(),
		gen.AnyString(),
	))

	properties.Property("delete removes item", prop.ForAll(
		func(key string, value string) bool {
			if key == "" {
				key = "test"
			}

			cache := NewMemoryCache()
			defer cache.(*memoryCache).Stop()

			cache.Set(key, value, 1*time.Minute)
			cache.Delete(key)
			_, exists := cache.Get(key)

			return !exists
		},
		gen.AlphaString(),
		gen.AnyString(),
	))

	properties.Property("clear removes all items", prop.ForAll(
		func(keys []string) bool {
			if len(keys) == 0 {
				keys = []string{"test"}
			}
			if len(keys) > 10 {
				keys = keys[:10]
			}

			cache := NewMemoryCache()
			defer cache.(*memoryCache).Stop()

			// 设置多个项
			for _, key := range keys {
				if key != "" {
					cache.Set(key, "value", 1*time.Minute)
				}
			}

			// 清空缓存
			cache.Clear()

			// 验证所有项都被删除
			return cache.Size() == 0
		},
		gen.SliceOf(gen.AlphaString()),
	))

	properties.Property("size reflects item count", prop.ForAll(
		func(count int) bool {
			if count < 0 || count > 20 {
				count = 5
			}

			cache := NewMemoryCache()
			defer cache.(*memoryCache).Stop()

			// 添加指定数量的项
			for i := 0; i < count; i++ {
				cache.Set(string(rune(i)), i, 1*time.Minute)
			}

			return cache.Size() == count
		},
		gen.IntRange(0, 20),
	))

	properties.Property("concurrent access is safe", prop.ForAll(
		func(operations int) bool {
			if operations < 1 || operations > 50 {
				operations = 10
			}

			cache := NewMemoryCache()
			defer cache.(*memoryCache).Stop()

			var wg sync.WaitGroup
			errors := make(chan bool, operations*3)

			// 并发读写
			for i := 0; i < operations; i++ {
				wg.Add(3)

				// 写
				go func(id int) {
					defer wg.Done()
					key := string(rune(id))
					cache.Set(key, id, 1*time.Minute)
					errors <- false
				}(i)

				// 读
				go func(id int) {
					defer wg.Done()
					key := string(rune(id))
					cache.Get(key)
					errors <- false
				}(i)

				// 删除
				go func(id int) {
					defer wg.Done()
					key := string(rune(id))
					cache.Delete(key)
					errors <- false
				}(i)
			}

			wg.Wait()
			close(errors)

			// 检查是否有错误
			for err := range errors {
				if err {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 50),
	))

	properties.Property("overwriting value updates cache", prop.ForAll(
		func(key string, value1 string, value2 string) bool {
			if key == "" {
				key = "test"
			}

			cache := NewMemoryCache()
			defer cache.(*memoryCache).Stop()

			// 设置第一个值
			cache.Set(key, value1, 1*time.Minute)

			// 覆盖为第二个值
			cache.Set(key, value2, 1*time.Minute)

			// 获取值
			retrieved, exists := cache.Get(key)

			return exists && retrieved == value2
		},
		gen.AlphaString(),
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestCacheBasic 测试基本功能
func TestCacheBasic(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	// 测试 Set 和 Get
	cache.Set("key1", "value1", 1*time.Minute)
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// 测试不存在的键
	_, exists = cache.Get("nonexistent")
	assert.False(t, exists)
}

// TestCacheExpiration 测试过期
func TestCacheExpiration(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	// 设置短 TTL
	cache.Set("key1", "value1", 50*time.Millisecond)

	// 立即获取应该成功
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// 等待过期
	time.Sleep(100 * time.Millisecond)

	// 过期后获取应该失败
	_, exists = cache.Get("key1")
	assert.False(t, exists)
}

// TestCacheDelete 测试删除
func TestCacheDelete(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Delete("key1")

	_, exists := cache.Get("key1")
	assert.False(t, exists)
}

// TestCacheClear 测试清空
func TestCacheClear(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	// 添加多个项
	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)
	cache.Set("key3", "value3", 1*time.Minute)

	assert.Equal(t, 3, cache.Size())

	// 清空
	cache.Clear()

	assert.Equal(t, 0, cache.Size())
}

// TestCacheSize 测试大小
func TestCacheSize(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	assert.Equal(t, 0, cache.Size())

	cache.Set("key1", "value1", 1*time.Minute)
	assert.Equal(t, 1, cache.Size())

	cache.Set("key2", "value2", 1*time.Minute)
	assert.Equal(t, 2, cache.Size())

	cache.Delete("key1")
	assert.Equal(t, 1, cache.Size())
}

// TestCacheKeys 测试获取所有键
func TestCacheKeys(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)
	cache.Set("key3", "value3", 1*time.Minute)

	keys := cache.Keys()
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
}

// TestCacheConcurrent 测试并发访问
func TestCacheConcurrent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	var wg sync.WaitGroup
	operations := 100

	// 并发写入
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := string(rune(id))
			cache.Set(key, id, 1*time.Minute)
		}(i)
	}

	// 并发读取
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := string(rune(id))
			cache.Get(key)
		}(i)
	}

	// 并发删除
	for i := 0; i < operations/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := string(rune(id))
			cache.Delete(key)
		}(i)
	}

	wg.Wait()

	// 验证缓存仍然可用
	cache.Set("final", "test", 1*time.Minute)
	value, exists := cache.Get("final")
	assert.True(t, exists)
	assert.Equal(t, "test", value)
}

// TestCacheOverwrite 测试覆盖值
func TestCacheOverwrite(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key1", "value2", 1*time.Minute)

	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value2", value)
}

// TestCacheNoTTL 测试无 TTL
func TestCacheNoTTL(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.(*memoryCache).Stop()

	// 设置无 TTL（永不过期）
	cache.Set("key1", "value1", 0)

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 应该仍然存在
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)
}

// TestCacheStats 测试统计信息
func TestCacheStats(t *testing.T) {
	cache := NewMemoryCache().(*memoryCache)
	defer cache.Stop()

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)

	stats := cache.Stats()
	assert.Equal(t, 2, stats.Size)
	assert.Len(t, stats.Keys, 2)
	assert.Contains(t, stats.Keys, "key1")
	assert.Contains(t, stats.Keys, "key2")
}
