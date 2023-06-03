package users

import (
	"context"

	"github.com/boichique/movie-reviews/internal/log"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, user *UserWithPassword) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetExistingUserWithPasswordByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	return s.repo.GetExistingUserWithPasswordByEmail(ctx, email)
}

func (s *Service) GetExistingUserByID(ctx context.Context, userID int) (*User, error) {
	return s.repo.GetExistingUserByID(ctx, userID)
}

func (s *Service) GetExistingUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.GetExistingUserByUsername(ctx, username)
}

func (s *Service) UpdateBio(ctx context.Context, userID int, bio string) error {
	if err := s.repo.UpdateBio(ctx, userID, bio); err != nil {
		return err
	}

	log.FromContext(ctx).Info("user bio updated", "userID", userID)
	return nil
}

func (s *Service) UpdateRole(ctx context.Context, userID int, role string) error {
	if err := s.repo.UpdateRole(ctx, userID, role); err != nil {
		return err
	}

	log.FromContext(ctx).Info("user role updated", "userID", userID, "role", role)
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, userID int) error {
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	log.FromContext(ctx).Info("user deleted", "userID", userID)
	return nil
}
