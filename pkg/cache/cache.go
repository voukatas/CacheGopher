package cache

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Cache struct {
	store map[string]string
	lock  sync.RWMutex
	size  int64
}

func NewCache() *Cache {
	return &Cache{
		store: make(map[string]string, 0),
		size:  0,
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
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		cmd := strings.Split(scanner.Text(), " ")
		switch cmd[0] {
		case "SET":
			if len(cmd) != 3 {
				fmt.Fprintf(conn, "ERROR: Usage: SET <key> <value>\n")
				continue
			}
			c.Set(cmd[1], cmd[2])
			fmt.Fprintf(conn, "OK\n")
		case "GET":
			if len(cmd) != 2 {
				fmt.Fprintf(conn, "ERROR: Usage: GET <key>\n")
				continue
			}
			v, ok := c.Get(cmd[1])
			if !ok {
				fmt.Fprintf(conn, "ERROR: Key not found\n")
				continue
			}
			fmt.Fprintf(conn, "%s\n", v)

		case "DELETE":
			if len(cmd) != 2 {
				fmt.Fprintf(conn, "ERROR: Usage: DELETE <key>\n")
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
			}

			keys := c.Keys()
			if len(keys) == 0 {
				fmt.Fprintf(conn, "No keys found\n")
				continue

			}

			for _, key := range keys {
				fmt.Fprintf(conn, "%s\n", key)
			}

		case "EXIT":

			fmt.Fprintf(conn, "Goodbye!\n")
			return

		default:
			fmt.Fprintf(conn, "ERROR: Unknown command\n")

		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(conn, "ERROR: Failed to read command: %s\n", err)

	}

}
