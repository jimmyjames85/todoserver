package main

import (
	"log"

	"github.com/jimmyjames85/todoserver/web"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	AdminKey string `envconfig:"ADMIN_KEY" required:"true"` // is used to add new users

	Port int `envconfig:"PORT" required:"false" default:"8080"` // port to run web server on

	//todo use os.hostname ?
	hHost       string `envconfig:"HOST" required:"false" default:"localhost"` // for the web form data to know which host to hit
	ResourceDir string `envconfig:"RESOURCE_DIR" required:"false" default:""`  // where static resources reside
	TodoHost    string `envconfig:"TODO_HOST" required "true"`
	TodoPort    int    `envconfig:"TODO_PORT" required "true"`
}

func main() {

	c := &config{}
	envconfig.MustProcess("WEB", c)

	ws := web.NewServer(c.Port, c.AdminKey, c.ResourceDir, c.TodoHost, c.TodoPort)

	log.Printf("todo web server listening on %d\n", c.Port)
	err := ws.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
