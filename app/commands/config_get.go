package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

type ConfigGetCommand struct{}

// Execute implements Command.
func (c *ConfigGetCommand) Execute(args []string, cache storage.Cache) string {
	return "This function shoudn't be called"
}

// Validate implements Command.
func (c *ConfigGetCommand) Validate(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("wrong number of arguments to `config get` command")
	}
	return nil
}

func (c *ConfigGetCommand) ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) []string {
	param := strings.ToLower(args[2])
	result := make([]string, 0)
	switch param {
	case "dir":
		result = append(result, param)
		result = append(result, metadata.Dir)
	case "dbfilename":
		result = append(result, param)
		result = append(result, metadata.DbFileName)
	default:
		result = append(result, "Invalid parameter")
	}
	return []string{protocol.BuildArray(utility.ConvertStringArrayToAny(result))}
}
