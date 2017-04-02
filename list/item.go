package list

import (
	"strings"
	"time"
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

type ByItem []Item

func (m ByItem) Len() int           { return len(m) }
func (m ByItem) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m ByItem) Less(i, j int) bool { return strings.Compare(m[i].Item, m[j].Item) < 0 }

type ByPriority []Item

func (p ByPriority) Len() int           { return len(p) }
func (p ByPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByPriority) Less(i, j int) bool { return p[i].Priority < p[j].Priority }

type ByCreatedAt []Item

func (c ByCreatedAt) Len() int           { return len(c) }
func (c ByCreatedAt) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCreatedAt) Less(i, j int) bool { return c[i].CreatedAt < c[j].CreatedAt }

type ByDueDate []Item

func (d ByDueDate) Len() int           { return len(d) }
func (d ByDueDate) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ByDueDate) Less(i, j int) bool { return d[i].DueDate < d[j].DueDate }
