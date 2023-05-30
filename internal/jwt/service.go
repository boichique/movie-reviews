package jwt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Service struct {
	secret           string
	accessExpiration time.Duration
}

func NewService(secret string, accessExpiration time.Duration) *Service {
	return &Service{
		secret:           secret,
		accessExpiration: accessExpiration,
	}
}

func (s *Service) GenerateToken(id int, role string) (string, error) {
	now := time.Now()

	claims := &AccessClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        uuid.New().String(),
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(s.accessExpiration).Unix(),
			Subject:   strconv.Itoa(id),
		},
		UserID: id,
		Role:   role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}
