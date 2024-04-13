package cache

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
)

type Cache struct {
	store  map[string]string
	lock   sync.RWMutex
	logger *slog.Logger
	size   int64
}

func NewCache(logger *slog.Logger) *Cache {
	return &Cache{
		store:  make(map[string]string, 0),
		logger: logger,
		size:   0,
	}
}

func (c *Cache) Set(key, value string) int64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, exists := c.store[key]; !exists {

		c.size++
	}
	c.store[key] = value

	return c.size
}

func (c *Cache) Get(key string) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	v, exists := c.store[key]

	return v, exists
}

func (c *Cache) Delete(key string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, exists := c.store[key]; exists {

		delete(c.store, key)
		c.size--
		return true
	}

	return false
}

func (c *Cache) Flush() bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.store = make(map[string]string)
	c.size = 0
	return true
}

func (c *Cache) Keys() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	keys := make([]string, 0, len(c.store))

	for k := range c.store {
		keys = append(keys, k)
	}

	return keys

}

func HandleConnection(conn net.Conn, c *Cache) {
	defer conn.Close()

	const maxTokenSize = 1 * 64 * 1024 // force 64KB to be the max
	const initBufSize = 4 * 1024

	buf := make([]byte, initBufSize, maxTokenSize) // 64KB

	scanner := bufio.NewScanner(conn) // Can read up to 64KB by default
	scanner.Buffer(buf, maxTokenSize)

	for scanner.Scan() {
		cmd := strings.SplitN(scanner.Text(), " ", 3)
		switch cmd[0] {
		case "SET":
			if len(cmd) != 3 {
				fmt.Fprintf(conn, "ERROR: Usage: SET <key> <value>\n")
				c.logger.Error("ERROR: Usage: SET <key> <value>")
				continue
			}
			c.Set(cmd[1], cmd[2])
			fmt.Fprintf(conn, "OK\n")
			c.logger.Debug("SET OK")
		case "GET":
			if len(cmd) != 2 {
				fmt.Fprintf(conn, "ERROR: Usage: GET <key>\n")
				c.logger.Debug("ERROR: Usage: GET <key>")
				continue
			}
			v, ok := c.Get(cmd[1])
			if !ok {
				fmt.Fprintf(conn, "ERROR: Key not found\n")
				c.logger.Debug("ERROR: Key not found")
				continue
			}
			fmt.Fprintf(conn, "%s\n", v)
			c.logger.Debug("GET", "value", v)

		case "DELETE":
			if len(cmd) != 2 {
				fmt.Fprintf(conn, "ERROR: Usage: DELETE <key>\n")
				continue
			}

			res := c.Delete(cmd[1])
			if res {
				fmt.Fprintf(conn, "OK\n")
			} else {

				fmt.Fprintf(conn, "ERROR: No such key\n")
			}

		case "FLUSH":
			if len(cmd) != 1 {

				fmt.Fprintf(conn, "ERROR: Usage: FLUSH\n")
				continue
			}

			res := c.Flush()
			if res {
				fmt.Fprintf(conn, "OK\n")
			} else {

				fmt.Fprintf(conn, "ERROR: WTF (What a Terrible Failure)\n")
			}

		case "KEYS":
			if len(cmd) != 1 {

				fmt.Fprintf(conn, "ERROR: Usage: KEYS\n")
				continue
			}

			keys := c.Keys()
			if len(keys) == 0 {
				fmt.Fprintf(conn, "No keys found\n")
				continue

			}

			for _, key := range keys {
				fmt.Fprintf(conn, "%s\n", key)
			}

		case "PING":

			fmt.Fprintf(conn, "PONG\n")
			c.logger.Debug("PONG")
		case "EXIT":

			fmt.Fprintf(conn, "Goodbye!\n")
			return

		default:
			fmt.Fprintf(conn, "ERROR: Unknown command: %s\n", cmd[0])

		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(conn, "ERROR: Failed to read command: %s\n", err)

	}

	c.logger.Debug("HandleConnection finished")
}
