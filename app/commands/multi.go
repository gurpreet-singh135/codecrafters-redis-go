package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type MultiCommand struct{}

type ExecCommand struct{}

// Execute implements Command.
func (e *ExecCommand) Execute(args []string, cache storage.Cache) string {
	return protocol.BuildError(protocol.EXEC_BEFORE_MULTI)
}

// Validate implements Command.
func (e *ExecCommand) Validate(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid arguments")
	}

	return nil
}

// Execute implements Command.
func (m *MultiCommand) Execute(args []string, cache storage.Cache) string {
	return protocol.BuildBulkString("OK")
}

// Validate implements Command.
func (m *MultiCommand) Validate(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid arguments")
	}

	return nil
}
