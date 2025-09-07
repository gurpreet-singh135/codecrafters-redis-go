package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type QueueCommand struct {
	Cmd       Command
	Args      []string
	Timestamp int64
}

func (q *QueueCommand) Execute(cache storage.Cache) string {
	if err := q.Cmd.Validate(q.Args); err != nil {
		return protocol.BuildError(err.Error())
	}

	return q.Cmd.Execute(q.Args, cache)
}
