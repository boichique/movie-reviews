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

func (s *Service) GetStarsPaginated(ctx context.Context, offset int, limit int) ([]*Star, int, error) {
	return s.repo.GetStarsPaginated(ctx, offset, limit)
}

func (s *Service) GetByID(ctx context.Context, starID int) (*Star, error) {
	return s.repo.GetByID(ctx, starID)
}

func (s *Service) Update(ctx context.Context, star *Star) error {
	if err := s.repo.Update(ctx, star); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"star updated",
		"starID", star.ID,
	)

	return nil
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
