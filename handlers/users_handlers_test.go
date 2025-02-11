package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	ResetCode   string `json:"reset_code"`
	NewPassword string `json:"new_password"`
}

type ResetPasswordResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
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

	t.Run("Invalid_JSON", func(t *testing.T) {
		reqBody := `{"username":}` // ogiltig JSON
		req := httptest.NewRequest(http.MethodPost, "/create-user", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "invalid input format\n" {
			t.Errorf("expected response 'invalid input format', got '%s'", string(body))
		}
	})

	t.Run("Username_Already_Exists", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser", // samma som tidigare
			"email":    "duplicate@test.com",
			"password": "password123",
		}
		payloadBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/create-user", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "username or email already exists\n" {
			t.Errorf("expected response 'username or email already exists', got '%s'", string(body))
		}
	})
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

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}

		var resp CheckPasswordResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Valid {
			t.Errorf("expected invalid password, got valid")
		}
		if resp.Error != "invalid username or password" {
			t.Errorf("expected error message 'invalid username or password', got '%s'", resp.Error)
		}
	})

	t.Run("MissingFields", func(t *testing.T) {
		reqBody := CheckPasswordRequest{Username: ""}
		rr := makeRequest(reqBody)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		body := rr.Body.String()
		if body != "both username and password are required\n" {
			t.Errorf("expected response 'both username and password are required', got %s", body)
		}
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		reqBody := CheckPasswordRequest{Username: "nonexistent", Password: password}
		rr := makeRequest(reqBody)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
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

		body := rr.Body.String()
		if body != "invalid JSON format\n" {
			t.Errorf("expected response 'invalid JSON format', got %s", body)
		}
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/checkpassword", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}

		body := rr.Body.String()
		if body != "invalid request method\n" {
			t.Errorf("expected response 'invalid request method', got %s", body)
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
		t.Fatalf("failed to create user table: %v", err)
	}

	username := "testuser"
	currentPassword := "oldpassword"
	newPassword := "newpassword"
	err = database.AddUser(db, username, "testuser@mail.com", currentPassword)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	handler := handlers.ChangePasswordHandler(db)

	makeRequest := func(payload ChangePasswordRequest) *httptest.ResponseRecorder {
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

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
		{
			name:           "Invalid JSON",
			payload:        ChangePasswordRequest{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid JSON format",
		},
		{
			name:           "Invalid request method",
			payload:        ChangePasswordRequest{},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "invalid request method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rr *httptest.ResponseRecorder

			if tt.name == "Invalid JSON" {
				req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader([]byte("invalid-json")))
				req.Header.Set("Content-Type", "application/json")
				rr = httptest.NewRecorder()
				handler.ServeHTTP(rr, req)
			} else if tt.name == "Invalid request method" {
				req := httptest.NewRequest(http.MethodGet, "/change-password", nil)
				rr = httptest.NewRecorder()
				handler.ServeHTTP(rr, req)
			} else {
				rr = makeRequest(tt.payload)
			}

			if rr.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp ChangePasswordResponse
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				if err != nil {
					t.Fatalf("[%s] failed to unmarshal response: %v", tt.name, err)
				}
				if resp.Message != tt.expectedBody {
					t.Errorf("[%s] expected message %q, got %q", tt.name, tt.expectedBody, resp.Message)
				}
			} else {
				if rr.Body.String() != tt.expectedBody+"\n" {
					t.Errorf("[%s] expected response body %q, got %q", tt.name, tt.expectedBody, rr.Body.String())
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

func TestRemoveUserHandler(t *testing.T) {
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
		t.Fatalf("failed to setup topics table: %v", err)
	}

	err = database.CreateMessageTable(db)
	if err != nil {
		t.Fatalf("failed to setup messages table: %v", err)
	}

	username := "testuser"
	email := "testuser@test.com"
	password := "password123"

	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	handler := handlers.RemoveUserHandler(db)

	t.Run("Successfully_remove_user", func(t *testing.T) {
		reqBody := `{"username":"testuser"}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-user", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]string
		err := json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if responseBody["message"] != "user removed successfully" {
			t.Errorf("expected message 'user removed successfully', got %s", responseBody["message"])
		}
	})

	t.Run("User_not_found", func(t *testing.T) {
		reqBody := `{"username":"nonexistent"}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-user", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "user not found\n" {
			t.Errorf("expected response 'user not found', got %s", string(body))
		}
	})

	t.Run("Invalid_JSON", func(t *testing.T) {
		reqBody := `{"username":}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-user", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "invalid JSON format\n" {
			t.Errorf("expected response 'invalid JSON format', got %s", string(body))
		}
	})

	t.Run("Missing_fields", func(t *testing.T) {
		reqBody := `{}`
		req := httptest.NewRequest(http.MethodDelete, "/remove-user", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "all fields (username) are required\n" {
			t.Errorf("expected response 'all fields (username) are required', got %s", string(body))
		}
	})
}

func TestGeneratePasswordResetCodeHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to setup user table: %v", err)
	}

	email := "testuser@test.com"
	username := "testuser"
	password := "password123"

	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	handler := handlers.GeneratePasswordResetCodeHandler(db)

	t.Run("Successfully_generate_reset_code", func(t *testing.T) {
		reqBody := `{"email":"testuser@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/generate-reset-code", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]string
		err := json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if responseBody["message"] != "password reset code generated successfully" {
			t.Errorf("expected message 'password reset code generated successfully', got %s", responseBody["message"])
		}

		if responseBody["reset_code"] == "" {
			t.Errorf("expected a reset code, got empty string")
		}
	})

	t.Run("Email_not_found", func(t *testing.T) {
		reqBody := `{"email":"nonexistent@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/generate-reset-code", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "email not found\n" {
			t.Errorf("expected response 'email not found', got %s", string(body))
		}
	})

	t.Run("Invalid_JSON", func(t *testing.T) {
		reqBody := `{"email":}`
		req := httptest.NewRequest(http.MethodPost, "/generate-reset-code", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "invalid JSON format\n" {
			t.Errorf("expected response 'invalid JSON format', got %s", string(body))
		}
	})

	t.Run("Missing_email_field", func(t *testing.T) {
		reqBody := `{}`
		req := httptest.NewRequest(http.MethodPost, "/generate-reset-code", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "email field is required\n" {
			t.Errorf("expected response 'email field is required', got %s", string(body))
		}
	})
}

func TestResetPasswordHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to create user table: %v", err)
	}

	username := "testuser"
	email := "testuser@mail.com"
	password := "oldpassword"
	resetCode := "123456"
	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	_, err = db.Exec("UPDATE users SET reset_code = ? WHERE email = ?", resetCode, email)
	if err != nil {
		t.Fatalf("failed to update reset code for test user: %v", err)
	}

	handler := handlers.ResetPasswordHandler(db)

	t.Run("InvalidJSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/reset-password", bytes.NewBuffer([]byte("invalid-json")))
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

	t.Run("InvalidMethod", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/reset-password", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}

		body := rr.Body.String()
		if body != "invalid request method\n" {
			t.Errorf("expected response 'invalid request method', got %s", body)
		}
	})

	t.Run("SuccessfulPasswordReset", func(t *testing.T) {
		reqBody := ResetPasswordRequest{
			Email:       email,
			ResetCode:   resetCode,
			NewPassword: "newpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var resp ResetPasswordResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp.Message != "password reset successfully" {
			t.Errorf("expected message 'password reset successfully', got '%s'", resp.Message)
		}
	})

	t.Run("InvalidResetCode", func(t *testing.T) {
		reqBody := ResetPasswordRequest{
			Email:       email,
			ResetCode:   "wrongcode",
			NewPassword: "newpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		bodyStr := rr.Body.String()
		if bodyStr != "invalid reset code\n" {
			t.Errorf("expected response 'invalid reset code', got %s", bodyStr)
		}
	})

	t.Run("MissingFields", func(t *testing.T) {
		reqBody := ResetPasswordRequest{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		bodyStr := rr.Body.String()
		if bodyStr != "all fields (email, reset_code, new_password) are required\n" {
			t.Errorf("expected response 'all fields (email, reset_code, new_password) are required', got %s", bodyStr)
		}
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		reqBody := ResetPasswordRequest{
			Email:       "nonexistent@mail.com",
			ResetCode:   resetCode,
			NewPassword: "newpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		bodyStr := rr.Body.String()
		if bodyStr != "invalid reset code\n" {
			t.Errorf("expected response 'invalid reset code', got %s", bodyStr)
		}
	})
}

func TestGetAllUsersHandler(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	err = database.CreateUserTable(db)
	if err != nil {
		t.Fatalf("failed to create user table: %v", err)
	}

	username := "testuser"
	email := "testuser@mail.com"
	password := "password123"
	err = database.AddUser(db, username, email, password)
	if err != nil {
		t.Fatalf("failed to add test user: %v", err)
	}

	handler := handlers.GetAllUsersHandler(db)

	t.Run("SuccessfullyFetchAllUsers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/get-users", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var users []map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("expected 1 user, got %d", len(users))
		}

		if users[0]["username"] != username {
			t.Errorf("expected username '%s', got '%s'", username, users[0]["username"])
		}
	})

	t.Run("InvalidRequestMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/get-users", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}

		body := rr.Body.String()
		if body != "invalid request method\n" {
			t.Errorf("expected response 'invalid request method', got %s", body)
		}
	})

	t.Run("FailedToFetchUsers", func(t *testing.T) {
		db.Close()

		req := httptest.NewRequest(http.MethodGet, "/get-users", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		body := rr.Body.String()
		if body != "failed to fetch users\n" {
			t.Errorf("expected response 'failed to fetch users', got %s", body)
		}

		db, err = sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatalf("failed to open in-memory database after closing: %v", err)
		}
	})
}
