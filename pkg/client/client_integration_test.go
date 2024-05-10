package client

import (
	"strings"
	"testing"

	"github.com/voukatas/CacheGopher/pkg/config"
)

func TestRealServerInteraction(t *testing.T) {

	// setup server
	listener, err := startTestServer(t)
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
