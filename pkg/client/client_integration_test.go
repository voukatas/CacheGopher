package client

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/voukatas/CacheGopher/pkg/config"
)

func TestRealServerInteraction(t *testing.T) {

	// setup server
	listener, err := startTestServer(t, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer (*listener).Close()

	// setup client
	// pool := NewConnPool(1, "localhost:12345")
	// client, err := NewClient(pool, true)
	// if err != nil {
	// 	t.Fatalf("Failed to create client: %v", err)
	// }

	pool := NewConnPool(1, "localhost:12345", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode := NewCacheNode("testNode", true, pool)

	ring := NewHashRing()
	balancers := map[string]*ReadBalancer{}

	ring.AddNode(newNode)
	newBalancer := NewReadBalancer()
	newBalancer.addCacheNode(newNode)
	balancers["testNode"] = newBalancer

	client := &Client{
		ring:      ring,
		balancers: balancers,
	}
	// Test Ping
	if resp, err := client.Ping(newNode); err != nil || resp != "PONG" {
		t.Errorf("Ping failed: resp=%s, err=%v", resp, err)
	}

	// Test Set
	if resp, err := client.Set("testkey", "testvalue"); err != nil || resp != "OK" {
		t.Errorf("Set failed: resp=%s, err=%v", resp, err)
	}

	// Test Get
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}

	// Test Delete
	if resp, err := client.Delete("testkey"); err != nil || resp != "OK" {
		t.Errorf("Delete failed: resp=%s, err=%v", resp, err)
	}

	if resp, err := client.Get("testkey"); err == nil || !strings.Contains(err.Error(), "Key not found") {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}

}

func TestRealServerInteractionConcurrently(t *testing.T) {

	// setup server
	listener, err := startTestServer(t, 500)
	if err != nil {
		t.Fatal(err)
	}
	defer (*listener).Close()

	pool := NewConnPool(5, "localhost:12345", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode := NewCacheNode("testNode", true, pool)

	ring := NewHashRing()
	balancers := map[string]*ReadBalancer{}

	ring.AddNode(newNode)
	newBalancer := NewReadBalancer()
	newBalancer.addCacheNode(newNode)
	balancers["testNode"] = newBalancer

	client := &Client{
		ring:      ring,
		balancers: balancers,
	}
	// Test Ping
	if resp, err := client.Ping(newNode); err != nil || resp != "PONG" {
		t.Errorf("Ping failed: resp=%s, err=%v", resp, err)
	}

	var wg sync.WaitGroup
	numGoroutines := 500

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("testkey%d", id)
			value := fmt.Sprintf("testvalue%d", id)

			if resp, err := client.Set(key, value); err != nil || resp != "OK" {
				t.Errorf("Set failed: key=%s, resp=%s, err=%v", key, resp, err)
			}
			// else {
			// 	t.Logf("Set done: key=%s, resp=%s, err=%v", key, resp, err)
			// }

			if resp, err := client.Get(key); err != nil || resp != value {
				t.Errorf("Get failed: key=%s, expected=%s, got=%s, err=%v", key, value, resp, err)
			}

			if resp, err := client.Delete(key); err != nil || resp != "OK" {
				t.Errorf("Delete failed: key=%s, resp=%s, err=%v", key, resp, err)
			}

			if resp, err := client.Get(key); err == nil || !strings.Contains(err.Error(), "Key not found") {
				t.Errorf("Post-Delete Get should fail: key=%s, resp=%s, err=%v", key, resp, err)
			}
		}(i)
	}

	wg.Wait()
}
