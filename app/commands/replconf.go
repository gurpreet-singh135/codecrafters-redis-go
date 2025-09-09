package commands

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

type ReplConfCommand struct{}

// Execute implements Command.
func (r *ReplConfCommand) Execute(args []string, cache storage.Cache) string {
	return "This REPLCONF method shouldn't be called"
}

// Validate implements Command.
func (r *ReplConfCommand) Validate(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("wrong number of arguments to 'replconf' command")
	}
	return nil
}

func (r *ReplConfCommand) ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) []string {
	log.Println("Inside the ExecuteWithMetadata of REPLCONF command")

	resp := make([]string, 0)
	if strings.ToUpper(args[1]) == "GETACK" {
		response := utility.ConvertStringArrayToAny([]string{"REPLCONF", "ACK", fmt.Sprintf("%d", metadata.CommandProcessed)})
		resp = append(resp, protocol.BuildArray(response))
		return resp
	} else if strings.ToUpper(args[1]) == "ACK" {
		// This is an ACK response from a replica - process it for WAIT commands
		offsetStr := args[2]
		offset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err == nil {
			// We need the connection ID, but we don't have access to the connection here
			// This will be handled in the handler.go when processing REPLCONF ACK
			log.Printf("Received ACK with offset: %d", offset)
		}
		// Don't return a response for ACK (replicas don't expect a response to their ACK)
		return []string{}
	}
	resp = append(resp, protocol.BuildSimpleString(protocol.RESPONSE_OK))
	log.Printf("REPLCONF returning response: %v", resp)
	return resp
}
