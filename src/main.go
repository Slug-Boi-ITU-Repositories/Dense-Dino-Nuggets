package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	openapi "minitwit/src/apimodels/go"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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
	Flashes []interface{}
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
	DB *sql.DB
}

var store = sessions.NewCookieStore([]byte("your-secret-key-here-at-least-32-bytes"))

// Get the logged in user from the user session.
//
// If the user pointer is nil and the error is nil then no user is logged in.
func getUser(r *http.Request) (*User, error) {
	user_session, err := store.Get(r, "user-session")
	if err != nil {
		return nil, err
	}
	if _, exists := user_session.Values["user"]; !exists {
		return nil, nil
	}
	user := &User{}
	err = json.Unmarshal(user_session.Values["user"].([]byte), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func connect_db() *sql.DB {
	db, err := sql.Open("sqlite3", DATABASE)
	if err != nil {
		panic(err)
	}
	return db
}

func getFlashes(r *http.Request, w http.ResponseWriter) ([]interface{}, error) {
	session, err := store.Get(r, "app-session")
	if err != nil {
		return nil, err
	}

	flashes := session.Flashes()
	session.Save(r, w)

	return flashes, nil
}

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
		return nil, sql.ErrNoRows
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

	user, err := getUser(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("We got a visitor from: %s\n", r.RemoteAddr)
	if user == nil {
		http.Redirect(w, r, "/public", http.StatusFound)
		return
	}
	data, err := query_db(`
		SELECT message.*, user.* FROM message, user
		WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
			user.user_id = ? OR
			user.user_id IN (SELECT whom_id FROM follower
								WHERE who_id = ?)
		) ORDER BY message.pub_date DESC LIMIT ?`, false, user.UserID, user.UserID, PER_PAGE)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messages := createTimelineMessages(data)

	flashes, err := getFlashes(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateData := TimelineData{
		BaseTemplateData: BaseTemplateData{
			User:    user, // Pass the current user (nil in this case)
			Flashes: flashes,
		},
		Messages:    messages,
		ProfileUser: user,
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
	user, err := getUser(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	g.DB = connect_db()
	defer g.DB.Close()

	data, err := query_db(`
		SELECT message.*, user.* FROM message, user
		WHERE message.flagged = 0 AND message.author_id = user.user_id
		ORDER BY message.pub_date DESC LIMIT ?`, false, PER_PAGE)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	messages := createTimelineMessages(data)

	flashes, err := getFlashes(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateData := TimelineData{
		BaseTemplateData: BaseTemplateData{
			User:    user, // Pass the current user (nil in this case)
			Flashes: flashes,
		},
		Messages:    messages,
		ProfileUser: user,
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
	user, err := getUser(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	pageUser := &User{
		UserID:   int(userId),
		Username: username,
		Email:    userEmail,
	}
	// Get messages data
	data, err = query_db(`
		select message.*, user.* from message, user where
        user.user_id = message.author_id and user.user_id = ?
        order by message.pub_date desc limit ?`, false, userId, PER_PAGE)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	messages := createTimelineMessages(data)

	follows := false
	if user != nil {
		queryCheckUserIsFollowed, err := query_db(
			`select 1 from follower
		where follower.who_id = ?
		and follower.whom_id = ?`, true,
			user.UserID, pageUser.UserID,
		)

		if err == nil {
			if len(queryCheckUserIsFollowed) > 0 {
				follows = true
			}
		}
	}

	flashes, err := getFlashes(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateData := TimelineData{
		BaseTemplateData: BaseTemplateData{
			User:    user, // Pass the current user (nil in this case)
			Flashes: flashes,
		},
		Messages:    messages,
		ProfileUser: pageUser,
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

// Adds the current user as follower of the given user.
func FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	g.DB = connect_db()
	defer g.DB.Close()

	// Check if user is logged in
	if user == nil {
		http.Error(w, http.StatusText(401), 401)
		return
	}
	// Get id of user to follow
	username := mux.Vars(r)["username"]
	whom_id, err := get_user_id(username)

	if err != nil {
		http.Error(w, http.StatusText(401), 401)
		return
	}
	//Insert follow into database
	_, err = g.DB.Exec("insert into follower (who_id, whom_id) values (?, ?)", user.UserID, whom_id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	session.AddFlash(fmt.Sprintf("You are now following \"%s\"", username))
	session.Save(r, w)

	url := "/" + username
	http.Redirect(w, r, url, http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if user != nil {
		http.Redirect(w, r, "/"+user.Username, http.StatusFound)
		return
	}

	g.DB = connect_db()
	defer g.DB.Close()

	// Create session to add flashes
	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
			//loginData.Error = "Invalid username or password"
		}

		user_session, _ := store.Get(r, "user-session")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// THIS IS SO BAD FOR SECURITY HOLY HELL
		//TODO: FIX THIS ASAP WHEN WE ACTUALLY REFACTOR FOR REAL
		if user.Username == "" {
			err = errors.New("Invalid username")
			session.AddFlash("Invalid username")
			session.Save(r, w)
		} else if !check_password_hash(password, password_hash) {
			err = errors.New("Invalid password")
			session.AddFlash("Invalid password")
			session.Save(r, w)
		} else {
			session.AddFlash("You were logged in")
			session.Save(r, w)

			userJson, err := json.Marshal(user)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			user_session.Values["user"] = userJson
			user_session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

	}

	flashes, err := getFlashes(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Figure out if we ever actually use the data error if not then just remove it all
	var err_str string
	if err != nil {
		err_str = err.Error()
	} else {
		err_str = ""
	}

	loginData := LoginData{
		BaseTemplateData: BaseTemplateData{
			User:    user,
			Flashes: flashes,
		},
		Error: err_str,
		Form: struct {
			Username string
		}{},
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
	user, err := getUser(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user != nil {
		http.Redirect(w, r, "/"+user.Username, http.StatusSeeOther)
		return
	}

	g.DB = connect_db()
	defer g.DB.Close()

	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var username, email string

	if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		username = r.FormValue("username")
		email = r.FormValue("email")

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

			session.AddFlash("You were successfully registered and can login now")
			session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	}

	flashes, err := getFlashes(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	registerData := RegisterData{
		BaseTemplateData: BaseTemplateData{
			Flashes: flashes,
		},
		Error: "",
		Form: struct {
			Username string
			Email    string
		}{username, email},
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
	user, err := getUser(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		log.Println("Tried to add message but no user is set")
		http.Error(w, "No user is logged in", http.StatusUnauthorized)
		return
	}

	// TEMPORARY DATABASE CONNECTION
	g.DB = connect_db()
	defer g.DB.Close()

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messageText := r.FormValue("text")
	if messageText != "" {
		g.DB.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)",
			user.UserID, messageText, int(time.Now().Unix()))
	}

	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	session.AddFlash("Your message was recorded")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//TODO: Add logout message
	if user == nil {
		http.Error(w, "No user is logged in", http.StatusConflict)
		return
	}
	user_session, err := store.Get(r, "user-session")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	delete(user_session.Values, "user")
	user_session.Save(r, w)

	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	session.AddFlash("You were logged out")
	session.Save(r, w)
	http.Redirect(w, r, "/public", http.StatusFound)
}

// Removes the current user as follower of the given user.
func UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	g.DB = connect_db()
	defer g.DB.Close()

	// Check if user is logged in
	if user == nil {
		http.Error(w, http.StatusText(401), 401)
		return
	}

	// Get id of user to unfollow
	username := mux.Vars(r)["username"]
	whom_id, err := get_user_id(username)

	_, err = g.DB.Exec("delete from follower where who_id=? and whom_id=?", user.UserID, whom_id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	session.AddFlash(fmt.Sprintf("You are no longer following \"%s\"", username))
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	url := "/" + username
	http.Redirect(w, r, url, http.StatusFound)
}

func main() {
	log.Printf("Server started")

	MinitwitAPIService := openapi.NewMinitwitAPIService()
	MinitwitAPIController := openapi.NewMinitwitAPIController(MinitwitAPIService)

	router := openapi.NewRouter(MinitwitAPIController)

	ensureDB()
	// g.DB = connect_db()
	// _, err := query_db("SELECT * FROM user WHERE user_id = 1", true)
	// if err != nil {
	// 	panic(err)
	// }
	// g.DB.Close()

	// g.User = &User{
	// 	UserID:   int(userData[0]["user_id"].(int64)),
	// 	Username: userData[0]["username"].(string),
	// 	Email:    userData[0]["email"].(string),
	// }

	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	router.HandleFunc("/", timeline).Methods("GET")
	router.HandleFunc("/public", public).Methods("GET")
	router.HandleFunc("/add_message", addMessage).Methods("POST")
	router.HandleFunc("/login", login).Methods("GET", "POST")
	router.HandleFunc("/register-user", register).Methods("GET", "POST")
	router.HandleFunc("/logout", logoutHandler).Methods("GET")
	router.PathPrefix("/static/").Handler(s).Methods("GET")

  	router.HandleFunc("/{username}/follow", FollowUserHandler).Methods("GET")
	router.HandleFunc("/{username}/unfollow", UnfollowUserHandler).Methods("GET")
	router.HandleFunc("/{username}", UserTimelineHandler).Methods("GET")
	// defer g.db.Close()

	println(gravatar_url("augustbrandt170@gmail.com", 80))

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8080", router))
}
