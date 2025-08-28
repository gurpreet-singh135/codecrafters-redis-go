package commands

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// XAddCommand implements the XADD command
type XAddCommand struct{}

func (c *XAddCommand) Execute(args []string, cache storage.Cache) string {
	streamKey := args[1]
	streamID := args[2]

	// Parse field-value pairs
	streamFields := make(map[string]string)
	for i := 3; i < len(args); i += 2 {
		if i+1 < len(args) {
			field := args[i]
			value := args[i+1]
			streamFields[field] = value
		}
	}
	
	streamEntry := storage.StreamEntry{
		ID:     streamID,
		Fields: streamFields,
	}
	
	// Use thread-safe atomic operation to add entry to stream
	err := cache.AddToStream(streamKey, &streamEntry)
	if err != nil {
		log.Printf("Error adding to stream %s: %v", streamKey, err)
		return protocol.BuildError(err.Error())
	}
	
	return protocol.BuildBulkString(streamID)
}

func (c *XAddCommand) Validate(args []string) error {
	if len(args) < 5 {
		return errors.New("wrong number of arguments for 'xadd' command")
	}
	
	// Check that we have pairs of field-value arguments
	if (len(args)-3)%2 != 0 {
		return errors.New("wrong number of arguments for 'xadd' command")
	}

	// validate entry ID
	entryId := args[2]
	parts := strings.Split(entryId, "-")
	if len(parts) != 2 {
		return errors.New("invalid entry ID")
	}
	millisecondsTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return errors.New("invalid entry ID (millisecondTime is not a number)")
	}

	sequenceNumber, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return errors.New("invalid entry ID (sequenceNumber is not a number)")
	}

	valid := IsGreaterThanIdentityId(millisecondsTime, sequenceNumber)
	if !valid {
		return errors.New(protocol.INVALID_MIN_ID)
	}

	return nil
}


func IsGreaterThanIdentityId(millisecondsTime int64, sequenceNumber int64) bool {
	if millisecondsTime <= 0 {
		if millisecondsTime == 0 && sequenceNumber > 0 {
			return true
		}
		return false
	}
	return true
}

