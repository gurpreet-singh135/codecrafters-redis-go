package commands

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type IncrCommand struct{}

// Execute implements Command.
func (i *IncrCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]
	redisValue, ok := cache.Get(key)
	if !ok {
		// Insert Key
		intValue := storage.IntValue{Val: 0, Expiration: time.Now()}
		cache.Set(key, &intValue)
	}

	redisValue, ok = cache.Get(key)

	value, isInteger := redisValue.(*storage.IntValue)
	if !isInteger {
		// TODO
	}
	value.Val += 1

	return protocol.BuildInt(value.Val)
}

// Validate implements Command.
func (i *IncrCommand) Validate(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("invalid number of arguments")
	}
	return nil
}
