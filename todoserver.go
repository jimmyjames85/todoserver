package todoserver

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/justinas/alice"
)

type todoserver struct {
	adminKey    string
	host        string
	port        int
	resourceDir string
	mysqlCfg    mysql.Config
	db          *sql.DB
	endpoints   []string
}

type Server interface {
	// Serve is a blocking function
	Serve() error
}

func NewServer(host string, port int, adminPass, resourceDir string, dsn mysql.Config) Server {

	c := &todoserver{
		host:        host,
		port:        port,
		adminKey:    adminPass,
		resourceDir: resourceDir,
	}

	//todo remove or scrub dsn
	fmt.Printf("DSN: %s\n", dsn.FormatDSN())
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		log.Fatalf("could not open %v", err)
	}
	c.db = db

	return c
}

// Serve is a blocking function
func (ts *todoserver) Serve() error {

	commonHandlers := alice.New(ts.aliceParseIncomingRequest)
	authenticatedHandlers := alice.New(ts.aliceParseIncomingRequest, ts.aliceParseIncomingUser)

	endpoints := map[string]http.Handler{
		"/admin/create/user":     commonHandlers.ThenFunc(ts.handleAdminCreateUser),
		"/user/create/sessionid": authenticatedHandlers.ThenFunc(ts.handleUserCreateSessionID),
		"/user/create/apikey":    authenticatedHandlers.ThenFunc(ts.handleUserCreateApikey),
		"/add":                   authenticatedHandlers.ThenFunc(ts.handleListAdd),
		"/get":                   authenticatedHandlers.ThenFunc(ts.handleListGet),
		"/getall":                authenticatedHandlers.ThenFunc(ts.handleListGetAll),
		"/remove":                authenticatedHandlers.ThenFunc(ts.handleListRemove),
		"/healthcheck":           commonHandlers.ThenFunc(ts.handleHealthcheck),
	}

	//TODO look at noodle and gorillamux

	for ep, fn := range endpoints {
		http.Handle(ep, fn)
		ts.endpoints = append(ts.endpoints, ep)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", ts.port), nil)
}
