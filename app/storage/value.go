package storage

import "time"

// RedisValue represents any value that can be stored in Redis
type RedisValue interface {
	Type() string
	IsExpired(currentTime time.Time) bool
}
