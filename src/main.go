package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	//"html/template"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/gorilla/mux"
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
func query_db(query string, one bool, args ...any) ([]map[string]any, error) {
	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out := []map[string]any{}

	for rows.Next() {
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		out = append(out, row)
		// Terminate early if we only want one result
		if one {
			break
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	println(out[0]["username"].(string))

	return out, nil
}

func get_user_id(username string) (int, error) {
	var id int
	err := g.db.QueryRow("select user_id from user where username = ?", username).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func format_datetime(timestamp time.Time) string {
	return timestamp.Format("%Y-%m-%d @ %H:%M")
}

func gravatar_url(email string, size int) string {
	emailHash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("http://www.gravatar.com/avatar/%s?d=identicon&s=%d", hex.EncodeToString(emailHash[:]), size)
}

func TimelineHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
	println("We got a visitor from: ", r.RemoteAddr)
    // TODO: Render timeline template 0-0
}

func main() {

	g.db = connect_db()
	r := mux.NewRouter()
	r.HandleFunc("/", TimelineHandler).Methods("GET")
	/*r.HandleFunc("/public", PublicTimelineHandler).Methods("GET")
	r.HandleFunc("/{username}", UserTimelineHandler).Methods("GET")
	r.HandleFunc("/{username}/follow", FollowUserHandler).Methods("POST")
	r.HandleFunc("/{username}/unfollow", UnfollowUserHandler).Methods("POST")
	r.HandleFunc("/add_message", AddMessageHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", RegisterHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", LogoutHandler).Methods("GET")*/
	// defer g.db.Close()
	query_db("SELECT * FROM user", false)

	println(gravatar_url("augustbrandt170@gmail.com", 80))
	
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
