package models

import "time"

type Message struct {
	ID        int       `json:"id"`
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content"`
	UserID    int       `json:"user_id"`
	TopicID   int       `json:"topic_id"`
	ParentID  *int      `json:"parent_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
