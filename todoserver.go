package todoserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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
		collection:    list.NewCollection(),
	}
	return c
}

// this function blocks
func (ts *todoserver) Serve() error {

	ts.endpoints = map[string]func(http.ResponseWriter, *http.Request){
		"/add":                 ts.handleListAdd, //todo save on every modification (shrug)
		"/get":                 ts.handleListGet,
		"/getall":              ts.handleListGetAll,
		"/remove":              ts.handleListRemove,
		"/web/add":             ts.handleWebAdd,
		"/web/add_redirect":    ts.handleWebAddWithRedirect,
		"/web/remove_redirect": ts.handleWebRemoveWithRedirect,
		"/web/getall":          ts.handleWebGetAll,
		"/healthcheck":         ts.handleHealthcheck,
		//"/test":                     ts.handleTest,
	}

	for ep, fn := range ts.endpoints {
		http.HandleFunc(ep, fn)
	}

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
