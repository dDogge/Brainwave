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
)

func TestAddMessageHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to setup user table: %v", err)
	}

	err = database.CreateTopicTable(db)
	if err != nil {
		t.Fatalf("failed to setup topic table: %v", err)
	}

	err = database.CreateMessageTable(db)
	if err != nil {
		t.Fatalf("failed to setup message table: %v", err)
	}

	username := "testuser"
	email := "testuser@example.com"
	password := "password123"
	topicTitle := "Test Topic"

	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	err = database.AddTopic(db, topicTitle, username)
	if err != nil {
		t.Fatalf("failed to add test topic: %v", err)
	}

	err = database.AddMessage(db, topicTitle, "bla bla", username)
	if err != nil {
		t.Fatalf("failed to add test message: %v", err)
	}

	handler := handlers.AddMessageHandler(db)

	makeRequest := func(reqBody struct {
		Topic    string `json:"topic"`
		Message  string `json:"message"`
		Username string `json:"username"`
	}) *httptest.ResponseRecorder {
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/add-message", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	t.Run("Successfully_Add_Message", func(t *testing.T) {
		reqBody := struct {
			Topic    string `json:"topic"`
			Message  string `json:"message"`
			Username string `json:"username"`
		}{
			Topic:    topicTitle,
			Message:  "This is a test message",
			Username: username,
		}

		rr := makeRequest(reqBody)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
		}

		var resp map[string]string
		err := json.NewDecoder(rr.Body).Decode(&resp)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["message"] != "message added successfully" {
			t.Errorf("expected response 'message added successfully', got %s", resp["message"])
		}
	})

	t.Run("User_Not_Found", func(t *testing.T) {
		reqBody := struct {
			Topic    string `json:"topic"`
			Message  string `json:"message"`
			Username string `json:"username"`
		}{
			Topic:    topicTitle,
			Message:  "This is a test message",
			Username: "nonexistentuser",
		}

		rr := makeRequest(reqBody)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}

		body := rr.Body.String()
		if body != "user not found\n" {
			t.Errorf("expected response 'user not found', got %s", body)
		}
	})

	t.Run("Topic_Not_Found", func(t *testing.T) {
		reqBody := struct {
			Topic    string `json:"topic"`
			Message  string `json:"message"`
			Username string `json:"username"`
		}{
			Topic:    "Nonexistent Topic",
			Message:  "This is a test message",
			Username: username,
		}

		rr := makeRequest(reqBody)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}

		body := rr.Body.String()
		if body != "topic not found\n" {
			t.Errorf("expected response 'topic not found', got %s", body)
		}
	})

	t.Run("Missing_Fields", func(t *testing.T) {
		reqBody := struct {
			Topic    string `json:"topic"`
			Message  string `json:"message"`
			Username string `json:"username"`
		}{
			Topic:    "",
			Message:  "",
			Username: "",
		}

		rr := makeRequest(reqBody)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		body := rr.Body.String()
		if body != "all fields (topic, message, username) are required\n" {
			t.Errorf("expected response 'all fields (topic, message, username) are required', got %s", body)
		}
	})

	t.Run("Invalid_JSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/add-message", bytes.NewBuffer([]byte("invalid-json")))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		body := rr.Body.String()
		if body != "invalid JSON format\n" {
			t.Errorf("expected response 'invalid JSON format', got %s", body)
		}
	})
}
