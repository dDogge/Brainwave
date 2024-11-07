package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	TopicsOpened int       `json:"topics_opened"`
	MessagesSent int       `json:"messages_sent"`
	CreationDate time.Time `json:"creation_date"`
}
