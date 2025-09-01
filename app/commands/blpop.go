package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type BLPopCommand struct{}

func (b *BLPopCommand) handleBlockCommand(cache storage.Cache, key string) string {
	for {
		time.Sleep(50 * time.Millisecond)

		redisValue, exists := cache.Get(key)
		if !exists {
			continue
		}

		listValue, correctType := redisValue.(*storage.ListValue)
		if !correctType {
			continue
		}

		if val := listValue.Lpop(); val != nil {
			return protocol.BuildArray([]any{key, val.Value})
		}
	}
}

// Execute implements Command.
func (b *BLPopCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]
	timeout, _ := strconv.ParseFloat(args[2], 64)
	delta := int(timeout * 1000)

	redisValue, exists := cache.Get(key)
	if !exists {
		if delta == 0 {
			return b.handleBlockCommand(cache, key) // Block indefinitely
		} else {
			time.Sleep(time.Duration(delta) * time.Millisecond)
		}
	}

	redisValue, exists = cache.Get(key)
	if !exists {
		return protocol.BuildNullArray()
	}
	listValue, correctType := redisValue.(*storage.ListValue)
	if !correctType {
		return protocol.BuildError("wrongtype of BLPOP command")
	}

	if val := listValue.Lpop(); val != nil {
		return protocol.BuildArray([]any{key, val.Value})
	} else {
		return protocol.BuildNullArray()
	}
}

// Validate implements Command.
func (b *BLPopCommand) Validate(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("invalid number of argument to BLPOP")
	}
	return nil
}
