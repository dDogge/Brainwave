package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	ResetCode    *string   `json:"reset_code,omitempty"`
	TopicsOpened int       `json:"topics_opened"`
	MessagesSent int       `json:"messages_sent"`
	CreationDate time.Time `json:"creation_date"`
}
