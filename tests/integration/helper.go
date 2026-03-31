package integration

import (
	"minitwit/src/model"
	"net/http"
	"net/url"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	testDB.Exec("PRAGMA foreign_keys = ON")

	// Run migrations
	err = testDB.AutoMigrate(&model.User{}, &model.Message{}, &model.Follower{})
	if err != nil {
		t.Fatalf("failed to migrate DB: %v", err)
	}
	return testDB
}

func registerUser(t *testing.T, client *http.Client, serverURL, username, email, password string) {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("password2", password)
	data.Set("email", email)

	resp, err := client.PostForm(serverURL+"/register-user", data)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}
