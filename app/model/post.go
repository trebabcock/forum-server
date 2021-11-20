package model

import "time"

type Post struct {
	ID         string    `json:"id"`
	AuthorID   string    `json:"author_id"`
	BoardID    string    `json:"board_id"`
	Content    string    `json:"content"`
	CreateDate time.Time `json:"create_date"`
}
