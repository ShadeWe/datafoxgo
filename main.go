package main

import (
	"database/sql"
	"datafox/service"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {

	open, err := sql.Open("sqlite3", "./meta.db")
	if err != nil {
		log.Panicln(err)
		return
	}

	handler := service.NewRouteHandler(open)

	log.Println(handler.Run())
}
