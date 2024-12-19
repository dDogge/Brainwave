package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dDogge/Brainwave/database"
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

type ChangeUsernameRequest struct {
	Username    string `json:"username"`
	NewUsername string `json:"new_username"`
}

type ChangeUsernameResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type GenerateResetCodeRequest struct {
	Email string `json:"email"`
}

type GenerateResetCodeResponse struct {
	Message    string `json:"message"`
	ResetCode  string `json:"reset_code,omitempty"`
	StatusCode int    `json:"-"`
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

func CreateUserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var reqBody struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "invalid input format", http.StatusBadRequest)
		return
	}

	if reqBody.Username == "" || reqBody.Email == "" || reqBody.Password == "" {
		http.Error(w, "all fields (username, email, password) are required", http.StatusBadRequest)
		return
	}

	err = database.AddUser(db, reqBody.Username, reqBody.Email, reqBody.Password)
	if err != nil {
		if err.Error() == "username or email already exists" {
			http.Error(w, "username or email already exists", http.StatusConflict)
		} else {
			http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "user created successfully",
	})
}

func CheckPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody CheckPasswordRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Username == "" || reqBody.Password == "" {
			http.Error(w, "both username and password are required", http.StatusBadRequest)
			return
		}

		valid, err := database.CheckPassword(db, reqBody.Username, reqBody.Password)
		if err != nil {
			http.Error(w, "error checking password", http.StatusInternalServerError)
			return
		}

		resp := CheckPasswordResponse{
			Valid: valid,
		}
		if !valid {
			resp.Error = "invalid username or password"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func ChangePasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody ChangePasswordRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Username == "" || reqBody.CurrentPassword == "" || reqBody.NewPassword == "" {
			http.Error(w, "all fields (username, current_password, new_password) are required", http.StatusBadRequest)
			return
		}

		err = database.ChangePassword(db, reqBody.Username, reqBody.CurrentPassword, reqBody.NewPassword)
		if err != nil {
			var statusCode int
			if err.Error() == "user not found" || err.Error() == "incorrect current password" {
				statusCode = http.StatusUnauthorized
			} else {
				statusCode = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), statusCode)
			return
		}

		resp := ChangePasswordResponse{
			Message: "password changed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func ChangeEmailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody ChangeEmailRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Username == "" || reqBody.Email == "" {
			http.Error(w, "all fields (username, email) are required", http.StatusBadRequest)
			return
		}

		err = database.ChangeEmail(db, reqBody.Username, reqBody.Email)
		if err != nil {
			if err.Error() == "email is already in use" {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, "failed to change email", http.StatusInternalServerError)
			return
		}

		resp := ChangeEmailResponse{
			Message: "email changed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func ChangeUsernameHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody ChangeUsernameRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Username == "" || reqBody.NewUsername == "" {
			http.Error(w, "all fields (username, new_username) are required", http.StatusBadRequest)
			return
		}

		err = database.ChangeUsername(db, reqBody.Username, reqBody.NewUsername)
		if err != nil {
			if err.Error() == "username is already in use" {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			} else if err.Error() == "username does not meet requirements" {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to change username", http.StatusInternalServerError)
			return
		}

		resp := ChangeUsernameResponse{
			Message: "username changed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func RemoveUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			Username string `json:"username"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Username == "" {
			http.Error(w, "all fields (username) are required", http.StatusBadRequest)
			return
		}

		err = database.RemoveUser(db, reqBody.Username)
		if err != nil {
			if err.Error() == "user not found" {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, "failed to remove user", http.StatusInternalServerError)
			return
		}

		resp := struct {
			Message string `json:"message"`
		}{
			Message: "user removed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func GeneratePasswordResetCodeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody GenerateResetCodeRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Email == "" {
			http.Error(w, "email field is required", http.StatusBadRequest)
			return
		}

		resetCode, err := database.GeneratePasswordResetCode(db, reqBody.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || err.Error() == "email not found" {
				http.Error(w, "email not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to generate password reset code", http.StatusInternalServerError)
			return
		}

		resp := GenerateResetCodeResponse{
			Message:   "password reset code generated successfully",
			ResetCode: resetCode,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func ResetPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			if r.Method != http.MethodPost {
				http.Error(w, "invalid reguest method", http.StatusMethodNotAllowed)
				return
			}
		}

		var reqBody ResetPasswordRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Email == "" || reqBody.ResetCode == "" || reqBody.NewPassword == "" {
			http.Error(w, "all fields (email, reset_code, new_password) are required", http.StatusBadRequest)
			return
		}

		err = database.ResetPassword(db, reqBody.Email, reqBody.ResetCode, reqBody.NewPassword)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || err.Error() == "invalid reset code" {
				http.Error(w, "invalid reset code", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to reset password", http.StatusInternalServerError)
			return
		}

		resp := ResetPasswordResponse{
			Message: "password reset successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func GetAllUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		users, err := database.GetAllUsers(db)
		if err != nil {
			http.Error(w, "failed to fetch users", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}
}
