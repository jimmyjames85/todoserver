package web

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/justinas/alice"
)

type webServer struct {
	adminKey    string
	host        string
	port        int
	resourceDir string
	endpoints   []string
	todoHost    string
	todoPort    int
}



type Server interface {
	// Serve is a blocking function
	Serve() error
}

func NewServer(port int, adminPass, resourceDir string, todoHost string, todoPort int) Server {

	ws := &webServer{
		adminKey:    adminPass,
		host:        "localhost",
		port:        port,
		resourceDir: resourceDir,
		todoHost:    todoHost,
		todoPort:    todoPort,
	}

	hostname, err := os.Hostname()
	if err == nil {
		ws.host = hostname
	}

	return ws
}

// Serve is a blocking function
func (ws *webServer) Serve() error {

	commonHandlers := alice.New(ws.aliceParseIncomingRequest)
	authenticatedHandlers := alice.New(ws.aliceParseIncomingRequest, ws.aliceParseIncomingUser)

	endpoints := map[string]http.Handler{
		"/healthcheck": commonHandlers.ThenFunc(ws.handleHealthcheck),

		"/web/login":           commonHandlers.ThenFunc(ws.handleWebLogin),
		//"/web/login_submit":    authenticatedHandlers.ThenFunc(ws.handleWebLoginSubmit),
		"/web/logout_submit":   authenticatedHandlers.ThenFunc(ws.handleWebLogoutSubmit),
		"/web/getall":          authenticatedHandlers.ThenFunc(ws.handleWebGetAll),
		"/web/add":             authenticatedHandlers.ThenFunc(ws.handleWebAdd),
		"/web/add_redirect":    authenticatedHandlers.ThenFunc(ws.handleWebAddWithRedirect),
		"/web/remove_redirect": authenticatedHandlers.ThenFunc(ws.handleWebRemoveWithRedirect),
	}

	for ep, fn := range endpoints {
		http.Handle(ep, fn)
		ws.endpoints = append(ws.endpoints, ep)
	}

	// this should not be in the list of available endpoints
	// this is just to serve anything inside resourceDir todo which should be configurable or resources need to be embedded in the binary
	// current use case is serving up images
	if _, err := os.Stat(ws.resourceDir); err == nil {
		http.Handle("/", http.FileServer(http.Dir(ws.resourceDir)))
	} else {
		log.Println(qm{"error": err, "info": "unable to serve files from resource directory"})
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", ws.port), nil)
}
