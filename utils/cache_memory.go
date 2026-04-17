package utils

import (
	"sync"
	"time"
)

// 简单内存缓存

type CacheMemoryData struct {
	lastTime time.Time // 最后一次写入时间
	data     any
}

type CacheMemory struct {
	rw          *sync.RWMutex
	sustainable time.Duration // 缓存持续时间
	memoryData  map[int]*CacheMemoryData
}

// NewCacheMemory -
// second 缓存持续时间(秒)
func NewCacheMemory(second int) *CacheMemory {
	if second <= 0 {
		return &CacheMemory{
			rw:         &sync.RWMutex{},
			memoryData: make(map[int]*CacheMemoryData),
		}
	}
	return &CacheMemory{
		rw:          &sync.RWMutex{},
		sustainable: time.Duration(second) * time.Second,
		memoryData:  make(map[int]*CacheMemoryData),
	}
}

func (cm *CacheMemory) Set(key int, value any) {
	cm.rw.Lock()
	defer cm.rw.Unlock()

	if _, ok := cm.memoryData[key]; ok {
		cm.memoryData[key].lastTime = time.Now()
		cm.memoryData[key].data = value
		return
	}
	cm.memoryData[key] = &CacheMemoryData{
		lastTime: time.Now(),
		data:     value,
	}
}

func (cm *CacheMemory) Get(key int) any {
	cm.rw.RLock()
	defer cm.rw.RUnlock()

	if _, ok := cm.memoryData[key]; ok {
		if cm.sustainable <= 0 {
			return cm.memoryData[key].data
		}
		remaining := time.Until(cm.memoryData[key].lastTime.Add(cm.sustainable))
		if remaining > 0 {
			return cm.memoryData[key].data
		}
	}
	return nil
}
