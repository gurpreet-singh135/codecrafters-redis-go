package storage

import "time"

type IntValue struct {
	Val        int
	Expiration time.Time
}

func (*IntValue) Type() string {
	return "integer"
}

func (*IntValue) IsExpired(t time.Time) bool {
	return false
}
