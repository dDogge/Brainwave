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
}

func TestGetMessagesByTopicHandler(t *testing.T) {
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
	err = database.AddTopic(db, topicTitle, "testuser")
	if err != nil {
		t.Fatalf("failed to add test topic: %v", err)
	}

	var topicID int
	err = db.QueryRow("SELECT id FROM topics WHERE title = ?", topicTitle).Scan(&topicID)
	if err != nil {
		t.Fatalf("failed to fetch topic ID: %v", err)
	}

	messages := []string{"Message 1", "Message 2", "Message 3"}
	for _, msg := range messages {
		err = database.AddMessage(db, topicTitle, msg, "testuser")
		if err != nil {
			t.Fatalf("failed to add message '%s': %v", msg, err)
		}
	}

	handler := handlers.GetMessagesByTopicHandler(db)

	t.Run("Valid Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/messages?topic_id=%d", topicID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response []map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		expectedMessages := []map[string]interface{}{
			{"id": int64(1), "message": "Message 1", "likes": int64(0), "user_id": int64(1), "parent_id": nil},
			{"id": int64(2), "message": "Message 2", "likes": int64(0), "user_id": int64(1), "parent_id": nil},
			{"id": int64(3), "message": "Message 3", "likes": int64(0), "user_id": int64(1), "parent_id": nil},
		}

		if len(response) != len(expectedMessages) {
			t.Fatalf("expected %d messages, got %d", len(expectedMessages), len(response))
		}

		for i, expected := range expectedMessages {
			actual := response[i]

			for key, expectedValue := range expected {
				actualValue, exists := actual[key]
				if !exists {
					t.Errorf("expected key '%s' to exist in response but it did not", key)
				}

				switch ev := expectedValue.(type) {
				case int64:
					av, ok := actualValue.(float64)
					if !ok || int64(av) != ev {
						t.Errorf("for key '%s', expected %v, got %v", key, ev, actualValue)
					}
				case nil:
					if actualValue != nil {
						t.Errorf("expected key '%s' to be nil, got %v", key, actualValue)
					}
				default:
					if expectedValue != actualValue {
						t.Errorf("for key '%s', expected %v, got %v", key, expectedValue, actualValue)
					}
				}
			}
		}
	})

	t.Run("Missing Topic ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/messages", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		expected := "topic_id is required\n"
		if w.Body.String() != expected {
			t.Errorf("expected response '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("Invalid Topic ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/messages?topic_id=invalid", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		expected := "invalid topic_id\n"
		if w.Body.String() != expected {
			t.Errorf("expected response '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("Non-Existent Topic", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/messages?topic_id=9999", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response []string
		err := json.NewDecoder(w.Body).Decode(&response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(response) != 0 {
			t.Errorf("expected empty response, got %v", response)
		}
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/messages?topic_id=1", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}

		expected := "invalid request method\n"
		if w.Body.String() != expected {
			t.Errorf("expected response '%s', got '%s'", expected, w.Body.String())
		}
	})
}

func TestLikeMessageHandler(t *testing.T) {
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

	messageContent := "Test Message"
	err = database.AddMessage(db, topicTitle, messageContent, username)
	if err != nil {
		t.Fatalf("failed to add test message: %v", err)
	}

	var messageID int
	err = db.QueryRow("SELECT id FROM messages WHERE message = ?", messageContent).Scan(&messageID)
	if err != nil {
		t.Fatalf("failed to fetch message ID: %v", err)
	}

	handler := handlers.LikeMessageHandler(db)

	t.Run("Valid Like", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"message_id": %d}`, messageID)
		req := httptest.NewRequest(http.MethodPost, "/like", strings.NewReader(reqBody))
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

		expectedMessage := "like added successfully"
		if resp["message"] != expectedMessage {
			t.Errorf("expected message '%s', got '%s'", expectedMessage, resp["message"])
		}
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/like", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}

		expectedMessage := "invalid request method\n"
		if w.Body.String() != expectedMessage {
			t.Errorf("expected response '%s', got '%s'", expectedMessage, w.Body.String())
		}
	})
}
