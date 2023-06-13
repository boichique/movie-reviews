package movies

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

func (s *Service) Create(ctx context.Context, movie *MovieDetails) error {
	if err := s.repo.Create(ctx, movie); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"movie created",
		"movieID", movie.ID,
		"movieTitle", movie.Title)
	return nil
}

func (s *Service) GetMoviesPaginated(ctx context.Context, offset int, limit int) ([]*Movie, int, error) {
	return s.repo.GetMoviesPaginated(ctx, offset, limit)
}

func (s *Service) GetByID(ctx context.Context, movieID int) (movie *MovieDetails, err error) {
	return s.repo.GetByID(ctx, movieID)
}

func (s *Service) Update(ctx context.Context, movie *MovieDetails) error {
	if err := s.repo.Update(ctx, movie); err != nil {
		return err
	}

	log.FromContext(ctx).Info("movie updated",
		"movieTitle", movie.Title)
	return nil
}

func (s *Service) Delete(ctx context.Context, movieID int) error {
	if err := s.repo.Delete(ctx, movieID); err != nil {
		return err
	}

	log.FromContext(ctx).Info("movie deleted",
		"movieID", movieID)
	return nil
}
