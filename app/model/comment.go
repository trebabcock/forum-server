package model

type Comment struct {
	ID         string `gorm:"UNIQUE;PRIMARY_KEY" json:"id"`
	AuthorID   string `json:"author_id"`
	PostID     string `json:"post_id"`
	ParentID   string `json:"parent_id"`
	Content    string `json:"content"`
	CreateDate string `json:"create_date"`
}

type NewComment struct {
	PostID  string `json:"post_id"`
	Content string `json:"content"`
}
