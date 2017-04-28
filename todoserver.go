package todoserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/todoserver/list"
	"github.com/jimmyjames85/todoserver/util"
)

type todoserver struct {

	//todo remove these
	pass          string
	saveFile      string
	saveFrequency time.Duration
	collection    list.Collection

	host          string
	port          int
	resourceDir   string
	mysqlCfg      mysql.Config
	db            *sql.DB
	endpoints     map[string]func(http.ResponseWriter, *http.Request)
}

func NewTodoServer(host string, port int, pass, savefile string, resourceDir string, saveFrequency time.Duration, dsn mysql.Config) *todoserver {

	c := &todoserver{
		host:          host,
		port:          port,
		pass:          pass,
		saveFile:      savefile,
		resourceDir:   resourceDir,
		saveFrequency: saveFrequency,
		collection:    list.NewCollection(),
	}

	fmt.Printf("DSN: %s\n", dsn.FormatDSN())
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		log.Fatalf("could not open %v", err)
	}
	c.db = db
	rows, err := db.Query("select id, username, sessionid from users")
	if err != nil {
		log.Fatalf("no query %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		id := sql.NullInt64{}
		username := sql.NullString{}
		sessionId := sql.NullString{}

		err = rows.Scan(&id, &username, &sessionId)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%#v %#v %#v\n", id, username, sessionId)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return c
}

// this function blocks
func (ts *todoserver) Serve() error {

	ts.endpoints = map[string]func(http.ResponseWriter, *http.Request){
		"/add":                 ts.handleListAdd, //todo save on every modification (shrug)
//		"/get":                 ts.handleListGet,
		"/v2/get":              ts.handleListGetV2,
		"/getall":              ts.handleListGetAll,
		"/remove":              ts.handleListRemove,
//		"/update":              ts.handleListUpdate,
		"/web/add":             ts.handleWebAdd,
		"/web/add_redirect":    ts.handleWebAddWithRedirect,
		"/web/remove_redirect": ts.handleWebRemoveWithRedirect,
		"/web/getall":          ts.handleWebGetAll,
		"/web/login":           ts.handleWebLogin,
		"/web/login_submit":    ts.handleWebLoginSubmit,
		"/healthcheck":         ts.handleHealthcheck,

		//"/test":                     ts.handleTest,
	}

	for ep, fn := range ts.endpoints {
		http.HandleFunc(ep, fn)
	}

	// this should not be in the list of available endpoints
	// this is just to serve anything inside resourceDir todo which should be configurable or resources need to be embedded in the binary
	// current use case is serving up images
	if _, err := os.Stat(ts.resourceDir); err == nil {
		http.Handle("/", http.FileServer(http.Dir(ts.resourceDir)))
	} else {
		log.Println(util.ToJSON(map[string]interface{}{"err": err, "info": "unable to server files from resource directory"}))
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
