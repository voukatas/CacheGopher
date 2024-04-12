package client

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type PoolConn struct {
	conn      net.Conn
	scanner   *bufio.Scanner
	createdAt time.Time
}

type ConnPool struct {
	pool    chan *PoolConn
	address string
	// size    int
}

func (pc *PoolConn) isExpired() bool {
	const maxValidTime = 5 * time.Minute
	return time.Since(pc.createdAt) > maxValidTime
}

func (pc *PoolConn) isValid() bool {
	// can be set from config if we need to use this fc

	_, err := pc.conn.Write([]byte("PING\n"))
	if err != nil {
		fmt.Println("isValid is false 1")
		return false
	}

	if ok := pc.scanner.Scan(); !ok {
		if err := pc.scanner.Err(); err != nil {
			fmt.Println("Scanner error on reading:", err)
			return false
		}
		fmt.Println("isValid is false 2: No data read")
		return false
	}

	if pc.scanner.Text() != "PONG" {
		fmt.Println("isValid is false 3: Unexpected response", pc.scanner.Text())
		return false
	}

	fmt.Println("isValid is true")
	return true

}

func (pc *PoolConn) Close() {
	pc.conn.Close()
}

func NewConnPool(size int, address string) *ConnPool {
	return &ConnPool{
		pool:    make(chan *PoolConn, size),
		address: address,
		// size:    size,
	}
}

func (cp *ConnPool) Get() (*PoolConn, error) {
	fmt.Println("Get called")

	for {
		select {
		case poolConn := <-cp.pool:
			if poolConn.isExpired() || !poolConn.isValid() {
				fmt.Println("isExpired or is invalid")
				poolConn.Close()
				continue
			}
			fmt.Println("found poolConn")
			return poolConn, nil

		default:

			return cp.dialWithBackOff()
		}
	}
}

func (cp *ConnPool) dialWithBackOff() (*PoolConn, error) {
	fmt.Println("dialWithBackOff")
	maxAttempts := 3
	baseTime := 100 * time.Millisecond
	maxBackoff := 1 * time.Second

	var conn net.Conn
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {

		conn, err = net.Dial("tcp", cp.address)

		if err == nil {
			poolConn := &PoolConn{conn: conn, scanner: bufio.NewScanner(conn), createdAt: time.Now()}
			fmt.Println("Successfully Created poolConn")
			return poolConn, nil
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

func (cp *ConnPool) Return(poolConn *PoolConn) error {
	// fmt.Println("Return called")
	// if _, err := conn.Write([]byte("PING")); err != nil {
	// 	conn.Close()
	// 	return err
	// }

	select {
	case cp.pool <- poolConn:
	// return the connection to the pool
	default:
		// pool is full so drop it
		poolConn.Close()
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

func validateCommand(cmdBytes []byte) error {
	const maxTokenSize = 64 * 1024

	if len(cmdBytes) > maxTokenSize {
		fmt.Println("data exceeds the maximum allowed size of 64KB")
		return fmt.Errorf("command exceeds the maximum allowed size of 64KB")

	}

	if bytes.Contains(cmdBytes[:len(cmdBytes)-1], []byte("\n")) {
		fmt.Println("Invalid Data, data can't contain the newline chars")
		return fmt.Errorf("command cannot contain newline characters")
	}

	return nil
}

func sendCommand(c *Client, cmd string) (string, error) {

	cmdBytes := []byte(cmd + "\n")

	if err := validateCommand(cmdBytes); err != nil {
		fmt.Println(err)
		return "", err

	}

	attempts := 2
	var poolConn *PoolConn
	var err error

	for attempts > 0 {
		poolConn, err = c.pool.Get()
		if err != nil {
			return "", err
		}

		_, err = poolConn.conn.Write(cmdBytes)

		if err != nil {
			poolConn.Close()
			attempts--
			if attempts <= 0 {
				return "", err
			}

			continue

		}

		break
	}
	defer c.pool.Return(poolConn)

	//poolConn.scanner = bufio.NewScanner(poolConn.conn)

	if poolConn.scanner.Scan() {
		fmt.Println("Data from read: ", poolConn.scanner.Text())
		return poolConn.scanner.Text(), nil
	}

	if err := poolConn.scanner.Err(); err != nil {
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

func (c *Client) Ping() (string, error) {

	res, err := sendCommand(c, "PING")

	return res, err

}

// pending to implement the other commands, FLUSH DELETE
