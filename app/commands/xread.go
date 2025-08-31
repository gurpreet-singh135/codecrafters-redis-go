package commands

import (
	"errors"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type XReadCommand struct{}

// Execute implements Command.
func (x *XReadCommand) Execute(args []string, cache storage.Cache) string {
	err := x.Validate(args)
	if err != nil {
		return protocol.BuildError(err.Error())
	}

	var startEntryId storage.EntryID
	var entries []any
	keysCount := (len(args) - 2) / 2

	for i := 0; i < keysCount; i += 1 {
		key := args[2 + i]
		redisValueForKey, exists := cache.Get(key)
		if !exists {
			continue
		}

		streamValue := redisValueForKey.(*storage.StreamValue)

		// parts := strings.Split(args[2 + i + keysCount], "-")
		startEntryId = *ParseStreamEntryID(args[2 + i + keysCount])

		keysEntries := streamValue.GetEntriesGreaterThan(&startEntryId)
		if len(keysEntries) > 0 {
			entries = append(entries, XReadEntriesConversion(key, keysEntries))
		}
		// entries = append(entries, keysEntries)
	}



	return protocol.BuildArray(entries)
}

// Validate implements Command.
func (x *XReadCommand) Validate(args []string) error {
	if len(args) < 4 || strings.ToUpper(args[1]) != "STREAMS" {
		return errors.New("Invalid argument to (XREAD) command")
	}

	if (len(args) - 2) % 2 != 0 {
      return errors.New("Unbalanced XREAD: number of keys must equal number of IDs")
  }

	return nil
}

func XReadEntriesConversion(key string, entries []storage.StreamEntry) []any {
	var result []any

	for _, entry := range entries {
		result = append(result, entry.ToArray())
	}

	return []any {
		key,
		result,
	}
}