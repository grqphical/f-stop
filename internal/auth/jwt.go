package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
)

var JWTExpiryDuration = time.Hour * 24 * 7 // 7 days

func GenerateJWT(userId int) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "f-stop.authentication",
			"sub": userId,
			"iat": time.Now(),
			"exp": time.Now().Add(JWTExpiryDuration),
		},
	)

	return token.SignedString([]byte(os.Getenv("SECRET")))
}
