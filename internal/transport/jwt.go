package transport

import (
	"notes-go/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWTAccessToken(username string) (string, error) {
	expiredTime := time.Now().Add(time.Duration(config.JWT_ACCESS_TIME) * time.Minute)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(config.JWT_SECRET)
}

func ParseJWSToken(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return config.JWT_SECRET, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
