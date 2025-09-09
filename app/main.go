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

func Send(conn net.Conn, command []any) {
	// Send data to the server
	_, err := conn.Write([]byte(protocol.BuildArray(command)))
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}
	fmt.Printf("Sent: %s", command[0])

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

func Handshake(address string, sInstancePort string) (net.Conn, error) {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return nil, err
	}
	// Step 1: PING
	command := []any{"PING"}
	Send(conn, command)

	// Step 2a: REPLCONF listening-port <PORT>
	command = []any{"REPLCONF", "listening-port", sInstancePort}
	Send(conn, command)

	// Step 2b: REPLCONF capa psync2
	command = []any{"REPLCONF", "capa", "psync2"}
	Send(conn, command)

	// Step 3: PSYNC ? -1 - Send but DON'T read response
	// The response (FULLRESYNC + RDB + future commands) will be handled by replication handler
	_, err = conn.Write([]byte(protocol.BuildArray([]any{"PSYNC", "?", "-1"})))
	if err != nil {
		fmt.Println("Error sending PSYNC:", err)
		return nil, err
	}
	fmt.Printf("Sent: PSYNC")

	return conn, nil
}

func ProcessServerMetadata() (*types.ServerMetadata, []string) {
	var portFlag, replicaOf, role, master_replid, masterPort string
	flag.StringVar(&portFlag, "port", "6379", "This flag is used for specifying the port")
	flag.StringVar(&replicaOf, "replicaof", "", "This flag is used to metion the master redis instance")
	flag.Parse()
	if replicaOf == "" {
		role = "master"
		master_replid = utility.GenerateRandomString(40)
	} else {
		addrs := strings.Split(replicaOf, " ")
		masterPort = addrs[1]
		// Handshake(addrs[0]+":"+addrs[1], portFlag)
		role = "slave"
	}
	metadata := types.NewServerMetadata(role)
	metadata.MasterReplID = master_replid

	return metadata, []string{masterPort, portFlag}
}

func main() {

	fmt.Println("Starting Redis server...")

	metadata, ports := ProcessServerMetadata()
	var redisServer *server.RedisServer

	// Create and start the Redis server
	if strings.ToLower(metadata.Role) == "slave" {
		addr := config.DefaultHost + ":" + ports[1]
		redisServer = server.NewRedisServer(addr, metadata)
		conn, err := Handshake(config.DefaultHost+":"+ports[0], ports[1])
		if err != nil {
			redisServer.Stop()
			os.Exit(1)
		}
		redisServer.StartSlave(conn)
	} else {
		redisServer = server.NewRedisServer(config.DefaultHost+":"+ports[1], metadata)
	}

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
