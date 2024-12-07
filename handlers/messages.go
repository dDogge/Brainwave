package handlers

import (
	"database/sql"
	"encoding/json"
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
