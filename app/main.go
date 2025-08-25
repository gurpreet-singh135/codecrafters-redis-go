package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	CLRF = "\r\n"
)

var (
	store = make(map[string]string)
	storeMutex sync.RWMutex
)


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

func buildRESPBulkString(s string) string {
	if s == "" {
		return "$" + "-1" + CLRF
	}
	s = strings.TrimSpace(s)
	strLen := len(s)
	RESPString := "$" + strconv.Itoa(strLen) + CLRF + s + CLRF;
	return RESPString
}

func buildRESPSimpleString(s string) string {
	s = strings.TrimSpace(s)
	RESPString := "+" + s + CLRF;
	return RESPString	
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
			// _, _ = reader.ReadString('\n')
			// count, _ := strconv.Atoi(line[1:])
			
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
				_, _ = reader.ReadString('\n')
				echoStr, _ := reader.ReadString('\n')
				echoStr = strings.TrimSpace(echoStr)

				conn.Write([]byte(buildRESPBulkString(echoStr)))
			case "GET":
				log.Println("GET command is executed")
				_, _ = reader.ReadString('\n')
				key, _ := reader.ReadString('\n')
				key = strings.TrimSpace(key)
				storeMutex.RLock()
				value := store[key]

				log.Println("GET: key and value are: ", key, value)

				conn.Write([]byte(buildRESPBulkString(value)))
				storeMutex.RUnlock()
			case "SET":
				_, _ = reader.ReadString('\n')
				key, _ := reader.ReadString('\n')
				key = strings.TrimSpace(key)
				_, _ = reader.ReadString('\n')
				value, _ := reader.ReadString('\n')
				value = strings.TrimSpace(value)

				log.Println("GET: key and value are: ", key, value)
				
				storeMutex.Lock()
				store[key] = value
				storeMutex.Unlock()

				conn.Write([]byte(buildRESPSimpleString("OK")))
			}
		}
	}
}