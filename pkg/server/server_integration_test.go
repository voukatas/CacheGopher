package server

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/config"
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

	myServer := NewServer(localCache, logger, mockReplicator, false, "")
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on TCP port: %v", err)
	}
	defer listener.Close()

	//ready := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Logf("Failed to accept connection: %v", err)
			return
		}
		//close(ready)
		myServer.HandleConnection(conn)
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer clientConn.Close()

	//<-ready

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

// ToDo: This thing needs refactor...
func TestKeyReplicationAndRecoveryForTheSecondary(t *testing.T) {
	log := logger.SetupDebugLogger()

	localCachePrimary, _ := cache.NewCache("LRU", 100)
	localCacheSecondary, _ := cache.NewCache("LRU", 100)
	primaryConfig := config.ServerConfig{ID: "primary", Address: "localhost:8000", Role: "PRIMARY", Secondaries: []string{"secondary"}}
	secondaryConfig := config.ServerConfig{ID: "secondary", Address: "localhost:8001", Role: "SECONDARY", Primary: "primary"}

	// start secondary server
	secondaryReplicator, _ := replication.NewReplicator("secondary", &config.Configuration{Servers: []config.ServerConfig{primaryConfig, secondaryConfig}}, log)
	secondaryServer := NewServer(localCacheSecondary, log, secondaryReplicator, false, "localhost:8000")

	secondaryListener, err := net.Listen("tcp", secondaryConfig.Address)
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
	}

	defer secondaryListener.Close()

	done := make(chan struct{})
	go func() {
		for {
			conn, err := secondaryListener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
				}
				continue
			}
			go secondaryServer.HandleConnection(conn)
		}
	}()

	// start primary server

	primaryReplicator, _ := replication.NewReplicator("primary", &config.Configuration{Servers: []config.ServerConfig{primaryConfig, secondaryConfig}}, log)

	primaryServer := NewServer(localCachePrimary, log, primaryReplicator, true, "")
	primaryListener, err := net.Listen("tcp", primaryConfig.Address)
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
	}

	defer primaryListener.Close()

	go func() {
		for {
			conn, err := primaryListener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
				}
				continue
			}
			go primaryServer.HandleConnection(conn)
		}
	}()

	clientConn, err := net.Dial("tcp", primaryConfig.Address)
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer clientConn.Close()

	// set value on primary
	fmt.Fprintf(clientConn, "SET myKey myValue\n")

	reader := bufio.NewReader(clientConn)
	res, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	if string(res) != "OK" {
		t.Fatalf(`res= %q; want "OK"`, res)
	}

	time.Sleep(1 * time.Second)

	// verify that the key-value was written on primary
	_, exists := primaryServer.cache.Get("myKey")
	//fmt.Println("exists? ", exists, v)

	// verify that the key-value was replicated on the secondary
	_, exists = secondaryServer.cache.Get("myKey")
	//fmt.Println("exists? ", exists)
	if !exists {
		t.Error("Secondary should have the key 'myKey'")
	}

	// delete the key from the secondary
	_ = secondaryServer.cache.Delete("myKey")

	// verify that it was deleted from the secondary
	_, exists = secondaryServer.cache.Get("myKey")
	fmt.Println("exists? ", exists)
	if exists {
		t.Error("Secondary should not have the key 'myKey'")
	}

	// trigger the recovery procedure on the secondary
	secondaryServer.HandleRecovery(secondaryConfig)
	time.Sleep(1 * time.Second)

	// verify that the key was recovered
	_, exists = secondaryServer.cache.Get("myKey")
	//fmt.Println("exists? ", exists)
	if !exists {
		t.Error("Secondary should have the key 'myKey'")
	}

	// test recovery of the primary
	// fmt.Println("Testing primary")
	//
	// _ = primaryServer.cache.Delete("myKey")
	// _, exists = primaryServer.cache.Get("myKey")
	// fmt.Println("exists? ", exists)
	// if exists {
	// 	t.Error("Primary should not have the key 'myKey'")
	// }
	//
	// primaryServer.HandleRecovery(primaryConfig)
	// time.Sleep(1 * time.Second)
	//
	// _, exists = primaryServer.cache.Get("myKey")
	// fmt.Println("exists? ", exists)
	// if !exists {
	// 	t.Error("Primary should have the key 'myKey'")
	// }

	// close the channel and the servers
	close(done)
}

func TestKeyReplicationAndRecoveryForThePrimary(t *testing.T) {
	log := logger.SetupDebugLogger()

	localCachePrimary, _ := cache.NewCache("LRU", 100)
	localCacheSecondary, _ := cache.NewCache("LRU", 100)
	primaryConfig := config.ServerConfig{ID: "primary", Address: "localhost:8000", Role: "PRIMARY", Secondaries: []string{"secondary"}}
	secondaryConfig := config.ServerConfig{ID: "secondary", Address: "localhost:8001", Role: "SECONDARY", Primary: "primary"}

	// start secondary server
	secondaryReplicator, _ := replication.NewReplicator("secondary", &config.Configuration{Servers: []config.ServerConfig{primaryConfig, secondaryConfig}}, log)
	secondaryServer := NewServer(localCacheSecondary, log, secondaryReplicator, false, "localhost:8000")

	secondaryListener, err := net.Listen("tcp", secondaryConfig.Address)
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
	}

	defer secondaryListener.Close()

	done := make(chan struct{})
	go func() {
		for {
			conn, err := secondaryListener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
				}
				continue
			}
			go secondaryServer.HandleConnection(conn)
		}
	}()

	// start primary server

	primaryReplicator, _ := replication.NewReplicator("primary", &config.Configuration{Servers: []config.ServerConfig{primaryConfig, secondaryConfig}}, log)

	primaryServer := NewServer(localCachePrimary, log, primaryReplicator, true, "")
	primaryListener, err := net.Listen("tcp", primaryConfig.Address)
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
	}

	defer primaryListener.Close()

	go func() {
		for {
			conn, err := primaryListener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
				}
				continue
			}
			go primaryServer.HandleConnection(conn)
		}
	}()

	clientConn, err := net.Dial("tcp", primaryConfig.Address)
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer clientConn.Close()

	// set value on primary
	fmt.Fprintf(clientConn, "SET myKey myValue\n")

	reader := bufio.NewReader(clientConn)
	res, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	if string(res) != "OK" {
		t.Fatalf(`res= %q; want "OK"`, res)
	}

	time.Sleep(1 * time.Second)

	// verify that the key-value was written on primary
	_, exists := primaryServer.cache.Get("myKey")
	//fmt.Println("exists? ", exists, v)

	// verify that the key-value was replicated on the secondary
	_, exists = secondaryServer.cache.Get("myKey")
	//fmt.Println("exists? ", exists)
	if !exists {
		t.Error("Secondary should have the key 'myKey'")
	}

	// test recovery of the primary
	fmt.Println("Testing primary")

	_ = primaryServer.cache.Delete("myKey")
	_, exists = primaryServer.cache.Get("myKey")
	fmt.Println("exists? ", exists)
	if exists {
		t.Error("Primary should not have the key 'myKey'")
	}

	primaryServer.HandleRecovery(primaryConfig)
	time.Sleep(1 * time.Second)

	_, exists = primaryServer.cache.Get("myKey")
	fmt.Println("exists? ", exists)
	if !exists {
		t.Error("Primary should have the key 'myKey'")
	}

	// close the channel and the servers
	close(done)
}
