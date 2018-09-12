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
	row := db.QueryRow(`INSERT INTO users(name, email) VALUES($1, $2) RETURNING id`, "Desmond Ng", "dnyy@email.com")
	err = row.Scan(&id)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	db.Close()
}
