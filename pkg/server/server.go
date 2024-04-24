package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/logger"
	"github.com/voukatas/CacheGopher/pkg/replication"
)

type Server struct {
	cache      cache.Cache
	logger     logger.Logger
	replicator replication.ReplicationService
}

func NewServer(cache cache.Cache, logger logger.Logger, replicator replication.ReplicationService) *Server {
	return &Server{
		cache:      cache,
		logger:     logger,
		replicator: replicator,
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
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
				s.logger.Error("ERROR: Usage: SET <key> <value>")
				continue
			}
			s.cache.Set(cmd[1], cmd[2])
			fmt.Fprintf(conn, "OK\n")
			s.logger.Debug("SET OK")
			s.replicator.AddWriteEvent(replication.WriteEvent{Key: cmd[1], Value: cmd[2], Cmd: cmd[0]})
		case "GET":
			if len(cmd) != 2 {
				fmt.Fprintf(conn, "ERROR: Usage: GET <key>\n")
				s.logger.Debug("ERROR: Usage: GET <key>")
				continue
			}
			v, ok := s.cache.Get(cmd[1])
			if !ok {
				fmt.Fprintf(conn, "ERROR: Key not found\n")
				s.logger.Debug("ERROR: Key not found: " + cmd[1])
				continue
			}
			fmt.Fprintf(conn, "%s\n", v)
			s.logger.Debug("GET" + " value:" + v)

		case "DELETE":
			if len(cmd) != 2 {
				fmt.Fprintf(conn, "ERROR: Usage: DELETE <key>\n")
				continue
			}

			res := s.cache.Delete(cmd[1])
			if res {
				fmt.Fprintf(conn, "OK\n")
				s.replicator.AddWriteEvent(replication.WriteEvent{Key: cmd[1], Cmd: cmd[0]})
			} else {

				fmt.Fprintf(conn, "ERROR: No such key\n")
			}

		case "FLUSH":
			if len(cmd) != 1 {

				fmt.Fprintf(conn, "ERROR: Usage: FLUSH\n")
				continue
			}

			s.cache.Flush()
			fmt.Fprintf(conn, "OK\n")

		case "KEYS":
			if len(cmd) != 1 {

				fmt.Fprintf(conn, "ERROR: Usage: KEYS\n")
				continue
			}

			keys := s.cache.Keys()
			if len(keys) == 0 {
				fmt.Fprintf(conn, "No keys found\n")
				continue

			}

			for _, key := range keys {
				fmt.Fprintf(conn, "%s\n", key)
			}

		case "PING":

			fmt.Fprintf(conn, "PONG\n")
			s.logger.Debug("PONG")
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

	s.logger.Debug("HandleConnection finished")
}
