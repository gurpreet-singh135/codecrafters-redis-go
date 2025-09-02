package commands

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// GetCommand implements the GET command
type GetCommand struct{}

func (c *GetCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]
	value, exists := cache.Get(key)

	if !exists {
		return protocol.BuildBulkString("")
	}

	// Type assertion to get string value
	if stringVal, ok := value.(*storage.StringValue); ok {
		return protocol.BuildBulkString(stringVal.GetValue())
	}

	if intValue, ok := value.(*storage.IntValue); ok {
		return protocol.BuildBulkString(strconv.Itoa(intValue.Val))
	}

	return protocol.BuildBulkString("")
}

func (c *GetCommand) Validate(args []string) error {
	if len(args) < 2 {
		return errors.New("wrong number of arguments for 'get' command")
	}
	return nil
}

// SetCommand implements the SET command
type SetCommand struct{}

func (c *SetCommand) Execute(args []string, cache storage.Cache) string {
	key := args[1]
	value := args[2]
	var expirationTime time.Time

	// Check for expiration arguments (SET key value PX milliseconds)
	if len(args) > 4 && strings.ToUpper(args[3]) == "PX" {
		expDelta, err := strconv.Atoi(args[4])
		if err == nil {
			expirationTime = time.Now().Add(time.Duration(expDelta) * time.Millisecond)
		}
	}

	val, err := strconv.Atoi(value)
	var redisValue storage.RedisValue
	if err != nil {
		redisValue = c.SetStringValue(key, value, expirationTime)
	} else {
		redisValue = c.SetIntValue(key, val, expirationTime)
	}

	cache.Set(key, redisValue)
	return protocol.BuildSimpleString(protocol.RESPONSE_OK)
}

func (c *SetCommand) SetStringValue(key string, val string, expirationTime time.Time) storage.RedisValue {
	return &storage.StringValue{
		Val:        val,
		Expiration: expirationTime,
	}
}

func (c *SetCommand) SetIntValue(key string, val int, expirationTime time.Time) storage.RedisValue {
	return &storage.IntValue{
		Val:        val,
		Expiration: expirationTime,
	}
}

func (c *SetCommand) Validate(args []string) error {
	if len(args) < 3 {
		return errors.New("wrong number of arguments for 'set' command")
	}

	// Validate PX syntax if present
	if len(args) > 3 {
		if len(args) < 5 || strings.ToUpper(args[3]) != "PX" {
			return errors.New("syntax error")
		}
		if _, err := strconv.Atoi(args[4]); err != nil {
			return errors.New("value is not an integer or out of range")
		}
	}

	return nil
}
