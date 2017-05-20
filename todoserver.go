package todoserver

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/todoserver/util"
	"github.com/justinas/alice"
)

type todoserver struct {
	pass string

	host        string
	port        int
	resourceDir string
	mysqlCfg    mysql.Config
	db          *sql.DB
	endpoints   []string //map[string]func(http.ResponseWriter, *http.Request) todo remove this COMMENT if []string works or after if cur date is after 6/19/2017
}

func NewTodoServer(host string, port int, pass, resourceDir string, dsn mysql.Config) *todoserver {

	c := &todoserver{
		host:        host,
		port:        port,
		pass:        pass,
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

//
//func (ts *todoserver) withUser(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
//
//	user, ok := r.Context().Value("user").(*auth.User)
//	if !ok {
//		ts.handleError(fmt.Errorf("no user in context %#v", user), "", http.StatusInternalServerError, w)
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}) //noop
//	}
//
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//
//	})
//
//
//}

// this function blocks
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

		"/web/login":         commonHandlers.ThenFunc(ts.handleWebLogin),
		"/web/login_submit":  authenticatedHandlers.ThenFunc(ts.handleWebLoginSubmit),
		"/web/logout_submit": authenticatedHandlers.ThenFunc(ts.handleWebLogoutSubmit),
		"/web/getall":        authenticatedHandlers.ThenFunc(ts.handleWebGetAll),

		//		"/update":              ts.handleListUpdate,
		"/web/add":             authenticatedHandlers.ThenFunc(ts.handleWebAdd),
		"/web/add_redirect":    authenticatedHandlers.ThenFunc(ts.handleWebAddWithRedirect),
		"/web/remove_redirect": authenticatedHandlers.ThenFunc(ts.handleWebRemoveWithRedirect),

		"/healthcheck": commonHandlers.ThenFunc(ts.handleHealthcheck),

		// "/test": commonHandlers.ThenFunc(ts.handleTest),

	}

	//TODO look at noodle and gorillamux

	for ep, fn := range endpoints {
		http.Handle(ep, fn)
		ts.endpoints = append(ts.endpoints, ep)
	}

	//
	//ts.endpoints = map[string]func(http.ResponseWriter, *http.Request){
	//	"/admin/create/user":   ts.handleAdminCreateUser,
	//	"/admin/create/apikey": ts.handleAdminCreateApikey,
	//	"/add":                 ts.handleListAdd,
	//	"/get":                 ts.handleListGet,
	//	"/getall":              ts.handleListGetAll,
	//	"/remove":              ts.handleListRemove,
	//
	//	"/web/login":        ts.handleWebLogin,
	//	"/web/login_submit": ts.handleWebLoginSubmit,
	//	"/web/getall":       ts.handleWebGetAll,
	//
	//	//		"/update":              ts.handleListUpdate,
	//	"/web/add":             ts.handleWebAdd,
	//	"/web/add_redirect":    ts.handleWebAddWithRedirect,
	//	"/web/remove_redirect": ts.handleWebRemoveWithRedirect,
	//
	//	"/healthcheck": ts.handleHealthcheck,
	//
	//	//"/test":                     ts.handleTest,
	//}
	//
	//for ep, fn := range ts.endpoints {
	//	http.HandleFunc(ep, fn)
	//}

	// this should not be in the list of available endpoints
	// this is just to serve anything inside resourceDir todo which should be configurable or resources need to be embedded in the binary
	// current use case is serving up images
	if _, err := os.Stat(ts.resourceDir); err == nil {
		http.Handle("/", http.FileServer(http.Dir(ts.resourceDir)))
	} else {
		log.Println(util.ToJSON(map[string]interface{}{"err": err, "info": "unable to serve files from resource directory"}))
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", ts.port), nil)
}
