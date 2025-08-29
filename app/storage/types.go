package storage

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

// StringValue represents a Redis string value with optional expiration
type StringValue struct {
	Val        string
	Expiration time.Time
}

func (s *StringValue) Type() string {
	return "string"
}

func (s *StringValue) IsExpired(currentTime time.Time) bool {
	return !s.Expiration.IsZero() && currentTime.Compare(s.Expiration) >= 0
}

// GetValue returns the string value
func (s *StringValue) GetValue() string {
	return s.Val
}

// StreamEntry represents a single entry in a Redis stream
type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// StreamValue represents a Redis stream value
type StreamValue struct {
	Entries []StreamEntry
	// EntriesMap map[string]StreamEntry thinking about it
}

func (s *StreamValue) Type() string {
	return "stream"
}

func (s *StreamValue) IsExpired(currentTime time.Time) bool {
	// Streams don't expire by default
	return false
}

// GetEntries returns the stream entries
func (s *StreamValue) GetEntries() []StreamEntry {
	return s.Entries
}

// func (s *StreamValue) GetEntry(entryId string) (*StreamEntry, bool) {
// 	val, ok := s.EntriesMap[entryId]

// 	if !ok {
// 		return nil, false
// 	}

// 	return &val, true 
// }

// AddEntry adds a new entry to the stream
func (s *StreamValue) AddEntry(entry *StreamEntry) (string, error) {
	newEntryID, err := s.IsValidNewEntryID(entry.ID)
	if err != nil {
		return protocol.EMPTY_STRING, errors.New(protocol.INVALID_ENTRY_ID)
	}
	entry.ID = newEntryID
	s.Entries = append(s.Entries, *entry)
	return newEntryID, nil
}

// IsValidNewEntryID validates that a new entry ID is greater than the last entry ID
func (s *StreamValue) IsValidNewEntryID(newEntryID string) (string, error) {
	var lastEntryID string
	if len(s.Entries) == 0 {
		lastEntryID = "0-0"
	} else {
		lastEntryID = s.Entries[len(s.Entries)-1].ID
	}
	return ValidateEntryIDOrder(newEntryID, lastEntryID)
}

// ValidateEntryIDOrder validates that entryID is greater than lastEntryID
func ValidateEntryIDOrder(entryID, lastEntryID string) (string, error) {
	// Parse entry ID parts
	parts := strings.Split(entryID, "-")
	lastParts := strings.Split(lastEntryID, "-")

	if len(parts) != 2 {
		return protocol.EMPTY_STRING, errors.New("invalid entry ID")
	}

	millisecondsTime, err := strconv.ParseInt(parts[0], 10, 64)
	lastMillisecondsTime, _ := strconv.ParseInt(lastParts[0], 10, 64)
	if err != nil {
		return protocol.EMPTY_STRING, errors.New("invalid entry ID (millisecondTime is not a number)")
	}
	var sequenceNumber int64
	lastSequenceNumber, _ := strconv.ParseInt(lastParts[1], 10, 64)

	if parts[1] == "*" {
		if millisecondsTime == lastMillisecondsTime {
			sequenceNumber = lastSequenceNumber + 1
		} else if millisecondsTime < lastMillisecondsTime {
			return protocol.EMPTY_STRING, errors.New("invalid entry ID") 
		} else {
			sequenceNumber = 0
			if millisecondsTime == 0 {
				sequenceNumber += 1
			}
		}
	} else {
		sequenceNumber, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return protocol.EMPTY_STRING, errors.New("invalid entry ID (sequenceNumber is not a number)")
		}
	}


	if (millisecondsTime > lastMillisecondsTime) || ((millisecondsTime == lastMillisecondsTime) && (sequenceNumber > lastSequenceNumber)) {
		return strconv.Itoa(int(millisecondsTime)) + "-" + strconv.Itoa(int(sequenceNumber)), nil
	}

	return protocol.EMPTY_STRING, errors.New("invalid entry ID")
}