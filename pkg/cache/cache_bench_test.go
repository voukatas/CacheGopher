package cache

import (
	"strconv"
	"testing"
)

func BenchmarkCacheSet(b *testing.B) {
	c := NewTestCache()
	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := NewTestCache()

	for i := 0; i < 100; i++ {
		key := "key" + strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i%100)
		c.Get(key)

	}
}

func BenchmarkCacheDelete(b *testing.B) {
	c := NewTestCache()
	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		c.Delete(key)
	}
}
