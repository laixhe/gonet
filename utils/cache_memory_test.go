package utils

import (
	"testing"
	"time"
)

func TestCacheMemory_Basic(t *testing.T) {
	cm := NewCacheMemory(0) // 永不过期
	cm.Set(1, "hello")
	cm.Set(2, 42)

	// Get 存在
	if v := cm.Get(1); v != "hello" {
		t.Fatalf("Get(1) = %v, want hello", v)
	}
	if v := cm.Get(2); v != 42 {
		t.Fatalf("Get(2) = %v, want 42", v)
	}

	// Get 不存在
	if v := cm.Get(999); v != nil {
		t.Fatalf("Get(999) = %v, want nil", v)
	}

	// Len
	if cm.Len() != 2 {
		t.Fatalf("Len = %d, want 2", cm.Len())
	}

	// Overwrite
	cm.Set(1, "world")
	if v := cm.Get(1); v != "world" {
		t.Fatalf("Get(1) after overwrite = %v, want world", v)
	}
}

func TestCacheMemory_Expire(t *testing.T) {
	cm := NewCacheMemory(1) // 1 秒过期
	cm.Set(1, "data")

	// 立即读取应成功
	if v := cm.Get(1); v != "data" {
		t.Fatalf("Get(1) = %v, want data", v)
	}

	// 等待过期
	time.Sleep(1100 * time.Millisecond)

	// 过期后应返回 nil
	if v := cm.Get(1); v != nil {
		t.Fatalf("Get(1) after expire = %v, want nil", v)
	}

	// 过期后应被惰性删除
	if cm.Len() != 0 {
		t.Fatalf("Len after expire = %d, want 0", cm.Len())
	}
}

func TestCacheMemory_Del(t *testing.T) {
	cm := NewCacheMemory(0)
	cm.Set(1, "data")
	cm.Set(2, "data2")

	cm.Del(1)

	if v := cm.Get(1); v != nil {
		t.Fatalf("Get(1) after Del = %v, want nil", v)
	}
	if v := cm.Get(2); v != "data2" {
		t.Fatalf("Get(2) after Del(1) = %v, want data2", v)
	}
	if cm.Len() != 1 {
		t.Fatalf("Len = %d, want 1", cm.Len())
	}

	// 删除不存在的 key 不 panic
	cm.Del(999)
}

func TestCacheMemory_ZeroExpire(t *testing.T) {
	cm := NewCacheMemory(0) // 永不过期
	cm.Set(1, "data")

	time.Sleep(10 * time.Millisecond)
	if v := cm.Get(1); v != "data" {
		t.Fatalf("zero-expire cache lost data: got %v, want data", v)
	}
}
