package ingesters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/maps"
	"github.com/boichique/movie-reviews/internal/slices"
	"github.com/boichique/movie-reviews/scrapper/models"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type MovieIngester struct {
	c                *client.Client
	token            string
	genreIDConverter func(string) (int, bool)
	starIDConverter  func(string) (int, bool)
	logger           *slog.Logger
}

func NewMovieIngester(c *client.Client, token string, genreIDConverter func(string) (int, bool), starIDConverter func(string) (int, bool), logger *slog.Logger) *MovieIngester {
	return &MovieIngester{
		c:                c,
		token:            token,
		genreIDConverter: genreIDConverter,
		starIDConverter:  starIDConverter,
		logger:           logger.With("ingester", "movie"),
	}
}

func (i *MovieIngester) Ingest(movies map[string]*models.Movie, casts map[string]*models.Cast) error {
	existingMovies, err := client.Paginate(&contracts.GetMoviesPaginatedRequest{}, i.c.GetMovies)
	if err != nil {
		return err
	}

	type movieCommonIdentifier struct {
		Title       string
		ReleaseDate time.Time
	}

	getID := func(m *contracts.Movie) movieCommonIdentifier {
		return movieCommonIdentifier{
			Title:       m.Title,
			ReleaseDate: m.ReleaseDate,
		}
	}

	idToMovieMap := slices.ToMap(existingMovies, getID, slices.NoChangeFunc[*contracts.Movie]())
	var mx sync.RWMutex

	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(8)

	for _, movie := range movies {
		movie := movie
		commonID := movieCommonIdentifier{movie.Title, movie.ReleaseDate}

		if maps.ExistsLocked(idToMovieMap, commonID, &mx) {
			continue
		}

		group.Go(func() error {
			var created bool
			_, created, err = maps.GetOrCreateLocked(idToMovieMap, commonID, &mx, func(name movieCommonIdentifier) (*contracts.Movie, error) {
				req := &contracts.CreateMovieRequest{
					Title:       movie.Title,
					ReleaseDate: movie.ReleaseDate,
					Description: movie.Description,
				}

				// Prepare genres
				for _, genre := range movie.Genres {
					genreID, ok := i.genreIDConverter(genre)
					if !ok {
						i.logger.With("genre", genre).Error("Cannot convert genre")
						continue
					}

					req.GenresID = append(req.GenresID, genreID)
				}

				// Prepare cast
				cast, ok := casts[movie.ID]
				if !ok {
					i.logger.With("movie_id", movie.ID).Error("Cast not found")
					cast = &models.Cast{}
				}

				for _, credit := range cast.Cast {
					starID, ok := i.starIDConverter(credit.StarID)
					if !ok {
						i.logger.With("star_id", credit.StarID).Warn("Cannot convert star id")
						continue
					}

					creditInfo := &contracts.MovieCreditInfo{
						StarID: starID,
						Role:   credit.Role,
					}
					if credit.Details != "" {
						creditInfo.Details = &credit.Details
					}

					req.Cast = append(req.Cast, creditInfo)
				}

				var md *contracts.MovieDetails
				md, err = i.c.CreateMovie(contracts.NewAuthenticated(req, i.token))
				if err != nil {
					return nil, fmt.Errorf("create movie: %w", err)
				}

				return &md.Movie, nil
			})
			if err != nil {
				return err
			}

			if created {
				i.logger.
					With("movie_id", movie.ID).
					With("movie_common_id", commonID).
					Debug("Created movie")
			}

			return nil
		})
	}

	if err = group.Wait(); err != nil {
		return fmt.Errorf("ingest movies: %w", err)
	}

	i.logger.Info("Successfully ingested movies")
	return nil
}
