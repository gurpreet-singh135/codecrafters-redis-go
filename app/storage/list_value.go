package storage

import (
	"sync"
	"time"
)

type ListValue struct {
	Items []ListItem
	mu    sync.RWMutex
}

type ListItem struct {
	Value string
}

// LIST_VALUE methods

func (l *ListValue) Type() string {
	return "list"
}

func (l *ListValue) IsExpired(t time.Time) bool {
	return false
}

func (l *ListValue) Size() int {
	return len(l.Items)
}

func (l *ListValue) Append(listEntry *ListItem) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Items = append(l.Items, *listEntry)
	return l.Size()
}

func (l *ListValue) Prepend(listEntry *ListItem) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Items = append([]ListItem{*listEntry}, l.Items...)
	return l.Size()
}

func (l *ListValue) Lpop() *ListItem {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.Size() == 0 {
		return nil
	}
	item := l.Items[0]
	l.Items = l.Items[1:]
	return &item
}

func (l *ListValue) GetRangeInclusive(start, end int) []ListItem {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]ListItem, 0)
	if start < 0 {
		if -start > l.Size() {
			start = 0
		} else {
			start = l.Size() + start
		}
	}
	if end < 0 {
		if -end > l.Size() {
			end = 0
		} else {
			end = l.Size() + end
		}
	}

	if start > end || start >= l.Size() {
		return result
	} else if end > l.Size() {
		end = l.Size() - 1
	}

	for i := start; i <= end; i += 1 {
		result = append(result, l.Items[i])
	}

	return result
}

func NewListValue() *ListValue {
	return &ListValue{}
}

// LIST_ENTRIES methods
func NewListItem(entryStr string) *ListItem {
	return &ListItem{
		Value: entryStr,
	}
}
