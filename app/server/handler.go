package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
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
	}
}

func (h *ConnectionHandler) IsGetAck(respRequest []string) bool {
	if len(respRequest) != 3 {
		return false
	}
	return strings.ToUpper(respRequest[0]) == "REPLCONF" && strings.ToUpper(respRequest[1]) == "GETACK"
}

// Handle processes commands from the client connection
func (h *ConnectionHandler) Handle() {
	defer h.conn.Close()
	fmt.Printf("New connection from %s\n", h.conn.RemoteAddr())

	for {
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

		// if len(respRequest) == 0 {
		// 	log.Println("Empty request received")
		// 	break
		// }

		// Execute command
		command := strings.ToUpper(respRequest[0])
		response := h.processCommand(command, respRequest)
		// response := h.registry.Execute(command, respRequest, h.cache)
		if command == "PSYNC" {
			// add connection to replicas
			h.AddReplicasConnection()
		}

		// Send response
		for _, res := range response {
			if h.isReplicationConn {
				h.metadata.AddCommandProcessed(n)
				if !h.IsGetAck(respRequest) {
					continue
				}
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
	switch cmdName {
	case "MULTI":
		return h.processMultiCommand()
	case "EXEC":
		return h.processExecCommand()
	case "DISCARD":
		return h.processDiscardCommand()
	}

	Command, err := h.registry.GetCommand(cmdName)
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
	// Always consume exactly 88 bytes (empty RDB file size)
	rdbData := make([]byte, 88)
	_, err := io.ReadFull(h.reader, rdbData)
	if err != nil {
		log.Printf("Error consuming empty RDB: %v", err)
		return
	}
	log.Printf("Consumed empty RDB file (88 bytes)")
}
