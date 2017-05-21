package todo

import (
	"sort"
	"strings"
)

type Item struct {
	Details   string `json:"details"`
	Item      string `json:"item"`
	Priority  int    `json:"priority"`
	CreatedAt string `json:"created_at"`
	DueDate   string `json:"due_date"`
	Id        int64  `json:"id"`
}

type List struct {
	Id        int64  `json:"id"`
	UserId    int64  `json:"user_id"`
	Title     string `json:"title"`
	Priority  int64  `json:"priority"`
	CreatedAt string `json:"created_at"`
	Items     []Item `json:"items"`
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
