package cache

import (
	"fmt"
	"sync"
)

type CacheItem struct {
	key   string
	value string
	prev  *CacheItem
	next  *CacheItem
}

func NewCacheItem(key string, value string) *CacheItem {
	return &CacheItem{
		key:   key,
		value: value,
		prev:  nil,
		next:  nil,
	}

}

type LRUCache struct {
	store    map[string]*CacheItem
	capacity int
	head     *CacheItem
	tail     *CacheItem
	lock     sync.RWMutex
	//logger   logger.Logger
}

func NewLRUCache(capacity int) Cache {
	return &LRUCache{
		store:    make(map[string]*CacheItem, capacity),
		capacity: capacity,
		head:     nil,
		tail:     nil,
	}
}

// removeItemFromQ
// Note: This method does not handle synchronization and expects the caller to manage locking
func (lru *LRUCache) removeItemFromQ(item *CacheItem) {

	//fmt.Println("removeItemFromQ")
	if item.prev != nil {
		item.prev.next = item.next
	} else {

		lru.head = item.next
	}

	if item.next != nil {

		item.next.prev = item.prev
	} else {

		lru.tail = item.prev
	}

	item.prev = nil
	item.next = nil

}

// addItemToFrontOfQ
// Note: This method does not handle synchronization and expects the caller to manage locking
func (lru *LRUCache) addItemToFrontOfQ(item *CacheItem) {
	//fmt.Println("addItemToFrontOfQ")
	if lru.head == nil {
		lru.head = item
		lru.tail = item
		return
	}

	lru.head.prev = item
	item.next = lru.head
	lru.head = item

}

// moveToFrontOfQ
// Note: This method does not handle synchronization and expects the caller to manage locking
func (lru *LRUCache) moveToFrontOfQ(item *CacheItem) {
	//fmt.Println("moveToFrontOfQ")
	if lru.head == item {
		//fmt.Println("moveToFrontOfQ item is already head, return")
		return
	}

	lru.removeItemFromQ(item)
	lru.addItemToFrontOfQ(item)
}

// Set
func (lru *LRUCache) Set(key string, value string) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if item, exists := lru.store[key]; exists {
		//fmt.Println("SET item exists")
		lru.moveToFrontOfQ(item)
		item.value = value
		return

	}

	//fmt.Println("SET item doesn't exists")
	newItem := NewCacheItem(key, value)
	if len(lru.store) >= lru.capacity {
		//fmt.Println("SET item capacity reached, evict")
		// evict the tail
		delete(lru.store, lru.tail.key)
		lru.removeItemFromQ(lru.tail)
	}

	lru.store[key] = newItem
	lru.addItemToFrontOfQ(newItem)

}

func (lru *LRUCache) Lock() {
	lru.lock.Lock()
}

func (lru *LRUCache) Unlock() {
	lru.lock.Unlock()
}

// Get
func (lru *LRUCache) Get(key string) (string, bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if item, exists := lru.store[key]; exists {
		//fmt.Println("GET key found")
		lru.moveToFrontOfQ(item)
		return item.value, true
	}

	//fmt.Println("GET key wasn't found")
	return "", false

}

// GetSnapshot key-value records
func (lru *LRUCache) GetSnapshot() map[string]string {
	lru.lock.RLock()
	defer lru.lock.RUnlock()

	keyValMap := make(map[string]string, 0)

	for k, v := range lru.store {
		keyValMap[k] = v.value
	}

	return keyValMap
}

// Delete
func (lru *LRUCache) Delete(key string) bool {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	item, exists := lru.store[key]
	if !exists {
		return false
	}

	delete(lru.store, key)
	lru.removeItemFromQ(item)

	return true

}

// Flush
func (lru *LRUCache) Flush() {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	// This might prevent potential memory leaks but it will slow down signifigantly the performance. Tradeoffs... consider a revisit on this
	current := lru.head
	for current != nil {
		next := current.next
		current.prev = nil
		current.next = nil
		current = next
	}
	// end of consideration

	lru.store = make(map[string]*CacheItem) // Reinitialize the map
	lru.head = nil
	lru.tail = nil
}

// Keys
func (lru *LRUCache) Keys() []string {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	keys := make([]string, 0, len(lru.store))
	for key := range lru.store {
		keys = append(keys, key)
	}
	return keys
}

// func (lru *LRUCache) SetLogger(logger logger.Logger) {
// 	lru.logger = logger
// }
//
// func (lru *LRUCache) GetLogger() logger.Logger {
// 	return lru.logger
// }

func (lru *LRUCache) PrintLRU() {
	fmt.Println("LRU Q contents: ")
	currentCacheItem := lru.head
	for currentCacheItem != nil {
		fmt.Println(currentCacheItem.key)
		currentCacheItem = currentCacheItem.next
	}

	fmt.Printf("LRU store contents: %v\n", lru.store)
}
