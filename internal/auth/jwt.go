package auth

import (
	"fmt"
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
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(JWTExpiryDuration).Unix(),
		},
	)

	return token.SignedString([]byte(os.Getenv("SECRET")))
}

// Checks if a JWT is valid. If the token is valid, it's claims are returned
func ValidateJWT(tokenString string) map[string]any {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v\n", t.Header)
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) >= claims["exp"].(float64) {
			return nil
		}

		return claims
	} else {
		return nil
	}
}
