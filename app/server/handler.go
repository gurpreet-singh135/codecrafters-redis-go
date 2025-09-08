package server

import (
	"bufio"
	"fmt"
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
	conn             net.Conn
	cache            storage.Cache
	registry         *commands.CommandRegistry
	reader           *bufio.Reader
	transactionState *commands.TransactionState
	metadata         *types.ServerMetadata
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(conn net.Conn, cache storage.Cache, registry *commands.CommandRegistry, metadata *types.ServerMetadata) *ConnectionHandler {
	return &ConnectionHandler{
		conn:             conn,
		cache:            cache,
		registry:         registry,
		reader:           bufio.NewReader(conn),
		transactionState: commands.NewTransactionState(),
		metadata:         metadata,
	}
}

// Handle processes commands from the client connection
func (h *ConnectionHandler) Handle() {
	defer h.conn.Close()
	fmt.Printf("New connection from %s\n", h.conn.RemoteAddr())

	for {
		// Parse RESP request
		respRequest, err := protocol.ParseRequest(h.reader)
		if err != nil {
			log.Printf("Connection closed or error parsing request: %v", err)
			break
		}

		log.Printf("RESP Request: %v", respRequest)

		if len(respRequest) == 0 {
			log.Println("Empty request received")
			break
		}

		// Execute command
		command := strings.ToUpper(respRequest[0])
		response := h.processCommand(command, respRequest)
		// response := h.registry.Execute(command, respRequest, h.cache)

		// Send response
		for _, res := range response {
			_, err = h.conn.Write([]byte(res))
			if err != nil {
				log.Printf("Error writing response: %v", err)
				break
			}
		}
	}

	fmt.Printf("Connection from %s closed\n", h.conn.RemoteAddr())
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

	return queueCommand.Execute(h.cache)
}

func (h *ConnectionHandler) processMultiCommand() []string {
	if h.transactionState.IsInTransaction() {
		return []string{protocol.BuildError(protocol.MULTI_IN_MULTI)}
	}

	h.transactionState.StartTransaction()
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

	// Use the new function specifically designed for already-formatted RESP responses
	return []string{protocol.BuildArrayFromResponses(res)}
}

func (h *ConnectionHandler) processDiscardCommand() []string {
	if !h.transactionState.IsInTransaction() {
		return []string{protocol.BuildError(protocol.DISCARD_WITHOUT_MULTI)}
	}

	h.transactionState.Reset()
	return []string{protocol.BuildSimpleString(protocol.RESPONSE_OK)}
}
