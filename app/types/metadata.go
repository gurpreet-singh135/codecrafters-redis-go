package types

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

type ServerMetadata struct {
	// Replication info
	Role                       string `json:"role"`
	ConnectedSlaves            int    `json:"connected_slaves"`
	MasterReplID               string `json:"master_replid"`
	MasterReplOffset           int64  `json:"master_repl_offset"`
	SecondReplOffset           int64  `json:"second_repl_offset"`
	ReplBacklogActive          int    `json:"repl_backlog_active"`
	ReplBacklogSize            int64  `json:"repl_backlog_size"`
	ReplBacklogFirstByteOffset int64  `json:"repl_backlog_first_byte_offset"`
	ReplBacklogHistlen         int64  `json:"repl_backlog_histlen"`

	mutex sync.RWMutex
}

type ServerMetadataCopy struct {
	// Replication info
	Role                       string `json:"role"`
	ConnectedSlaves            int    `json:"connected_slaves"`
	MasterReplID               string `json:"master_replid"`
	MasterReplOffset           int64  `json:"master_repl_offset"`
	SecondReplOffset           int64  `json:"second_repl_offset"`
	ReplBacklogActive          int    `json:"repl_backlog_active"`
	ReplBacklogSize            int64  `json:"repl_backlog_size"`
	ReplBacklogFirstByteOffset int64  `json:"repl_backlog_first_byte_offset"`
	ReplBacklogHistlen         int64  `json:"repl_backlog_histlen"`
}

func NewServerMetadata(role string) *ServerMetadata {
	return &ServerMetadata{
		Role: role,
	}
}

func (m *ServerMetadata) ToStringArray() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	log.Println("Inside ServerMetadata's ToStringArray method")

	result := make([]string, 0)
	result = append(result, "# Replication")
	values := reflect.ValueOf(*(m.Copy()))
	types := values.Type()

	for i := 0; i < values.NumField(); i += 1 {
		field := types.Field(i)
		value := values.Field(i)
		val := fmt.Sprint(value.Interface())
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Name
		}
		result = append(result, tag+":"+val)
	}
	return result
}

func (m *ServerMetadata) String() string {
	result := m.ToStringArray()
	res := ""
	for _, str := range result {
		res += str
	}
	return res
}

func (m *ServerMetadata) Copy() *ServerMetadataCopy {
	copy := ServerMetadataCopy{
		m.Role,
		m.ConnectedSlaves,
		m.MasterReplID,
		m.MasterReplOffset,
		m.SecondReplOffset,
		m.ReplBacklogActive,
		m.ReplBacklogSize,
		m.ReplBacklogFirstByteOffset,
		m.ReplBacklogHistlen,
	}
	return &copy
}
