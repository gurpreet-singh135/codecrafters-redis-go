package types

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

// AckResponse represents a REPLCONF ACK response from a replica
type AckResponse struct {
	ConnID string
	Offset int64
}

// WaitRequest represents an active WAIT command
type WaitRequest struct {
	ID            string
	TargetOffset  int64
	RequiredCount int
	ResponseChan  chan int
	ReceivedAcks  map[string]int64 // ConnID -> Offset
	StartTime     time.Time
}

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
	Dir                        string        `json:"-"`
	DbFileName                 string        `json:"-"`

	// WAIT command support
	AckResponseChannel chan AckResponse        `json:"-"`
	WaitRequests       map[string]*WaitRequest `json:"-"`
	ConnIDMap          map[net.Conn]string     `json:"-"`

	mutex sync.RWMutex
}

func (m *ServerMetadata) Replicate(Cmd []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	activeConnections := make([]net.Conn, 0)

	// Calculate the byte size of the command being replicated
	cmdBytes := protocol.BuildArray(utility.ConvertStringArrayToAny(Cmd))
	cmdSize := int64(len(cmdBytes))

	for _, conn := range m.ReplActiveConnection {
		err := m.Send(conn, Cmd)
		if err == nil {
			activeConnections = append(activeConnections, conn)
			// removing inactive connection
		}
	}
	m.ReplActiveConnection = activeConnections

	// Update master offset only if we successfully sent to at least one replica
	if len(activeConnections) > 0 {
		m.MasterReplOffset += cmdSize
	}
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

// UpdateMasterOffset updates the master replication offset
func (m *ServerMetadata) UpdateMasterOffset(bytes int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.MasterReplOffset += bytes
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

	// Generate connection ID for this replica
	connID := fmt.Sprintf("replica-%p", conn)
	m.ConnIDMap[conn] = connID
}

// GetConnectionID returns the connection ID for a given connection
func (m *ServerMetadata) GetConnectionID(conn net.Conn) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.ConnIDMap[conn]
}

// RegisterWaitRequest registers a new WAIT request
func (m *ServerMetadata) RegisterWaitRequest(req *WaitRequest) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.WaitRequests[req.ID] = req
}

// UnregisterWaitRequest removes a WAIT request
func (m *ServerMetadata) UnregisterWaitRequest(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.WaitRequests, id)
}

// SendAckResponse sends an ACK response to the channel
func (m *ServerMetadata) SendAckResponse(connID string, offset int64) {
	ackResp := AckResponse{
		ConnID: connID,
		Offset: offset,
	}

	select {
	case m.AckResponseChannel <- ackResp:
	default:
		// Channel full, log warning but don't block
		log.Printf("Warning: ACK response channel full, dropping ACK from %s", connID)
	}
}

// ProcessAckResponses processes incoming ACK responses from replicas
func (m *ServerMetadata) ProcessAckResponses() {
	for ackResp := range m.AckResponseChannel {
		m.mutex.Lock()

		// Check all active wait requests
		for _, waitReq := range m.WaitRequests {
			// Check if this ACK satisfies the wait condition
			if ackResp.Offset >= waitReq.TargetOffset {
				// Add to received ACKs if not already present or update with higher offset
				if existingOffset, exists := waitReq.ReceivedAcks[ackResp.ConnID]; !exists || ackResp.Offset > existingOffset {
					waitReq.ReceivedAcks[ackResp.ConnID] = ackResp.Offset

					// Check if we have enough ACKs
					if len(waitReq.ReceivedAcks) >= waitReq.RequiredCount {
						select {
						case waitReq.ResponseChan <- len(waitReq.ReceivedAcks):
						default:
						}
					}
				}
			}
		}

		m.mutex.Unlock()
	}
}

func NewServerMetadata(role string) *ServerMetadata {
	// ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	metadata := ServerMetadata{
		Role:                 role,
		ReplChannel:          make(chan []string, 1_000),
		ReplActiveConnection: make([]net.Conn, 0),
		AckResponseChannel:   make(chan AckResponse, 100),
		WaitRequests:         make(map[string]*WaitRequest),
		ConnIDMap:            make(map[net.Conn]string),
	}
	go metadata.ReplicateCommandToReplicas()
	go metadata.ProcessAckResponses()
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
