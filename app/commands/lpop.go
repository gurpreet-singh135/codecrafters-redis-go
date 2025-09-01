package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type LPopCommand struct{}

func (l *LPopCommand) PopItems(num int, listValue *storage.ListValue) []any {
	result := make([]any, num)
	for i := range num {
		result[i] = listValue.Lpop().Value
	}
	return result
}

// Execute implements Command.
func (l *LPopCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]

	redisValue, ok := cache.Get(key)
	if !ok {
		return protocol.BuildBulkString("")
	}
	listValue, ok := redisValue.(*storage.ListValue)
	if !ok {
		return protocol.BuildError("wrong type for LPOP")
	}

	if len(args) == 2 {
		item := l.PopItems(1, listValue)[0].(string)
		return protocol.BuildBulkString(item)
	}
	num, err := strconv.Atoi(args[2])
	if err != nil {
		return protocol.BuildError(err.Error())
	}
	items := l.PopItems(num, listValue)

	return protocol.BuildArray(items)
}

// Validate implements Command.
func (l *LPopCommand) Validate(args []string) error {
	if len(args) > 3 {
		return fmt.Errorf("invalid number of argument to LPOP")
	}

	return nil
}
