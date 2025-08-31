package commands

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// PingCommand implements the PING command
type PingCommand struct{}

func (c *PingCommand) Execute(args []string, cache storage.Cache) string {
	return "+PONG" + protocol.CRLF
}

func (c *PingCommand) Validate(args []string) error {
	// PING command takes no arguments
	if len(args) > 1 {
		return errors.New("wrong number of arguments for 'ping' command")
	}
	return nil
}

// EchoCommand implements the ECHO command
type EchoCommand struct{}

func (c *EchoCommand) Execute(args []string, cache storage.Cache) string {
	if len(args) < 2 {
		return protocol.BuildError("wrong number of arguments for 'echo' command")
	}
	return protocol.BuildBulkString(args[1])
}

func (c *EchoCommand) Validate(args []string) error {
	if len(args) < 2 {
		return errors.New("wrong number of arguments for 'echo' command")
	}
	return nil
}
