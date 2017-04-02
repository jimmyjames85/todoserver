package list
//
//import (
//	"io/ioutil"
//	"encoding/json"
//	"fmt"
//	"sort"
//	"github.com/jimmyjames85/todoserver/util"
//)
//
//
//type Todos map[string][]string
//
//func NewTodos() Todos {
//	return make(map[string][]string)
//}
//
//func (t Todos) AddItems(listname string, items ...string) {
//	t[listname] = append(t[listname], items...)
//}
//
//func (t Todos) RemoveItems(listname string, indicies ...int) ([]string, error) {
//	if _, ok := t[listname]; !ok {
//		return nil, fmt.Errorf("no such list: %s", listname)
//	}
//
//	sort.Sort(sort.Reverse(sort.IntSlice(indicies))) // TODO why does this work; i think sort.Reverse defines diff sort methods maybe
//
//	removed := make([]string, 0)
//
//	for _, index := range indicies {
//		if index < 0 || index >= len(t[listname]) {
//			continue
//		}
//		removed = append(removed, t[listname][index])
//		t[listname] = append(t[listname][:index], t[listname][index+1:]...)
//	}
//
//	if len(t[listname]) == 0 {
//		delete(t, listname)
//	}
//	return removed, nil
//}
//
//func (t Todos) SetPriority(listname string, index int, newIndex int) {
//
//	list, ok := t[listname]
//	if !ok {
//		return
//	} else if index < 0 || index >= len(list) || newIndex < 0 || newIndex >= len(list) {
//		return
//	}
//
//	// by default we move [index] to the right
//	direction := 1
//	if newIndex < index {
//		direction = -1
//	}
//
//	for index != newIndex {
//		t[listname][index], t[listname][index+direction] = t[listname][index+direction], t[listname][index]
//		index += direction
//
//	}
//}
//
//func (t Todos) GetLists(listnames ...string) map[string][]string {
//	ret := make(map[string][]string, 0)
//	for _, listname := range listnames {
//		if list, ok := t[listname]; ok {
//			ret[listname] = append(ret[listname], list...)
//		}
//	}
//	return ret
//}
//
//func (t Todos) GetAllLists() map[string][]string {
//	ret := make(map[string][]string, 0)
//	for listname, list := range t {
//		ret[listname] = append(ret[listname], list...)
//	}
//	return ret
//}
//
//func (t Todos) SavetoDisk(fileloc string) error {
//	d := []byte(util.ToJSON(t))
//	return ioutil.WriteFile(fileloc, d, 0644)
//}
//
//func (t Todos) LoadFromDisk(fileloc string) error {
//	var lists map[string][]string
//	bytes, err := ioutil.ReadFile(fileloc)
//	if err != nil {
//		return err
//	}
//	err = json.Unmarshal(bytes, &lists)
//	if err != nil {
//		return err
//	}
//
//	for listname, _ := range t {
//		delete(t, listname)
//	}
//	for listname, list := range lists {
//		t[listname] = list
//	}
//
//	return nil
//}
//
/////////////////////////////////////////////////////////////////////////////////////
//
////func QuickCopy(s string) string {
////	ret := copydata
////	copydata = s
////	return ret
////}
////
////func QuickPaste() string {
////	return copydata
////}
//
