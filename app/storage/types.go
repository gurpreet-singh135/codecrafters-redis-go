package storage

import "time"

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

// AddEntry adds a new entry to the stream
func (s *StreamValue) AddEntry(entry StreamEntry) {
	s.Entries = append(s.Entries, entry)
}