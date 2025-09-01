package commands

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type XReadCommand struct{}

func waitForNewData(cache storage.Cache, keys, ids []string) {
	newDataChan := make(chan bool)
	go func(cache storage.Cache, keys, ids []string) {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			<-ticker.C
			for i, key := range keys {
				redisValueForKey, exists := cache.Get(key)
				if !exists {
					continue
				}
				entityId := ParseStreamEntryID(ids[i])
				newLen := len(redisValueForKey.(*storage.StreamValue).GetEntriesGreaterThan(entityId))
				if newLen > 0 {
					newDataChan <- true
					return
				}
			}
		}
	}(cache, keys, ids)

	<-newDataChan
}

func handleBlockCommand(args []string, cache storage.Cache) []string {
	if len(args) > 2 && strings.ToUpper(args[1]) == "BLOCK" {
		waitPeriod, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			log.Printf("Invalid BLOCK period: %s", args[2])
			return args
		}
		args = append([]string{args[0]}, args[3:]...) // Remove BLOCK and timeout
		keysCount := (len(args) - 2) / 2
		keys := args[2 : 2+keysCount]
		ids := args[2+keysCount:]

		if waitPeriod == 0 {
			waitForNewData(cache, keys, ids)
		} else {
			time.Sleep(time.Duration(waitPeriod) * time.Millisecond)
		}
	}
	return args
}

func processStreams(args []string, cache storage.Cache) string {
	var startEntryId storage.EntryID
	var entries []any
	keysCount := (len(args) - 2) / 2

	for i := 0; i < keysCount; i += 1 {
		key := args[2+i]
		redisValueForKey, exists := cache.Get(key)
		if !exists {
			continue
		}

		streamValue := redisValueForKey.(*storage.StreamValue)

		// parts := strings.Split(args[2 + i + keysCount], "-")
		startEntryId = *ParseStreamEntryID(args[2+i+keysCount])
		log.Println("value of entryid is: ", startEntryId, args[2+i+keysCount], args)

		keysEntries := streamValue.GetEntriesGreaterThan(&startEntryId)
		if len(keysEntries) > 0 {
			entries = append(entries, XReadEntriesConversion(key, keysEntries))
		}
		// entries = append(entries, keysEntries)
	}

	return protocol.BuildArray(entries)

}

// Execute implements Command.
func (x *XReadCommand) Execute(args []string, cache storage.Cache) string {
	log.Println("value of args is: ", args)
	processedArgs := handleBlockCommand(args, cache)

	return processStreams(processedArgs, cache)

}

func preprocessXReadArgs(args []string) ([]string, error) {
	log.Println("preprocessXReadArgs input:", args)
	if len(args) > 2 && strings.ToUpper(args[1]) == "BLOCK" {
		// Create a new clean slice instead of using append
		result := make([]string, 0, len(args)-2)
		result = append(result, args[0])     // Add "XREAD"
		result = append(result, args[3:]...) // Add everything after "1000"
		return result, nil
	}
	log.Println("preprocessXReadArgs output (no change):", args)
	return args, nil
}

// Validate implements Command.
func (x *XReadCommand) Validate(args []string) error {
	processedArgs, err := preprocessXReadArgs(args)
	if err != nil {
		return err
	}

	if len(processedArgs) < 4 || strings.ToUpper(processedArgs[1]) != "STREAMS" {
		return errors.New("invalid argument to (XREAD) command")
	}

	if (len(processedArgs)-2)%2 != 0 {
		return errors.New("unbalanced XREAD: number of keys must equal number of IDs")
	}
	return nil
}

func XReadEntriesConversion(key string, entries []storage.StreamEntry) []any {
	var result []any

	for _, entry := range entries {
		result = append(result, entry.ToArray())
	}

	return []any{
		key,
		result,
	}
}
