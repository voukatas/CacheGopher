package client

import (
	"testing"
)

func TestRealServerInteraction(t *testing.T) {

	// setup server
	listener, err := startTestServer(t)
	if err != nil {
		t.Fatal(err)
	}
	defer (*listener).Close()

	// setup client
	pool := NewConnPool(1, "localhost:12345")
	client, err := NewClient(pool)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test Ping
	if resp, err := client.Ping(); err != nil || resp != "PONG" {
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

}
