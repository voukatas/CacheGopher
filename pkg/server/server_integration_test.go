package server

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/logger"
	"github.com/voukatas/CacheGopher/pkg/replication"
)

func TestServerIntegration(t *testing.T) {
	logger := logger.SetupDebugLogger()
	localCache, err := cache.NewCache("LRU", 10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	mockReplicator := &replication.MockReplicator{}

	myServer := NewServer(localCache, logger, mockReplicator)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on TCP port: %v", err)
	}
	defer listener.Close()

	ready := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Logf("Failed to accept connection: %v", err)
			return
		}
		close(ready)
		myServer.HandleConnection(conn)
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer clientConn.Close()

	<-ready

	fmt.Fprintf(clientConn, "SET myKey myValue\n")

	reader := bufio.NewReader(clientConn)
	res, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	if string(res) != "OK" {
		t.Fatalf(`res= %q; want "OK"`, res)
	}
}
