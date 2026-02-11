package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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

func public(w http.ResponseWriter, r *http.Request) {
	g.DB = connect_db()
	defer g.DB.Close()


	data, err := query_db(`
		SELECT message.*, user.* FROM message, user
		WHERE message.flagged = 0 AND message.author_id = user.user_id 
		ORDER BY message.pub_date DESC LIMIT ?`, false, PER_PAGE)
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

func errorGen(err string) error {
	return errors.New(err)
}

func register(w http.ResponseWriter, r *http.Request) {
	if g.User != nil {
		http.Redirect(w, r, "/"+g.User.Username, http.StatusOK)
	}

	g.DB = connect_db()
	defer g.DB.Close()

	registerData := RegisterData{
		Error: "",
		Form: struct {
			Username string
			Email    string
		}{},
	}

	var err error
	if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")

		registerData.Form.Username = username
		registerData.Form.Email = email

		if username == "" {
			err = errorGen("You have to enter a username")
		} else if email  == "" || !strings.Contains(email, "@") {
			err = errorGen("You have to enter a valid email address")
		} else if r.FormValue("password") == "" {
			err = errorGen("You have to enter a password")
		} else if r.FormValue("password") != r.FormValue("password2") {
			err = errorGen("The two passwords do not match")
		} else if val, _ := get_user_id(username); val != -1 {
			err = errorGen("The username is already taken")
		} else {
			pw_hash, err := genereate_password_hash(r.FormValue("password"))
			if err != nil {
				panic(err)
			}
			g.DB.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, pw_hash)
			//TODO: Add notfication popup here
			http.Redirect(w, r, "/", http.StatusOK)
			return
		}
		print(err.Error())
	}
}

func main() {
	// TEMPORARY loading of a user
	g.DB = connect_db()
	_, err := query_db("SELECT * FROM user WHERE user_id = 1", true)
	if err != nil {
		panic(err)
	}
	g.DB.Close()

	// g.User = &User{
	// 	UserID:   int(userData[0]["user_id"].(int64)),
	// 	Username: userData[0]["username"].(string),
	// 	Email:    userData[0]["email"].(string),
	// }

	r := mux.NewRouter()
	r.HandleFunc("/", timeline).Methods("GET")
	r.HandleFunc("/public", public).Methods("GET")
	// r.HandleFunc("/{username}", UserTimelineHandler).Methods("GET")
	// r.HandleFunc("/{username}/follow", FollowUserHandler).Methods("POST")
	// r.HandleFunc("/{username}/unfollow", UnfollowUserHandler).Methods("POST")
	// r.HandleFunc("/add_message", AddMessageHandler).Methods("POST")
	// r.HandleFunc("/login", LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", register).Methods("GET", "POST")
	// r.HandleFunc("/logout", LogoutHandler).Methods("GET")
	// defer g.db.Close()

	println(gravatar_url("augustbrandt170@gmail.com", 80))

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
