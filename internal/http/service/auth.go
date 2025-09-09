package service

import (
	"event_service/internal/cfg"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(jwtSecretKey string) string {

	// expTime := time.Now().Add(time.Hour * 9999).Unix()

	payload := jwt.MapClaims{
		"username": "event_service",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	t, _ := token.SignedString([]byte(jwtSecretKey))
	return t
}

func CheckHeader(header string) error {
	strArr := strings.Split(header, "Bearer ")
	cfg := cfg.GetConfig()

	_, err := ParseToken(strArr[1], cfg.JwtSecretKey)

	if err != nil {
		return err
	}

	return nil
}

func ParseToken(tokenString string, key string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
}
