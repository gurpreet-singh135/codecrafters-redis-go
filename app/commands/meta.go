package commands

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// TypeCommand implements the TYPE command
type TypeCommand struct{}

func (c *TypeCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]
	dataType := cache.Type(key)
	return protocol.BuildSimpleString(dataType)
}

func (c *TypeCommand) Validate(args []string) error {
	if len(args) < 2 {
		return errors.New("wrong number of arguments for 'type' command")
	}
	return nil
}