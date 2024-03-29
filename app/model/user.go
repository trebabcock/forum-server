package model

type User struct {
	ID         string `gorm:"UNIQUE;PRIMARY_KEY" json:"id"`
	Username   string `gorm:"UNIQUE" json:"username"`
	Email      string `gorm:"UNIQUE" json:"email"`
	Password   string `json:"password"`
	Bio        string `json:"bio"`
	Reputation int    `json:"reputation"`
	AvatarURL  string `json:"avatar_url"`
	Role       string `json:"role"`
	Active     bool   `json:"active"`
	CreateDate string `json:"create_date"`
}

type PublicUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Bio        string `json:"bio"`
	Reputation int    `json:"reputation"`
	AvatarURL  string `json:"avatar_url"`
	Role       string `json:"role"`
	CreateDate string `json:"create_date"`
}

type RegisterCredentials struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TODO: user either email or username, not just one
type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	PublicUser PublicUser `json:"public_user"`
	Token      string     `json:"token"`
}
