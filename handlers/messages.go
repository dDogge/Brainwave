package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dDogge/Brainwave/database"
)

func AddMessageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			Topic    string `json:"topic"`
			Message  string `json:"message"`
			Username string `json:"username"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Topic == "" || reqBody.Message == "" || reqBody.Username == "" {
			http.Error(w, "all fields (topic, message, username) are required", http.StatusBadRequest)
			return
		}

		err = database.AddMessage(db, reqBody.Topic, reqBody.Message, reqBody.Username)
		if err != nil {
			if err.Error() == "user not found" {
				http.Error(w, "user not found", http.StatusNotFound)
			} else if err.Error() == "topic not found" {
				http.Error(w, "topic not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to add message", http.StatusInternalServerError)
			}
			return
		}

		resp := map[string]string{
			"message": "message added successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}
