package model

import "time"

type Board struct {
	ID          string    `gorm:"UNIQUE;PRIMARY_KEY" json:"id"`
	Name        string    `gorm:"UNIQUE" json:"name"`
	Description string    `json:"description"`
	CreateDate  time.Time `json:"create_date"`
}

type NewBoard struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
