package main

import (
	"database/sql"
	"datafox/service"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

//type ResponseFunction func(w http.ResponseWriter, r *http.Request)
//
//const TableInitQuery = `
//	CREATE TABLE IF NOT EXISTS users (
//	    id INTEGER PRIMARY KEY AUTOINCREMENT,
//	    username TEXT NOT NULL,
//	    password TEXT NOT NULL
//	)
//`
//
//const InitAdminQuery = `
//	INSERT INTO users (username, password) VALUES ('admin', 'admin');
//`

func main() {

	open, err := sql.Open("sqlite3", "./meta.db")
	if err != nil {
		log.Panicln(err)
		return
	}

	handler := service.NewRouteHandler(open)

	log.Println(handler.Run())
}
