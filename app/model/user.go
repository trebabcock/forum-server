package model

import (
	"time"
)

type User struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	Bio        string    `json:"bio"`
	Reputation int       `json:"reputation"`
	AvatarURL  string    `json:"avatar_url"`
	Role       string    `json:"role"`
	Active     bool      `json:"active"`
	CreateDate time.Time `json:"create_date"`
}

type PublicUser struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Bio        string    `json:"bio"`
	Reputation int       `json:"reputation"`
	AvatarURL  string    `json:"avatar_url"`
	CreateDate time.Time `json:"create_date"`
}
