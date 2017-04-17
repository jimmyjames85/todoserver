package list

import (
	"strings"
	"time"
	"sort"
)

type Item struct {
	Item      string `json:"item"`
	Priority  int    `json:"priority"`
	CreatedAt int64  `json:"created_at"`
	DueDate   int64  `json:"due_date"`
}

func fmtUnixTime(sec int64) string {
	return time.Unix(sec, 0).Format(time.ANSIC)
}

func (i *Item) DueDateString() string {
	if i.DueDate == 0 {
		return ""
	}
	return fmtUnixTime(i.DueDate)
}

func (i *Item) CreatedAtDateString() string {
	return fmtUnixTime(i.CreatedAt)
}

func SortItemsByItem(items []Item) {
	sort.Slice(items, func(i, j int) bool { return strings.Compare(items[i].Item, items[j].Item) < 0 })
}

func SortItemsByPriority(items []Item) {
	sort.Slice(items, func(i, j int) bool { return items[i].Priority < items[j].Priority })
}

func SortItemsByCreatedAt(items []Item) {
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt < items[j].CreatedAt })
}

func SortItemsByDueDate(items []Item) {
	sort.Slice(items, func(i, j int) bool { return items[i].DueDate < items[j].DueDate })
}
