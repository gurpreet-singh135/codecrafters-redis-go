package commands

import (
	"encoding/hex"
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

func (p *PSyncCommand) ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) []string {
	replicationId := metadata.MasterReplID
	offset := metadata.MasterReplOffset
	masterResponse := protocol.BuildSimpleString(fmt.Sprintf("FULLRESYNC %s %d", replicationId, offset))
	emptyRDBFileHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	binaryData, err := hex.DecodeString(emptyRDBFileHex)
	if err != nil {
		return []string{protocol.BuildError("Failed to decode RDB file")}
	}
	RDBResponse := fmt.Sprintf("$%d\r\n%s", len(binaryData), binaryData)
	return []string{masterResponse, RDBResponse}
}
