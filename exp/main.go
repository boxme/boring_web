package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	post   = 5432
	user   = "desmond"
	dbname = "postgres"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, post, user, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	var id int
	var name, email string
	row := db.QueryRow(`SELECT id, name, email FROM users WHERE id=$1`, 1)
	err = row.Scan(&id, &name, &email)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	fmt.Println("ID:", id, "Name:", name, "Email:", email)
	db.Close()
}
