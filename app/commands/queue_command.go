package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

type QueueCommand struct {
	Cmd       Command
	Args      []string
	Timestamp int64
	Metadata  *types.ServerMetadata
}

func (q *QueueCommand) Execute(cache storage.Cache) []string {
	if err := q.Cmd.Validate(q.Args); err != nil {
		return []string{protocol.BuildError(err.Error())}
	}

	if serverCmd, ok := q.Cmd.(ServerAwareCommand); ok {
		return serverCmd.ExecuteWithMetadata(q.Args, cache, q.Metadata)
	}

	return []string{q.Cmd.Execute(q.Args, cache)}
}
