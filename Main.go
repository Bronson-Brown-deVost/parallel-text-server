package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	db, err = sql.Open("mysql", "bronson:none@/Text_Comparison?charset=utf8")
	checkErr(err)
	db.SetMaxOpenConns(100)
}

func main() {
	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
