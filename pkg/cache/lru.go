package cache

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/voukatas/CacheGopher/pkg/logger"
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
	lock     sync.Mutex
	logger   logger.Logger
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

	// This might prevent potential memory leaks but it will slow down signifigantly the performance. Tradeoffs... consider revisit this
	// current := lru.head
	// for current != nil {
	// 	next := current.next
	// 	current.prev = nil
	// 	current.next = nil
	// 	current = next
	// }

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

// Also an implementation of doubly linked list from the std lib. I need to investigate its performance since it is battle tested and probably much more tested than mine implementation

// Update: Seems that mine implementation is faster in Get, Set methods but slower in Delete...std lib seems to avoid checking/setting the head/tail with their implementation so it is faster... Nevermind, I'll keep using mine

// Also needs testing...
type LRUCache2 struct {
	capacity int
	store    map[string]*list.Element
	queue    *list.List
	lock     sync.Mutex
	logger   logger.Logger
}

type entry struct {
	key   string
	value string
}

func NewLRUCache2(capacity int) Cache {
	return &LRUCache2{
		capacity: capacity,
		store:    make(map[string]*list.Element),
		queue:    list.New(),
	}
}

func (lru *LRUCache2) Get(key string) (string, bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if elem, found := lru.store[key]; found {
		lru.queue.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return "", false
}

func (lru *LRUCache2) Set(key string, value string) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if elem, found := lru.store[key]; found {

		lru.queue.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	if lru.queue.Len() == lru.capacity {
		oldest := lru.queue.Back()
		if oldest != nil {
			lru.queue.Remove(oldest)
			delete(lru.store, oldest.Value.(*entry).key)
		}
	}

	newElem := lru.queue.PushFront(&entry{key, value})
	lru.store[key] = newElem
}

func (lru *LRUCache2) Delete(key string) bool {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if elem, found := lru.store[key]; found {
		lru.queue.Remove(elem)
		delete(lru.store, key)
		return true
	}

	return false
}

func (c *LRUCache2) Flush() {
	c.queue.Init()
	c.store = make(map[string]*list.Element)
}

func (c *LRUCache2) Keys() []string {
	keys := make([]string, 0, len(c.store))
	for key := range c.store {
		keys = append(keys, key)
	}
	return keys
}

// func (lru *LRUCache2) SetLogger(logger logger.Logger) {
// 	lru.logger = logger
// }
// func (lru *LRUCache2) GetLogger() logger.Logger {
// 	return lru.logger
// }
