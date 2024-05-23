package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/logger"
	"github.com/voukatas/CacheGopher/pkg/replication"
)

type Server struct {
	cache          cache.Cache
	logger         logger.Logger
	replicator     replication.ReplicationService
	isPrimary      bool
	primaryAddress string
	queuedWrites   []*LogEvent
	writeLock      sync.Mutex // lock to protect the queuedWrites
	isRecovering   bool
}

func NewServer(cache cache.Cache, logger logger.Logger, replicator replication.ReplicationService, isPrimary bool, primaryAddress string) *Server {
	return &Server{
		cache:          cache,
		logger:         logger,
		replicator:     replicator,
		isPrimary:      isPrimary,
		primaryAddress: primaryAddress,
	}
}

type LogEvent struct {
	//Timestamp time.Time
	Key   string
	Value string
	Op    string
	//SeqNum    int64 // maybe i need it in future
}

func (s *Server) SendCurrentState(conn net.Conn) {
	s.logger.Debug("Sending current state")

	data := s.cache.GetSnapshot()
	for k, v := range data {
		fmt.Fprintf(conn, "SET %s %s\n", k, v)
		s.logger.Debug("Key: " + k + "Value: " + v + "\n")

	}

}

func (s *Server) StopWriteOpsAndEnableQueuedWrites() {
	s.cache.Lock()
	defer s.cache.Unlock()
	// s.writeLock.Lock()
	// defer s.writeLock.Unlock()
	s.logger.Debug("isRecovering = true")

	s.isRecovering = true
}

func (s *Server) SendQueuedWrites(conn net.Conn) {
	s.cache.Lock()
	defer s.cache.Unlock()

	s.writeLock.Lock()
	defer s.writeLock.Unlock()

	s.logger.Debug("inside SendQueuedWrites")

	for _, v := range s.queuedWrites {

		if v.Op == "SET" {

			fmt.Fprintf(conn, "SET %s %s\n", v.Key, v.Value)
			s.logger.Debug("Key: " + v.Key + "Value: " + v.Value + "\n")

		} else if v.Op == "DELETE" {

			fmt.Fprintf(conn, "DELETE %s\n", v.Key)
			s.logger.Debug("Key: " + v.Key + "\n")

		}

	}

	s.queuedWrites = []*LogEvent{}
	s.isRecovering = false

}

func (s *Server) IsRecovering(cmd []string) {

	if s.isRecovering {
		var event *LogEvent

		if cmd[0] == "SET" {
			event = &LogEvent{
				Key:   cmd[1],
				Value: cmd[2],
				Op:    cmd[0],
			}
		} else if cmd[0] == "DELETE" {
			event = &LogEvent{
				Key: cmd[1],
				Op:  cmd[0],
			}
		} else {
			s.logger.Error("Unknown command: " + cmd[0])
			return
		}

		s.writeLock.Lock()
		defer s.writeLock.Unlock()
		s.queuedWrites = append(s.queuedWrites, event)
	}

}

func (s *Server) HandleRecovery(myConfig config.ServerConfig) error {

	if !s.replicator.IsPrimary() {
		conn, err := net.Dial("tcp", s.primaryAddress)
		if err != nil {

			return fmt.Errorf("failed to connect to primary, recovery failed")
		}

		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		replConn := &replication.ReplConn{Conn: conn, Scanner: scanner}
		err = s.startRecovery(replConn, myConfig.ID)
		if err != nil {
			return fmt.Errorf("failed to recover")
		}

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

			err = s.startRecovery(replConn, myConfig.ID)
			if err != nil {
				return fmt.Errorf("failed to recover")
			}

			// if we reach here then there is no need to continue with the rest servers

			return nil
		}
	}

	return nil

}

func (s *Server) startRecovery(replConn *replication.ReplConn, serverId string) error {
	s.logger.Debug("Initiating recovery process")
	fmt.Fprintf(replConn.Conn, "RECOVER "+serverId+"\n")
	//scanner := bufio.NewScanner(replConn.Scanner)
	for replConn.Scanner.Scan() {
		s.logger.Debug("Line: " + replConn.Scanner.Text())
		if replConn.Scanner.Text() == "RECOVEREND" {
			break
		}
		parts := strings.SplitN(replConn.Scanner.Text(), " ", 3)

		switch parts[0] {
		case "SET":
			if len(parts) != 3 {
				return fmt.Errorf("failed to parse recover key value")
			}

			s.cache.Set(parts[1], parts[2])
			fmt.Fprintf(replConn.Conn, "OK\n")
		case "DELETE":
			if len(parts) != 2 {
				return fmt.Errorf("failed to parse recover key value")
			}

			s.cache.Delete(parts[1])
			fmt.Fprintf(replConn.Conn, "OK\n")

		}
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
	s.logger.Debug("inside HandleConnection")

	for scanner.Scan() {

		s.logger.Debug("inside scanner: " + scanner.Text())

		cmd := strings.SplitN(strings.TrimSpace(scanner.Text()), " ", 3)
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

			// check if we recover and do your thing
			s.IsRecovering(cmd)

			//fmt.Println("Keys SET : ", s.cache.Keys())

		case "GET":
			//fmt.Println("Keys GET : ", s.cache.Keys())
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
			//fmt.Println("Keys DELETE : ", s.cache.Keys())
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

			// check if we recover and do your thing
			s.IsRecovering(cmd)

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
			// This command should be received one at a time, no two servers should send the recover command at the same time
			s.logger.Debug("RECOVER")
			s.logger.Debug("serverId: " + cmd[1])
			s.replicator.RemoveConn(cmd[1])
			// get the lock for write to block any write operation and start keeping in memory the write changes in a log slice
			if s.isPrimary {
				s.logger.Debug("It is a Primary node")
				// do something
				// since it is a primary node we need to stop briefly the write operations and enable queued writes
				s.StopWriteOpsAndEnableQueuedWrites()
			}

			// lock in read mode and send all the current keys
			s.SendCurrentState(conn)

			s.logger.Debug("state send")
			//time.Sleep(10 * 1000 * time.Millisecond)

			// lock in write mode to send all the remaining keys if it is the primary server
			// if s.replicator.IsPrimary() {
			// 	s.logger.Debug("IsPrimary true")
			// } else {
			// 	s.logger.Debug("Not a primary node")
			// }
			if s.isPrimary {
				//s.logger.Debug("IsPrimary true")
				s.SendQueuedWrites(conn)
			}

			fmt.Fprintf(conn, "RECOVEREND\n")
			return

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
