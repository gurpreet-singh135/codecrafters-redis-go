package commands

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type XRangeCommand struct{}

func (c *XRangeCommand) Execute(args []string, cache storage.Cache) string {
	var startEntryID, endEntryID storage.EntryID

	err := c.Validate(args)
	if err != nil {
		return protocol.BuildError(err.Error())
	}

	key := args[1]

	startEntryID = *ParseStreamEntryID(args[2])
	endEntryID = *ParseStreamEntryID(args[3])

	if !strings.Contains(args[3], "-") {
		endEntryID.SequenceNumber = math.MaxInt64
	}

	streamValue, exists := cache.Get(key)
	if !exists {
		return protocol.BuildEmptyArray()
	}

	inRangeEntries := streamValue.(*storage.StreamValue).GetEntriesByRange(&startEntryID, &endEntryID)
	var entries []any
	for _, entry := range inRangeEntries {
		entries = append(entries, entry.ToArray())
	}

	return protocol.BuildArray(entries)
}

func (c *XRangeCommand) Validate(args []string) error {
	if len(args) != 4 {
		return errors.New("wrong number of arguments for 'XRANGE' command")
	}

	return nil
}

func ParseStreamEntryID(idStr string) *storage.EntryID {
	var entryId storage.EntryID

	switch idStr {
	case "-":
		entryId.Milliseconds = 0
		entryId.SequenceNumber = 0
	case "+":
		entryId.Milliseconds = math.MaxInt64
		entryId.SequenceNumber = math.MaxInt64
	default:
		parts := strings.Split(idStr, "-")
		if len(parts) == 1 {
			entryId.Milliseconds, _ = strconv.ParseInt(parts[0], 10, 64)
			entryId.SequenceNumber = 0
		} else {
			entryId.Milliseconds, _ = strconv.ParseInt(parts[0], 10, 64)
			entryId.SequenceNumber, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}
	return &entryId
}
