package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

type WaitCommand struct{}

// Execute implements Command.
func (w *WaitCommand) Execute(args []string, cache storage.Cache) string {
	return "This method of `wait` shouldn't be called"
}

// Validate implements Command.
func (w *WaitCommand) Validate(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("wrong number of arguments to 'wait' command")
	}
	return nil
}

func (w *WaitCommand) ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) []string {
	resp := make([]string, 0)
	resp = append(resp, protocol.BuildInt(metadata.NumberOfActiveConnections()))
	return resp
}
