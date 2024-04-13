// main_test.go
package client

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}

func startTestServer(t *testing.T) (*net.Listener, error) {
	tlogger := logger.SetupTestLogger()
	localCache := cache.NewCache(tlogger)

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
