package users

import "context"

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

func (s *Service) UpdateBio(ctx context.Context, userID int, bio string) error {
	return s.repo.UpdateBio(ctx, userID, bio)
}

func (s *Service) UpdateRole(ctx context.Context, userID int, role string) error {
	return s.repo.UpdateRole(ctx, userID, role)
}

func (s *Service) DeleteUser(ctx context.Context, userID int) error {
	return s.repo.DeleteUser(ctx, userID)
}
