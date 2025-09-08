package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

func ConnectToPrimaryInstance(address string) {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close() // Close the connection when the main function exits

	// Send data to the server
	message := make([]any, 0)
	message = append(message, "PING")
	_, err = conn.Write([]byte(protocol.BuildArray(message)))
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}
	fmt.Printf("Sent: %s", message)

	// Receive response from the server
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	response := string(buffer[:n])
	fmt.Printf("Received: %s", response)
}

func ProcessServerMetadata() (*types.ServerMetadata, string) {
	var portFlag, replicaOf, role, master_replid string
	flag.StringVar(&portFlag, "port", "6379", "This flag is used for specifying the port")
	flag.StringVar(&replicaOf, "replicaof", "", "This flag is used to metion the master redis instance")
	flag.Parse()
	if replicaOf == "" {
		role = "master"
		master_replid = utility.GenerateRandomString(40)
	} else {
		addrs := strings.Split(replicaOf, " ")
		ConnectToPrimaryInstance(addrs[0] + ":" + addrs[1])
		role = "slave"
	}
	metadata := types.NewServerMetadata(role)
	metadata.MasterReplID = master_replid

	return metadata, config.DefaultHost + ":" + portFlag
}

func main() {
	fmt.Println("Starting Redis server...")

	metadata, address := ProcessServerMetadata()
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
