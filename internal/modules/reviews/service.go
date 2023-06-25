package reviews

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

func (s *Service) Create(ctx context.Context, review *Review) error {
	if err := s.repo.Create(ctx, review); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"review created",
		"reviewId", review.ID,
	)

	return nil
}

func (s *Service) GetByID(ctx context.Context, reviewID int) (*Review, error) {
	return s.repo.GetByID(ctx, reviewID)
}

func (s *Service) GetReviewsPaginated(ctx context.Context, movieID, userID *int, offset int, limit int) ([]*Review, int, error) {
	return s.repo.GetReviewsPaginated(ctx, movieID, userID, offset, limit)
}

func (s *Service) Update(ctx context.Context, reviewID, userID int, title, content string, rating int) error {
	if err := s.repo.Update(ctx, reviewID, userID, title, content, rating); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"review updated",
		"reviewId", reviewID,
	)

	return nil
}

func (s *Service) Delete(ctx context.Context, reviewID, userID int) error {
	if err := s.repo.Delete(ctx, reviewID, userID); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"review deleted",
		"reviewId", reviewID,
	)

	return nil
}
