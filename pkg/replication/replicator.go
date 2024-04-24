package replication

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

// only for testing
// maybe I should create a new pkg and put every common Mock in there and use a build test tag ...
type MockReplicator struct{}

func (mr *MockReplicator) AddWriteEvent(we WriteEvent) {
}

type ReplicationService interface {
	AddWriteEvent(WriteEvent)
}

type ReplConn struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

type WriteEvent struct {
	Cmd   string
	Key   string
	Value string
}

type Replicator struct {
	connMap     map[string]*ReplConn
	secondaries []config.ServerConfig
	writeCh     chan WriteEvent
	logger      logger.Logger
}

func NewReplicator(currentServerId string, cfg *config.Configuration, logger logger.Logger) (*Replicator, error) {

	connMap := make(map[string]*ReplConn)
	secondariesConfig := make([]config.ServerConfig, 0)
	writeCh := make(chan WriteEvent, 100)

	for _, server := range cfg.Servers {
		if currentServerId == server.Primary {
			conn, err := establishConnection(server.Address)
			if err != nil {
				return nil, err
			}
			connMap[server.ID] = conn
			secondariesConfig = append(secondariesConfig, server)
		}
	}

	rep := &Replicator{
		//secondaries: secondaries,
		connMap:     connMap,
		secondaries: secondariesConfig,
		writeCh:     writeCh,
		logger:      logger,
	}

	// Single goroutine to keep the order as much as possible
	// Might be performance bottleneck though....
	go func() {
		for we := range writeCh {

			rep.replicateTask(we)
		}
	}()

	return rep, nil
}

func establishConnection(address string) (*ReplConn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conn)
	return &ReplConn{conn: conn, scanner: scanner}, nil

}

func (r *Replicator) checkResponse(conn *ReplConn) error {
	err := conn.checkConnResp()
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	r.logger.Info("Task replicated")
	return nil

}

func (rc *ReplConn) checkConnResp() error {
	if rc.scanner.Scan() {
		if rc.scanner.Text() != "OK" {
			return fmt.Errorf("received: " + rc.scanner.Text() + " instead of OK")
		}
	}

	if err := rc.scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (r *Replicator) AddWriteEvent(we WriteEvent) {
	r.writeCh <- we
}

func sendCommand(replConn *ReplConn, we WriteEvent) error {

	var cmd string

	switch we.Cmd {
	case "SET":
		cmd = fmt.Sprintf("%s %s %s\n", we.Cmd, we.Key, we.Value)
	case "DELETE":
		cmd = fmt.Sprintf("%s %s\n", we.Cmd, we.Key)
	}
	//cmd = fmt.Sprintf("%s %s %s\n", we.Cmd, we.Key, we.Value)
	_, err := replConn.conn.Write([]byte(cmd))
	return err
}

func reEstablishConnection(address string) (*ReplConn, error) {
	time.Sleep(2 * time.Second)
	return establishConnection(address)
}

// If one secondary server fails continue trying with the others
func (r *Replicator) replicateTask(we WriteEvent) {
	for _, server := range r.secondaries {
		replConn, exists := r.connMap[server.ID]
		if !exists {
			r.logger.Error("failed to replicate, no connection to server: " + server.ID)
			continue
		}

		// cmd := fmt.Sprintf("%s %s %s\n", we.cmd, we.key, we.value)
		// _, err := replConn.conn.Write([]byte(cmd))
		err := sendCommand(replConn, we)
		currentConn := replConn

		if err != nil {
			replConn.conn.Close()

			newConn, err := reEstablishConnection(server.Address)
			if err != nil {
				r.logger.Error("failed to establish connection" + err.Error())
				continue
			}

			r.connMap[server.ID] = newConn

			//_, err = newConn.conn.Write([]byte(cmd))
			err = sendCommand(newConn, we)

			if err != nil {
				newConn.conn.Close()
				r.logger.Error(err.Error())
				continue
			}

			currentConn = newConn
			// cehck the response
			// err = newConn.checkConnResp()
			// if err != nil {
			// 	r.logger.Error(err.Error())
			// 	return
			// }
			// r.logger.Info("Task replicated")

		}

		if err := r.checkResponse(currentConn); err != nil {
			r.logger.Error(fmt.Sprintf("Response check failed for %s: %v", server.ID, err.Error()))
		}

		// else {
		// 	err = replConn.checkConnResp()
		// 	if err != nil {
		// 		r.logger.Error(err.Error())
		// 		return
		// 	}
		// 	r.logger.Info("Task replicated")
		//
		// }

	}

}
