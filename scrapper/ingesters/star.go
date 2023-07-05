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

type StarIngester struct {
	c      *client.Client
	token  string
	logger *slog.Logger

	conversionMap map[string]int
}

func NewStarIngester(c *client.Client, token string, logger *slog.Logger) *StarIngester {
	return &StarIngester{
		c:      c,
		token:  token,
		logger: logger.With("ingester", "star"),
	}
}

func (i *StarIngester) Ingest(stars map[string]*models.Star, bios map[string]*models.Bio) error {
	existingStars, err := client.Paginate(&contracts.GetStarsPaginatedRequest{}, i.c.GetStars)
	if err != nil {
		return err
	}

	type starCommonIdentifier struct {
		FirstName string
		LastName  string
		BirthDate time.Time
	}

	getID := func(s *contracts.Star) starCommonIdentifier {
		return starCommonIdentifier{
			FirstName: s.FirstName,
			LastName:  s.LastName,
			BirthDate: s.BirthDate,
		}
	}

	idToStarMap := slices.ToMap(existingStars, getID, slices.NoChangeFunc[*contracts.Star]())
	var mx sync.RWMutex

	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(8)

	for _, star := range stars {
		star := star
		commonID := starCommonIdentifier{star.FirstName, star.LastName, star.BirthDate}

		if maps.ExistsLocked(idToStarMap, commonID, &mx) {
			continue
		}

		group.Go(func() error {
			var created bool
			_, created, err = maps.GetOrCreateLocked(idToStarMap, commonID, &mx, func(name starCommonIdentifier) (*contracts.Star, error) {
				bio, ok := bios[star.ID]
				if !ok {
					i.logger.With("star_id", star.ID).Error("Bio not found")
					bio = &models.Bio{}
				}

				req := &contracts.CreateStarRequest{
					FirstName: star.FirstName,
					LastName:  star.LastName,
					BirthDate: star.BirthDate,
					DeathDate: star.DeathDate,
				}
				if bio.Bio != "" {
					req.Bio = &bio.Bio
				}
				if bio.BirthPlace != "" {
					req.BirthPlace = &bio.BirthPlace
				}

				var sd *contracts.StarDetails
				sd, err = i.c.CreateStar(contracts.NewAuthenticated(req, i.token))
				if err != nil {
					return nil, fmt.Errorf("create star %q: %w", name, err)
				}

				return &sd.Star, nil
			})
			if err != nil {
				return err
			}

			if created {
				i.logger.
					With("star_id", star.ID).
					With("star_common_id", commonID).
					Debug("Created star")
			}

			return nil
		})
	}

	if err = group.Wait(); err != nil {
		return fmt.Errorf("ingest stars: %w", err)
	}

	i.conversionMap = make(map[string]int, len(idToStarMap))
	for _, star := range stars {
		commonID := starCommonIdentifier{star.FirstName, star.LastName, star.BirthDate}
		s, ok := idToStarMap[commonID]
		if !ok {
			i.logger.With("star_id", star.ID).Error("Cannot find star for conversion map creation")
			continue
		}

		i.conversionMap[star.ID] = s.ID
	}

	i.logger.Info("Successfully ingested stars")
	return nil
}

func (i *StarIngester) Converter(imdbID string) (int, bool) {
	id, ok := i.conversionMap[imdbID]
	return id, ok
}
