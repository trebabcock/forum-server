package model

type Post struct {
	ID         string `gorm:"UNIQUE;PRIMARY_KEY" json:"id"`
	AuthorID   string `json:"author_id"`
	BoardID    string `json:"board_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreateDate string `json:"create_date"`
}

type NewPost struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
