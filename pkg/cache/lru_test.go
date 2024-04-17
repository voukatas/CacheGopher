package cache

import (
	"strconv"
	"sync"
	"testing"
)

func TestLRUEviction(t *testing.T) {
	lru := NewTestLRUCache(2)
	lru.Set("a", "a")
	lru.Set("b", "b")
	lru.Set("c", "c")

	if _, ok := lru.Get("a"); ok {
		t.Fatal("Expected a to be evicted")
	}

	if val, ok := lru.Get("c"); !ok || val != "c" {
		t.Fatalf("Expected c , got '%s'", val)
	}
}

func TestLRUCacheSetGetDelete(t *testing.T) {
	lru := NewTestLRUCache(2)
	key, value := "testKey", "testValue"
	key2, value2 := "testKey2", "testValue2"
	key3, value3 := "testKey3", "testValue3"
	key4, value4 := "testKey4", "testValue4"
	lru.Set(key, value)
	//lru.PrintLRU()
	//fmt.Println(lru.head.key == key)
	if lru.head.key != key {
		t.Fatalf(`lru head = %s; want %s`, lru.head.key, key)
	}
	lru.Set(key2, value2)
	//lru.PrintLRU()
	if lru.head.key != key2 {
		t.Fatalf(`lru head = %s; want %s`, lru.head.key, key2)
	}
	lru.Set(key2, "resetTestValue2")
	//lru.PrintLRU()
	if lru.head.key != key2 {
		t.Fatalf(`lru head = %s; want %s`, lru.head.key, key2)
	}
	if lru.head.value != "resetTestValue2" {
		t.Fatalf(`lru head value = %s; want %s`, lru.head.value, "resetTestValue2")
	}

	if v, ok := lru.Get(key); !ok || v != value {
		t.Fatalf(`lru.Get("%s", "%s") = %s; want %s`, key, value, v, value)
	}
	//lru.PrintLRU()

	if v, ok := lru.Get(key2); !ok || v != "resetTestValue2" {
		t.Fatalf(`lru.Get("%s", "%s") = %s; want %s`, key2, value2, v, "resetTestValue2")
	}
	//lru.PrintLRU()
	lru.Set(key3, value3)
	if lru.head.key != key3 {
		t.Fatalf(`lru head = %s; want %s`, lru.head.key, key3)
	}
	//lru.PrintLRU()

	lru.Delete(key3)
	if _, exists := lru.store[key3]; exists {
		t.Fatalf(`lru store key = %s shouldn't exist`, key3)
	}
	//lru.PrintLRU()

	lru.Set(key4, value4)
	if lru.head.key != key4 {
		t.Fatalf(`lru head = %s; want %s`, lru.head.key, key4)
	}
	//lru.PrintLRU()
	//fmt.Println("---------------here")
}

func TestFlushAndKeys(t *testing.T) {
	lru := NewTestLRUCache(3)
	lru.Set("a", "a")
	lru.Set("b", "b")
	lru.Set("c", "c")

	keysBefore := lru.Keys()
	if len(keysBefore) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keysBefore))
	}

	lru.Flush()
	keysAfter := lru.Keys()
	if len(keysAfter) != 0 {
		t.Errorf("Expected 0 keys after flush, got %d", len(keysAfter))
	}

	if lru.head != nil || lru.tail != nil {
		t.Errorf("Expected head and tail to be nil after flush")
	}
}

func TestLRUConcurrency(t *testing.T) {
	lru := NewTestLRUCache(26)
	var wg sync.WaitGroup
	actions := 100000

	for i := 0; i < actions; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string("key_" + strconv.Itoa(i%26))
			value := "value_" + strconv.Itoa(i)
			lru.Set(key, value)
			if _, ok := lru.Get(key); !ok {
				t.Errorf("Expected key '%s' to exist", key)
			}
		}(i)
	}

	wg.Wait()

}
