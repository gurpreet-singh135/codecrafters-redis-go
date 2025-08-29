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
	newEntryID, err := cache.AddToStream(streamKey, &streamEntry)
	if err != nil {
		log.Printf("Error adding to stream %s: %v", streamKey, err)
		return protocol.BuildError(err.Error())
	}
	
	return protocol.BuildBulkString(newEntryID)
}

func (c *XAddCommand) Validate(args []string) error {
	if len(args) < 5 {
		return errors.New("wrong number of arguments for 'xadd' command")
	}
	
	// Check that we have pairs of field-value arguments
	if (len(args)-3)%2 != 0 {
		return errors.New("wrong number of arguments for 'xadd' command")
	}

	err := IsGreaterThanIdentityId(args[2])
	if err != nil {
		return errors.New(protocol.INVALID_MIN_ID)
	}

	return nil
}

func IsGreaterThanIdentityId(entryId string) error {
	if entryId == "*" {
		return nil
	}
	// validate entry ID format and minimum ID constraint
	parts := strings.Split(entryId, "-")
	if len(parts) != 2 {
		return errors.New("invalid entry ID")
	}
	millisecondsTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return errors.New("invalid entry ID (millisecondTime is not a number)")
	}

	if parts[1] == "*" {
		if millisecondsTime >= 0 {
			return nil
		} 	
		return errors.New("invalid entry ID") 
	}
	sequenceNumber, err := strconv.ParseInt(parts[1], 10, 64)

	if err != nil {
		return errors.New("invalid entry ID (sequenceNumber is not a number)")
	}


	if millisecondsTime <= 0 {
		if millisecondsTime == 0 && sequenceNumber > 0 {
			return nil 
		}
		return errors.New("invalid entry ID") 
	}
	return nil 
}

