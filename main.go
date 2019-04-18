package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type human struct {
	Name string
	Age  int
}

func main() {
	connStr := "user=postgres password=secret host=postgres-service dbname=conmansys sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM human")
	if err != nil {
		log.Fatal(err)
	}

	var h human

	for rows.Next() {
		err := rows.Scan(&h.Name, &h.Age)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%+v\n", h)
	}
}
