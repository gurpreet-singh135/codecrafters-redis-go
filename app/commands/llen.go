package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type LLenCommand struct{}

// Execute implements Command.
func (l *LLenCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]

	redisValue, ok := cache.Get(key)
	if !ok {
		return protocol.BuildInteger("0")
	}
	listValue, ok := redisValue.(*storage.ListValue)
	if !ok {
		return protocol.BuildError("wrongtype for listValue")
	}
	return protocol.BuildInteger(strconv.Itoa(listValue.Size()))
}

// Validate implements Command.
func (l *LLenCommand) Validate(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("wrong number of arguments to LLEN command")
	}
	return nil
}
