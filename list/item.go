package list

import (
	"sort"
	"strings"
	"time"
)

type Item struct {
	Details   string `json:"details"`
	Item      string `json:"item"`
	Priority  int    `json:"priority"`
	CreatedAt string `json:"created_at"`
	DueDate   string `json:"due_date"`
	Id        int64  `json:"id"`
}

func fmtUnixTime(sec int64) string {
	return time.Unix(sec, 0).Format(time.ANSIC)
}

func (i *Item) DueDateString() string {
	return i.DueDate
	//TODO
	//if i.DueDate == 0 {
	//	return ""
	//}
	//return fmtUnixTime(i.DueDate)
}

func (i *Item) CreatedAtDateString() string {
	return i.CreatedAt
	// TODO
	// return fmtUnixTime(i.CreatedAt)
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
