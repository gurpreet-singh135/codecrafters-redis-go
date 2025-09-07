package main

import (
	"flag"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting Redis server...")
	var portFlag string
	flag.StringVar(&portFlag, "port", "6379", "This flag is used for specifying the port")

	flag.Parse()

	// Create and start the Redis server
	redisServer := server.NewRedisServer(config.DefaultHost + ":" + portFlag)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("\nReceived shutdown signal")
		redisServer.Stop()
		os.Exit(0)
	}()

	// Start the server (this blocks)
	redisServer.Start()
}
