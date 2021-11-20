package model

import "time"

type Board struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	CreateDate  time.Time `json:"create_date"`
}
