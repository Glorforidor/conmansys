package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Glorforidor/conmansys/confservice/handler"
	"github.com/Glorforidor/conmansys/confservice/storage/postgres"
)

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

func main() {
	conf := dbConfig()
	p, err := postgres.New(
		conf[dbhost], conf[dbport],
		conf[dbuser], conf[dbpass],
		conf[dbname],
	)
	if err != nil {
		panic(err)
	}
	defer p.Close()

	r := handler.New(p)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
		// lets make sure we timeout
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func dbConfig() map[string]string {
	conf := make(map[string]string)

	host, ok := os.LookupEnv(dbhost)
	if !ok {
		panic("DBHOST environment variable required but not set")
	}

	port, ok := os.LookupEnv(dbport)
	if !ok {
		panic("DBPORT environment variable required but not set")
	}

	user, ok := os.LookupEnv(dbuser)
	if !ok {
		panic("DBUSER environment variable required but not set")
	}

	password, ok := os.LookupEnv(dbpass)
	if !ok {
		panic("DBPASS environment variable required but not set")
	}

	name, ok := os.LookupEnv(dbname)
	if !ok {
		panic("DBNAME environment variable required but not set")
	}

	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name

	return conf
}
