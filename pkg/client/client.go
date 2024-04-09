package client

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Client struct {
	conn    net.Conn
	scanner *bufio.Scanner
	lock    sync.RWMutex
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, err
	}

	return &Client{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
		lock:    sync.RWMutex{},
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func sendCommand(c *Client, cmd string) (string, error) {
	// fmt.Fprintf(c.conn, cmd+"\n")
	cmdBytes := []byte(cmd + "\n")
	_, err := c.conn.Write(cmdBytes)
	if err != nil {
		return "", err
	}

	if c.scanner.Scan() {
		return c.scanner.Text(), nil
	}

	if err := c.scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no response")
}

func (c *Client) Set(k, v string) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	cmd := fmt.Sprintf("SET %s %s", k, v)
	resp, err := sendCommand(c, cmd)

	return resp, err
}

func (c *Client) Get(k string) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	cmd := fmt.Sprintf("GET %s", k)
	res, err := sendCommand(c, cmd)

	return res, err

}
