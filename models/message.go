package models

import "time"

type Message struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	UserID    int       `json:"user_id"`
	TopicID   int       `json:"topic_id"`
	ParentID  *int      `json:"parent_id,omitempty"`
	Likes     int       `json:"likes"`
	Timestamp time.Time `json:"timestamp"`
}
