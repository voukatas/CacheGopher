package cache

import (
	"fmt"
	"strings"
)

type Cache interface {
	Set(key string, value string)
	Get(key string) (string, bool)
	Delete(key string) bool
	Flush()
	Keys() []string
	GetAll() map[string]string
	// SetLogger(logger.Logger)
	// GetLogger() logger.Logger
}

// type Cache struct {
// 	store  map[string]string
// 	lock   sync.RWMutex
// 	logger logger.Logger
// 	size   int64
// }

func NewCache(cacheType string, capacity int) (Cache, error) {
	if capacity < 1 {
		return nil, fmt.Errorf("capacity should be more than 1")
	}
	//
	// if logger == nil {
	// 	return nil, fmt.Errorf("logger cannot be nil")
	// }

	var c Cache

	switch strings.ToUpper(cacheType) {
	case "LRU":
		c = NewLRUCache(capacity)
		//c = NewLRUCache2(capacity)
		//c.SetLogger(logger)

	//
	default:
		return nil, fmt.Errorf("Unknown cache type: %s", cacheType)
	}

	return c, nil
}

// func HandleConnection(conn net.Conn, c Cache, logger logger.Logger) {
// 	defer conn.Close()
//
// 	const maxTokenSize = 1 * 64 * 1024 // force 64KB to be the max
// 	const initBufSize = 4 * 1024
//
// 	buf := make([]byte, initBufSize, maxTokenSize) // 64KB
//
// 	scanner := bufio.NewScanner(conn) // Can read up to 64KB by default
// 	scanner.Buffer(buf, maxTokenSize)
//
// 	for scanner.Scan() {
// 		cmd := strings.SplitN(scanner.Text(), " ", 3)
// 		switch cmd[0] {
// 		case "SET":
// 			if len(cmd) != 3 {
// 				fmt.Fprintf(conn, "ERROR: Usage: SET <key> <value>\n")
// 				logger.Error("ERROR: Usage: SET <key> <value>")
// 				continue
// 			}
// 			c.Set(cmd[1], cmd[2])
// 			fmt.Fprintf(conn, "OK\n")
// 			logger.Debug("SET OK")
// 		case "GET":
// 			if len(cmd) != 2 {
// 				fmt.Fprintf(conn, "ERROR: Usage: GET <key>\n")
// 				logger.Debug("ERROR: Usage: GET <key>")
// 				continue
// 			}
// 			v, ok := c.Get(cmd[1])
// 			if !ok {
// 				fmt.Fprintf(conn, "ERROR: Key not found\n")
// 				logger.Debug("ERROR: Key not found")
// 				continue
// 			}
// 			fmt.Fprintf(conn, "%s\n", v)
// 			logger.Debug("GET" + " value:" + v)
//
// 		case "DELETE":
// 			if len(cmd) != 2 {
// 				fmt.Fprintf(conn, "ERROR: Usage: DELETE <key>\n")
// 				continue
// 			}
//
// 			res := c.Delete(cmd[1])
// 			if res {
// 				fmt.Fprintf(conn, "OK\n")
// 			} else {
//
// 				fmt.Fprintf(conn, "ERROR: No such key\n")
// 			}
//
// 		case "FLUSH":
// 			if len(cmd) != 1 {
//
// 				fmt.Fprintf(conn, "ERROR: Usage: FLUSH\n")
// 				continue
// 			}
//
// 			c.Flush()
// 			fmt.Fprintf(conn, "OK\n")
//
// 		case "KEYS":
// 			if len(cmd) != 1 {
//
// 				fmt.Fprintf(conn, "ERROR: Usage: KEYS\n")
// 				continue
// 			}
//
// 			keys := c.Keys()
// 			if len(keys) == 0 {
// 				fmt.Fprintf(conn, "No keys found\n")
// 				continue
//
// 			}
//
// 			for _, key := range keys {
// 				fmt.Fprintf(conn, "%s\n", key)
// 			}
//
// 		case "PING":
//
// 			fmt.Fprintf(conn, "PONG\n")
// 			logger.Debug("PONG")
// 		case "EXIT":
//
// 			fmt.Fprintf(conn, "Goodbye!\n")
// 			return
//
// 		default:
// 			fmt.Fprintf(conn, "ERROR: Unknown command: %s\n", cmd[0])
//
// 		}
// 	}
//
// 	if err := scanner.Err(); err != nil {
// 		fmt.Fprintf(conn, "ERROR: Failed to read command: %s\n", err)
//
// 	}
//
// 	logger.Debug("HandleConnection finished")
// }
