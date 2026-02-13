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

type BaseTemplateData struct {
	User    *User
	Flashes []string
}

type RegisterData struct {
	BaseTemplateData
	Error string
	Form  struct {
		Username string
		Email    string
	}
}

type LoginData struct {
	BaseTemplateData
	Error string
	Form  struct {
		Username string
	}
}

type TimelineData struct {
	BaseTemplateData
	Messages    []*Message
	ProfileUser *User
	Follows     bool
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

	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		panic(err)
	}
	// TEMPORARY: Insert a user and a message for testing
	_, err = db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", "testuser", "testuser@hotmail.com", "testpassword")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, ?)", 1, "Hello world!", time.Now().Unix(), 0)
	if err != nil {
		panic(err)
	}
}

func ensureDB() {
	if _, err := os.Stat(DATABASE); os.IsNotExist(err) {
		fmt.Println("Database does not exist. Initializing...")
		init_db()
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

	if len(out) == 0 {
		return nil, errors.New("Query returned no rows")
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

func genereate_password_hash(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 14)
	return string(bytes), err
}

func check_password_hash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func format_datetime(timestamp time.Time) string {
	// This example time is the reference time for golang time formatting
	return timestamp.Format("2006-01-02 @ 15:04")
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
			Username: template.HTMLEscapeString(message["username"].(string)),
			Email:    template.HTMLEscapeString(message["email"].(string)),
		}
		newMessage := &Message{
			MessageID: int(message["message_id"].(int64)),
			Author:    messageAuthor,
			Text:      template.HTMLEscapeString(message["text"].(string)),
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
		BaseTemplateData: BaseTemplateData{
			User:    g.User, // Pass the current user (nil in this case)
			Flashes: Flashes,
		},
		Messages:    messages,
		ProfileUser: g.User,
		Endpoint:    "timeline",
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
		BaseTemplateData: BaseTemplateData{
			User:    g.User, // Pass the current user (nil in this case)
			Flashes: Flashes,
		},
		Messages:    messages,
		ProfileUser: g.User,
		Endpoint:    "public_timeline",
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

func UserTimelineHandler(w http.ResponseWriter, r *http.Request) {
	// TEMPORARY DATABASE CONNECTION CREATION
	g.DB = connect_db()
	defer g.DB.Close()

	// Get username from path
	username := mux.Vars(r)["username"]

	// Check existance of user in database
	data, err := query_db("select user_id, email from user where username = ?", true, username)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	userId := data[0]["user_id"].(int64)
	userEmail := data[0]["email"].(string)

	// Get messages data
	data, err = query_db(`
		select message.*, user.* from message, user where
        user.user_id = message.author_id and user.user_id = ?
        order by message.pub_date desc limit ?`, false, userId, PER_PAGE)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	messages := createTimelineMessages(data)

	User := &User{
		UserID:   int(userId),
		Username: username,
		Email:    userEmail,
	}

	follows := false
	if g.User != nil {
		queryCheckUserIsFollowed, err := query_db(
			`select 1 from follower
		where follower.who_id = ?
		and follower.whom_id = ?`, true,
			g.User.UserID, User.UserID,
		)

		if err == nil {
			if queryCheckUserIsFollowed[0]["user_id"] != nil {
				follows = true
			}
		}
	}

	templateData := TimelineData{
		BaseTemplateData: BaseTemplateData{
			User:    g.User, // Pass the current user (nil in this case)
			Flashes: Flashes,
		},
		Messages:    messages,
		ProfileUser: User,
		Endpoint:    "user_timeline",
		Follows:     follows,
	}

	template, err := template.New("layout.html").Funcs(template.FuncMap{
		"gravatar":        gravatar_url,
		"format_datetime": format_datetime,
	}).
		ParseFiles("templates/layout.html", "templates/timeline.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = template.Execute(w, templateData)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func errorGen(err string) error {
	return errors.New(err)
}

func login(w http.ResponseWriter, r *http.Request) {
	if g.User != nil {
		http.Redirect(w, r, "/"+g.User.Username, http.StatusFound)
		return
	}

	g.DB = connect_db()
	defer g.DB.Close()

	loginData := LoginData{
		BaseTemplateData: BaseTemplateData{
			User:    g.User,
			Flashes: []string{},
		},
		Error: "",
		Form: struct {
			Username string
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
		password := r.FormValue("password")

		user := User{}
		var password_hash string

		err := g.DB.QueryRow("SELECT * FROM user WHERE username = ?", username).Scan(&user.UserID, &user.Username, &user.Email, &password_hash)
		if err != nil {
			// This line is kinda redudundant since we override it based on what was wrong later down
			loginData.Error = "Invalid username or password"
		}

		// THIS IS SO BAD FOR SECURITY HOLY HELL
		//TODO: FIX THIS ASAP WHEN WE ACTUALLY REFACTOR FOR REAL
		if user.Username == "" {
			loginData.Error = "Invalid username"
		} else if !check_password_hash(password, password_hash) {
			loginData.Error = "Invalid password "
		} else {
			//TODO: Add flash login message
			g.User = &user
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

	}
	tmpl, err := template.New("layout.html").
		Funcs(template.FuncMap{
			"gravatar":        gravatar_url,
			"format_datetime": format_datetime,
		}).
		ParseFiles("templates/layout.html", "templates/login.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, loginData)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	if g.User != nil {
		http.Redirect(w, r, "/"+g.User.Username, http.StatusSeeOther)
		return
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
		} else if email == "" || !strings.Contains(email, "@") {
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
			_, err = g.DB.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, pw_hash)
			if err != nil {
				panic(err)
			}
			//TODO: Add notfication popup here
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	}
	// Parse and execute template
	tmpl, err := template.New("layout.html").
		Funcs(template.FuncMap{
			"gravatar":        gravatar_url,
			"format_datetime": format_datetime,
		}).
		ParseFiles("templates/layout.html", "templates/register.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, registerData)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func addMessage(w http.ResponseWriter, r *http.Request) {
	if g.User == nil {
		log.Println("Tried to add message but no user is set")
		http.Error(w, "No user is logged in", http.StatusUnauthorized)
		return
	}

	// TEMPORARY DATABASE CONNECTION
	g.DB = connect_db()
	defer g.DB.Close()

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messageText := r.FormValue("text")
	if messageText != "" {
		g.DB.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)",
			g.User.UserID, messageText, int(time.Now().Unix()))
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: Add logout message
	if g.User == nil {
		http.Error(w, "No user is logged in", http.StatusConflict)
		return
	}
	g.User = nil
	http.Redirect(w, r, "/public", http.StatusFound)
}

func main() {
	ensureDB()
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
	// r.HandleFunc("/{username}/follow", FollowUserHandler).Methods("POST")
	// r.HandleFunc("/{username}/unfollow", UnfollowUserHandler).Methods("POST")
	r.HandleFunc("/add_message", addMessage).Methods("POST")
	r.HandleFunc("/login", login).Methods("GET", "POST")
	r.HandleFunc("/register", register).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/{username}", UserTimelineHandler).Methods("GET")
	// defer g.db.Close()

	println(gravatar_url("augustbrandt170@gmail.com", 80))

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
