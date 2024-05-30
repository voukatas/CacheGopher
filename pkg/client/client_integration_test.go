package client

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voukatas/CacheGopher/pkg/config"
)

func TestRealServerInteraction(t *testing.T) {

	// setup server
	listener, err := startTestServer(t, 10, 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer (listener).Close()

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
	listener, err := startTestServer(t, 500, 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer (listener).Close()

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
	numGoroutines := 1

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

// A round robin selection is assumed so the next server is expected to be queried in each request
func TestRoundRobinInMultipleRealServersInteraction(t *testing.T) {

	// setup primary server
	// mapPrimary := map[string]string{
	// 	"hello": "world",
	// }
	listener1, err := startTestServer(t, 10, 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer (listener1).Close()

	// start secondary server 1
	mapSecondary1 := map[string]string{
		"testkey": "testvalue1",
	}
	listener2, err := startTestServer(t, 10, 12346, mapSecondary1)
	if err != nil {
		t.Fatal(err)
	}
	defer (listener2).Close()

	// start secondary server 2
	mapSecondary2 := map[string]string{
		"testkey": "testvalue2",
	}
	listener3, err := startTestServer(t, 10, 12347, mapSecondary2)
	if err != nil {
		t.Fatal(err)
	}
	defer (listener3).Close()

	ring := NewHashRing()
	balancers := map[string]*ReadBalancer{}

	// setup primary config
	pool := NewConnPool(1, "localhost:12345", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode := NewCacheNode("testPrimary", true, pool)

	ring.AddNode(newNode)
	newBalancer := NewReadBalancer()
	newBalancer.addCacheNode(newNode)
	balancers["testPrimary"] = newBalancer

	// setup secondary config
	pool2 := NewConnPool(1, "localhost:12346", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode2 := NewCacheNode("testSecondary1", true, pool2)
	readBalancer := balancers["testPrimary"]
	readBalancer.addCacheNode(newNode2)

	// setup secondary config
	pool3 := NewConnPool(1, "localhost:12347", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode3 := NewCacheNode("testSecondary2", true, pool3)
	//readBalancer := balancers["testPrimary"]
	readBalancer.addCacheNode(newNode3)

	client := &Client{
		ring:      ring,
		balancers: balancers,
	}

	// Test Set
	if resp, err := client.Set("testkey", "testvalue"); err != nil || resp != "OK" {
		t.Errorf("Set failed: resp=%s, err=%v", resp, err)
	}

	// Test Get, This should query the primary server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// Test Get, This should query the first secondary server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue1" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// Test Get, This should query the second secondary server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue2" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// Test Get, This should query the primary server again
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}

	// delete this
	// if resp, err := client.Get("testkey"); err != nil {
	// 	t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	// }

}

// A round robin selection is assumed so the next server is expected to be queried in each request
func TestRoundRobinInMultipleRealServersInteractionWithBlacklisting(t *testing.T) {

	// setup primary server
	// mapPrimary := map[string]string{
	// 	"hello": "world",
	// }
	listener1, err := startTestServer(t, 10, 12345, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer (listener1).Close()

	// start secondary server 1
	mapSecondary1 := map[string]string{
		"testkey": "testvalue1",
	}
	listener2, err := startTestServer(t, 10, 12346, mapSecondary1)
	if err != nil {
		t.Fatal(err)
	}
	defer (listener2).Close()

	//start secondary server 2
	mapSecondary2 := map[string]string{
		"testkey": "testvalue2",
	}
	listener3, err := startTestServer(t, 10, 12347, mapSecondary2)
	if err != nil {
		t.Fatal(err)
	}
	//defer (*listener3).Close()

	ring := NewHashRing()
	balancers := map[string]*ReadBalancer{}

	// setup primary config
	pool := NewConnPool(1, "localhost:12345", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode := NewCacheNode("testPrimary", true, pool)

	ring.AddNode(newNode)

	newBalancer := NewReadBalancer()
	newBalancer.addCacheNode(newNode)
	balancers["testPrimary"] = newBalancer

	// setup secondary config
	pool2 := NewConnPool(1, "localhost:12346", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode2 := NewCacheNode("testSecondary1", true, pool2)
	readBalancer := balancers["testPrimary"]
	readBalancer.addCacheNode(newNode2)

	// setup secondary config
	pool3 := NewConnPool(1, "localhost:12347", config.ClientConfig{ConnectionTimeout: 300, KeepAliveInterval: 15})
	newNode3 := NewCacheNode("testSecondary2", true, pool3)
	//readBalancer := balancers["testPrimary"]
	readBalancer.addCacheNode(newNode3)

	client := &Client{
		ring:      ring,
		balancers: balancers,
	}

	// Test Set
	if resp, err := client.Set("testkey", "testvalue"); err != nil || resp != "OK" {
		t.Errorf("Set failed: resp=%s, err=%v", resp, err)
	}

	// Test Get, This should query the primary server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// Test Get, This should query the first secondary server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue1" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}

	(listener3).Close()
	// Test Get, This should query the primary server since the secondary went down and the round robin will start again from the first server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// Test Get, This should query the first secondary server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue1" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// Test Get, This should query the primary again since the second secondary is still blacklisted
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// Test Get, This should query the first secondary server
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue1" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
	// add a delay so the blacklisted node is whitelisted again
	time.Sleep(31 * time.Second)
	// Test Get, This should try to query the second secondary first and since it is down the primary should queried again
	if resp, err := client.Get("testkey"); err != nil || resp != "testvalue" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}

}
