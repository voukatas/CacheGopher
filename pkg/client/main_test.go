// main_test.go
package client

import (
	"fmt"
	"net"
	"os"
	"sync"
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

type TestServer struct {
	listener  net.Listener
	myServer  *server.Server
	address   string
	stopChan  chan struct{}
	stopMutex sync.Mutex
	running   bool
}

func startTestServer(t *testing.T, cap int, port int, cacheValues map[string]string) (*TestServer, error) {
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

	ts := &TestServer{
		listener: listener,
		myServer: myServer,
		address:  address,
		stopChan: make(chan struct{}),
		running:  true,
	}

	go ts.acceptConnections(t)

	// go func() {
	// 	defer listener.Close()
	// 	for {
	// 		conn, err := listener.Accept()
	// 		if err != nil {
	// 			t.Log("Server stopped accepting connections")
	// 			return
	// 		}
	// 		t.Logf("--------- Server %d will handle the connection", port)
	// 		go myServer.HandleConnection(conn)
	// 	}
	// }()

	// small delay
	time.Sleep(100)
	return ts, nil
	//return listener, nil
}

func (ts *TestServer) acceptConnections(t *testing.T) {
	for {
		select {
		case <-ts.stopChan:
			return
		default:
			conn, err := ts.listener.Accept()
			if err != nil {
				t.Log("Server stopped accepting connections")
				return
			}
			t.Logf("--------- Server %s will handle the connection", ts.address)
			go ts.myServer.HandleConnection(conn)
		}
	}
}

func (ts *TestServer) Stop() {
	ts.stopMutex.Lock()
	defer ts.stopMutex.Unlock()

	if ts.running {
		close(ts.stopChan)
		ts.listener.Close()
		ts.running = false
	}
}

func (ts *TestServer) Resume(t *testing.T) error {
	ts.stopMutex.Lock()
	defer ts.stopMutex.Unlock()

	if !ts.running {
		t.Logf("------------------- server resumed:%s ", ts.address)
		listener, err := net.Listen("tcp", ts.address)
		if err != nil {
			return fmt.Errorf("failed to resume server: %w", err)
		}
		ts.listener = listener
		ts.stopChan = make(chan struct{})
		ts.running = true
		go ts.acceptConnections(t)
	}
	return nil
}
