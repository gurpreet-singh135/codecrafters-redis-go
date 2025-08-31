package storage

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type EntryID struct {
	Milliseconds   int64
	SequenceNumber int64
	StreamEntryID  string
}

func NewEntryID(streamEntryID string) *EntryID {
	return &EntryID{0, 0, streamEntryID}
}

func (e *EntryID) ParseStreamEntryID() {
	parts := strings.Split(e.StreamEntryID, "-")
	e.Milliseconds, _ = strconv.ParseInt(parts[0], 10, 64)
	e.SequenceNumber, _ = strconv.ParseInt(parts[1], 10, 64)
}

func (e *EntryID) GetEntryID() string {
	// return strconv.FormatInt(e.Milliseconds, 10) + "-" + strconv.FormatInt(e.SequenceNumber, 10)
	return e.StreamEntryID
}

func (e *EntryID) IsGreater(otherEntry *EntryID) bool {
	return (e.Milliseconds > otherEntry.Milliseconds) ||
		((e.Milliseconds == otherEntry.Milliseconds) && (e.SequenceNumber > otherEntry.SequenceNumber))
}

func (e *EntryID) IsEqual(otherEntry *EntryID) bool {
	return (e.Milliseconds == otherEntry.Milliseconds) && (e.SequenceNumber == otherEntry.SequenceNumber)
}

func (e *EntryID) IsSmaller(otherEntry *EntryID) bool {
	return (e.Milliseconds < otherEntry.Milliseconds) ||
		((e.Milliseconds == otherEntry.Milliseconds) && (e.SequenceNumber < otherEntry.SequenceNumber))
}

func (e *EntryID) IsInRange(start, end *EntryID) bool {
	return (e.IsGreater(start) || e.IsEqual(start)) && (e.IsSmaller(end) || e.IsEqual(end))
}

// StreamEntry represents a single entry in a Redis stream
type StreamEntry struct {
	ID     EntryID
	Fields map[string]string
}

func (s *StreamEntry) ToArray() []any {
	flattenedArray := make([]any, len(s.Fields)*2)

	i := 0
	for key, value := range s.Fields {
		flattenedArray[i] = key
		flattenedArray[i+1] = value
		i += 2
	}

	return []any{
		s.ID.GetEntryID(),
		flattenedArray,
	}
}

// StreamValue represents a Redis stream value
type StreamValue struct {
	Entries []StreamEntry
	// EntriesMap map[string]StreamEntry thinking about it
}

func (s *StreamValue) GetEntriesByRange(start, end *EntryID) []StreamEntry {
	var entries []StreamEntry

	for _, entry := range s.Entries {
		if entry.ID.IsInRange(start, end) {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (s *StreamValue) GetEntriesGreaterThan(start *EntryID) []StreamEntry {
	var entries []StreamEntry

	for _, entry := range s.Entries {
		if entry.ID.IsGreater(start) {
			entries = append(entries, entry)
		}
	}

	return entries
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
func (s *StreamValue) AddEntry(entry *StreamEntry) (*EntryID, error) {
	newEntryID, err := s.IsValidNewEntryID(entry.ID.GetEntryID())
	if err != nil {
		return NewEntryID(""), errors.New(protocol.INVALID_ENTRY_ID)
	}
	entry.ID = *newEntryID
	// (&entry.ID).ParseStreamEntryID()
	s.Entries = append(s.Entries, *entry)
	return &entry.ID, nil
}

// IsValidNewEntryID validates that a new entry ID is greater than the last entry ID
func (s *StreamValue) IsValidNewEntryID(newEntryID string) (*EntryID, error) {
	var lastEntryID string

	if len(s.Entries) == 0 {
		lastEntryID = "0-0"
	} else {
		lastEntryID = (&s.Entries[len(s.Entries)-1].ID).GetEntryID()
	}
	if newEntryID == "*" {
		unixTime := time.Now().UnixMilli()
		newEntryID = strconv.FormatInt(unixTime, 10) + "-*"
	}

	return ValidateEntryIDOrder(newEntryID, lastEntryID)
}

// ValidateEntryIDOrder validates that entryID is greater than lastEntryID
func ValidateEntryIDOrder(entryID, lastEntryID string) (*EntryID, error) {
	// Parse entry ID parts
	parts := strings.Split(entryID, "-")
	lastParts := strings.Split(lastEntryID, "-")

	if len(parts) != 2 {
		return NewEntryID(""), errors.New("invalid entry ID")
	}

	millisecondsTime, err := strconv.ParseInt(parts[0], 10, 64)
	lastMillisecondsTime, _ := strconv.ParseInt(lastParts[0], 10, 64)
	if err != nil {
		return NewEntryID(""), errors.New("invalid entry ID (millisecondTime is not a number)")
	}
	var sequenceNumber int64
	lastSequenceNumber, _ := strconv.ParseInt(lastParts[1], 10, 64)

	if parts[1] == "*" {
		if millisecondsTime == lastMillisecondsTime {
			sequenceNumber = lastSequenceNumber + 1
		} else if millisecondsTime < lastMillisecondsTime {
			return NewEntryID(""), errors.New("invalid entry ID")
		} else {
			sequenceNumber = 0
			if millisecondsTime == 0 {
				sequenceNumber += 1
			}
		}
	} else {
		sequenceNumber, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return NewEntryID(""), errors.New("invalid entry ID (sequenceNumber is not a number)")
		}
	}

	if (millisecondsTime > lastMillisecondsTime) || ((millisecondsTime == lastMillisecondsTime) && (sequenceNumber > lastSequenceNumber)) {
		// return strconv.FormatInt(millisecondsTime, 10)+ "-" + strconv.FormatInt(sequenceNumber, 10), nil
		streamEntryId := strconv.FormatInt(millisecondsTime, 10) + "-" + strconv.FormatInt(sequenceNumber, 10)

		entry := EntryID{millisecondsTime, sequenceNumber, streamEntryId}
		return &entry, nil

	}

	return NewEntryID(""), errors.New("invalid entry ID")
}
