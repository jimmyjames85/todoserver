package todoserver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"log"
	"os"

	"github.com/jimmyjames85/todoserver/list"
)

type todoserver struct {
	host          string
	port          int
	pass          string
	saveFile      string
	saveFrequency time.Duration
	collection    list.Collection
	endpoints     map[string]func(http.ResponseWriter, *http.Request)
}

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
func (ts *todoserver) Serve() error {

	ts.endpoints = map[string]func(http.ResponseWriter, *http.Request){
		"/todo/web/add":             ts.handleWebAdd,
		"/todo/web/add_redirect":    ts.handleWebAddWithRedirect,
		"/todo/web/remove_redirect": ts.handleWebRemoveWithRedirect,
		"/todo/web/getall":          ts.handleWebGetAll,
		"/test":                     ts.handleTest,
		"/todo/add":                 ts.handleListAdd,
		"/todo/get":                 ts.handleListGet,
		"/todo/getall":              ts.handleListGetAll,
		"/todo/remove":              ts.handleListRemove,
		"/todo/save":                ts.handleSaveListsToDisk,   //todo save on every modification (shrug)
		"/todo/load":                ts.handleLoadListsFromDisk, //todo remove why do we need this
		"/healthcheck":              ts.handleHealthcheck,
	}

	for ep, fn := range ts.endpoints {
		http.HandleFunc(ep, fn)
	}
	//
	//
	//http.HandleFunc("/todo/web/add", ts.handleWebAdd)
	//http.HandleFunc("/todo/web/add_redirect", ts.handleWebAddWithRedirect)
	//http.HandleFunc("/todo/web/remove_redirect", ts.handleWebRemoveWithRedirect)
	//http.HandleFunc("/todo/web/getall", ts.handleWebGetAll)
	//http.HandleFunc("/test", ts.handleTest)
	//http.HandleFunc("/todo/add", ts.handleListAdd)
	//http.HandleFunc("/todo/get", ts.handleListGet)
	//http.HandleFunc("/todo/getall", ts.handleListGetAll)
	//http.HandleFunc("/todo/remove", ts.handleListRemove)
	//http.HandleFunc("/todo/save", ts.handleSaveListsToDisk) //todo save on every modification (shrug)
	//http.HandleFunc("/todo/load", ts.handleLoadListsFromDisk) //todo remove why do we need this
	//http.HandleFunc("/healthcheck", ts.handleHealthcheck)

	if _, err := os.Stat(ts.saveFile); err == nil {
		err := ts.loadFromDisk()
		if err != nil {
			log.Printf("unable to load from previous file: %s\n", ts.saveFile)
		}
	}

	// save on a cron
	go func() {
		saveTimer := time.Tick(ts.saveFrequency)
		for _ = range saveTimer {
			err := ts.saveToDisk()
			if err != nil {
				fmt.Printf(outcomeMessage(false, fmt.Sprintf("%s", err))) //todo notify
				return
			}
		}
	}()
	return http.ListenAndServe(fmt.Sprintf(":%d", ts.port), nil)
}

func (ts *todoserver) saveToDisk() error {
	b := ts.collection.Serialize()
	return ioutil.WriteFile(ts.saveFile, b, 0644)
}

func (ts *todoserver) loadFromDisk() error {
	b, err := ioutil.ReadFile(ts.saveFile)
	if err != nil {
		return err
	}
	col, err := list.DeserializeCollection(b)
	if err != nil {
		return err
	}
	ts.collection = col
	return err
}
