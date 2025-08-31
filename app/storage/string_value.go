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
