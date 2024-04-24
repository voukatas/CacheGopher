package client

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func startMockServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleMockConnection(conn)
	}
}

func handleMockConnection(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		switch {
		case text == "PING":
			fmt.Fprintln(conn, "PONG")
		case strings.HasPrefix(text, "SET"):
			fmt.Fprintln(conn, "OK")
		case strings.HasPrefix(text, "GET"):
			fmt.Fprintln(conn, "VALUE")
		case strings.HasPrefix(text, "DELETE"):
			fmt.Fprintln(conn, "OK")
		default:
			fmt.Fprintln(conn, "ERROR")
		}
	}
	conn.Close()
}

func TestClient(t *testing.T) {
	go startMockServer("3333")
	time.Sleep(time.Second)

	pool := NewConnPool(1, "localhost:3333")
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

	//client, err := NewClient(true)
	// if err != nil {
	// 	t.Fatalf("Failed to create client: %v", err)
	// }

	// Test Ping
	if resp, err := client.Ping(newNode); err != nil || resp != "PONG" {
		t.Errorf("Ping failed: resp=%s, err=%v", resp, err)
	}

	// Test Set
	if resp, err := client.Set("key", "VALUE"); err != nil || resp != "OK" {
		t.Errorf("Set failed: resp=%s, err=%v", resp, err)
	}

	// Test Get
	if resp, err := client.Get("key"); err != nil || resp != "VALUE" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}

	// Test Delete
	if resp, err := client.Delete("key"); err != nil || resp != "OK" {
		t.Errorf("Delete failed: resp=%s, err=%v", resp, err)
	}

	if resp, err := client.Get("key"); err != nil || resp != "VALUE" {
		t.Errorf("Get failed: resp=%s, err=%v", resp, err)
	}
}

func TestValidateCommand(t *testing.T) {
	const maxTokenSize = 64 * 1024

	maxCommand := make([]byte, maxTokenSize)
	for i := range maxCommand {
		maxCommand[i] = 'a'
	}
	maxCommand[len(maxCommand)-1] = '\n'

	tests := []struct {
		name      string
		cmdBytes  []byte
		wantError bool
		errMsg    string
	}{
		{
			name:      "Exact Size Limit",
			cmdBytes:  maxCommand,
			wantError: false,
		},
		{
			name:      "Below Size Limit",
			cmdBytes:  []byte("SET key value" + "\n"),
			wantError: false,
		},
		{
			name:      "Above Size Limit",
			cmdBytes:  append(maxCommand, 'b'),
			wantError: true,
			errMsg:    "command exceeds the maximum allowed size of 64KB",
		},
		{
			name:      "Contains Illegal Newline",
			cmdBytes:  []byte("SET key value\nanother"),
			wantError: true,
			errMsg:    "command cannot contain newline characters",
		},
		{
			name:      "Can Contain Escaped Illegal Newline",
			cmdBytes:  []byte("SET key value\\nanother"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.cmdBytes)
			if (err != nil) != tt.wantError {
				t.Errorf("validateCommand() error = %v, wantError %v", err, tt.wantError)
			} else if err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateCommand() error = %v, want errMsg %v", err.Error(), tt.errMsg)
			}
		})
	}
}
