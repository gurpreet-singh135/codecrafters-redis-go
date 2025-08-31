package server

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// ConnectionHandler handles individual client connections
type ConnectionHandler struct {
	conn     net.Conn
	cache    storage.Cache
	registry *commands.CommandRegistry
	reader   *bufio.Reader
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(conn net.Conn, cache storage.Cache, registry *commands.CommandRegistry) *ConnectionHandler {
	return &ConnectionHandler{
		conn:     conn,
		cache:    cache,
		registry: registry,
		reader:   bufio.NewReader(conn),
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
		command := respRequest[0]
		response := h.registry.Execute(command, respRequest, h.cache)

		// Send response
		_, err = h.conn.Write([]byte(response))
		if err != nil {
			log.Printf("Error writing response: %v", err)
			break
		}
	}

	fmt.Printf("Connection from %s closed\n", h.conn.RemoteAddr())
}
