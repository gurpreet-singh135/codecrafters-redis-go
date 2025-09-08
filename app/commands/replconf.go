package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type ReplConfCommand struct{}

// Execute implements Command.
func (r *ReplConfCommand) Execute(args []string, cache storage.Cache) string {
	return protocol.BuildSimpleString(protocol.RESPONSE_OK)
}

// Validate implements Command.
func (r *ReplConfCommand) Validate(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("wrong number of arguments to 'replconf' command")
	}
	return nil
}
