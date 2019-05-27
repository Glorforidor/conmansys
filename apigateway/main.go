package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	confserviceHost, ok := os.LookupEnv("CONFSERVICE_HOST")
	if !ok {
		log.Fatal("Missing environment variable CONFSERIVCE_HOST")
	}

	insserviceHost, ok := os.LookupEnv("INSSERVICE_HOST")
	if !ok {
		log.Fatal("Missing environment variable INSSERIVCE_HOST")
	}

	confserviceURL := fmt.Sprintf("http://%v", confserviceHost)
	insserviceURL := fmt.Sprintf("http://%v", insserviceHost)
	log.Println(confserviceURL)
	log.Println(insserviceURL)

	r := mux.NewRouter()
	r.HandleFunc("/api", api)
	r.HandleFunc("/api/items", proxyHandler(confserviceURL))
	r.HandleFunc("/api/items/{id}", proxyHandler(confserviceURL))
	r.HandleFunc("/api/modules", proxyHandler(confserviceURL))
	r.HandleFunc("/api/modules/{id}", proxyHandler(confserviceURL))
	r.HandleFunc("/api/itemmodules", proxyHandler(confserviceURL))
	r.HandleFunc("/api/itemmodules/{id}", proxyHandler(confserviceURL))
	r.HandleFunc("/api/moduledependencies", proxyHandler(confserviceURL))
	r.HandleFunc("/api/moduledependencies/dependent/{id}", proxyHandler(confserviceURL))
	r.HandleFunc("/api/moduledependencies/dependee/{id}", proxyHandler(confserviceURL))
	r.HandleFunc("/api/itemmodules/{id}", proxyHandler(confserviceURL))
	r.HandleFunc("/api/insfile", proxyHandler(insserviceURL))
	r.HandleFunc("/api/insfile/text", proxyHandler(insserviceURL))

	srv := http.Server{
		Addr:         "",
		Handler:      r,
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

const (
	usage = `
GET, POST /api/items
GET, DELETE /api/items/:id
`
)

func api(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(usage))
}

// proxyHandler is used for reverse proxying. It will ask the target for the
// given resource.
func proxyHandler(target string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		remote, err := url.Parse(target)
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Printf("%#v", remote)

		proxy := httputil.NewSingleHostReverseProxy(remote)

		r.URL.Path = strings.Replace(r.URL.Path, "/api/", "/", 1)

		proxy.ServeHTTP(w, r)
	}
}
