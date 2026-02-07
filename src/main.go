package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
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

// THIS FUNCTION IS DISGUSTING
func query_db(query string, one bool, args ...any) []map[string]any {
	var err error
	var rows *sql.Rows
	if args == nil {
		rows, err = g.db.Query(query)
	} else {
		rows, err = g.db.Query(query, args)
	}
	if err != nil {
		panic(err)
	}

	print()

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	out := []map[string]any{}

	for rows.Next() {
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			panic(err)
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		out = append(out, row)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	print(out[0]["username"])

	if one {
		return []map[string]any{out[0]}
	}
	return out
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

func gravatar_url(email string, size int) string {
	emailHash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("http://www.gravatar.com/avatar/%s?d=identicon&s=%d", hex.EncodeToString(emailHash[:]), size)
}

func main() {

	// g.db = connect_db()

	// defer g.db.Close()
	// query_db("SELECT * FROM user", false)
	print(gravatar_url("augustbrandt170@gmail.com", 80))
}
