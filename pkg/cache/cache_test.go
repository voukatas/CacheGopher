package cache

import (
	"strconv"
	"sync"
	"testing"
)

func TestCacheSetAndGet(t *testing.T) {
	cache := NewTestCache()
	key, value := "testKey", "testValue"
	cache.Set(key, value)

	if v, ok := cache.Get(key); !ok || v != value {
		t.Fatalf(`Cache.Set("%s", "%s") = %s; want %s`, key, value, v, value)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewTestCache()
	key := "testKey"
	cache.Set(key, "value")
	cache.Delete(key)

	if _, ok := cache.Get(key); ok {
		t.Fatalf(`Cache.Delete("%s") failed; key still exists`, key)
	}
}

func TestCacheFlush(t *testing.T) {
	cache := NewTestCache()
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Flush()

	if len(cache.Keys()) != 0 {
		t.Fatal(`Cache.Flush() failed; cache is not empty`)
	}
}

func TestCacheKeys(t *testing.T) {
	cache := NewTestCache()
	keys := []string{"key1", "key2", "key3"}

	for _, key := range keys {
		cache.Set(key, "value")
	}

	resultKeys := cache.Keys()
	if len(resultKeys) != len(keys) {
		t.Fatalf(`Cache.Keys() returned %d keys; want %d`, len(resultKeys), len(keys))
	}

	keyMap := make(map[string]bool)
	for _, k := range resultKeys {
		keyMap[k] = true
	}

	for _, k := range keys {
		if !keyMap[k] {
			t.Fatalf(`Cache.Keys() missing key "%s"`, k)
		}
	}
}

func TestCacheConcurrentSetAndGet(t *testing.T) {
	cache := NewTestCache()
	var wg sync.WaitGroup
	itemCount := 100
	keyBase := "key"
	valueBase := "value"

	for i := 0; i < itemCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Set(keyBase+strconv.Itoa(i), valueBase+strconv.Itoa(i))
		}(i)

	}

	wg.Wait()

	for i := 0; i < itemCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := keyBase + strconv.Itoa(i)
			expectedValue := valueBase + strconv.Itoa(i)

			if v, ok := cache.Get(key); !ok || v != expectedValue {
				t.Errorf("Cache.Get(%s) = %s; want %s", key, v, expectedValue)
			}
		}(i)
	}

	wg.Wait()

}
