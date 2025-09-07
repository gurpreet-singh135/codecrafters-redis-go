package protocol

import (
	"bufio"
	"log"
	"strconv"
	"strings"
)

// ParseRequest parses a RESP array request from the reader
func ParseRequest(reader *bufio.Reader) ([]string, error) {
	respRequest := []string{}
	numOfLine, err := reader.ReadString('\n')
	if err != nil {
		return RESP_ZERO_REQUEST, err
	}

	numOfLine = strings.TrimSpace(numOfLine)
	log.Println("numOfLine is: ", numOfLine)

	// RESP format check
	if !strings.HasPrefix(numOfLine, "*") {
		log.Println("invalid RESP command")
		return RESP_ZERO_REQUEST, nil
	}

	lines, err := strconv.Atoi(numOfLine[1:])
	if err != nil {
		log.Printf("Error converting to int: %v, input was: %q", err, numOfLine[1:])
		return RESP_ZERO_REQUEST, err
	}

	for i := 0; i < lines; i++ {
		// Read bulk string length line
		_, err := reader.ReadString('\n')
		if err != nil {
			return RESP_ZERO_REQUEST, err
		}

		// Read bulk string data
		command, err := reader.ReadString('\n')
		if err != nil {
			return RESP_ZERO_REQUEST, err
		}

		command = strings.TrimSpace(command)
		respRequest = append(respRequest, command)
	}

	return respRequest, nil
}

// BuildBulkString creates a RESP bulk string
func BuildBulkString(s string) string {
	if s == "" {
		return "$-1" + CRLF
	}
	s = strings.TrimSpace(s)
	strLen := len(s)
	return "$" + strconv.Itoa(strLen) + CRLF + s + CRLF
}

// BuildSimpleString creates a RESP simple string
func BuildSimpleString(s string) string {
	s = strings.TrimSpace(s)
	return "+" + s + CRLF
}

// BuildError creates a RESP error message
func BuildError(msg string) string {
	return "-ERR " + msg + CRLF
}

// BuildEmptyArray
func BuildEmptyArray() string {
	return "*0" + CRLF
}

// Build Null Array
func BuildNullArray() string {
	return "*-1" + CRLF
}

// Build Integer Response
func BuildInteger(s string) string {
	return ":" + s + CRLF
}

func BuildInt(s int) string {
	return ":" + strconv.Itoa(s) + CRLF
}

// Build RESP Array from raw values (original function)
func BuildArray(entries []any) string {
	length := len(entries)

	if length == 0 {
		return BuildNullArray()
	}

	resp := ""
	resp += "*" + strconv.Itoa(length) + CRLF

	for _, entry := range entries {
		switch v := entry.(type) {
		case []any:
			res := BuildArray(v)
			resp += res
		case string:
			// Format raw strings as bulk strings
			length := len(v)
			resp += "$" + strconv.Itoa(length) + CRLF
			resp += v + CRLF
		}
	}

	return resp
}

// BuildArrayFromResponses creates a RESP array from already-formatted RESP responses
// This is specifically for transaction EXEC results where each element is already a complete RESP response
func BuildArrayFromResponses(responses []string) string {
	if len(responses) == 0 {
		return BuildEmptyArray()
	}

	result := "*" + strconv.Itoa(len(responses)) + CRLF

	for _, response := range responses {
		// Each response is already a complete RESP-formatted string
		// Just concatenate them directly
		result += response
	}

	return result
}
