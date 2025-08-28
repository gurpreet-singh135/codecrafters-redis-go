package server

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// RedisServer represents the Redis server instance
type RedisServer struct {
	address  string
	cache    storage.Cache
	registry *commands.CommandRegistry
}

// NewRedisServer creates a new Redis server instance
func NewRedisServer(address string) *RedisServer {
	return &RedisServer{
		address:  address,
		cache:    storage.NewCache(),
		registry: commands.NewCommandRegistry(),
	}
}

// Start starts the Redis server and begins accepting connections
func (s *RedisServer) Start() {
	fmt.Println("Starting Redis server on", s.address)
	
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		fmt.Printf("Failed to bind to %s: %v\n", s.address, err)
		os.Exit(1)
	}
	defer listener.Close()
	
	fmt.Printf("Redis server listening on %s\n", s.address)
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue // Don't exit on connection errors, just log and continue
		}
		
		// Handle each connection in a separate goroutine
		go NewConnectionHandler(conn, s.cache, s.registry).Handle()
	}
}

// Stop gracefully stops the Redis server
func (s *RedisServer) Stop() {
	fmt.Println("Stopping Redis server...")
	// TODO: Implement graceful shutdown
}