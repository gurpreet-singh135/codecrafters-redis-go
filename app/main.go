package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	fmt.Println("Starting Redis server...")

	// Create and start the Redis server
	redisServer := server.NewRedisServer(config.DefaultAddress)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nReceived shutdown signal")
		redisServer.Stop()
		os.Exit(0)
	}()

	// Start the server (this blocks)
	redisServer.Start()
}
