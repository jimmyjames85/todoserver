package main

import (
	"encoding/base64"
	"log"

	"time"

	"github.com/jimmyjames85/todoserver"
	"github.com/kelseyhightower/envconfig"
	"github.com/jimmyjames85/todoserver/list"

	"fmt"
)

type config struct {
	Host             string `envconfig:"HOST" required:"false" default:"localhost"`               // for the web form data to know which host to hit
	Port             int    `envconfig:"PORT" required:"false" default:"1234"`                    // port to run on
	Pass64           string `envconfig:"PASS64" required:"false" default:""`                      // base64 password
	SaveFileloc      string `envconfig:"SAVEFILE" required:"false" default:"/tmp/todolists.json"` // where to save the to-do list
	SaveFrequencySec int    `envconfig:"SAVE_FREQUENCY_SEC" required:"false" default:"60"`        // how often to save the to-do list
}

func main() {

	return
	myList := list.NewList()
	myList.AddItems("jimbo", "rikki", "mateo", "buddy", "mama", "pops")
	myList.Serialize()

	b := []byte("here\x00'sjohnie\n")
	for _, c := range b{
		fmt.Printf("char='%c' int=%d str='%s'\n", c,c,c)
	}


//	fileloc := "/tmp/todo.list"



	//d := []byte(ToJSON(t))
	//if err := ioutil.WriteFile(fileloc, d, 0644) ; err !=nil {
	//	log.Fatal(err)
	//}


	//if _, err := os.Stat(c.saveFile); err == nil {
	//	err := c.todolists.LoadFromDisk(c.saveFile)
	//	if err != nil {
	//		log.Fatalf("unable to load from previous file: %s\n", c.saveFile)
	//	}
	//}

	//	err := c.todolists.SavetoDisk(c.saveFile)

}
func main2() {
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
