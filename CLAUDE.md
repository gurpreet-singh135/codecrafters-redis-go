# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Redis clone implementation built for the CodeCrafters "Build Your Own Redis" challenge. The project implements a subset of Redis functionality including basic commands like PING, ECHO, GET, SET, TYPE, and XADD (streams).

## Architecture

The codebase follows a modular architecture with clear separation of concerns:

- **`app/main.go`**: Entry point that creates the server instance and handles graceful shutdown
- **`app/server/`**: Core server implementation with connection handling and request processing
- **`app/commands/`**: Command system with a registry pattern for Redis commands
- **`app/protocol/`**: RESP (REdis Serialization Protocol) parsing and formatting
- **`app/storage/`**: In-memory cache implementation with different value types
- **`app/config/`**: Configuration management

### Key Components

- **CommandRegistry**: Uses the registry pattern to manage all Redis commands. New commands are registered in `commands/registry.go:NewCommandRegistry()`
- **Storage Layer**: Implements different Redis data types (strings, streams) with TTL support
- **RESP Protocol**: Full implementation of Redis protocol parsing and response formatting
- **Connection Handling**: Each client connection runs in its own goroutine

## Development Commands

### Build and Run
```bash
# Build and run the Redis server locally
./your_program.sh

# Or manually build and run
go build -o /tmp/codecrafters-build-redis-go app/*.go
/tmp/codecrafters-build-redis-go
```

### Testing with Redis CLI
```bash
# Connect to the server (runs on localhost:6379 by default)
redis-cli
# Then test commands like: PING, SET key value, GET key
```

### CodeCrafters Workflow
```bash
# Submit solution to CodeCrafters
git commit -am "your message"
git push origin master
```

## Implementation Patterns

### Adding New Commands
1. Create command struct implementing the `Command` interface in appropriate file in `app/commands/`
2. Implement `Execute(args, cache)` and `Validate(args)` methods
3. Register the command in `commands/registry.go:NewCommandRegistry()`

### RESP Protocol
- Use `protocol.ParseRESP()` to parse incoming client data
- Use `protocol.Build*()` functions to format responses (BuildSimpleString, BuildBulkString, BuildArray, BuildError)

### Storage Types
- String values: Use `storage.StringValue`
- Stream values: Use `storage.StreamValue` 
- All values support TTL through `storage.Value` interface

The server runs on localhost:6379 by default and handles multiple concurrent connections.