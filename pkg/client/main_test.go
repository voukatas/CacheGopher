// main_test.go
package client

import (
	"github.com/voukatas/CacheGopher/pkg/cache"
	"net"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}

func startTestServer(t *testing.T) (*net.Listener, error) {
	localCache := cache.NewCache()
	listener, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		t.Fatalf("Failed to start server: %s", err)
		return nil, err
	}

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log("Server stopped accepting connections")
				return
			}
			go cache.HandleConnection(conn, localCache)
		}
	}()

	// small delay
	time.Sleep(time.Second)
	return &listener, nil
}

func TestPing(t *testing.T) {
	listener, err := startTestServer(t)
	if err != nil {
		t.Fatal(err)
	}
	defer (*listener).Close()

	// Your client setup and test execution logic here
}
