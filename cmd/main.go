package main

import (
	"encoding/base64"
	"log"

	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/todoserver"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Host             string `envconfig:"HOST" required:"false" default:"localhost"`          // for the web form data to know which host to hit
	Port             int    `envconfig:"PORT" required:"false" default:"1234"`               // port to run on
	Pass64           string `envconfig:"PASS64" required:"false" default:""`                 // base64 password
	SaveFileloc      string `envconfig:"SAVEFILE" required:"false" default:"/tmp/todolists"` // where to save the to-do list
	SaveFrequencySec int    `envconfig:"SAVE_FREQUENCY_SEC" required:"false" default:"60"`   // how often to save the to-do list
	ResourceDir      string `envconfig:"RESOURCE_DIR" required:"false" default:""`           // where static resources reside
	DBuser           string `envconfig:"DB_USER" required "false" default:"todouser"`
	DBPswd           string `envconfig:"DB_PSWD" required "false" default:"todopswd"`
	DBHost           string `envconfig:"DB_HOST" required "false" default:"localhost"`
	DBPort           int    `envconfig:"DB_PORT" required "false" default:"3306"`
	DBName           string `envconfig:"DB_NAME" required "false" default:"todolists"`
}

func main() {

	c := &config{}
	envconfig.MustProcess("TODO", c)
	pass, err := base64.StdEncoding.DecodeString(c.Pass64)
	if err != nil {
		log.Fatal("unable to decode PASS64")
	}

	dsn := mysql.Config{}
	dsn.Addr = fmt.Sprintf("%s:%d", c.DBHost, c.DBPort)
	dsn.Passwd = c.DBPswd
	dsn.User = c.DBuser
	dsn.DBName = c.DBName
	dsn.Net = "tcp"

	fmt.Printf("dbname: %s %s\n", dsn.DBName, c.DBName)
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	ts := todoserver.NewTodoServer(c.Host, c.Port, string(pass), c.ResourceDir, dsn)
	err = ts.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
