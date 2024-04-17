package cache

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}

func NewTestCache(cap int) Cache {

	//tlogger := logger.SetupDebugLogger()
	cache, err := NewCache("LRU", cap)
	if err != nil {
		fmt.Println("failed to start cache: ", err.Error())
		os.Exit(1)
	}
	return cache
}

func NewTestLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		store:    make(map[string]*CacheItem, capacity),
		capacity: capacity,
		head:     nil,
		tail:     nil,
	}
}
