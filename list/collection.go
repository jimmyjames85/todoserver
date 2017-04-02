package list

import (
	"bytes"
	"encoding/base64"

	"github.com/jimmyjames85/todoserver/util"
)

type Collection map[string]List

const serialDelimiter = byte(0)

func NewCollection() Collection {
	return make(map[string]List)
}

func (c Collection) Serialize() []byte {
	var buf bytes.Buffer
	for listName, list := range c {
		buf.WriteString(util.ToBase64(listName))
		buf.WriteByte(serialDelimiter)
		buf.WriteString(list.Serialize())
		buf.WriteByte(serialDelimiter)
	}
	return buf.Bytes()
}

func DeserializeCollection(c []byte) (Collection, error) {
	ret := make(map[string]List)
	buf := bytes.NewBuffer(c)
	readingName := true
	var name []byte
	for line, err := util.ReadStringTrimDelim(buf, serialDelimiter); err == nil; line, err = util.ReadStringTrimDelim(buf, serialDelimiter) {
		if readingName {
			name, err = base64.StdEncoding.DecodeString(line)
			if err != nil {
				return ret, err
			}
		} else {
			lst, err := DeserializeList(line)
			if err != nil {
				return ret, err
			}
			ret[string(name)] = lst
		}
		readingName = !readingName
	}
	return ret, nil
}

func (c Collection) Keys() []string {
	ret := make([]string, 0)
	for key, _ := range c {
		ret = append(ret, key)
	}
	return ret
}

func (c Collection) GetOrCreateList(listName string) List {
	if _, ok := c[listName]; !ok {
		c[listName] = NewList()
	}
	return c[listName]
}

func (c Collection) Names() []string {
	var ret []string
	for name := range c {
		ret = append(ret, name)
	}
	return ret
}

func (c Collection) GetList(listName string) List {
	l, ok := c[listName]
	if !ok {
		return nil
	}
	return l
}

func (c Collection) SubSet(listNames ...string) Collection {
	ret := make(map[string]List)

	for _, listName := range listNames {
		if l := c.GetList(listName); l != nil {
			ret[listName] = l
		}
	}
	return ret
}

func (c Collection) ToJSON() string {
	m := make(map[string][]Item)
	for lname, lst := range c {
		m[lname] = lst.Items()
	}
	return util.ToJSON(m)
}
