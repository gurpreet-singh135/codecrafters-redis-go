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
	"time"
)

const (
	CRLF = "\r\n"
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

type RedisValue interface {
	Type() string
	IsExpired(currentTime time.Time) bool
}

type StringValue struct {
	val string
	expiration time.Time
}

func (s StringValue) Type() string {
	return "string"
}

func (s StringValue) IsExpired(currentTime time.Time) bool {
	return !s.expiration.IsZero() && currentTime.Compare(s.expiration) >= 0
}

type StreamEntry struct {
	ID string
	Fields map[string]string	
}

type StreamValue struct {
	entries []StreamEntry
}

func (entry StreamValue) IsExpired(currentTime time.Time) bool {
	return false
}

func (entry StreamValue) Type() string {
	return "stream" 
} 

type Cache struct {
	data map[string]RedisValue
	mu sync.RWMutex
}

func NewCache() *Cache {
    return &Cache{
        data: make(map[string]RedisValue),
    }
}


var (
	store = NewCache()
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
		return "$" + "-1" + CRLF
	}
	s = strings.TrimSpace(s)
	strLen := len(s)
	RESPString := "$" + strconv.Itoa(strLen) + CRLF + s + CRLF
	return RESPString
}

func buildRESPSimpleString(s string) string {
	s = strings.TrimSpace(s)
	RESPString := "+" + s + CRLF;
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
		  currentTime := time.Now()
			log.Println("GET command is executed")
			key := respRequest[1]
			var response string

			store.mu.RLock()	
			value, exists := store.data[key]
			if exists && value.IsExpired(currentTime) {
				log.Println("Key exists, but might have expired")
				store.mu.RUnlock()
				store.mu.Lock()
				log.Println("Key exists, but has expired")
				if val2, stillExists := store.data[key]; stillExists {
					if val2.IsExpired(currentTime) {
						delete(store.data, key)
					} else if sv, ok := val2.(StringValue); ok {
						response = sv.val
					}
				store.mu.Unlock()
				}
			} else if exists {
				log.Println("Key exists, but didn't expired")
				if sv, ok := value.(StringValue); ok {
					response = sv.val
				}
				store.mu.RUnlock()
			} else {
				log.Println("Key does not exists")
				// key doesn't exist
				store.mu.RUnlock()
			}
			conn.Write([]byte(buildRESPBulkString(response)))

		case "SET":
			key := respRequest[1]
			value := respRequest[2]
			var expirationTime time.Time
			if len(respRequest) > 4 {
				// we have to parse expiration as well
				expirationDuration := respRequest[4]
				expDelta, _ := strconv.Atoi(expirationDuration)
				expirationTime = time.Now().Add(time.Duration(expDelta) * time.Millisecond)
			}
			store.mu.Lock()
			store.data[key] = StringValue{value, expirationTime}
			store.mu.Unlock()

			conn.Write([]byte(buildRESPSimpleString(RESPONSE_OK)))

		case "TYPE":
			var response string
			key := respRequest[1]
			store.mu.RLock()
			value, exists := store.data[key]
			if !exists {
				response = RESPONSE_NONE
			} else {
				response = value.Type() 
			}
			store.mu.RUnlock()
			conn.Write([]byte(buildRESPSimpleString(response)))

		case "XADD":
			streamKey := respRequest[1]
			streamID := respRequest[2]
			streamFields := map[string]string{}

			for i := 3; i < len(respRequest); i += 2 {
				key := respRequest[i]
				value := respRequest[i + 1]
				streamFields[key] = value
			}

			streamEntry := StreamEntry{streamID, streamFields}

			store.mu.Lock()
			existingEntries, exists := store.data[streamKey]
			if se, ok := existingEntries.(StreamValue); exists {
				if ok {
					se.entries = append(se.entries, streamEntry)
				} else {
					log.Fatal("duplicate key with stream and other type")
				}
			} else {
				store.data[streamKey] = StreamValue{[]StreamEntry{streamEntry}}
			}
			store.mu.Unlock()
			conn.Write([]byte(buildRESPBulkString(streamID)))

		default:
			conn.Write([]byte("-ERR unknown command\r\n"))
		}

	}
}