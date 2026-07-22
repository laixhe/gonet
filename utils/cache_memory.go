package utils

import (
	"sync"
	"time"
)

// CacheMemory 带过期时间的简单内存缓存
// 缓存项在过期后不会立即删除，而是在 Get 时惰性清理并在返回 nil 时删除
type CacheMemory struct {
	mu          sync.RWMutex
	sustainable time.Duration // 缓存过期时间，0 表示永不过期
	memoryData  map[int]*CacheMemoryData
}

// CacheMemoryData 缓存数据项
type CacheMemoryData struct {
	lastTime time.Time // 最后一次写入时间
	data     any
}

// NewCacheMemory 创建内存缓存
// second 缓存过期时间（秒），<= 0 表示永不过期
func NewCacheMemory(second int) *CacheMemory {
	cm := &CacheMemory{
		memoryData: make(map[int]*CacheMemoryData),
	}
	if second > 0 {
		cm.sustainable = time.Duration(second) * time.Second
	}
	return cm
}

// Set 写入缓存
func (cm *CacheMemory) Set(key int, value any) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.memoryData[key] = &CacheMemoryData{
		lastTime: time.Now(),
		data:     value,
	}
}

// Get 读取缓存，若过期或不存在返回 nil
func (cm *CacheMemory) Get(key int) any {
	cm.mu.RLock()
	entry, ok := cm.memoryData[key]
	cm.mu.RUnlock()

	if !ok {
		return nil
	}

	// 永不过期
	if cm.sustainable <= 0 {
		return entry.data
	}

	// 检查是否过期
	if time.Since(entry.lastTime) < cm.sustainable {
		return entry.data
	}

	// 惰性删除过期项
	cm.mu.Lock()
	// 双重检查：可能在等锁期间被其他 goroutine 刷新
	if e, ok := cm.memoryData[key]; ok && time.Since(e.lastTime) >= cm.sustainable {
		delete(cm.memoryData, key)
	}
	cm.mu.Unlock()

	return nil
}

// Del 删除缓存项
func (cm *CacheMemory) Del(key int) {
	cm.mu.Lock()
	delete(cm.memoryData, key)
	cm.mu.Unlock()
}

// Len 返回当前缓存项数量（含尚未被 Get 惰性清理的过期项）
func (cm *CacheMemory) Len() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.memoryData)
}
