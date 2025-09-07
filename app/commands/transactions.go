package commands

import (
	"fmt"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type TransactionState struct {
	InTransaction bool
	QueueCommands []*QueueCommand
	MaxQueueSize  int
	StartTime     int64
	mutex         sync.RWMutex
}

func NewTransactionState() *TransactionState {
	return &TransactionState{
		InTransaction: false,
		QueueCommands: make([]*QueueCommand, 0, 200),
		MaxQueueSize:  200,
		// StartTime:     time.Now().UnixNano(),
	}
}

func (t *TransactionState) StartTransaction() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.InTransaction = true
	t.StartTime = time.Now().UnixNano()
}

func (t *TransactionState) QueueCommand(Cmd Command, args []string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(t.QueueCommands) >= t.MaxQueueSize {
		return fmt.Errorf("transaction queue full")
	}

	queueCommand := QueueCommand{
		Cmd:       Cmd,
		Args:      args,
		Timestamp: time.Now().UnixNano(),
	}

	t.QueueCommands = append(t.QueueCommands, &queueCommand)
	return nil
}

func (t *TransactionState) ExecuteTransaction(cache storage.Cache) []string {
	t.mutex.RLock()
	commands := make([]*QueueCommand, len(t.QueueCommands))
	copy(commands, t.QueueCommands)
	t.mutex.RUnlock()

	result := make([]string, 0, t.QueueSize())
	for _, queuedCommand := range commands {
		result = append(result, queuedCommand.Execute(cache))
	}

	return result
}

func (t *TransactionState) GetQueueCommands() []*QueueCommand {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.QueueCommands
}

func (t *TransactionState) EndTransaction() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.InTransaction = false
	t.QueueCommands = t.QueueCommands[:0]
	t.StartTime = 0
}

func (t *TransactionState) IsInTransaction() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.InTransaction
}

func (t *TransactionState) QueueSize() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return len(t.QueueCommands)
}

func (t *TransactionState) Reset() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.InTransaction = false
	t.QueueCommands = nil
	t.StartTime = 0
	return true
}
