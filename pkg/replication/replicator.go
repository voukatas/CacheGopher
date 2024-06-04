package replication

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/errorutil"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

// only for testing
// maybe I should create a new pkg and put every common Mock in there and use a build test tag ...
type MockReplicator struct{}

func (mr *MockReplicator) AddWriteEvent(we WriteEvent) {
}
func (mr *MockReplicator) IsPrimary() bool {
	return true
}
func (mr *MockReplicator) GetSecondaryConn(id string) (*ReplConn, error) {
	return nil, nil
}
func (mr *MockReplicator) RemoveConn(id string) {
}

type ReplicationService interface {
	AddWriteEvent(WriteEvent)
	IsPrimary() bool
	GetSecondaryConn(string) (*ReplConn, error)
	RemoveConn(string)
}

type ReplConn struct {
	Conn    net.Conn
	Scanner *bufio.Scanner
}

type WriteEvent struct {
	Cmd   string
	Key   string
	Value string
}

type Replicator struct {
	connMap     map[string]*ReplConn
	connMapLock sync.RWMutex
	secondaries []config.ServerConfig
	writeCh     chan WriteEvent
	logger      logger.Logger
	isPrimary   bool
}

func NewReplicator(currentServerId string, cfg *config.Configuration, logger logger.Logger) (*Replicator, error) {

	connMap := make(map[string]*ReplConn)
	secondariesConfig := make([]config.ServerConfig, 0)
	writeCh := make(chan WriteEvent, 100)
	isPrimary := false

	for _, server := range cfg.Servers {
		// open connections to all the secondary servers
		if currentServerId == server.Primary {
			isPrimary = true

			secondariesConfig = append(secondariesConfig, server)

			conn, err := establishConnection(server.Address)
			if err != nil {
				logger.Error(errorutil.Wrap(err, "failed to establish connection").Error())
				continue
			}
			connMap[server.ID] = conn
		}
	}

	// i should refactor this for the case that is not primary, no need to keep the references and waste resources

	rep := &Replicator{
		connMap:     connMap,
		secondaries: secondariesConfig,
		writeCh:     writeCh,
		logger:      logger,
		isPrimary:   isPrimary,
	}

	// Single goroutine to keep the order as much as possible
	// Might be performance bottleneck though....
	// In case this is replaced from multiple goroutines then I need to lock the connMap
	go func() {
		for we := range writeCh {

			rep.replicateTask(we)
		}
	}()

	return rep, nil
}

func (rp *Replicator) RemoveConn(serverId string) {
	rp.connMapLock.Lock()
	defer rp.connMapLock.Unlock()
	rp.logger.Debug("inside RemoveConn")

	delete(rp.connMap, serverId)
}

func (rp *Replicator) GetSecondaryConn(id string) (*ReplConn, error) {
	rp.connMapLock.RLock()
	defer rp.connMapLock.RUnlock()

	replConn, exists := rp.connMap[id]
	if !exists {
		return nil, fmt.Errorf("connection to server not found")
	}
	return replConn, nil
}

func (rp *Replicator) IsPrimary() bool {
	return rp.isPrimary
}

func establishConnection(address string) (*ReplConn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conn)
	return &ReplConn{Conn: conn, Scanner: scanner}, nil

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
	if rc.Scanner.Scan() {

		if rc.Scanner.Text() != "OK" {
			return fmt.Errorf("received: " + rc.Scanner.Text() + " instead of OK")
		}
	} else {
		return fmt.Errorf("No response received")
	}

	if err := rc.Scanner.Err(); err != nil {
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
	_, err := replConn.Conn.Write([]byte(cmd))
	return err
}

func reEstablishConnection(address string) (*ReplConn, error) {
	time.Sleep(2 * time.Second)
	return establishConnection(address)
}

// ToDo: THIS THING NEEDS REFACTOR
// Maybe refactor this using a granular locking where a sync.Map keeps the locks or an actor model approach to increase performance and avoid memory sharing

// If one secondary server fails continue trying with the others
func (r *Replicator) replicateTask(we WriteEvent) {
	r.connMapLock.Lock()
	defer r.connMapLock.Unlock()

	for _, server := range r.secondaries {
		replConn, exists := r.connMap[server.ID]
		if !exists {
			conn, err := establishConnection(server.Address)
			if err != nil {
				//r.logger.Error("failed to establish connection" + err.Error())
				r.logger.Error(errorutil.Wrap(err, "failed to establish connection").Error())
				continue
			}
			r.connMap[server.ID] = conn

			replConn = conn

		}

		// There is a case where the connection drop is not detected immediatelly so no error is returned, we will handle it at the end
		err := sendCommand(replConn, we)
		currentConn := replConn

		if err != nil {
			replConn.Conn.Close()

			newConn, err := reEstablishConnection(server.Address)
			if err != nil {
				r.logger.Error(errorutil.Wrap(err, "failed to establish connection").Error())
				continue
			}

			r.connMap[server.ID] = newConn

			err = sendCommand(newConn, we)

			if err != nil {
				newConn.Conn.Close()
				r.logger.Error(errorutil.Wrap(err, "").Error())
				continue
			}

			currentConn = newConn

		}

		if err := r.checkResponse(currentConn); err != nil {
			r.logger.Error(fmt.Sprintf("Response check failed for %s: %v", server.ID, err.Error()))
			r.logger.Error(fmt.Sprintf("Retry one more time to connect to %s", server.ID))

			// This is for the case that the dropped connection is not imediatelly detected

			conn, err := establishConnection(server.Address)
			if err != nil {
				r.logger.Error(errorutil.Wrap(err, "failed to establish connection").Error())
				return
			}
			r.connMap[server.ID] = conn
			err = sendCommand(conn, we)
			if err != nil {
				conn.Conn.Close()
				r.logger.Error(err.Error())
				continue
			}

		}

	}

}
