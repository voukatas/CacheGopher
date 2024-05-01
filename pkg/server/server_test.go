package server

import (
	"bufio"
	"net"
	"testing"

	"github.com/voukatas/CacheGopher/pkg/replication"
)

type MockCache struct {
	SetCalled    bool
	GetCalled    bool
	DeleteCalled bool
	FlushCalled  bool
	KeysCalled   bool
}

func (m *MockCache) Set(key string, value string) {
	m.SetCalled = true
}

func (m *MockCache) Get(key string) (string, bool) {
	m.GetCalled = true
	return "value", true
}

func (m *MockCache) Delete(key string) bool {
	m.DeleteCalled = true
	return true
}

func (m *MockCache) Flush() {
	m.FlushCalled = true
}

func (m *MockCache) Keys() []string {
	m.KeysCalled = true
	return []string{"key1", "key2"}
}
func (m *MockCache) GetSnapshot() map[string]string {
	return map[string]string{}
}

func (lru *MockCache) Lock() {
}

func (lru *MockCache) Unlock() {
}

type MockLogger struct {
	DebugMessages []string
	ErrorMessages []string
}

func (m *MockLogger) Debug(msg string) { m.DebugMessages = append(m.DebugMessages, msg) }
func (m *MockLogger) Info(msg string)  {}
func (m *MockLogger) Warn(msg string)  {}
func (m *MockLogger) Error(msg string) { m.ErrorMessages = append(m.ErrorMessages, msg) }

func TestHandleConnection(t *testing.T) {
	mockCache := &MockCache{}
	mockLogger := &MockLogger{}
	mockReplicator := &replication.MockReplicator{}
	server := NewServer(mockCache, mockLogger, mockReplicator, false, "")

	// start a network pipe that is like having real network connections
	clientConn, serverConn := net.Pipe()
	//defer clientConn.Close()
	defer serverConn.Close()

	done := make(chan struct{})

	go func() {
		server.HandleConnection(serverConn)
		//fmt.Println("after HandleConnectioON")
		close(done)
	}()

	//fmt.Println("before write")
	// A small sleep to ensure the server started
	//time.Sleep(1)
	clientConn.Write([]byte("GET key\n"))
	scanner := bufio.NewScanner(clientConn)
	scanner.Scan()
	if scanner.Text() != "value" {
		t.Error("Expected on GET to receive value 'value'")

	}
	clientConn.Write([]byte("SET key value\n"))
	scanner.Scan()
	if scanner.Text() != "OK" {
		t.Error("Expected on SET to receive 'OK'")

	}
	//fmt.Println("after write 2:", scanner.Text())
	clientConn.Close() // signal EOF to close the connection on HandleConnection

	// this channel is needed to avoid race conditions
	<-done
	//fmt.Println("after done")

	if !mockCache.GetCalled {
		t.Error("Expected Get to be called")

	}
	if !mockCache.SetCalled {
		t.Error("Expected Set to be called")
	}

	if len(mockLogger.DebugMessages) == 0 {
		t.Error("Expected info messages to be logged")
	}
}
