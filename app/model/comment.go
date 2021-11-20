package model

import "time"

type Comment struct {
	ID         string    `json:"id"`
	AuthorID   string    `json:"author_id"`
	BoardID    string    `json:"board_id"`
	ParentID   string    `json:"parent_id"`
	Content    string    `json:"content"`
	CreateDate time.Time `json:"create_date"`
}
