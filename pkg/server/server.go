package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/logger"
	"github.com/voukatas/CacheGopher/pkg/replication"
)

type Server struct {
	cache      cache.Cache
	logger     logger.Logger
	replicator replication.ReplicationService
	isPrimary  bool
	// logEvent LogEvent
}

func NewServer(cache cache.Cache, logger logger.Logger, replicator replication.ReplicationService) *Server {
	return &Server{
		cache:      cache,
		logger:     logger,
		replicator: replicator,
	}
}

func (s *Server) SendCurrentState(conn net.Conn) {
	s.logger.Debug("Sending current state")

	data := s.cache.GetAll()
	for k, v := range data {
		fmt.Fprintf(conn, "%s %s\n", k, v)
		s.logger.Debug("Key: " + k + "Value: " + v + "\n")

	}

}

func (s *Server) HandleRecovery(myConfig config.ServerConfig) error {

	if !s.replicator.IsPrimary() {

	} else {

		for _, serverId := range myConfig.Secondaries {
			fmt.Println("\n", serverId)
			replConn, err := s.replicator.GetSecondaryConn(serverId)
			//conn, err := net.Dial("tcp", "localhost:31338")
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}
			//defer conn.Close()

			err = s.startRecovery(replConn)
			if err != nil {
				return fmt.Errorf("failed to recover")
			}

			// if we reach here then there is no need to continue with the rest servers

			return nil
		}
	}

	return nil

}

func (s *Server) startRecovery(replConn *replication.ReplConn) error {
	s.logger.Debug("Initiating recovery process")
	fmt.Fprintf(replConn.Conn, "RECOVER\n")
	//scanner := bufio.NewScanner(replConn.Scanner)
	for replConn.Scanner.Scan() {
		s.logger.Debug("Line: " + replConn.Scanner.Text())
		if replConn.Scanner.Text() == "RECOVEREND" {
			break
		}
		parts := strings.SplitN(replConn.Scanner.Text(), " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("failed to parse recover key value")
		}

		s.cache.Set(parts[0], parts[1])
	}

	if err := replConn.Scanner.Err(); err != nil {
		return fmt.Errorf("failed, errors during reading data")

	}

	return nil
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
		case "RECOVER":
			s.logger.Debug("RECOVER")
			// get the lock for write to block any write operation and start keeping in memory the write changes in a log slice
			s.SendCurrentState(conn)

			// lock in read mode and send all the current keys

			// lock in write mode to send all the remaining keys if it is the primary server
			if s.replicator.IsPrimary() {
				s.logger.Debug("IsPrimary true")
			} else {
				s.logger.Debug("Not a primary node")
			}
			fmt.Fprintf(conn, "RECOVEREND\n")
			//return
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
