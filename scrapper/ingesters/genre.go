package ingesters

import (
	"context"
	"fmt"
	"sync"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/maps"
	"github.com/boichique/movie-reviews/internal/slices"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type GenreIngester struct {
	c      *client.Client
	token  string
	logger *slog.Logger

	conversionMap map[string]int
}

func NewGenreIngester(c *client.Client, token string, logger *slog.Logger) *GenreIngester {
	return &GenreIngester{
		c:      c,
		token:  token,
		logger: logger.With("ingester", "genre"),
	}
}

func (i *GenreIngester) Ingest(genres []string) error {
	existingGenres, err := i.c.GetGenres()
	if err != nil {
		return fmt.Errorf("get genres: %w", err)
	}

	getGenreName := func(g *contracts.Genre) string {
		return g.Name
	}

	nameToGenreMap := slices.ToMap(existingGenres, getGenreName, slices.NoChangeFunc[*contracts.Genre]())
	var mx sync.RWMutex

	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(8)

	for _, genre := range genres {
		genre := genre

		if maps.ExistsLocked(nameToGenreMap, genre, &mx) {
			continue
		}

		group.Go(func() error {
			var created bool
			_, created, err = maps.GetOrCreateLocked(nameToGenreMap, genre, &mx, func(name string) (*contracts.Genre, error) {
				req := &contracts.CreateGenreRequest{Name: name}
				return i.c.CreateGenre(contracts.NewAuthenticated[*contracts.CreateGenreRequest](req, i.token))
			})
			if err != nil {
				return fmt.Errorf("create genre %q: %w", genre, err)
			}

			if created {
				i.logger.With("genre", genre).Debug("Created genre")
			}
			return nil
		})
	}

	if err = group.Wait(); err != nil {
		return fmt.Errorf("ingest genres: %w", err)
	}

	i.conversionMap = make(map[string]int, len(nameToGenreMap))
	for _, genre := range nameToGenreMap {
		i.conversionMap[genre.Name] = genre.ID
	}

	i.logger.Info("Successfully ingested genres")
	return nil
}

func (i *GenreIngester) Converter(genre string) (int, bool) {
	id, ok := i.conversionMap[genre]
	return id, ok
}
