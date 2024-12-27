package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dDogge/Brainwave/database"
)

func AddTopicHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		}

		var reqBody struct {
			Title    string `json:"title"`
			Username string `json:"username"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Title == "" || reqBody.Username == "" {
			http.Error(w, "both title and username are required", http.StatusBadRequest)
			return
		}

		err = database.AddTopic(db, reqBody.Title, reqBody.Username)
		if err != nil {
			if err.Error() == "user not found" {
				http.Error(w, "user not found", http.StatusNotFound)
			} else if err.Error() == "topic title already exists" {
				http.Error(w, "topic title already exists", http.StatusConflict)
			} else {
				http.Error(w, "failed to add topic", http.StatusInternalServerError)
			}
			return
		}

		resp := struct {
			Message string `json:"message"`
		}{
			Message: "topic added successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

func RemoveTopicHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			Title string `json:"title"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Title == "" {
			http.Error(w, "title field is required", http.StatusBadRequest)
			return
		}

		err = database.RemoveTopic(db, reqBody.Title)
		if err != nil {
			if err.Error() == "no topic found with title" {
				http.Error(w, "topic not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to remove topic", http.StatusInternalServerError)
			}
			return
		}

		resp := struct {
			Message string `json:"message"`
		}{
			Message: "topic removed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func UpVoteTopicHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			Title    string `json:"title"`
			Username string `json:"username"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.Title == "" || reqBody.Username == "" {
			http.Error(w, "all fields (title, username) are required", http.StatusBadRequest)
			return
		}

		err = database.UpVoteTopic(db, reqBody.Title, reqBody.Username)
		if err != nil {
			http.Error(w, "failed to upvote topic", http.StatusInternalServerError)
			return
		}

		resp := map[string]string{
			"message": "upvote added successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
