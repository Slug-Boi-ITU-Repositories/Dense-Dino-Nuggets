package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"minitwit/src/db"
	"minitwit/src/model"
	"minitwit/src/repository"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	openapi "minitwit/src/apimodels/go"
	"minitwit/src/monitor"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
    Messages    []model.Message
    ProfileUser *User
    Follows     bool
    Endpoint    string
    Page        int
    TotalPages  int
}

const PER_PAGE = 30
const DEBUG = true
const SECRET_KEY = "development key"

var store = sessions.NewCookieStore([]byte("your-secret-key-here-at-least-32-bytes"))

var GormDB *gorm.DB

// Add repositories as globals
var UserRepo *repository.UserRepository
var MessageRepo *repository.MessageRepository
var FollowerRepo *repository.FollowerRepository

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

func getFlashes(r *http.Request, w http.ResponseWriter) ([]interface{}, error) {
	session, err := store.Get(r, "app-session")
	if err != nil {
		return nil, err
	}

	flashes := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		return nil, err
	}

	return flashes, nil
}

func init_db() {
	// Create test user with hashed password
	pwHash, err := generate_password_hash("testpassword")
	if err != nil {
		log.Printf("Warning: failed to hash password for test user: %v\n", err)
		return
	}
	testUser := model.User{
		Username: "testuser",
		Email:    "testuser@hotmail.com",
		PwHash:   pwHash,
	}
	err = UserRepo.Create(&testUser)
	if err != nil {
		log.Printf("Warning: failed to create test user: %v\n", err)
		return
	}

	// Create test message
	testMessage := model.Message{
		AuthorID: testUser.UserID,
		Text:     "Hello world!",
		PubDate:  time.Now().Unix(),
		Flagged:  0,
	}
	err = MessageRepo.Create(&testMessage)
	if err != nil {
		log.Printf("Warning: failed to create test message: %v\n", err)
		return
	}
}

func generate_password_hash(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 10)
	return string(bytes), err
}

func check_password_hash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func format_datetime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 @ 15:04")
}

func gravatar_url(email string, size int) string {
	emailHash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("http://www.gravatar.com/avatar/%s?d=identicon&s=%d", hex.EncodeToString(emailHash[:]), size)
}

func timeline(w http.ResponseWriter, r *http.Request) {
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

    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    offset := (page - 1) * PER_PAGE

    messages, err := MessageRepo.GetPersonalTimeline(uint(user.UserID), PER_PAGE, offset)
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    totalMessages, err := MessageRepo.CountPersonalTimeline(uint(user.UserID))
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    totalPages := int(math.Ceil(float64(totalMessages) / float64(PER_PAGE)))

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
        Page:        page,
        TotalPages:  totalPages,
    }

	tmpl, err := template.New("layout.html").
		Funcs(template.FuncMap{
			"gravatar":        gravatar_url,
			"format_datetime": format_datetime,
			"pred": func(i int) int { return i - 1 },
			"succ": func(i int) int { return i + 1 },
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

    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    offset := (page - 1) * PER_PAGE

    messages, err := MessageRepo.GetPublicTimeline(PER_PAGE, offset)
    if err != nil {
        log.Println(err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    totalMessages, err := MessageRepo.CountPublicTimeline()
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    totalPages := int(math.Ceil(float64(totalMessages) / float64(PER_PAGE)))

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
        Page:        page,
        TotalPages:  totalPages,
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

	// Get username from path
	username := mux.Vars(r)["username"]

    // Check existence of user in database
    data, err := UserRepo.GetUserByUsername(username)
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    userId := data.UserID
    userEmail := data.Email
    pageUser := &User{
        UserID:   int(userId),
        Username: username,
        Email:    userEmail,
    }

    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    offset := (page - 1) * PER_PAGE

    // Get messages data
    messages, err := MessageRepo.GetUserTimeline(uint(userId), PER_PAGE, offset)
    if err != nil {
        log.Println(err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    totalMessages, err := MessageRepo.CountUserTimeline(uint(userId))
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    totalPages := int(math.Ceil(float64(totalMessages) / float64(PER_PAGE)))

	follows := false
	if user != nil {
		queryCheckUserIsFollowed, err := UserRepo.IsFollowing(uint(user.UserID), uint(userId))
		if err == nil {
			if queryCheckUserIsFollowed {
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
        Page:        page,
        TotalPages:  totalPages,
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

	// Check if user is logged in
	if user == nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// Get id of user to follow
	username := mux.Vars(r)["username"]
	whom_id, err := UserRepo.GetUserIDByUsername(username)
	if err != nil {
		http.Error(w, "No user was found", http.StatusNotFound)
		return
	}
	//Insert follow into database
	err = FollowerRepo.Follow(uint(user.UserID), whom_id)
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
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	// Create session to add flashes
	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var loginErr error
	if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user_session, err := store.Get(r, "user-session")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		// Get user from repository
		modelUser, err := UserRepo.GetUserByUsername(username)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				loginErr = errors.New("invalid username")
				session.AddFlash("Invalid username")
				err = session.Save(r, w)
				if err != nil {
					log.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if !check_password_hash(password, modelUser.PwHash) {
			loginErr = errors.New("invalid password")
			session.AddFlash("Invalid password")
			err = session.Save(r, w)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Convert model.User to User for session
			user := User{
				UserID:   int(modelUser.UserID),
				Username: modelUser.Username,
				Email:    modelUser.Email,
			}

			session.AddFlash("You were logged in")
			err = session.Save(r, w)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			userJson, err := json.Marshal(user)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			user_session.Values["user"] = userJson
			err = user_session.Save(r, w)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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

	errStr := errString(loginErr)

	loginData := LoginData{
		BaseTemplateData: BaseTemplateData{
			User:    user,
			Flashes: flashes,
		},
		Error: errStr,
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

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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
		} else if val, _ := UserRepo.GetUserIDByUsername(username); val != 0 {
			err = errorGen("The username is already taken")
		} else {
			pw_hash, err := generate_password_hash(r.FormValue("password"))
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = UserRepo.RegisterUser(username, email, pw_hash)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			session.AddFlash("You were successfully registered and can login now")
			err = session.Save(r, w)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	}

	flashes, flashErr := getFlashes(r, w)
	if flashErr != nil {
		log.Println(flashErr.Error())
		http.Error(w, flashErr.Error(), http.StatusInternalServerError)
		return
	}

	registerData := RegisterData{
		BaseTemplateData: BaseTemplateData{
			Flashes: flashes,
		},
		Error: errString(err),
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

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messageText := r.FormValue("text")
	if messageText != "" {
		err = MessageRepo.AddMessage(uint(user.UserID), messageText)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	session, err := store.Get(r, "app-session")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	session.AddFlash("Your message was recorded")
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	err = user_session.Save(r, w)
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

	session.AddFlash("You were logged out")
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	// Check if user is logged in
	if user == nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get id of user to unfollow
	username := mux.Vars(r)["username"]
	whom_id, err := UserRepo.GetUserIDByUsername(username)
	if err != nil {
		log.Println(err)
		http.Error(w, "User could not be found", http.StatusNotFound)
		return
	}

	err = FollowerRepo.Unfollow(uint(user.UserID), whom_id)
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
	reg := prometheus.NewRegistry()
	// reg.MustRegister(

	// 	collectors.NewGoCollector(),
	// 	collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	// )

	log.Printf("Server started")
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // No ssl cert
		SameSite: http.SameSiteLaxMode,
	}

	MinitwitAPIService := openapi.NewMinitwitAPIService()
	MinitwitAPIController := openapi.NewMinitwitAPIController(MinitwitAPIService)

	router := openapi.NewRouter(MinitwitAPIController)
	router.Use(monitor.MetricsMiddleware(monitor.NewMetrics(reg)))

	// Check if DATABASE_URL is set in environment first
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Try loading from .env file
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file: ", err)
		}
    
		dsn = os.Getenv("DATABASE_URL")
		if dsn == "" {
			log.Fatal("DATABASE_URL environment variable is not set in environment or .env file!")
		}
	}

	// Create global GORM connection
	GormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatal("Failed to connect to database with GORM:", err)
	}
	// Initialize repositories
	UserRepo = repository.NewUserRepository(GormDB)
	MessageRepo = repository.NewMessageRepository(GormDB)
	FollowerRepo = repository.NewFollowerRepository(GormDB)
	// Seed database with initial data if empty
	var userCount int64
	GormDB.Model(&model.User{}).Count(&userCount)
	if userCount == 0 {
		init_db()
	}
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	router.Handle("/", openapi.Logger(http.HandlerFunc(timeline), "My timeline")).Methods("GET")
	router.Handle("/public", openapi.Logger(http.HandlerFunc(public), "Public timeline")).Methods("GET")
	router.Handle("/add_message", openapi.Logger(http.HandlerFunc(addMessage), "Posting tweet")).Methods("POST")
	router.Handle("/login", openapi.Logger(http.HandlerFunc(login), "Login")).Methods("GET", "POST")
	router.Handle("/register-user", openapi.Logger(http.HandlerFunc(register), "Register User")).Methods("GET", "POST")
	router.Handle("/logout", openapi.Logger(http.HandlerFunc(logoutHandler), "Logout")).Methods("GET")
	router.PathPrefix("/static/").Handler(s).Methods("GET")
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	router.Handle("/{username}/follow", openapi.Logger(http.HandlerFunc(FollowUserHandler), "Following")).Methods("GET")
	router.Handle("/{username}/unfollow", openapi.Logger(http.HandlerFunc(UnfollowUserHandler), "Unfollowing")).Methods("GET")
	router.Handle("/{username}", openapi.Logger(http.HandlerFunc(UserTimelineHandler), "User timeline")).Methods("GET")


	println(gravatar_url("augustbrandt170@gmail.com", 80))

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8080", router))
}
