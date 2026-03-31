package main

import (
	"fmt"
	"minitwit/src/repository"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
)

// Tests register, login, and logout.

// Test that Register handler creates user in database
func TestRegisterHandlerCreatesUser(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	store := sessions.NewCookieStore([]byte("test-secret"))

	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	email := username + "@example.com"

	mux := http.NewServeMux()
	mux.HandleFunc("/register-user", func(w http.ResponseWriter, r *http.Request) {
		register(w, r, userRepo, store)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := server.Client()
	registerUser(t, client, server.URL, username, email, "password123")

	// Verifiy user is in database
	user, err := userRepo.GetUserByUsername(username)
	if err != nil {
		t.Fatalf("Error querying user: %v", err)
	}
	if user == nil {
		t.Fatalf("Expected user does not exist in database")
	}
	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
}
