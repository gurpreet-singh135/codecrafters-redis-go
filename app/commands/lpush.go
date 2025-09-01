package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type LPushCommand struct{}

func (l *LPushCommand) PrependToList(listValue *storage.ListValue, argsValues []string) {
	for _, value := range argsValues {
		listValue.Prepend(storage.NewListItem(value))
	}
}

// Execute implements Command.
func (l *LPushCommand) Execute(args []string, cache storage.Cache) string {
	var listValue *storage.ListValue
	var ok bool
	key := args[1]
	argsValues := args[2:]
	redisValue, exists := cache.Get(key)
	if !exists {
		listValue = storage.NewListValue()
	} else {
		listValue, ok = redisValue.(*storage.ListValue)
		if !ok {
			return protocol.BuildError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	}
	l.PrependToList(listValue, argsValues)
	cache.Set(key, listValue)
	return protocol.BuildInteger(strconv.Itoa(listValue.Size()))
}

// Validate implements Command.
func (l *LPushCommand) Validate(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("invalid number of args for LPUSH")
	}

	return nil
}
