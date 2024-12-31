package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dDogge/Brainwave/database"
	"github.com/dDogge/Brainwave/handlers"
)

func TestAddTopicHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to create user table: %v", err)
	}

	err = database.CreateTopicTable(db)
	if err != nil {
		t.Fatalf("failed to create topic table: %v", err)
	}

	username := "testuser"
	email := "testuser@example.com"
	password := "password123"
	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	handler := handlers.AddTopicHandler(db)

	makeRequest := func(reqBody map[string]string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/add-topic", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	t.Run("Successfully_add_topic", func(t *testing.T) {
		reqBody := map[string]string{
			"title":    "Test Topic",
			"username": username,
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

		if resp["message"] != "topic added successfully" {
			t.Errorf("expected message 'topic added successfully', got %s", resp["message"])
		}
	})

	t.Run("Invalid_JSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/add-topic", bytes.NewBuffer([]byte("invalid-json")))
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

	t.Run("Missing_fields", func(t *testing.T) {
		reqBody := map[string]string{
			"title": "",
		}
		rr := makeRequest(reqBody)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		body := rr.Body.String()
		if body != "both title and username are required\n" {
			t.Errorf("expected response 'both title and username are required', got %s", body)
		}
	})

	t.Run("User_not_found", func(t *testing.T) {
		reqBody := map[string]string{
			"title":    "Test Topic",
			"username": "nonexistentuser",
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

	t.Run("Topic_title_already_exists", func(t *testing.T) {
		reqBody := `{"title":"Test Topic 2", "username":"testuser"}`
		req := httptest.NewRequest(http.MethodPost, "/add-topic", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}

		req = httptest.NewRequest(http.MethodPost, "/add-topic", strings.NewReader(reqBody))
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
		}

		body := w.Body.String()
		expectedBody := "topic title already exists\n"
		if body != expectedBody {
			t.Errorf("expected response body '%s', got '%s'", expectedBody, body)
		}
	})
}

func TestRemoveTopicHandler(t *testing.T) {
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
		t.Fatalf("failed to setup topics table: %v", err)
	}

	username := "testuser"
	email := "testuser@test.com"
	password := "password123"
	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	topicTitle := "Test Topic"
	err = database.AddTopic(db, topicTitle, username)
	if err != nil {
		t.Fatalf("failed to add test topic: %v", err)
	}

	handler := handlers.RemoveTopicHandler(db)

	t.Run("Successfully_remove_topic", func(t *testing.T) {
		reqBody := `{"title":"Test Topic"}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-topic", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var resp struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Message != "topic removed successfully" {
			t.Errorf("expected message 'topic removed successfully', got '%s'", resp.Message)
		}
	})

	t.Run("Topic_not_found", func(t *testing.T) {
		reqBody := `{"title":"Nonexistent Topic"}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-topic", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		body := w.Body.String()
		if body != "topic not found\n" {
			t.Errorf("expected response 'topic not found', got '%s'", body)
		}
	})

	t.Run("Missing_fields", func(t *testing.T) {
		reqBody := `{}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-topic", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		body := w.Body.String()
		if body != "title field is required\n" {
			t.Errorf("expected response 'title field is required', got '%s'", body)
		}
	})

	t.Run("Invalid_JSON", func(t *testing.T) {
		reqBody := `{"title":}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-topic", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		body := w.Body.String()
		if body != "invalid JSON format\n" {
			t.Errorf("expected response 'invalid JSON format', got '%s'", body)
		}
	})

	t.Run("Invalid_Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/remove-topic", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}

		body := w.Body.String()
		if body != "invalid request method\n" {
			t.Errorf("expected response 'invalid request method', got '%s'", body)
		}
	})
}
