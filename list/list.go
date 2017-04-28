package list

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/jimmyjames85/todoserver/util"
)

type List interface {
	AddItems(items ...string)
	RemoveItems(items ...string)
	Items() []Item
	String() string
	ToJSON() string
	GetItem(item string) (Item, bool)
	UpdateItem(item string, newItem Item) bool
	Serialize() string
}

type list struct {
	data map[string]Item
}

type List2 struct {
	Id        int64  `json:"id"`
	UserId    int64  `json:"user_id"`
	Title     string `json:"title"`
	Priority  int64  `json:"priority"`
	CreatedAt int64  `json:"created_at"`
	Items     []Item `json:"items"`
}

func DeserializeList(serializedList string) (list, error) {
	ret := NewList()
	jsonBytes, err := base64.StdEncoding.DecodeString(serializedList)
	if err != nil {
		return ret, err
	}

	data := make(map[string]Item)
	err = json.Unmarshal(jsonBytes, &data)
	if err != nil {
		return ret, err
	}

	ret.data = data
	return ret, nil
}

func (l list) Serialize() string {
	return util.ToBase64(util.ToJSON(l.data))
}

func NewList() list {
	return list{data: make(map[string]Item)}
}

func (l list) AddItems(items ...string) {
	for _, itm := range items {
		if _, ok := l.data[itm]; !ok && len(itm) > 0 {
			l.data[itm] = Item{CreatedAt: time.Now().Unix(), Item: itm}
		}
	}
}

func (l list) RemoveItems(items ...string) {
	for _, itm := range items {
		delete(l.data, itm)
	}
}

func (l list) Items() []Item {
	items := make([]Item, 0)
	for _, itm := range l.data {
		items = append(items, itm)
	}
	return items
}

func (l list) String() string {
	return l.ToJSON()
}
func (l list) ToJSON() string {
	return util.ToJSON(l.Items())
}

func (l list) GetItem(item string) (Item, bool) {
	ret, ok := l.data[item]
	return ret, ok
}

func (l list) UpdateItem(itm string, newItem Item) bool {
	if _, ok := l.data[itm]; !ok {
		return false
	}
	l.data[itm] = newItem
	return true
}
