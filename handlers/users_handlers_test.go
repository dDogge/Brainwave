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
	_ "modernc.org/sqlite"
)

type CheckPasswordRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CheckPasswordResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

type ChangePasswordRequest struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ChangeEmailRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ChangeEmailResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

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

func TestCheckPasswordHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to setup user table: %v", err)
	}

	username, email, password := "testuser", "test@example.com", "password123"
	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	handler := handlers.CheckPasswordHandler(db)

	makeRequest := func(reqBody CheckPasswordRequest) *httptest.ResponseRecorder {
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/checkpassword", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	t.Run("ValidPassword", func(t *testing.T) {
		reqBody := CheckPasswordRequest{Username: username, Password: password}
		rr := makeRequest(reqBody)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var resp CheckPasswordResponse
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if !resp.Valid {
			t.Errorf("expected valid password, got invalid")
		}
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		reqBody := CheckPasswordRequest{Username: username, Password: "wrongpassword"}
		rr := makeRequest(reqBody)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var resp CheckPasswordResponse
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp.Valid {
			t.Errorf("expected invalid password, got valid")
		}
		if resp.Error != "invalid username or password" {
			t.Errorf("expected error message, got %s", resp.Error)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/checkpassword", bytes.NewBuffer([]byte("invalid-json")))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

func TestChangePasswordHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	username := "testuser"
	currentPassword := "oldpassword"
	newPassword := "newpassword"
	err = database.AddUser(db, username, "testuser@mail.com", currentPassword)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	handler := handlers.ChangePasswordHandler(db)

	tests := []struct {
		name           string
		payload        ChangePasswordRequest
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful password change",
			payload: ChangePasswordRequest{
				Username:        username,
				CurrentPassword: currentPassword,
				NewPassword:     newPassword,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "password changed successfully",
		},
		{
			name: "Incorrect current password",
			payload: ChangePasswordRequest{
				Username:        username,
				CurrentPassword: "wrongpassword",
				NewPassword:     newPassword,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "incorrect current password",
		},
		{
			name: "Empty fields",
			payload: ChangePasswordRequest{
				Username:        "",
				CurrentPassword: "",
				NewPassword:     "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "all fields (username, current_password, new_password) are required",
		},
		{
			name: "Non-existent user",
			payload: ChangePasswordRequest{
				Username:        "nonexistent",
				CurrentPassword: currentPassword,
				NewPassword:     newPassword,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)

			req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp ChangePasswordResponse
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				if err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Message != tt.expectedBody {
					t.Errorf("expected response message %q, got %q", tt.expectedBody, resp.Message)
				}
			} else {
				if rr.Body.String() != tt.expectedBody+"\n" {
					t.Errorf("expected response body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			}
		})
	}
}

func TestChangeEmailHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to create user table: %v", err)
	}

	err = database.AddUser(db, "testuser", "oldemail@test.com", "password123")
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	handler := handlers.ChangeEmailHandler(db)

	t.Run("Successfully email change", func(t *testing.T) {
		reqBody := ChangeEmailRequest{
			Username: "testuser",
			Email:    "newemail@test.com",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/change-email", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status code 200, got %d", rr.Code)
		}

		var resp ChangeEmailResponse
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Message != "email changed successfully" {
			t.Errorf("expected success message, got %s", resp.Message)
		}
	})

	t.Run("Email already in use", func(t *testing.T) {
		err = database.AddUser(db, "otheruser", "newemail@test.com", "password123")
		if err != nil {
			t.Fatalf("failed to add other user: %v", err)
		}

		reqBody := ChangeEmailRequest{
			Username: "testuser",
			Email:    "newemail@test.com",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/change-email", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("expected status code 409, got %d", rr.Code)
		}

		expectedMessage := "email is already in use"
		if rr.Body.String() != expectedMessage+"\n" {
			t.Errorf("expected response body %s, got %s", expectedMessage, rr.Body.String())
		}
	})

	t.Run("Invalid JSON format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/change-email", bytes.NewReader([]byte(`{"username": "testuser", "email": "missing_quote}`)))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code 400, got %d", rr.Code)
		}

		expectedMessage := "invalid JSON format"
		if rr.Body.String() != expectedMessage+"\n" {
			t.Errorf("expected response body %s, got %s", expectedMessage, rr.Body.String())
		}
	})

	t.Run("Missing fields", func(t *testing.T) {
		reqBody := ChangeEmailRequest{
			Username: "",
			Email:    "newemail@test.com",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/change-email", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code 400, got %d", rr.Code)
		}

		expectedMessage := "all fields (username, email) are required"
		if rr.Body.String() != expectedMessage+"\n" {
			t.Errorf("expected response body %s, got %s", expectedMessage, rr.Body.String())
		}
	})

	t.Run("Internal server error", func(t *testing.T) {
		db.Close()

		reqBody := ChangeEmailRequest{
			Username: "testuser",
			Email:    "finalemail@test.com",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/change-email", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status code 500, got %d", rr.Code)
		}

		expectedMessage := "failed to change email"
		if rr.Body.String() != expectedMessage+"\n" {
			t.Errorf("expected response body %s, got %s", expectedMessage, rr.Body.String())
		}
	})
}

func TestChangeUsernameHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}
	defer db.Close()

	// Setup the database
	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to setup user table: %v", err)
	}

	handler := http.HandlerFunc(handlers.ChangeUsernameHandler(db))

	// Seed data
	username := "testuser"
	newUsername := "newtestuser"
	existingUsername := "existinguser"
	email := "testuser@test.com"
	password := "password123"

	// Add test users
	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	err = database.AddUser(db, existingUsername, "existinguser@test.com", "password456")
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	t.Run("Successfully change username", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"username":"%s","new_username":"%s"}`, username, newUsername) // Använd newUsername
		req := httptest.NewRequest(http.MethodPost, "/change-username", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var resp handlers.ChangeUsernameResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			t.Errorf("failed to decode response: %v", err)
		}

		if resp.Message != "username changed successfully" {
			t.Errorf("expected message 'username changed successfully', got '%s'", resp.Message)
		}

		// Kontrollera att användarnamnet faktiskt ändrats i databasen
		var updatedUsername string
		err = db.QueryRow("SELECT username FROM users WHERE username = ?", newUsername).Scan(&updatedUsername)
		if err != nil {
			t.Errorf("failed to verify updated username: %v", err)
		}
		if updatedUsername != newUsername {
			t.Errorf("expected username to be updated to '%s', got '%s'", newUsername, updatedUsername)
		}
	})

	t.Run("New username already in use", func(t *testing.T) {
		reqBody := `{"username":"testuser","new_username":"existinguser"}`
		req := httptest.NewRequest(http.MethodPost, "/change-username", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status code %d, got %d", http.StatusConflict, w.Code)
		}

		expected := "username is already in use"
		if strings.TrimSpace(w.Body.String()) != expected {
			t.Errorf("expected response body '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("Invalid new username format", func(t *testing.T) {
		reqBody := `{"username":"testuser","new_username":"!"}`
		req := httptest.NewRequest(http.MethodPost, "/change-username", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		expected := "username does not meet requirements"
		if strings.TrimSpace(w.Body.String()) != expected {
			t.Errorf("expected response body '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("Missing fields", func(t *testing.T) {
		reqBody := `{"username":"","new_username":""}`
		req := httptest.NewRequest(http.MethodPost, "/change-username", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		expected := "all fields (username, new_username) are required"
		if strings.TrimSpace(w.Body.String()) != expected {
			t.Errorf("expected response body '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("Internal server error", func(t *testing.T) {
		db.Close() // Simulate a database failure

		reqBody := `{"username":"testuser","new_username":"newtestuser"}`
		req := httptest.NewRequest(http.MethodPost, "/change-username", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		expected := "failed to change username"
		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("expected response body to contain '%s', got '%s'", expected, w.Body.String())
		}
	})
}
