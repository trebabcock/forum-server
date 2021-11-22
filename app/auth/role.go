package auth

type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
