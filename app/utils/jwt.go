package utils

import (
	"errors"
	"os"
	"time"

	models "crud-app/app/model"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(func() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	// default (untuk dev). Pastikan di production diganti.
	return "your-secret-key-here"
}())

type JwtClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	RoleID   string `json:"role_id"`
	jwt.RegisteredClaims
}

func GenerateToken(user models.User) (string, error) {
	claims := JwtClaims{
		UserID:   user.ID,
		Username: user.Username,
		RoleID:   user.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}