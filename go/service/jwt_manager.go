package service

import (
	"fmt"
	"log"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{secretKey, tokenDuration}
}

// Generate generates and signs a new token for a user
func (manager *JWTManager) Generate(user *User) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(manager.tokenDuration).Unix(),
	})

	tokenString, err := claims.SignedString([]byte(manager.secretKey))
	if err != nil {
		log.Printf("There was an error in signing the token: %v", err)
		return "", err
	}

	return tokenString, nil
}

// Verify verifies the access token string and return a user claim if the token is valid
func (manager *JWTManager) Verify(accessToken string) (*User, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token signing method")
		}
		return []byte(manager.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &User{
		Username: claims["username"].(string),
		Role:     claims["role"].(string),
	}, nil
}
