package stars

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

func (s *Service) Create(ctx context.Context, star *Star) error {
	if err := s.repo.Create(ctx, star); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"star created",
		"starFirstName", star.FirstName,
		"starLastName", star.LastName,
	)

	return nil
}

func (s *Service) GetStars(ctx context.Context) ([]*Star, error) {
	return s.repo.GetStars(ctx)
}

func (s *Service) GetByID(ctx context.Context, starID int) (*Star, error) {
	return s.repo.GetByID(ctx, starID)
}

func (s *Service) Delete(ctx context.Context, starID int) error {
	if err := s.repo.Delete(ctx, starID); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"star deleted",
		"starID", starID,
	)

	return nil
}
