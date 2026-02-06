package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DATABASE = "/tmp/minitwit.db"
const PER_PAGE = 30
const DEBUG = true
const SECRET_KEY = "development key"

var g struct {
	db   *sql.DB
	user string
}

func connect_db() *sql.DB {
	db, err := sql.Open("sqlite3", DATABASE)
	if err != nil {
		panic(err)
	}

	return db
}

func init_db() {
	db := connect_db()
	defer db.Close()

	g.db = db

	schema, err := os.ReadFile("../schema.sql")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		panic(err)
	}
}


func get_user_id(username string) int {
	rows, err := g.db.Query("select user_id from user where username = ?", username)
	if err != nil {
		panic(err)
	}
	var id int
	if rows.Next() {
		rows.Scan(&id)
		return id 
	}
	return -1
}

func format_datetime(timestamp time.Time) string {
	return timestamp.Format("%Y-%m-%d @ %H:%M")
}

func main() {

	g.db = connect_db()

	defer g.db.Close()
	query_db("SELECT * FROM user", false)

}
