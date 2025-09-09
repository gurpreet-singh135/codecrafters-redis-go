package commands

import (
	"fmt"
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
	resp := make([]string, 0)
	if strings.ToUpper(args[1]) == "GETACK" {
		response := utility.ConvertStringArrayToAny([]string{"REPLCONF", "ACK", fmt.Sprintf("%d", metadata.CommandProcessed)})
		resp = append(resp, protocol.BuildArray(response))
		return resp
	}
	resp = append(resp, protocol.BuildSimpleString(protocol.RESPONSE_OK))
	return resp
}
