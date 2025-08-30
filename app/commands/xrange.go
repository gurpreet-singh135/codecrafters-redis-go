package commands

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)


type XRangeCommand struct {}


func (c *XRangeCommand) Execute(args []string, cache storage.Cache) string {
	var startEntryId, endEntryID storage.EntryID

	err := c.Validate(args)
	if err != nil {
		return protocol.BuildError(err.Error()) 
	}

	key := args[1]
	sParts := strings.Split(args[2], "-")
	eParts := strings.Split(args[3], "-")

	if len(sParts) == 1 {
		startEntryId.Milliseconds, _ = strconv.ParseInt(sParts[0], 10, 64)
		startEntryId.SequenceNumber = 0
	} else {
		startEntryId.Milliseconds, _ = strconv.ParseInt(sParts[0], 10, 64)
		startEntryId.SequenceNumber, _ = strconv.ParseInt(sParts[1], 10, 64)
	}

	if len(eParts) == 1 {
		endEntryID.Milliseconds, _ = strconv.ParseInt(eParts[0], 10, 64)
		endEntryID.SequenceNumber = math.MaxInt64
	} else {
		endEntryID.Milliseconds, _ = strconv.ParseInt(eParts[0], 10, 64)
		endEntryID.SequenceNumber, _ = strconv.ParseInt(eParts[1], 10, 64)
	}

	streamValue, exists := cache.Get(key)
	if !exists {
		return protocol.BuildEmptyArray()
	}
	
	inRangeEntries := streamValue.(*storage.StreamValue).GetEntriesByRange(&startEntryId, &endEntryID)
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
