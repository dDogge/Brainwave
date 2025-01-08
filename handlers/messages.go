package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

func SetParentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			ParentID int `json:"parent_id"`
			ChildID  int `json:"child_id"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.ParentID == 0 || reqBody.ChildID == 0 {
			http.Error(w, "both parent_id and child_id are required", http.StatusBadRequest)
			return
		}

		err = database.SetParent(db, reqBody.ParentID, reqBody.ChildID)
		if err != nil {
			if err.Error() == fmt.Sprintf("parent message with ID %d not found", reqBody.ParentID) {
				http.Error(w, fmt.Sprintf("parent message with ID %d not found", reqBody.ParentID), http.StatusNotFound)
			} else if err.Error() == fmt.Sprintf("child message with ID %d not found", reqBody.ChildID) {
				http.Error(w, fmt.Sprintf("child message with ID %d not found", reqBody.ChildID), http.StatusNotFound)
			} else if err.Error() == fmt.Sprintf("messages are not in the same topic: parentID=%d, childID=%d", reqBody.ParentID, reqBody.ChildID) {
				http.Error(w, fmt.Sprintf("messages are not in the same topic: parentID=%d, childID=%d", reqBody.ParentID, reqBody.ChildID), http.StatusBadRequest)
			} else {
				http.Error(w, "failed to set parent", http.StatusInternalServerError)
			}
			return
		}

		resp := map[string]string{
			"message": "parent set successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func GetMessagesByTopicHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		topicIDStr := r.URL.Query().Get("topic_id")
		if topicIDStr == "" {
			http.Error(w, "topic_id is required", http.StatusBadRequest)
			return
		}

		topicID, err := strconv.Atoi(topicIDStr)
		if err != nil {
			http.Error(w, "invalid topic_id", http.StatusBadRequest)
			return
		}

		messages, err := database.GetMessagesByTopic(db, topicID)
		if err != nil {
			http.Error(w, "failed to fetch messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func LikeMessageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			MessageID int `json:"message_id"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.MessageID == 0 {
			http.Error(w, "message_id is required", http.StatusBadRequest)
			return
		}

		err = database.LikeMessage(db, reqBody.MessageID)
		if err != nil {
			http.Error(w, "failed to like message", http.StatusInternalServerError)
			return
		}

		resp := map[string]string{
			"message": "like added successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func DislikeMessageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			MessageID int `json:"message_id"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if reqBody.MessageID == 0 {
			http.Error(w, "message_id is required", http.StatusBadRequest)
			return
		}

		err = database.DislikeMessage(db, reqBody.MessageID)
		if err != nil {
			http.Error(w, "failed to dislike message", http.StatusInternalServerError)
			return
		}

		resp := map[string]string{
			"message": "dislike added successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
