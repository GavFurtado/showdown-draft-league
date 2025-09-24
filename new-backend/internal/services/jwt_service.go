package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secretKey []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secretKey: []byte(secret),
	}
}

func (s *JWTService) GenerateToken(userID uuid.UUID) (string, error) {
	const sessionTokenPeriod = (time.Hour * 24 * 3 * 30) // 90 days
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"expiry":  time.Now().Add(sessionTokenPeriod).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("(Error: GenerateToken) - Failed to sign token: %w", err)
	}

	return tokenString, err
}

func (s *JWTService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("(Error: jwt.Parse) - unexpected signing method: %s", t.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("Invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, fmt.Errorf("(Error: ValidateToken) - user_id claim is not a string")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("(Error: ValidateToken) - Invalid user_id UUID format: %w", err)
		}
		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("(Error: ValidateToken) - Invalid token claims")
}
