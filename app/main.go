package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

func processServerMetadata() (*types.ServerMetadata, string) {
	var portFlag, replicaOf, role string
	flag.StringVar(&portFlag, "port", "6379", "This flag is used for specifying the port")
	flag.StringVar(&replicaOf, "replicaof", "", "This flag is used to metion the master redis instance")
	flag.Parse()
	if replicaOf == "" {
		role = "master"
	} else {
		role = "slave"
	}
	metadata := types.NewServerMetadata(role)

	return metadata, config.DefaultHost + ":" + portFlag
}

func main() {
	fmt.Println("Starting Redis server...")

	metadata, address := processServerMetadata()
	flag.Parse()

	// Create and start the Redis server
	redisServer := server.NewRedisServer(address, metadata)

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

// func main_back() {
// 	fmt.Printf("vales of Server's Metadata is:\n %v", types.NewServerMetadata("master").String())
// }
