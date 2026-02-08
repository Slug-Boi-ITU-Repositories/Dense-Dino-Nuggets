package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	UserID   int
	Username string
	Email    string
}

type Message struct {
	MessageID int
	Author    *User
	Text      string
	PubTime   time.Time
	Flagged   int
}

type TimelineData struct {
	Messages    []*Message
	User        *User
	ProfileUser *User // For specific user profile pages e.g. (/helgecph)
	Follows     bool  // If the logged in user follows the profile user
	Flashes     []string
	Endpoint    string
}

const DATABASE = "/tmp/minitwit.db"
const PER_PAGE = 30
const DEBUG = true
const SECRET_KEY = "development key"

var g struct {
	DB   *sql.DB
	User *User
}

func connect_db() *sql.DB {
	db, err := sql.Open("sqlite3", DATABASE)
	if err != nil {
		panic(err)
	}

	return db
}

var Flashes []string

func init_db() {
	db := connect_db()
	defer db.Close()

	g.DB = db

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
	rows, err := g.DB.Query(query, args...)
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

	if one {
		return []map[string]any{out[0]}, nil
	}
	return out, nil
}

func get_user_id(username string) (int, error) {
	var id int
	err := g.DB.QueryRow("select user_id from user where username = ?", username).Scan(&id)
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

func createTimelineMessages(queryResult []map[string]any) []*Message {
	messages := make([]*Message, len(queryResult))
	for i, message := range queryResult {
		messageAuthor := &User{
			UserID:   int(message["author_id"].(int64)),
			Username: message["username"].(string),
			Email:    message["email"].(string),
		}
		newMessage := &Message{
			MessageID: int(message["message_id"].(int64)),
			Author:    messageAuthor,
			Text:      message["text"].(string),
			PubTime:   time.Unix(message["pub_date"].(int64), 0),
			Flagged:   int(message["flagged"].(int64)),
		}
		messages[i] = newMessage
	}
	return messages
}

func timeline(w http.ResponseWriter, r *http.Request) {
	// TEMPORARY DATABASE CONNECTION CREATION
	g.DB = connect_db()
	defer g.DB.Close()

	fmt.Printf("We got a visitor from: %s\n", r.RemoteAddr)
	if g.User == nil {
		http.Redirect(w, r, "/public", http.StatusOK)
		return
	}
	data, err := query_db(`
		SELECT message.*, user.* FROM message, user
		WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
			user.user_id = ? OR
			user.user_id IN (SELECT whom_id FROM follower
								WHERE who_id = ?)
		) ORDER BY message.pub_date DESC LIMIT ?`, false, g.User.UserID, g.User.UserID, PER_PAGE)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messages := createTimelineMessages(data)

	templateData := TimelineData{
		Messages:    messages,
		User:        g.User,
		ProfileUser: g.User,
		Flashes:     Flashes,
	}

	tmpl, err := template.New("layout.html").
		Funcs(template.FuncMap{
			"gravatar":        gravatar_url,
			"format_datetime": format_datetime,
		}).
		ParseFiles("templates/layout.html", "templates/timeline.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, templateData)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	// mux := http.NewServeMux()
	g.DB = connect_db()
	userData, err := query_db("SELECT * FROM user WHERE user_id = 1", false)
	if err != nil {
		panic(err)
	}
	g.User = &User{
		UserID:   int(userData[0]["user_id"].(int64)),
		Username: userData[0]["username"].(string),
		Email:    userData[0]["email"].(string),
	}
	g.DB.Close()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", timeline)

	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
