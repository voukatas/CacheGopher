package replication

import (
	"bufio"
	"fmt"
	"net"

	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

type ReplConn struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

type WriteEvent struct {
	cmd   string
	key   string
	value string
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
		if server.ID == server.Primary {
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

func (rc *ReplConn) checkResp() error {
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

func (r *Replicator) replicateTask(we WriteEvent) {
	for _, server := range r.secondaries {
		if replConn, exists := r.connMap[server.ID]; exists {
			cmd := fmt.Sprintf("%s %s %s\n", we.cmd, we.key, we.value)
			_, err := replConn.conn.Write([]byte(cmd))

			if err != nil {
				replConn.conn.Close()

				newConn, err := establishConnection(server.Address)
				if err != nil {
					r.logger.Error("failed to establish connection" + err.Error())
					return
				}

				r.connMap[server.ID] = newConn

				_, err = newConn.conn.Write([]byte(cmd))
				if err != nil {
					r.logger.Error(err.Error())
					return
				}
				// cehck the response
				err = newConn.checkResp()
				if err != nil {
					r.logger.Error(err.Error())
					return
				}
				r.logger.Info("Task replicated")

			} else {
				err = replConn.checkResp()
				if err != nil {
					r.logger.Error(err.Error())
					return
				}
				r.logger.Info("Task replicated")

			}

		} else {
			r.logger.Error("failed to replicate, no connection to server: " + server.ID)

		}
	}

}
