package main

import (
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/Glorforidor/conmansys/confservice/storage"
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

	// testItem(p)
	// testModule(p)
	testItemModule(p)
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

func testItem(p storage.Service) {
	// items
	it, err := p.GetItem(1)
	if err != nil {
		log.Fatal(err)
	}
	print(it)

	it, err = p.GetItem(2)
	if err != nil {
		log.Fatal(err)
	}
	print(it)

	ii, err := p.GetItems()
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range ii {
		print(i)
	}

	i, err := p.CreateItem("pension_calculator", "window", "0.0.1")
	if err != nil {
		log.Fatal(err)
	}

	it, err = p.GetItem(i)
	if err != nil {
		log.Fatal(err)
	}
	print(it)

	p.DeleteItem(i)
}

func testModule(p storage.Service) {
	//modules
	m, err := p.GetItem(1)
	if err != nil {
		log.Fatal(err)
	}
	print(m)

	m, err = p.GetItem(2)
	if err != nil {
		log.Fatal(err)
	}
	print(m)

	mm, err := p.GetModules()
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range mm {
		print(m)
	}
}

func testItemModule(p storage.Service) {
	it, err := p.CreateItem("test_system", "test", "0.0.1")
	if err != nil {
		log.Fatal(err)
	}
	print(it)

	m, err := p.CreateModule("test_system", "0.0.1")
	if err != nil {
		log.Fatal(err)
	}
	print(m)

	imid, err := p.CreateItemModule(it, m)
	if err != nil {
		log.Fatal(err)
	}

	im, err := p.GetItemModule(imid)
	if err != nil {
		log.Fatal(err)
	}
	print(im)

	ims, err := p.GetItemModules()
	if err != nil {
		log.Fatal(err)
	}
	for _, im := range ims {
		print(im)
	}
}

func print(i interface{}) {
	log.Println(strings.Repeat("-", 20))
	log.Println(reflect.TypeOf(i))
	log.Printf("%+v\n", i)
	log.Println(strings.Repeat("-", 20))
}
