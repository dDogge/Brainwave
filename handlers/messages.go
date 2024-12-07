package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dDogge/Brainwave/database"
)

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
		http.Error(w, "All fields (username, email, password) are required", http.StatusBadRequest)
		return
	}

	err = database.AddUser(db, reqBody.Username, reqBody.Email, reqBody.Password)
	if err != nil {
		if err.Error() == "username or email already exists" {
			http.Error(w, "username or email already exists", http.StatusConflict)
		} else {
			http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "user created successfully",
	})
}
