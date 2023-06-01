package auth

import (
	"context"
	"errors"

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/jwt"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userService *users.Service
	jwtService  *jwt.Service
}

func NewService(userService *users.Service, jwtService *jwt.Service) *Service {
	return &Service{
		userService: userService,
		jwtService:  jwtService,
	}
}

func (s *Service) Register(ctx context.Context, user *users.User, password string) error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.Internal(err)
	}

	userWithPassword := &users.UserWithPassword{
		User:         user,
		PasswordHash: string(passHash),
	}

	return s.userService.CreateUser(ctx, userWithPassword)
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userService.GetExistingUserWithPasswordByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", apperrors.Unauthorized("invalid password")
		}
		return "", apperrors.Internal(err)
	}

	accessToken, err := s.jwtService.GenerateToken(int(user.ID), user.Role)
	if err != nil {
		return "", apperrors.Internal(err)
	}

	return accessToken, nil
}
