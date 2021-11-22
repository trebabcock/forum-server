package auth

import (
	"forum-server/app/model"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/joho/godotenv/autoload"
)

type Claims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func GenerateToken(user *model.User) (string, error) {
	claims := Claims{
		ID:       user.ID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "kerrmetric.space",
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Unix() + 3600,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Println("ERROR SIGN:", err)
		return "", nil
	}
	return signedToken, nil
}
