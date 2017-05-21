package todoserver

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

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

func NewTodoServer(host string, port int, adminPass, resourceDir string, dsn mysql.Config) Server {

	c := &todoserver{
		host:        host,
		port:        port,
		adminKey:    adminPass,
		resourceDir: resourceDir,
	}

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
		"/admin/create/user":   commonHandlers.ThenFunc(ts.handleAdminCreateUser),
		"/admin/create/apikey": authenticatedHandlers.ThenFunc(ts.handleAdminCreateApikey),
		"/add":                 authenticatedHandlers.ThenFunc(ts.handleListAdd),
		"/get":                 authenticatedHandlers.ThenFunc(ts.handleListGet),
		"/getall":              authenticatedHandlers.ThenFunc(ts.handleListGetAll),
		"/remove":              authenticatedHandlers.ThenFunc(ts.handleListRemove),
		"/healthcheck":         commonHandlers.ThenFunc(ts.handleHealthcheck),

		// TODO remove web from this app
		"/web/login":           commonHandlers.ThenFunc(ts.handleWebLogin),
		"/web/login_submit":    authenticatedHandlers.ThenFunc(ts.handleWebLoginSubmit),
		"/web/logout_submit":   authenticatedHandlers.ThenFunc(ts.handleWebLogoutSubmit),
		"/web/getall":          authenticatedHandlers.ThenFunc(ts.handleWebGetAll),
		"/web/add":             authenticatedHandlers.ThenFunc(ts.handleWebAdd),
		"/web/add_redirect":    authenticatedHandlers.ThenFunc(ts.handleWebAddWithRedirect),
		"/web/remove_redirect": authenticatedHandlers.ThenFunc(ts.handleWebRemoveWithRedirect),
	}

	//TODO look at noodle and gorillamux

	for ep, fn := range endpoints {
		http.Handle(ep, fn)
		ts.endpoints = append(ts.endpoints, ep)
	}

	// this should not be in the list of available endpoints
	// this is just to serve anything inside resourceDir todo which should be configurable or resources need to be embedded in the binary
	// current use case is serving up images
	if _, err := os.Stat(ts.resourceDir); err == nil {
		http.Handle("/", http.FileServer(http.Dir(ts.resourceDir)))
	} else {
		log.Println(qm{"error": err, "info": "unable to serve files from resource directory"})
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", ts.port), nil)
}
