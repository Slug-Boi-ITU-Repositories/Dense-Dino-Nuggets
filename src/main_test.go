package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"
)

const BASE_URL = "http://localhost:8080"

func helper_register(client *http.Client, username, password string, password2 *string, email *string) (*http.Response, error) {
	//Helper function to register a user
	if password2 == nil {
		password2 = &password
	}

	if email == nil {
		tmp := username + "@example.com"
		email = &tmp
	}

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("password2", *password2)
	data.Set("email", *email)

	resp, err := client.PostForm(BASE_URL+"/register", data)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func helper_login(username string, password string) (*http.Response, *http.Client, error) {
	//Helper function to login
	http_session := &http.Client{}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, err
	}

	http_session.Jar = jar

	r, err := http_session.PostForm(BASE_URL+"/login", url.Values{
		"username": {username},
		"password": {password},
	})

	if err != nil {
		return nil, nil, err
	}

	return r, http_session, nil
}

func register_and_login(client *http.Client, username string, password string) (*http.Response, *http.Client, error) {
	//Registers and logs in in one go
	helper_register(client, username, password, nil, nil)
	return helper_login(username, password)
}

func logout(http_session *http.Client) (*http.Response, error) {
	//Helper function to logout
	return http_session.Get(BASE_URL + "/logout")
}

func add_message(http_session *http.Client, text string) (*http.Response, error) {
	//Records a message
	r, err := http_session.PostForm(BASE_URL+"/add_message", url.Values{
		"text": {text},
	})
	if err != nil {
		return nil, err
	}

	if text != "" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body.Close()

		if !strings.Contains(string(body), "Your message was recorded") {
			return nil, fmt.Errorf("success response not found")
		}
	}

	return r, nil
}

// Lil' helper function for assertions
func assertContains(t *testing.T, body, substr string) {
	t.Helper()
	if !strings.Contains(body, substr) {
		t.Fatalf("expected %q . \nGot:\n%s", substr, body)
	}
}

// lil' helper function for reading the response and converting it to a string
func readBody(t *testing.T, r *http.Response) string {
	t.Helper()
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func newTestClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

// Testing functions
func TestRegister(t *testing.T) {
	client := newTestClient()
	//Make sure registering works
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())

	r, err := helper_register(client, username, "default", nil, nil)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You were successfully registered and can login now")

	//_, _ = client.Get(BASE_URL + "/login")

	r, err = helper_register(client, username, "default", nil, nil)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	assertContains(t, readBody(t, r), "The username is already taken")

	r, err = helper_register(client, "", "default", nil, nil)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You have to enter a username")

	r, err = helper_register(client, "meh", "", nil, nil)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You have to enter a password")

	p2 := "y"
	r, err = helper_register(client, "meh", "x", &p2, nil)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	assertContains(t, readBody(t, r), "The two passwords do not match")

	email := "broken"
	r, err = helper_register(client, "meh", "foo", nil, &email)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You have to enter a valid email address")
}

func test_loging_logout() {
	//Make sure logging in and logging out works
}

func test_message_recording() {
	//Check if adding messages works
}

func test_timelines() {
	//Make sure that timelines work
}
