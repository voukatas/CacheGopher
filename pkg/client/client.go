package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/errorutil"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

// package-scope logger, use this for the whole lib, not only for the Client
var (
	libLoggerInstance logger.Logger = &NoOpLogger{}
)

// An empty logger that doesn't trigger any logging but it is convinient to use because we avoid the err != nil for every log
type NoOpLogger struct{}

// implement the Logger interface
func (n *NoOpLogger) Debug(msg string) {}
func (n *NoOpLogger) Info(msg string)  {}
func (n *NoOpLogger) Warn(msg string)  {}
func (n *NoOpLogger) Error(msg string) {}

// maybe i should consider using an Exported not bound to Client logger in the future
func setupLogger() {

	libLoggerInstance = logger.SetupDebugLogger()

}

func getLogger() logger.Logger {
	return libLoggerInstance
}

type PoolConn struct {
	conn      net.Conn
	scanner   *bufio.Scanner
	createdAt time.Time
}

type ConnPool struct {
	pool    chan *PoolConn
	address string
	cfg     config.ClientConfig
	// size    int
}

type ReadBalancer struct {
	nodes []*CacheNode
	index int
	lock  sync.Mutex
	cfg   config.ClientConfig
}

func NewReadBalancer(cfg config.ClientConfig) *ReadBalancer {
	return &ReadBalancer{
		nodes: make([]*CacheNode, 0),
		index: 0,
		cfg:   cfg,
	}
}

func (rb *ReadBalancer) addCacheNode(node *CacheNode) {
	rb.nodes = append(rb.nodes, node)
}

func (rb *ReadBalancer) getNextCacheNode() (*CacheNode, error) {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	total := len(rb.nodes)
	if total == 0 {
		return nil, fmt.Errorf("no cache nodes exist")
	}

	for i := 0; i < total; i++ {

		node := rb.nodes[rb.index%total]
		rb.index++

		node.HealthLock.Lock()
		if !node.Unhealthy || time.Now().After(node.RetryAt) {

			if node.Unhealthy {
				node.Unhealthy = false
			}

			node.HealthLock.Unlock()
			return node, nil
		}

		node.HealthLock.Unlock()

	}

	return nil, fmt.Errorf("no available cache nodes")
}

func (pc *PoolConn) isExpired(timeout int) bool {
	maxValidTime := time.Duration(timeout) * time.Second
	getLogger().Debug("isExpired timeout: " + maxValidTime.String())
	return time.Since(pc.createdAt) > maxValidTime
}

func (pc *PoolConn) Close() {
	pc.conn.Close()
}

func NewConnPool(size int, address string, cfg config.ClientConfig) *ConnPool {
	return &ConnPool{
		pool:    make(chan *PoolConn, size),
		address: address,
		cfg:     cfg,
	}
}

func (cp *ConnPool) Get() (*PoolConn, error) {
	getLogger().Debug("Get connection from pool called")

	for {
		select {
		case poolConn := <-cp.pool:
			if poolConn.isExpired(cp.cfg.ConnectionTimeout) {
				getLogger().Debug("isExpired or is invalid")
				poolConn.Close()
				continue
			}
			getLogger().Debug("found poolConn")
			return poolConn, nil

		default:

			return cp.dialWithBackOff()
		}
	}
}

func (cp *ConnPool) dialWithBackOff() (*PoolConn, error) {
	getLogger().Debug(" dialWithBackOff")
	maxAttempts := 3
	baseTime := 100 * time.Millisecond
	maxBackoff := 1 * time.Second

	var conn net.Conn
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {

		conn, err = net.Dial("tcp", cp.address)

		if err == nil {
			tcpConn, ok := conn.(*net.TCPConn)
			if !ok {
				getLogger().Debug("Connection is not TCP type")
				err = fmt.Errorf("expected TCP connection, got different type")
				continue

			}

			// disable Nagle's Algorithm
			// if err := tcpConn.SetNoDelay(true); err != nil {
			// 	return nil, fmt.Errorf("failed to set TCP_NODELAY: %s", err)
			// }

			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(time.Duration(cp.cfg.KeepAliveInterval) * time.Second)
			getLogger().Debug("KeepAlive: " + fmt.Sprint(cp.cfg.KeepAliveInterval))

			poolConn := &PoolConn{conn: tcpConn, scanner: bufio.NewScanner(tcpConn), createdAt: time.Now()}
			getLogger().Debug("Successfully Created poolConn")
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
	ring      HashRing
	balancers map[string]*ReadBalancer
}

func NewClient(enableLogging bool) (*Client, error) {
	cfg, err := config.LoadConfig("cacheGopherConfig.json")
	if err != nil {
		return nil, fmt.Errorf("Failed to read configuration: " + err.Error())
	}

	ring := NewHashRing()
	balancers := map[string]*ReadBalancer{}

	for _, node := range cfg.Servers {

		newPool := NewConnPool(5, node.Address, cfg.ClientConfig)

		if strings.ToUpper(node.Role) == "PRIMARY" {
			newNode := NewCacheNode(node.ID, true, newPool)

			ring.AddNode(newNode)
			newBalancer := NewReadBalancer(cfg.ClientConfig)
			newBalancer.addCacheNode(newNode)
			balancers[node.ID] = newBalancer

		} else if strings.ToUpper(node.Role) == "SECONDARY" {
			newNode := NewCacheNode(node.ID, false, newPool)

			readBalancer := balancers[node.Primary]
			readBalancer.addCacheNode(newNode)

		} else {
			return nil, fmt.Errorf("Unknown role: %s", node.Role)
		}
	}

	if enableLogging {

		setupLogger()

	}

	return &Client{
		ring:      ring,
		balancers: balancers,
	}, nil
}

func validateCommand(cmdBytes []byte) error {
	const maxTokenSize = 64 * 1024

	if len(cmdBytes) > maxTokenSize {
		getLogger().Debug("data exceeds the maximum allowed size of 64KB")
		return fmt.Errorf("command exceeds the maximum allowed size of 64KB")

	}

	if bytes.Contains(cmdBytes[:len(cmdBytes)-1], []byte("\n")) {
		getLogger().Debug("Invalid Data, data can't contain the newline chars")
		return fmt.Errorf("command cannot contain newline characters")
	}

	return nil
}

func (c *Client) sendCommand(node *CacheNode, cmd string) (string, error) {

	cmdBytes := []byte(strings.TrimSpace(cmd) + "\n")

	if err := validateCommand(cmdBytes); err != nil {

		getLogger().Debug("Error sending command:" + err.Error())
		return "", err

	}

	attempts := 2
	var poolConn *PoolConn
	var err error

	for attempts > 0 {
		poolConn, err = node.ConnPool.Get()
		if err != nil {
			getLogger().Debug("Error in conn pool" + err.Error())
			return "", err
		}

		getLogger().Debug("sendCommand: Before writing to the connection")
		_, err = poolConn.conn.Write(cmdBytes)
		getLogger().Debug("sendCommand: After writing to the connection")

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
	defer node.ConnPool.Return(poolConn)

	getLogger().Debug("sendCommand: Waiting for response")

	if poolConn.scanner.Scan() {
		getLogger().Debug("Data from read: " + poolConn.scanner.Text())
		if poolConn.scanner.Text() == "ERROR: Key not found" {
			return "", errorutil.ErrKeyNotFound
		} else if strings.Contains(poolConn.scanner.Text(), "ERROR:") {
			return "", fmt.Errorf(poolConn.scanner.Text())
		}
		return poolConn.scanner.Text(), nil
	}

	if err := poolConn.scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no response")
}

func (c *Client) Set(k, v string) (string, error) {
	getLogger().Debug("SET " + k + " " + v)
	cmd := fmt.Sprintf("SET %s %s", k, v)
	primaryNode, err := c.ring.GetNode(k)
	if err != nil {
		return "", err
	}
	getLogger().Debug("node selected to send the request: " + primaryNode.ID)
	resp, err := c.sendCommand(primaryNode, cmd)

	return resp, err
}

func (c *Client) Get(k string) (string, error) {
	getLogger().Debug("GET " + k)
	cmd := fmt.Sprintf("GET %s", k)
	primaryNode, err := c.ring.GetNode(k)
	if err != nil {
		return "", err
	}

	balancer := c.balancers[primaryNode.ID]

	for {

		node, err := balancer.getNextCacheNode()
		if err != nil {
			getLogger().Error(err.Error())
			return "", err
		}
		getLogger().Debug("node selected to send the request: " + node.ID)
		resp, err := c.sendCommand(node, cmd)

		if err != nil {
			switch {
			case errors.Is(err, errorutil.ErrKeyNotFound):
				return "", err
			default:
				// set unhealthy
				getLogger().Warn("node: " + node.ID + " set to UnHealthy")
				node.SetUnhealthy(time.Duration(balancer.cfg.UnHealthyInterval) * time.Second)
				continue
			}
		}

		return resp, err
	}

}

func (c *Client) Ping(node *CacheNode) (string, error) {

	res, err := c.sendCommand(node, "PING")

	return res, err

}

func (c *Client) Delete(k string) (string, error) {
	getLogger().Debug("DELETE " + k)
	cmd := fmt.Sprintf("DELETE %s", k)
	primaryNode, err := c.ring.GetNode(k)
	if err != nil {
		return "", err
	}
	getLogger().Debug("node selected to send the request: " + primaryNode.ID)
	res, err := c.sendCommand(primaryNode, cmd)

	return res, err

}

// pending to implement the other commands, FLUSH DELETE
