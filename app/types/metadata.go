package types

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

type ServerMetadata struct {
	// Replication info
	Role                       string        `json:"role"`
	ConnectedSlaves            int           `json:"connected_slaves"`
	MasterReplID               string        `json:"master_replid"`
	MasterReplOffset           int64         `json:"master_repl_offset"`
	SecondReplOffset           int64         `json:"second_repl_offset"`
	ReplBacklogActive          int           `json:"repl_backlog_active"`
	ReplBacklogSize            int64         `json:"repl_backlog_size"`
	ReplBacklogFirstByteOffset int64         `json:"repl_backlog_first_byte_offset"`
	ReplBacklogHistlen         int64         `json:"repl_backlog_histlen"`
	ReplActiveConnection       []net.Conn    `json:"-"`
	ReplChannel                chan []string `json:"-"`
	ShutdownChannel            chan struct{} `json:"-"`
	CommandProcessed           int64         `json:"-"`

	mutex sync.RWMutex
}

func (m *ServerMetadata) Replicate(Cmd []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	activeConnections := make([]net.Conn, 0)
	for _, conn := range m.ReplActiveConnection {
		err := m.Send(conn, Cmd)
		if err == nil {
			activeConnections = append(activeConnections, conn)
			// removing inactive connection
		}
	}
	m.ReplActiveConnection = activeConnections
}

func (m *ServerMetadata) NumberOfActiveConnections() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.ReplActiveConnection)
}

func (m *ServerMetadata) AddCommandProcessed(n int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.CommandProcessed += n
}

func (m *ServerMetadata) Send(conn net.Conn, Cmd []string) error {
	res := protocol.BuildArray(utility.ConvertStringArrayToAny(Cmd))
	_, err := conn.Write([]byte(res))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return fmt.Errorf("Seems like replica's not connected with errr: %s", err.Error())
	}
	return nil
}

func (m *ServerMetadata) ReplicateCommandToReplicas() {
	for {
		select {
		case Cmd := <-m.ReplChannel:
			// send the command to replicas
			m.Replicate(Cmd)
		case <-m.ShutdownChannel:
			log.Println("Replication work is shutting down")
			return
		}
	}
}

func (m *ServerMetadata) AddReplicasConnection(conn net.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ReplActiveConnection = append(m.ReplActiveConnection, conn)
}

func NewServerMetadata(role string) *ServerMetadata {
	// ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	metadata := ServerMetadata{
		Role:                 role,
		ReplChannel:          make(chan []string, 1_000),
		ReplActiveConnection: make([]net.Conn, 0),
	}
	go metadata.ReplicateCommandToReplicas()
	return &metadata
}

func (m *ServerMetadata) ToStringArray() []string {
	m.mutex.RLock()

	log.Println("Inside ServerMetadata's ToStringArray method")

	result := make([]string, 0)
	result = append(result, "# Replication")
	values := reflect.ValueOf(*(m.Copy()))
	types := values.Type()
	m.mutex.RUnlock()

	for i := 0; i < values.NumField(); i += 1 {
		field := types.Field(i)
		value := values.Field(i)
		val := fmt.Sprint(value.Interface())
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Name
		} else if tag == "-" {
			continue // ignore field tags with "-"
		}
		result = append(result, tag+":"+val)
	}
	return result
}

func (m *ServerMetadata) String() string {
	result := m.ToStringArray()
	res := ""
	for _, str := range result {
		res += str
	}
	return res
}

func (m *ServerMetadata) Copy() *ServerMetadataCopy {
	copy := ServerMetadataCopy{
		m.Role,
		m.ConnectedSlaves,
		m.MasterReplID,
		m.MasterReplOffset,
		m.SecondReplOffset,
		m.ReplBacklogActive,
		m.ReplBacklogSize,
		m.ReplBacklogFirstByteOffset,
		m.ReplBacklogHistlen,
	}
	return &copy
}

type ServerMetadataCopy struct {
	// Replication info
	Role                       string `json:"role"`
	ConnectedSlaves            int    `json:"connected_slaves"`
	MasterReplID               string `json:"master_replid"`
	MasterReplOffset           int64  `json:"master_repl_offset"`
	SecondReplOffset           int64  `json:"second_repl_offset"`
	ReplBacklogActive          int    `json:"repl_backlog_active"`
	ReplBacklogSize            int64  `json:"repl_backlog_size"`
	ReplBacklogFirstByteOffset int64  `json:"repl_backlog_first_byte_offset"`
	ReplBacklogHistlen         int64  `json:"repl_backlog_histlen"`
}
