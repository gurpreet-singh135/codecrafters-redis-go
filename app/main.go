package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}

}


func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Handling Connection")

	reader := bufio.NewReader(conn)
	
	for {
		// Read until \r\n
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Connection closed: %v\n", err)
			break
		}

		log.Println("Value of first line is: ", line)
		
		line = strings.TrimSpace(line)
		fmt.Printf("Data Received: %s\n", line)
		
		// Parse RESP array
		if strings.HasPrefix(line, "*") {
			count, _ := strconv.Atoi(line[1:])
			
			for i := 0; i < count; i++ {
				// Read bulk string length
				lengthLine, _ := reader.ReadString('\n')
				lengthLine = strings.TrimSpace(lengthLine)
				fmt.Printf("Length: %s\n", lengthLine)
				
				// Read bulk string data
				dataLine, _ := reader.ReadString('\n')
				command := strings.TrimSpace(dataLine)
				fmt.Printf("Command: %s\n", command)
				
				switch command {
				case "PING":
					conn.Write([]byte("+PONG\r\n"))
				case "ECHO":
					lenEchoStr, _ := reader.ReadString('\n')
					lenEchoStr = strings.TrimSpace(lenEchoStr)
					log.Println("Length of the echoed string: ", lenEchoStr)
					echoStr, _ := reader.ReadString('\n')
					echoStr = strings.TrimSpace(echoStr)

					conn.Write([]byte(lenEchoStr + "\r\n" + echoStr + "\r\n"))
				}
			}
		}
	}
}