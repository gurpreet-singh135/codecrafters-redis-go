package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CLRF = "\r\n"
	RESPONSE_OK = "OK"
	RESPONSE_NONE = "none"
)

var (
	RESP_ZERO_REQUEST = []string{}
)

type Value struct {
	value string
	expiration time.Time
}

var (
	store = make(map[string]Value)
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


func parseRESPRequest(reader *bufio.Reader) []string {
	respRequest := []string{}
	numOfLine, err := reader.ReadString('\n')
	numOfLine = strings.TrimSpace(numOfLine)
	log.Println("numOfLine is: ", numOfLine)
	
	// RESP format check
	if err != nil || !strings.HasPrefix(numOfLine, "*") {
		log.Println("invalid RESP command")
		return RESP_ZERO_REQUEST 
	}

	lines, err := strconv.Atoi(numOfLine[1:])
	if err != nil {
    log.Printf("Error converting to int: %v, input was: %q", err, numOfLine[1:])
		return RESP_ZERO_REQUEST
	}


	for i := 0; i < lines; i += 1 {
		_, err := reader.ReadString('\n')
		if err != nil {
			return RESP_ZERO_REQUEST 
		}

		command, er := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if er != nil {
			return RESP_ZERO_REQUEST 
		}

		respRequest = append(respRequest, command)
	}

	return respRequest
}

func isValueExpired(value Value, currentTime time.Time) bool {
	return !value.expiration.IsZero() && currentTime.Compare(value.expiration) >= 0 
}



func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Handling Connection")

	var respRequest []string
	reader := bufio.NewReader(conn)
	
	for {
		// currentTime := time.Now()
		respRequest = parseRESPRequest(reader)
		log.Println("RESP Request, ", respRequest)

		if len(respRequest) == 0 {
			// Client disconnected or no more commands
			log.Println("Connection closed or no data received")
			break
		}

		command := respRequest[0]

		switch command {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			echoStr := respRequest[1]
			conn.Write([]byte(buildRESPBulkString(echoStr)))
		case "GET":
			log.Println("GET command is executed")
			key := respRequest[1]
			var response string

			storeMutex.RLock()	
			value, ok := store[key]
			if ok {
				needsDelete := isValueExpired(value, time.Now()) 
				storeMutex.RUnlock()

				if needsDelete {
					storeMutex.Lock()
					if value2, exists := store[key]; exists && isValueExpired(value2, time.Now()) {
						delete(store, key)
						// response remains empty string for expired key

					} else {
						response = value2.value
					}
					// If !exists, key was deleted by another goroutine - response stays empty
					storeMutex.Unlock()
				} else {
					response = value.value
				}
			} else {
				storeMutex.Unlock()
				// response remains empty string for non-existent key
			}
			log.Println("GET: key and value are: ", key, value)
			conn.Write([]byte(buildRESPBulkString(response)))
		case "SET":
			key := respRequest[1]
			value := respRequest[2]
			var expirationTime time.Time
			if len(respRequest) > 4 {
				// we have to parse expiration as well
				expirationDuration := respRequest[4]
				expDelta, _ := strconv.Atoi(expirationDuration)
				expirationTime = time.Now().Add(time.Duration(expDelta * int(time.Millisecond)))
			}
			storeMutex.Lock()
			store[key] = Value{value, expirationTime}
			storeMutex.Unlock()

			conn.Write([]byte(buildRESPSimpleString(RESPONSE_OK)))
		case "TYPE":
			var response string
			key := respRequest[1]
			storeMutex.RLock()
			value, exists := store[key]
			if !exists {
				response = RESPONSE_NONE
			} else {
				response = reflect.TypeOf(value.value).String()
			}
			storeMutex.RUnlock()

			conn.Write([]byte(buildRESPSimpleString(response)))
			
		default:
			conn.Write([]byte("-ERR unknown command\r\n"))
		}

	}
}