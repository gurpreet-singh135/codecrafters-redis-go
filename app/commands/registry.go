package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

// Command interface defines the contract for Redis commands
type Command interface {
	Execute(args []string, cache storage.Cache) string
	Validate(args []string) error
}

type ServerAwareCommand interface {
	Command
	ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) string
}

// CommandRegistry manages all available Redis commands
type CommandRegistry struct {
	commands map[string]Command
}

type CommandExecutionResult struct {
	Response string
	Error    error
	Success  bool
}

func NewCommandExecutionResult(response string, err error) *CommandExecutionResult {
	return &CommandExecutionResult{
		Response: response,
		Error:    err,
		Success:  err == nil,
	}
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
	registry.Register("RPUSH", &RPushCommand{})
	registry.Register("LRANGE", &LRangeCommand{})
	registry.Register("LPUSH", &LPushCommand{})
	registry.Register("LLEN", &LLenCommand{})
	registry.Register("LPOP", &LPopCommand{})
	registry.Register("BLPOP", &BLPopCommand{})
	registry.Register("INCR", &IncrCommand{})
	registry.Register("MULTI", &MultiCommand{})
	registry.Register("EXEC", &ExecCommand{})
	registry.Register("INFO", &InfoCommand{})

	return registry
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(name string, cmd Command) {
	r.commands[strings.ToUpper(name)] = cmd
}

// // Execute runs a command with the given arguments
// func (r *CommandRegistry) Execute(cmdName string, args []string, cache storage.Cache) string {
// 	cmd, exists := r.commands[strings.ToUpper(cmdName)]
// 	if !exists {
// 		return protocol.BuildError("unknown command '" + cmdName + "'")
// 	}

// 	if err := cmd.Validate(args); err != nil {
// 		return protocol.BuildError(err.Error())
// 	}

// 	log.Printf("Values of cmd, and args are %s, %v", cmdName, args)
// 	return cmd.Execute(args, cache)
// }

// HasCommand checks if a command exists in the registry
func (r *CommandRegistry) HasCommand(cmdName string) bool {
	_, exists := r.commands[strings.ToUpper(cmdName)]
	return exists
}

func (r *CommandRegistry) GetCommand(cmdName string) (Command, error) {
	Cmd, exists := r.commands[strings.ToUpper(cmdName)]
	if !exists {
		return nil, fmt.Errorf("Given command: %s is not supported yet", cmdName)
	}
	return Cmd, nil
}
