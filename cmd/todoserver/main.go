package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/todoserver"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	AdminKey string `envconfig:"ADMIN_KEY" required:"true"` // is used to add new users

	Port int `envconfig:"PORT" required:"false" default:"1234"` // port to run on

	Host        string `envconfig:"HOST" required:"false" default:"localhost"` // for the web form data to know which host to hit
	ResourceDir string `envconfig:"RESOURCE_DIR" required:"false" default:""`  // where static resources reside

	DBuser string `envconfig:"DB_USER" required "true"`
	DBPswd string `envconfig:"DB_PASS" required "true"`
	DBHost string `envconfig:"DB_HOST" required "true"`
	DBPort int    `envconfig:"DB_PORT" required "true"`
	DBName string `envconfig:"DB_NAME" required "true"`
}

func main() {

	c := &config{}
	envconfig.MustProcess("TODO", c)

	dsn := mysql.Config{}
	dsn.Addr = fmt.Sprintf("%s:%d", c.DBHost, c.DBPort)
	dsn.Passwd = c.DBPswd
	dsn.User = c.DBuser
	dsn.DBName = c.DBName
	dsn.Net = "tcp"

	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		log.Fatalf("%v: Did you set environment variables?\n", err.Error())
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatalf("%v: Did you set environment variables?\n", err.Error())
	}
	ts := todoserver.NewServer(c.Host, c.Port, c.AdminKey, c.ResourceDir, dsn)

	log.Printf("listening on %d\n", c.Port)
	err = ts.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
