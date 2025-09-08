package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

type InfoCommand struct {
}

// Execute implements Command.
func (i *InfoCommand) Execute(args []string, cache storage.Cache) string {
	return ""
}

// Validate implements Command.
func (i *InfoCommand) Validate(args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("wrong number of arguments for 'info' command")
	}
	return nil
}

func (i *InfoCommand) ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) []string {
	// convert server_metadata to string
	return []string{protocol.BuildBulkString(metadata.String())}
}
