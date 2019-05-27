package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
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
		Addr:    "",
		Handler: r,
		// lets make sure we timeout
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// run server in own go routine.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// buffered channel
	c := make(chan os.Signal, 1)

	// notify the channel of any interrupts.
	signal.Notify(c, os.Interrupt)

	// block until a signal is recieved
	<-c

	// wait some time before shutting down to finish work.
	wait := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// if there are no connections, then it shutsdown immediately otherwise
	// block until wait time is over.
	srv.Shutdown(ctx)

	log.Println("shutting down")
	os.Exit(0)
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
