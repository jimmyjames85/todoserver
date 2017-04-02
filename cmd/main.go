package main

import (
	"encoding/base64"
	"log"
	"time"

	"github.com/jimmyjames85/todoserver"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Host             string `envconfig:"HOST" required:"false" default:"localhost"`          // for the web form data to know which host to hit
	Port             int    `envconfig:"PORT" required:"false" default:"1234"`               // port to run on
	Pass64           string `envconfig:"PASS64" required:"false" default:""`                 // base64 password
	SaveFileloc      string `envconfig:"SAVEFILE" required:"false" default:"/tmp/todolists"` // where to save the to-do list
	SaveFrequencySec int    `envconfig:"SAVE_FREQUENCY_SEC" required:"false" default:"60"`   // how often to save the to-do list
}

func main() {
	c := &config{}
	envconfig.MustProcess("TODO", c)
	pass, err := base64.StdEncoding.DecodeString(c.Pass64)
	if err != nil {
		log.Fatal("unable to decode PASS64")
	}

	ts := todoserver.NewTodoServer(c.Host, c.Port, string(pass), c.SaveFileloc, time.Duration(c.SaveFrequencySec)*time.Second)
	err = ts.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
