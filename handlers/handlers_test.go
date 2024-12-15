package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dDogge/Brainwave/database"
	"github.com/dDogge/Brainwave/handlers"
	_ "modernc.org/sqlite"
)

func TestCreateUserHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to setup tables: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUserHandler(w, r, db)
	})

	payload := map[string]string{
		"username": "testuser",
		"email":    "testuser@example.com",
		"password": "password123",
	}
	payloadBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/create-user", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if respBody["message"] != "user created successfully" {
		t.Errorf("expected message 'user created successfully', got '%s'", respBody["message"])
	}

	var username string
	err = db.QueryRow("SELECT username FROM users WHERE username = ?", "testuser").Scan(&username)
	if err != nil {
		t.Fatalf("user not found in database: %v", err)
	}
	if username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", username)
	}
}
