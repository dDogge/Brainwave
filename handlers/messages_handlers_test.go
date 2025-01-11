package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestSetParentHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to create user table: %v", err)
	}

	err = database.CreateTopicTable(db)
	if err != nil {
		t.Fatalf("failed to create topics table: %v", err)
	}

	err = database.CreateMessageTable(db)
	if err != nil {
		t.Fatalf("failed to create messages table: %v", err)
	}

	username := "testuser"
	err = database.AddUser(db, username, "testuser@test.com", "password123")
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	topicTitle := "Test Topic"
	err = database.AddTopic(db, topicTitle, username)
	if err != nil {
		t.Fatalf("failed to add test topic: %v", err)
	}

	parentMessage := "This is a parent message."
	childMessage := "This is a child message."

	err = database.AddMessage(db, topicTitle, parentMessage, username)
	if err != nil {
		t.Fatalf("failed to add parent message: %v", err)
	}
	err = database.AddMessage(db, topicTitle, childMessage, username)
	if err != nil {
		t.Fatalf("failed to add child message: %v", err)
	}

	var parentID, childID int
	err = db.QueryRow("SELECT id FROM messages WHERE message = ?", parentMessage).Scan(&parentID)
	if err != nil {
		t.Fatalf("failed to get parent message ID: %v", err)
	}
	err = db.QueryRow("SELECT id FROM messages WHERE message = ?", childMessage).Scan(&childID)
	if err != nil {
		t.Fatalf("failed to get child message ID: %v", err)
	}

	handler := handlers.SetParentHandler(db)

	t.Run("Successfully_Set_Parent", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"parent_id": %d, "child_id": %d}`, parentID, childID)
		req := httptest.NewRequest(http.MethodPost, "/set-parent", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["message"] != "parent set successfully" {
			t.Errorf("expected message 'parent set successfully', got '%s'", resp["message"])
		}
	})

	t.Run("Parent_Not_Found", func(t *testing.T) {
		reqBody := `{"parent_id": 9999, "child_id": 100}`
		req := httptest.NewRequest(http.MethodPost, "/set-parent", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		body := w.Body.String()
		if body != "parent message with ID 9999 not found\n" {
			t.Errorf("expected response 'parent message with ID 9999 not found', got '%s'", body)
		}
	})

	t.Run("Child_Not_Found", func(t *testing.T) {
		reqBody := `{"parent_id": 100, "child_id": 9999}`
		req := httptest.NewRequest(http.MethodPost, "/set-parent", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		body := w.Body.String()
		if body != "child message with ID 9999 not found\n" {
			t.Errorf("expected response 'child message with ID 9999 not found', got '%s'", body)
		}
	})

	t.Run("Different_Topics", func(t *testing.T) {
		anotherTopic := "Another Topic"
		err := database.AddTopic(db, anotherTopic, username)
		if err != nil {
			t.Fatalf("failed to add another topic: %v", err)
		}

		anotherParentMessage := "Parent in another topic"
		anotherChildMessage := "Child in another topic"
		err = database.AddMessage(db, anotherTopic, anotherParentMessage, username)
		if err != nil {
			t.Fatalf("failed to add another parent message: %v", err)
		}
		err = database.AddMessage(db, anotherTopic, anotherChildMessage, username)
		if err != nil {
			t.Fatalf("failed to add another child message: %v", err)
		}

		var newParentID, newChildID int
		err = db.QueryRow("SELECT id FROM messages WHERE message = ?", anotherParentMessage).Scan(&newParentID)
		if err != nil {
			t.Fatalf("failed to get new parent message ID: %v", err)
		}
		err = db.QueryRow("SELECT id FROM messages WHERE message = ?", anotherChildMessage).Scan(&newChildID)
		if err != nil {
			t.Fatalf("failed to get new child message ID: %v", err)
		}

		reqBody := fmt.Sprintf(`{"parent_id": %d, "child_id": %d}`, newParentID, newChildID)
		req := httptest.NewRequest(http.MethodPost, "/set-parent", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		body := w.Body.String()
		if body != fmt.Sprintf("messages are not in the same topic: parentID=%d, childID=%d\n", newParentID, newChildID) {
			t.Errorf("expected response 'messages are not in the same topic: parentID=%d, childID=%d', got '%s'", newParentID, newChildID, body)
		}
	})
}
