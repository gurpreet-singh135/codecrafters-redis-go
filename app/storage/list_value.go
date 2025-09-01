package storage

import "time"

type ListValue struct {
	Items []ListItem
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
	l.Items = append(l.Items, *listEntry)
	return len(l.Items)
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
