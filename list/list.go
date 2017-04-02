package list

import (
	"encoding/base64"
	"time"

	"encoding/json"

	"github.com/jimmyjames85/todoserver/util"
)

type List interface {
	AddItems(items ...string)
	RemoveItems(items ...string)
	Items() []Item
	String() string
	ToJSON() string
	GetItem(item string) (Item, bool)
	UpdateItem(item Item) bool
	Serialize() string
}

type Collective map[string]List

type list struct {
	data map[string]Item
}

func (c Collective) Keys() []string {
	ret := make([]string, 0)
	for key, _ := range c {
		ret = append(ret, key)
	}
	return ret
}

func (c Collective) GetOrCreateList(listName string) List {
	if _, ok := c[listName]; !ok {
		c[listName] = NewList()
	}
	return c[listName]
}

func (c Collective) Names() []string {
	var ret []string
	for name := range c {
		ret = append(ret, name)
	}
	return ret
}

func (c Collective) GetList(listName string) List {
	l, ok := c[listName]
	if !ok {
		return nil
	}
	return l
}

func (c Collective) SubSet(listNames ...string) Collective {
	ret := make(map[string]List)

	for _, listName := range listNames {
		if l := c.GetList(listName); l != nil {
			ret[listName] = l
		}
	}
	return ret
}

func (c Collective) ToJSON() string {
	m := make(map[string][]Item)
	for lname, lst := range c {
		m[lname] = lst.Items()
	}
	return util.ToJSON(m)
}

func (l list) Serialize() string {
	return util.ToBase64(util.ToJSON(l.data))
}

func Deserialize(serializedList string) (list, error) {
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

func NewList() list {
	return list{data: make(map[string]Item)}
}

type Item struct {
	Item      string `json:"item"`
	Priority  int    `json:"priority"`
	CreatedAt int64  `json:"created_at"`
	DueDate   int64  `json:"due_date"`
}

type ByPriority []Item

func (p ByPriority) Len() int           { return len(p) }
func (p ByPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByPriority) Less(i, j int) bool { return p[i].Priority < p[j].Priority }

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

func (l list) UpdateItem(item Item) bool {
	if _, ok := l.data[item.Item]; !ok {
		return false
	}
	l.data[item.Item] = item
	return true
}
