package models

import "time"

type Message struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"` // Ändrat från Content
	UserID    int       `json:"user_id"`
	TopicID   int       `json:"topic_id"`
	ParentID  *int      `json:"parent_id,omitempty"` // Nullability hanteras korrekt
	Likes     int       `json:"likes"`               // Lägga till Likes-fältet
	Timestamp time.Time `json:"timestamp"`
}
