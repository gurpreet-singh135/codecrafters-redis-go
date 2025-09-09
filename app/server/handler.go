package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

// ConnectionHandler handles individual client connections
type ConnectionHandler struct {
	conn              net.Conn
	cache             storage.Cache
	registry          *commands.CommandRegistry
	reader            *bufio.Reader
	transactionState  *commands.TransactionState
	metadata          *types.ServerMetadata // Metadata should have a list of active connections, it should get populated when PSYNC command is a success
	isReplicationConn bool
	connectionID      string
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(conn net.Conn, cache storage.Cache, registry *commands.CommandRegistry, metadata *types.ServerMetadata) *ConnectionHandler {
	return &ConnectionHandler{
		conn:              conn,
		cache:             cache,
		registry:          registry,
		reader:            bufio.NewReader(conn),
		transactionState:  commands.NewTransactionState(),
		metadata:          metadata,
		isReplicationConn: false,
		connectionID:      fmt.Sprintf("conn-%p", conn),
	}
}

func (h *ConnectionHandler) IsGetAck(respRequest []string) bool {
	if len(respRequest) != 3 {
		return false
	}
	log.Printf("IsGetAck method: %+v", respRequest)
	return strings.ToUpper(respRequest[0]) == "REPLCONF" && strings.ToUpper(respRequest[1]) == "GETACK"
}

// Handle processes commands from the client connection
func (h *ConnectionHandler) Handle() {
	// defer h.conn.Close()
	defer func() {
		log.Printf("Connection handler exiting for %s, isReplicationConn=%v",
			h.conn.RemoteAddr(), h.isReplicationConn)
		h.conn.Close()
	}()
	fmt.Printf("New connection from %s\n", h.conn.RemoteAddr())

	for {
		log.Printf("Waiting for next command...") // Add this
		// Parse RESP request
		respRequest, err, n := protocol.ParseRequest(h.reader)
		if err != nil {
			log.Printf("Connection closed or error parsing request: %v", err)
			break
		}

		if len(respRequest) == 0 {
			h.consumeEmptyRDB()
			continue
		}

		log.Printf("RESP Request: %v", respRequest)

		// Execute command
		command := strings.ToUpper(respRequest[0])
		response := h.processCommand(command, respRequest)

		// Handle REPLCONF ACK responses for WAIT commands
		if command == "REPLCONF" && len(respRequest) == 3 && strings.ToUpper(respRequest[1]) == "ACK" {
			offsetStr := respRequest[2]
			if offset, err := strconv.ParseInt(offsetStr, 10, 64); err == nil {
				connID := h.metadata.GetConnectionID(h.conn)
				if connID == "" {
					connID = h.connectionID
				}
				h.metadata.SendAckResponse(connID, offset)
			}
		}
		// response := h.registry.Execute(command, respRequest, h.cache)
		if command == "PSYNC" {
			h.AddReplicasConnection()

			// Send FULLRESYNC response first
			for _, res := range response {
				_, err = h.conn.Write([]byte(res))
				if err != nil {
					log.Printf("Error writing PSYNC response: %v", err)
					break
				}
			}

			// Send raw RDB data
			psyncCmd := &commands.PSyncCommand{}
			rdbData, err := psyncCmd.GetRDBData()
			if err != nil {
				log.Printf("Error getting RDB data: %v", err)
				break
			}

			// Send RDB with proper bulk string header + raw binary
			rdbHeader := fmt.Sprintf("$%d\r\n", len(rdbData))
			h.conn.Write([]byte(rdbHeader))
			h.conn.Write(rdbData)

			continue // Skip normal response processing
		}

		// Send response
		for _, res := range response {
			log.Printf("Response loop: isReplicationConn=%v, IsGetAck=%v, response=%s",
				h.isReplicationConn, h.IsGetAck(respRequest), res)
			shouldSendResponse := true

			if h.isReplicationConn {
				if h.IsGetAck(respRequest) {
					// GETACK: always send response, don't count bytes
					h.metadata.AddCommandProcessed(n)
					shouldSendResponse = true
				} else if strings.ToUpper(respRequest[0]) == "REPLCONF" {
					// Other REPLCONF: send response, don't count bytes
					shouldSendResponse = true
					h.metadata.AddCommandProcessed(n)
				} else {
					// Data commands: count bytes, don't send response
					h.metadata.AddCommandProcessed(n)
					shouldSendResponse = false
				}
			}

			if !shouldSendResponse {
				continue
			}
			_, err = h.conn.Write([]byte(res))
			if err != nil {
				log.Printf("Error writing response: %v", err)
				break
			}
		}
	}

	fmt.Printf("Connection from %s closed\n", h.conn.RemoteAddr())
}

func (h *ConnectionHandler) AddReplicasConnection() {
	h.metadata.AddReplicasConnection(h.conn)
}

func (h *ConnectionHandler) processCommand(cmdName string, args []string) []string {
	if cmdName == "REPLCONF" && !h.isReplicationConn {
		h.isReplicationConn = true
		log.Printf("Detected replication connection from %s", h.conn.RemoteAddr())
	}
	switch cmdName {
	case "MULTI":
		return h.processMultiCommand()
	case "EXEC":
		return h.processExecCommand()
	case "DISCARD":
		return h.processDiscardCommand()
	}

	Command, err := h.registry.GetCommand(cmdName)
	log.Printf("Command being processed is: %s", cmdName)
	if err != nil {
		return []string{protocol.BuildError("Invalid command")}
	}

	if h.transactionState.IsInTransaction() {
		// QUEUE commands
		err := h.transactionState.QueueCommand(Command, args)
		if err != nil {
			return []string{protocol.BuildError(err.Error())}
		}
		return []string{protocol.BuildSimpleString(protocol.RESPONSE_QUEUED)}
	}

	// Otherwise execute them
	queueCommand := commands.QueueCommand{
		Cmd:       Command,
		Args:      args,
		Timestamp: time.Now().UnixNano(),
		Metadata:  h.metadata,
	}
	// Here we can send the command to replica
	h.SendCommandToReplicas(args)
	return queueCommand.Execute(h.cache)
}

func (h *ConnectionHandler) processMultiCommand() []string {
	if h.transactionState.IsInTransaction() {
		return []string{protocol.BuildError(protocol.MULTI_IN_MULTI)}
	}

	h.transactionState.StartTransaction()
	// Send Command to Replicas from here
	h.SendCommandToReplicas([]string{"MULTI"})
	return []string{protocol.BuildSimpleString("OK")}
}

func (h *ConnectionHandler) processExecCommand() []string {
	if !h.transactionState.IsInTransaction() {
		return []string{protocol.BuildError(protocol.EXEC_BEFORE_MULTI)}
	}

	result := h.transactionState.ExecuteTransaction(h.cache)
	res := make([]string, 0, len(result))
	for _, val := range result {
		res = append(res, val...)
	}
	h.transactionState.EndTransaction()

	// Send the Command to Replicas from here
	h.SendCommandToReplicas([]string{"EXEC"})
	// Use the new function specifically designed for already-formatted RESP responses
	return []string{protocol.BuildArrayFromResponses(res)}
}

func (h *ConnectionHandler) processDiscardCommand() []string {
	if !h.transactionState.IsInTransaction() {
		return []string{protocol.BuildError(protocol.DISCARD_WITHOUT_MULTI)}
	}
	// Send to Replicas
	h.transactionState.Reset()
	h.SendCommandToReplicas([]string{"DISCARD"})
	return []string{protocol.BuildSimpleString(protocol.RESPONSE_OK)}
}

func (h *ConnectionHandler) SendCommandToReplicas(Cmd []string) {
	if h.metadata.Role == "master" && h.shouldReplicate(Cmd[0]) {
		h.metadata.ReplChannel <- Cmd
	}
}

func (h *ConnectionHandler) shouldReplicate(cmdName string) bool {
	writeCommands := map[string]bool{
		"SET": true, "DEL": true, "INCR": true, "DECR": true,
		"LPUSH": true, "RPUSH": true, "LPOP": true, "RPOP": true,
		"XADD": true, "MULTI": true, "EXEC": true, "DISCARD": true,
	}
	return writeCommands[cmdName]
}

func (h *ConnectionHandler) consumeEmptyRDB() {
	// Read bytes until we find the start of next RESP command
	for {
		b, err := h.reader.ReadByte()
		if err != nil {
			log.Printf("Error consuming RDB: %v", err)
			return
		}
		if b == '*' {
			// Put the '*' back for the next ParseRequest
			h.reader.UnreadByte()
			break
		}
	}
	log.Printf("Consumed RDB file and found next command")
}
