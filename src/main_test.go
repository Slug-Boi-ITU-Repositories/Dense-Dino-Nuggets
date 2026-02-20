package main

import (
	"net/http"
	"net/url"
)

const BASE_URL = "http://localhost:5000" // Should probably be changed

func test_register(username, password string, password2 *string, email *string) (*http.Response, error) {
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
