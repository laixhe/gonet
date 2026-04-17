package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestCacheMemoryGet(t *testing.T) {
	key := 123
	cm := NewCacheMemory(5)
	cm.Set(key, "A")

	fmt.Println("1", cm.Get(key))
	time.Sleep(2 * time.Second)
	fmt.Println("2", cm.Get(key))
	time.Sleep(2 * time.Second)
	fmt.Println("3", cm.Get(key))
	time.Sleep(2 * time.Second)
	fmt.Println("4", cm.Get(key))
}
