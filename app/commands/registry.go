package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// Command interface defines the contract for Redis commands
type Command interface {
	Execute(args []string, cache storage.Cache) string
	Validate(args []string) error
}

// CommandRegistry manages all available Redis commands
type CommandRegistry struct {
	commands map[string]Command
}

// NewCommandRegistry creates a new command registry with all commands
func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]Command),
	}
	
	// Register all commands
	registry.Register("PING", &PingCommand{})
	registry.Register("ECHO", &EchoCommand{})
	registry.Register("GET", &GetCommand{})
	registry.Register("SET", &SetCommand{})
	registry.Register("TYPE", &TypeCommand{})
	registry.Register("XADD", &XAddCommand{})
	registry.Register("XRANGE", &XRangeCommand{})
	registry.Register("XREAD", &XReadCommand{})
	
	return registry
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(name string, cmd Command) {
	r.commands[strings.ToUpper(name)] = cmd
}

// Execute runs a command with the given arguments
func (r *CommandRegistry) Execute(cmdName string, args []string, cache storage.Cache) string {
	cmd, exists := r.commands[strings.ToUpper(cmdName)]
	if !exists {
		return protocol.BuildError("unknown command '" + cmdName + "'")
	}
	
	if err := cmd.Validate(args); err != nil {
		return protocol.BuildError(err.Error())
	}
	
	return cmd.Execute(args, cache)
}

// HasCommand checks if a command exists in the registry
func (r *CommandRegistry) HasCommand(cmdName string) bool {
	_, exists := r.commands[strings.ToUpper(cmdName)]
	return exists
}