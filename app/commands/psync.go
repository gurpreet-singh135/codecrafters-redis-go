package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

type PSyncCommand struct {
}

// Execute implements Command.
func (p *PSyncCommand) Execute(args []string, cache storage.Cache) string {
	return "Accidently call execute of server aware command"
}

// Validate implements Command.
func (p *PSyncCommand) Validate(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("wrong number of arguments to 'psync' command")
	}
	return nil
}

func (p *PSyncCommand) ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) string {
	replicationId := metadata.MasterReplID
	offset := metadata.MasterReplOffset
	masterResponse := fmt.Sprintf("FULLRESYNC %s %d", replicationId, offset)
	return protocol.BuildSimpleString(masterResponse)
}
