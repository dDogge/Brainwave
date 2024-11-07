package models

import "time"

type Topic struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	CreatorID    int       `json:"creator_id"`
	Messages     int       `json:"messages"`
	CreationDate time.Time `json:"creation_date"`
}
