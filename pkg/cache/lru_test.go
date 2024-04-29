package cache

import (
	"reflect"
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

func TestGetAllEmptyCache(t *testing.T) {
	lru := NewTestLRUCache(2)
	res := lru.GetAll()
	if len(res) != 0 {
		t.Errorf("Expected empty map, got %v", res)
	}
}

func TestGetAllPartiallyFilledCache(t *testing.T) {
	lru := NewTestLRUCache(5)
	lru.Set("a", "a")
	lru.Set("b", "b")
	expected := map[string]string{"a": "a", "b": "b"}
	result := lru.GetAll()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected map %v, got %v", expected, result)
	}
}

func TestGetAllFullCacheWithEvictions(t *testing.T) {
	lru := NewTestLRUCache(2)
	lru.Set("a", "a")
	lru.Set("b", "b")
	lru.Set("c", "c") // This should evict a
	expected := map[string]string{"b": "b", "c": "c"}
	result := lru.GetAll()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected map %v, got %v", expected, result)
	}
}

func TestGetAllConcurrency(t *testing.T) {
	lru := NewTestLRUCache(10)
	actions := 1000
	var wg sync.WaitGroup
	wg.Add(actions)
	for i := 0; i < actions; i++ {
		go func(i int) {
			defer wg.Done()
			key := string("key_" + strconv.Itoa(i%10))
			value := "value_" + strconv.Itoa(i)
			lru.Set(key, value)
		}(i)
	}
	wg.Wait()

	result := lru.GetAll()
	if len(result) != 10 {
		t.Errorf("Expected 10 entries in the map, got %d", len(result))
	}
}
