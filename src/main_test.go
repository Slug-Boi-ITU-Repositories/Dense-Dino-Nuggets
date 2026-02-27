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
	r, err := helper_register(client, username, password, nil, nil)
	if err != nil {
		return nil, client, err
	}
	defer r.Body.Close()
	r, err = helper_login_with_client(client, username, password)
	if err != nil {
		return nil, client, err
	}
	return r, client, nil
}

func helper_login_with_client(client *http.Client, username, password string) (*http.Response, error) {
	r, err := client.PostForm(BASE_URL+"/login", url.Values{
		"username": {username},
		"password": {password},
	})
	if err != nil {
		return nil, err
	}
	return r, nil
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
		//r.Body.Close()

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

// Lil' helper function for assertions
func assertContainsNot(t *testing.T, body, substr string) {
	t.Helper()
	if strings.Contains(body, substr) {
		t.Fatalf(" %q in body. \n", substr)
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

func TestLogin_logout(t *testing.T) {
	//Make sure logging in and logging out works
	client := newTestClient()
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())

	r, http_client, err := register_and_login(client, username, "default")
	if err != nil {
		t.Fatalf("register_and_login failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You were logged in")

	r, err = logout(http_client)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You were logged out")

	r, http_client, err = helper_login(username, "wrongpassword")
	if err != nil {
		t.Fatalf("helper_login failed: %v", err)
	}
	assertContains(t, readBody(t, r), "Invalid password")

	// This has to use a username that is not in the db
	r, http_client, err = helper_login("notInDB", "wrongpassword")
	if err != nil {
		t.Fatalf("helper_login failed: %v", err)
	}
	assertContains(t, readBody(t, r), "Invalid username")
}

func TestMessage_recording(t *testing.T) {
	//Check if adding messages works
	client := newTestClient()
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())

	r, http_client, err := register_and_login(client, username, "default")
	if err != nil {
		t.Fatalf("register_and_login failed: %v", err)
	}
	io.ReadAll(r.Body) // Not sure if this is needed
	r.Body.Close()

	r, err = add_message(http_client, "test message 1")
	if err != nil {
		t.Fatalf("add_message failed: %v", err)
	}

	r, err = add_message(http_client, "<test message 2>")
	if err != nil {
		t.Fatalf("add_message failed: %v", err)
	}

	r, err = http_client.Get(BASE_URL + "/")
	if err != nil {
		t.Fatalf("get timeline failed: %v", err)
	}

	body := readBody(t, r)

	assertContains(t, body, "test message 1")
	assertContains(t, body, "&lt;test message 2&gt;")
}

func TestTimelines(t *testing.T) {
	//Make sure that timelines work
	client := newTestClient()
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())

	_, http_client, err := register_and_login(client, username, "default")
	if err != nil {
		t.Fatalf("register_and_login failed: %v", err)
	}
	_, err = add_message(http_client, "the message by foo")
	if err != nil {
		t.Fatalf("add_message failed: %v", err)
	}
	_, err = logout(http_client)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	client2 := newTestClient()
	username2 := fmt.Sprintf("user_%d", time.Now().UnixNano())
	_, http_client2, err := register_and_login(client2, username2, "default")

	_, err = add_message(http_client2, "the message by bar")
	if err != nil {
		t.Fatalf("add_message failed: %v", err)
	}

	r, err := http_client2.Get(BASE_URL + "/public")
	if err != nil {
		t.Fatalf("get public timeline failed: %v", err)
	}
	body := readBody(t, r)
	assertContains(t, body, "the message by foo")
	assertContains(t, body, "the message by bar")

	// bar's timeline should just show bar's message
	r, err = http_client2.Get(BASE_URL + "/")
	if err != nil {
		t.Fatalf("get timeline failed: %v", err)
	}

	body = readBody(t, r)
	assertContainsNot(t, body, "the message by foo")
	assertContains(t, body, "the message by bar")

	// now let's follow foo
	r, err = http_client2.Get(BASE_URL + "/" + username + "/follow")
	if err != nil {
		t.Fatalf("follow foo failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You are now following")

	// we should now see foo's message
	r, err = http_client2.Get(BASE_URL + "/")
	if err != nil {
		t.Fatalf("get timeline failed: %v", err)
	}

	body = readBody(t, r)
	assertContains(t, body, "the message by foo")
	assertContains(t, body, "the message by bar")

	//but on the user's page we only want the user's message
	r, err = http_client2.Get(BASE_URL + "/" + username2)
	body = readBody(t, r)
	assertContainsNot(t, body, "the message by foo")
	assertContains(t, body, "the message by bar")

	r, err = http_client2.Get(BASE_URL + "/" + username)
	body = readBody(t, r)
	assertContains(t, body, "the message by foo")
	assertContainsNot(t, body, "the message by bar")

	// now unfollow and check if that worked
	r, err = http_client2.Get(BASE_URL + "/" + username + "/unfollow")
	if err != nil {
		t.Fatalf("unfollow foo failed: %v", err)
	}
	assertContains(t, readBody(t, r), "You are no longer following")
	r, err = http_client2.Get(BASE_URL + "/")
	if err != nil {
		t.Fatalf("get timeline failed: %v", err)
	}

	body = readBody(t, r)
	assertContainsNot(t, body, "the message by foo")
	assertContains(t, body, "the message by bar")
}
