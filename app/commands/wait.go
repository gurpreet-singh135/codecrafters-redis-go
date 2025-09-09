package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

type WaitCommand struct{}

// Execute implements Command.
func (w *WaitCommand) Execute(args []string, cache storage.Cache) string {
	return "This method of `wait` shouldn't be called"
}

// Validate implements Command.
func (w *WaitCommand) Validate(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("wrong number of arguments to 'wait' command")
	}
	return nil
}

func (w *WaitCommand) ExecuteWithMetadata(args []string, cache storage.Cache, metadata *types.ServerMetadata) []string {
	// Parse arguments: WAIT numreplicas timeout
	numReplicasStr := args[1]
	timeoutStr := args[2]

	numReplicas, err := strconv.Atoi(numReplicasStr)
	if err != nil {
		return []string{protocol.BuildError("ERR invalid number of replicas")}
	}

	timeoutMs, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return []string{protocol.BuildError("ERR invalid timeout")}
	}

	// If no replicas requested, return 0
	if numReplicas <= 0 {
		return []string{protocol.BuildInt(0)}
	}

	// Get current number of active replicas
	activeReplicas := metadata.NumberOfActiveConnections()

	// If no replicas connected, return 0
	if activeReplicas == 0 {
		return []string{protocol.BuildInt(0)}
	}

	// Get current master offset - this is what replicas need to catch up to
	currentOffset := metadata.MasterReplOffset

	// If there are no pending writes (offset is 0), return the number of connected replicas immediately
	// This handles the "WAIT with no commands" case
	if currentOffset == 0 {
		// When no commands have been executed, all connected replicas are considered caught up
		// Return the total number of active replicas
		return []string{protocol.BuildInt(activeReplicas)}
	}

	// If we need more replicas than we have, wait for all available
	if numReplicas > activeReplicas {
		numReplicas = activeReplicas
	}

	// Create wait request
	waitReq := &types.WaitRequest{
		ID:            w.generateWaitID(),
		TargetOffset:  currentOffset,
		RequiredCount: numReplicas,
		ResponseChan:  make(chan int, 1),
		ReceivedAcks:  make(map[string]int64),
		StartTime:     time.Now(),
	}

	// Register wait request
	metadata.RegisterWaitRequest(waitReq)
	defer metadata.UnregisterWaitRequest(waitReq.ID)

	// Send REPLCONF GETACK * to all replicas
	getackCmd := []string{"REPLCONF", "GETACK", "*"}
	metadata.Replicate(getackCmd)

	// Wait for responses or timeout
	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case count := <-waitReq.ResponseChan:
		return []string{protocol.BuildInt(count)}
	case <-time.After(timeout):
		// Timeout - return current count of satisfied replicas
		satisfiedCount := 0
		for _, offset := range waitReq.ReceivedAcks {
			if offset >= waitReq.TargetOffset {
				satisfiedCount++
			}
		}
		return []string{protocol.BuildInt(satisfiedCount)}
	}
}

func (w *WaitCommand) generateWaitID() string {
	return fmt.Sprintf("wait-%d", time.Now().UnixNano())
}
