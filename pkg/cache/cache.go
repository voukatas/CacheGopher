package cache

import (
	"fmt"
	"strings"
)

type Cache interface {
	Set(key string, value string)
	Get(key string) (string, bool)
	Delete(key string) bool
	Flush()
	Keys() []string
	GetSnapshot() map[string]string
	Lock()
	Unlock()
}

func NewCache(cacheType string, capacity int) (Cache, error) {
	if capacity < 1 {
		return nil, fmt.Errorf("capacity should be more than 1")
	}

	var c Cache

	switch strings.ToUpper(cacheType) {
	case "LRU":
		c = NewLRUCache(capacity)
	default:
		return nil, fmt.Errorf("Unknown cache type: %s", cacheType)
	}

	return c, nil
}
