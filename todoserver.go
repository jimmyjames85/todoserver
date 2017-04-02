package todoserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jimmyjames85/todoserver/list"
)

type todoserver struct {
	host          string
	port          int
	pass          string
	saveFile      string
	saveFrequency time.Duration
	collection    list.Collective
	//lists        list.Collective
}

//TODO rename to savetodisk and return err ... i.e. this is for testing
//func (c *todoserver) Serialize() []byte {
//	var buf bytes.Buffer
//	for listName, list := range c.lists {
//		buf.WriteString(listName)
//		buf.WriteByte(0)
//		buf.WriteString(list.Serialize())
//		buf.WriteByte(0)
//		buf.WriteByte(0)
//	}
//	buf.WriteByte(0)
//	buf.WriteByte(0)
//	buf.WriteByte(0)
//	return buf.Bytes()
//}

//func (c *todoserver) getOrCreateList(listName string) list.List {
//	return c.collection.GetOrCreateList(listName)
//	if _, ok := c.lists[listName]; !ok {
//		c.lists[listName] = list.NewList()
//	}
//	return c.lists[listName]
//
//}

//func (c *todoserver) listNames() []string {
//	var ret []string
//	for name, _ := range c.collection {
//		ret = append(ret, name)
//	}
//	return ret
//}

func (c *todoserver) listSetToJSON(listNames ...string) string {

	subSet := c.collection.SubSet(listNames...)
	return subSet.ToJSON()
	//
	//
	//m := make(map[string][]list.Item)
	//for _, lname := range listNames {
	//	if l, ok := c.lists[lname]; ok {
	//		m[lname] = l.Items()
	//	}
	//}
	//return util.ToJSON(m)
}

//func (c *todoserver) saveToDisk() {
//	d := []byte(ToJSON(t))
//	return ioutil.WriteFile(fileloc, d, 0644)
//}
//func (t Todos) SavetoDisk(fileloc string) error {
//
//}
func NewTodoServer(host string, port int, pass, savefile string, saveFrequency time.Duration) *todoserver {
	c := &todoserver{
		host:          host,
		port:          port,
		pass:          pass,
		saveFile:      savefile,
		saveFrequency: saveFrequency,
		collection:    make(map[string]list.List),
	}
	return c
}

// this function blocks
func (c *todoserver) Serve() error {

	http.HandleFunc("/todo/web/add", c.handleWebAdd)
	http.HandleFunc("/todo/web/add_redirect", c.handleWebAddWithRedirect)
	http.HandleFunc("/todo/web/remove_redirect", c.handleWebRemoveWithRedirect)
	http.HandleFunc("/todo/web/getall", c.handleWebGetAll)
	http.HandleFunc("/test", c.handleTest)
	http.HandleFunc("/todo/add", c.handleListAdd)
	http.HandleFunc("/todo/get", c.handleListGet)
	http.HandleFunc("/todo/getall", c.handleListGetAll)
	http.HandleFunc("/todo/remove", c.handleListRemove)
	http.HandleFunc("/todo/save", c.handleSaveListsToDisk) //todo save on every modification
	http.HandleFunc("/todo/load", c.handleLoadListsFromDisk)
	http.HandleFunc("/healthcheck", c.handleHealthcheck)

	// TODO
	//if _, err := os.Stat(c.saveFile); err == nil {
	//	err := c.todolists.LoadFromDisk(c.saveFile)
	//	if err != nil {
	//		log.Fatalf("unable to load from previous file: %s\n", c.saveFile)
	//	}
	//}
	//
	//// save on a cron
	//go func() {
	//	saveTimer := time.Tick(c.saveFrequency)
	//	for _ = range saveTimer {
	//		err := c.todolists.SavetoDisk(c.saveFile)
	//		if err != nil {
	//			fmt.Printf(outcomeMessage(false, fmt.Sprintf("%s", err))) //todo notify
	//			return
	//		}
	//	}
	//}()
	return http.ListenAndServe(fmt.Sprintf(":%d", c.port), nil)
}
