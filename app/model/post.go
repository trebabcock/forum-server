package model

import "time"

type Post struct {
	ID         string    `gorm:"UNIQUE;PRIMARY_KEY" json:"id"`
	AuthorID   string    `json:"author_id"`
	BoardID    string    `json:"board_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	CreateDate time.Time `json:"create_date"`
}
