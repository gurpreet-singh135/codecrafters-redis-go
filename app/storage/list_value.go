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

func (l *ListValue) GetRangeInclusive(start, end int) []ListItem {
	result := make([]ListItem, 0)
	if start < 0 {
		start = l.Size() + start
	}
	if end < 0 {
		end = l.Size() + end
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
