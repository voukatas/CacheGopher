// main_test.go
package client

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/logger"
	"github.com/voukatas/CacheGopher/pkg/replication"
	"github.com/voukatas/CacheGopher/pkg/server"
)

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}

func startTestServer(t *testing.T, cap int, port int, cacheValues map[string]string) (net.Listener, error) {
	tlogger := logger.SetupDebugLogger()
	localCache, err := cache.NewCache("LRU", cap)
	if err != nil {
		fmt.Println("failed to start cache: ", err.Error())
		os.Exit(1)
	}
	// inject values in cache
	if cacheValues != nil {
		for k, v := range cacheValues {
			localCache.Set(k, v)
		}
	}
	fmt.Println(localCache.GetSnapshot())
	mockReplicator := &replication.MockReplicator{}
	// Create myServer
	myServer := server.NewServer(localCache, tlogger, mockReplicator, false, "")

	address := fmt.Sprintf("localhost:%d", port)

	listener, err := net.Listen("tcp", address)
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
			t.Logf("--------- Server %d will handle the connection", port)
			go myServer.HandleConnection(conn)
		}
	}()

	// small delay
	time.Sleep(time.Second)
	return listener, nil
}
