package client

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type ConnPool struct {
	pool    chan net.Conn
	address string
	// size    int
}

func NewConnPool(size int, address string) *ConnPool {
	return &ConnPool{
		pool:    make(chan net.Conn, size),
		address: address,
		// size:    size,
	}
}

func (cp *ConnPool) Get() (net.Conn, error) {

	select {
	case conn := <-cp.pool:
		return conn, nil

	default:

		return cp.dialWithBackOff()
	}
}

func (cp *ConnPool) dialWithBackOff() (net.Conn, error) {
	maxAttempts := 3
	baseTime := 100 * time.Millisecond
	maxBackoff := 1 * time.Second

	var conn net.Conn
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {

		conn, err = net.Dial("tcp", cp.address)

		if err == nil {
			return conn, nil
		}

		// consider adding a more sophisticated jitter approach
		jitter := time.Duration(rand.Int63n(100)) * time.Millisecond
		delay := time.Duration(1<<attempt)*baseTime + jitter

		if delay > maxBackoff {
			delay = maxBackoff
		}

		time.Sleep(delay)
	}

	return nil, err

}

func (cp *ConnPool) Return(conn net.Conn) error {
	// fmt.Println("Return called")
	// if _, err := conn.Write([]byte("PING")); err != nil {
	// 	conn.Close()
	// 	return err
	// }

	select {
	case cp.pool <- conn:
	// return the connection to the pool
	default:
		// pool is full so drop it
		conn.Close()
	}

	return nil
}

type Client struct {
	pool *ConnPool
	// conn    net.Conn
	// scanner *bufio.Scanner
	// lock    sync.RWMutex
}

func NewClient(pool *ConnPool) (*Client, error) {
	// conn, err := net.Dial("tcp", address)
	//
	// if err != nil {
	// 	return nil, err
	// }

	return &Client{
		pool: pool,
		// conn:    conn,
		// scanner: bufio.NewScanner(conn),
		// lock:    sync.RWMutex{},
	}, nil
}

// func (c *Client) Close() {
// 	c.conn.Close()
// }

func sendCommand(c *Client, cmd string) (string, error) {
	attempts := 2
	var conn net.Conn
	var err error

	for attempts > 0 {
		conn, err = c.pool.Get()
		if err != nil {
			return "", err
		}

		cmdBytes := []byte(cmd + "\n")
		_, err = conn.Write(cmdBytes)

		if err != nil {
			conn.Close()
			attempts--
			if attempts <= 0 {
				return "", err
			}

			continue

		}

		break
	}
	defer c.pool.Return(conn)

	scanner := bufio.NewScanner(conn)

	if scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no response")
}

func (c *Client) Set(k, v string) (string, error) {
	// c.lock.Lock()
	// defer c.lock.Unlock()
	cmd := fmt.Sprintf("SET %s %s", k, v)
	resp, err := sendCommand(c, cmd)

	return resp, err
}

func (c *Client) Get(k string) (string, error) {
	// c.lock.RLock()
	// defer c.lock.RUnlock()
	cmd := fmt.Sprintf("GET %s", k)
	res, err := sendCommand(c, cmd)

	return res, err

}
