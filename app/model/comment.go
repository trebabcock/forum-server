package model

import "time"

type Comment struct {
	ID         string    `gorm:"UNIQUE;PRIMARY_KEY" json:"id"`
	AuthorID   string    `json:"author_id"`
	PostID     string    `json:"post_id"`
	ParentID   string    `json:"parent_id"`
	Content    string    `json:"content"`
	CreateDate time.Time `json:"create_date"`
}
