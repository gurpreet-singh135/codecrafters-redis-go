package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type LRangeCommand struct{}

func (l *LRangeCommand) ConvertListItemToAny(items []storage.ListItem) []any {
	result := make([]any, 0)

	for _, item := range items {
		result = append(result, item.Value)
	}
	return result
}

// Execute implements Command.
func (l *LRangeCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]
	start, err := strconv.Atoi(args[2])
	if err != nil {
		return protocol.BuildError("invalid integer argument")
	}
	end, err := strconv.Atoi(args[3])
	if err != nil {
		return protocol.BuildError("invalid integer argument")
	}

	redisValue, ok := cache.Get(key)
	if !ok {
		return protocol.BuildEmptyArray()
	}

	listValue, ok := redisValue.(*storage.ListValue)
	if !ok {
		return protocol.BuildError("key's value is not of List Type")
	}

	listItems := listValue.GetRangeInclusive(start, end)

	if len(listItems) == 0 {
		return protocol.BuildEmptyArray()
	}

	anyItems := l.ConvertListItemToAny(listItems)

	return protocol.BuildArray(anyItems)
}

// Validate implements Command.
func (l *LRangeCommand) Validate(args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("invalid command arguments for LRANGE")
	}

	return nil
}
