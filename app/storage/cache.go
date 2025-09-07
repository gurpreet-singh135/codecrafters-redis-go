package storage

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

// Cache interface defines the operations for Redis storage
type Cache interface {
	Get(key string) (RedisValue, bool)
	Set(key string, value RedisValue)
	Delete(key string)
	Type(key string) string
	CleanupExpired()
	// Thread-safe stream operations
	AddToStream(key string, entry *StreamEntry) (string, error)
}

// InMemoryCache implements Cache interface with thread-safe operations
type InMemoryCache struct {
	data map[string]RedisValue
	mu   sync.RWMutex
}

// NewCache creates a new in-memory cache instance
func NewCache() Cache {
	return &InMemoryCache{
		data: make(map[string]RedisValue),
	}
}

// Get retrieves a value from the cache, handling expiration
func (c *InMemoryCache) Get(key string) (RedisValue, bool) {
	c.mu.RLock()
	value, exists := c.data[key]
	if !exists {
		c.mu.RUnlock()
		return nil, false
	}

	// Check if expired
	if value.IsExpired(time.Now()) {
		c.mu.RUnlock()
		// Upgrade to write lock to delete expired key
		c.mu.Lock()
		// Double-check after acquiring write lock
		if val2, stillExists := c.data[key]; stillExists {
			if val2.IsExpired(time.Now()) {
				delete(c.data, key)
				c.mu.Unlock()
				return nil, false
			} else if stillExists {
				c.mu.Unlock()
				return val2, true
			}
		}
		c.mu.Unlock()
		return value, false
	}

	c.mu.RUnlock()
	return value, true
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(key string, value RedisValue) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Delete removes a key from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Type returns the type of the value stored at key
func (c *InMemoryCache) Type(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.data[key]
	if !exists {
		return "none"
	}

	// Check if expired
	if value.IsExpired(time.Now()) {
		return "none"
	}

	return value.Type()
}

// CleanupExpired removes all expired keys from the cache
func (c *InMemoryCache) CleanupExpired() {
	currentTime := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, value := range c.data {
		if value.IsExpired(currentTime) {
			delete(c.data, key)
		}
	}
}

// AddToStream atomically adds an entry to a stream
func (c *InMemoryCache) AddToStream(key string, entry *StreamEntry) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, exists := c.data[key]
	var newEntryID *EntryID
	var err error
	if exists {
		// Check if it's actually a stream
		if streamVal, ok := value.(*StreamValue); ok {
			// Validate the new entry ID using the stream's validation method
			newEntryID, err = streamVal.AddEntry(entry)
			if err != nil {
				return protocol.EMPTY_STRING, err
			}
		} else {
			return protocol.EMPTY_STRING, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		// Create new stream
		log.Println("I'm inside the case where there is no entry for the streamKey")
		streamVal := StreamValue{Entries: []StreamEntry{}}
		c.data[key] = &streamVal
		newEntryID, err = streamVal.AddEntry(entry)
		if err != nil {
			return protocol.EMPTY_STRING, errors.New(protocol.INVALID_ENTRY_ID)
		}
	}

	log.Println("value of the inserted entryId is: ", newEntryID)

	return newEntryID.GetEntryID(), nil
}
