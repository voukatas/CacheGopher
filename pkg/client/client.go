package client

import (
	"bufio"
	"fmt"
	"net"
)

type ConnPool struct {
	pool    chan net.Conn
	address string
	size    int
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

		conn, err := net.Dial("tcp", cp.address)

		if err != nil {
			return nil, err
		}
		return conn, nil
	}
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
	conn, err := c.pool.Get()
	if err != nil {
		return "", err
	}

	cmdBytes := []byte(cmd + "\n")
	_, err = conn.Write(cmdBytes)
	if err != nil {
		conn.Close()
		return "", err
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
