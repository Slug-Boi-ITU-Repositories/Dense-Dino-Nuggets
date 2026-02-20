package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const BASE_URL = "http://localhost:5000" // Should probably be changed

func helper_register(username, password string, password2 *string, email *string) (*http.Response, error) {
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

	resp, err := http.PostForm(BASE_URL+"/register", data)
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

func register_and_login(username string, password string) (*http.Response, *http.Client, error) {
	//Registers and logs in in one go
	helper_register(username, password, nil, nil)
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
func assertContains(body, substr string) error {
	if !strings.Contains(body, substr) {
		return fmt.Errorf("expected %q in body\nGot:\n%s", substr, body)
	}
	return nil
}

// Testing functions
func test_register() {
	//Make sure registering works
	r, err := helper_register("user1", "default", nil, nil)
	if err != nil {
		fmt.Errorf("register failed: %v", err)
	}
	defer r.Body.Close()

	b, _ := io.ReadAll(r.Body)
	assertContains(string(b), "You were successfully registered and can login now")

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
