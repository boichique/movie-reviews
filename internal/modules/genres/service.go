package genres

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

func (s *Service) Create(ctx context.Context, name string) (*Genre, error) {
	genre, err := s.repo.Create(ctx, name)
	if err != nil {
		return nil, err
	}

	log.FromContext(ctx).Info(
		"genre created",
		"genreID", genre.ID,
		"genreName", genre.Name,
	)

	return genre, nil
}

func (s *Service) GetAll(ctx context.Context) ([]*Genre, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, genreID int) (*Genre, error) {
	return s.repo.GetByID(ctx, genreID)
}

func (s *Service) GetByMovieID(ctx context.Context, movieID int) ([]*Genre, error) {
	return s.repo.GetByMovieID(ctx, movieID)
}

func (s *Service) Update(ctx context.Context, genreID int, name string) error {
	if err := s.repo.Update(ctx, genreID, name); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"genre updated",
		"genreID", genreID,
		"genreName", name,
	)

	return nil
}

func (s *Service) Delete(ctx context.Context, genreID int) error {
	if err := s.repo.Delete(ctx, genreID); err != nil {
		return err
	}

	log.FromContext(ctx).Info(
		"genre deleted",
		"genreID", genreID,
	)

	return nil
}
