package collectors

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/boichique/movie-reviews/internal/maps"
	"github.com/boichique/movie-reviews/scrapper/models"
	"github.com/gocolly/colly/v2"
	"golang.org/x/exp/slog"
)

type StarCollector struct {
	c *colly.Collector
	l *slog.Logger

	starMap map[string]*models.Star
	mx      sync.RWMutex
}

func NewStarCollector(c *colly.Collector, bioCollector *BioCollector, logger *slog.Logger) *StarCollector {
	_ = c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	collector := &StarCollector{
		c:       c,
		l:       logger.With("collector", "star"),
		starMap: make(map[string]*models.Star),
	}

	c.OnHTML("html", func(e *colly.HTMLElement) {
		starID := getStarID(e.Request.URL.String())

		star := collector.getOrCreateStar(starID, e.Request.URL.String())

		type starInfo struct {
			Name        string `json:"name"`
			Image       string `json:"image"`
			Description string `json:"description"`
			MainEntity  struct {
				Name        string `json:"name"`
				BirthDate   string `json:"birthDate"`
				DeathDate   string `json:"deathDate"`
				Description string `json:"description"`
			} `json:"mainEntity"`
		}

		var info starInfo
		err := json.Unmarshal([]byte(e.ChildText("script[type='application/ld+json']")), &info)
		if err != nil {
			collector.l.
				With("star_id", starID).
				With("err", err).
				Error("error unmarshalling star info")
			return
		}

		star.Name = info.Name
		star.FirstName, star.LastName = splitName(info.Name)
		star.BirthDate = mustParseDate(info.MainEntity.BirthDate)
		if info.MainEntity.DeathDate != "" {
			deathDate := mustParseDate(info.MainEntity.DeathDate)
			star.DeathDate = &deathDate
		}

		collector.l.
			With("star_id", starID).
			With("star_name", star.Name).
			Debug("star collected")

		bioCollector.Visit(star.Link + "/bio")
	})

	return collector
}

func (c *StarCollector) Visit(link string) {
	visit(c.c, link, c.l)
}

func (c *StarCollector) Wait() {
	c.c.Wait()
}

func (c *StarCollector) Stars() map[string]*models.Star {
	return c.starMap
}

func (c *StarCollector) getOrCreateStar(starID, link string) *models.Star {
	star, _, _ := maps.GetOrCreateLocked(c.starMap, starID, &c.mx, func(key string) (*models.Star, error) {
		return &models.Star{
			ID:   key,
			Link: removeQueryPart(link),
		}, nil
	})

	return star
}

func splitName(name string) (string, string) {
	names := strings.Split(name, " ")
	switch len(names) {
	case 1:
		return names[0], ""
	case 2:
		return names[0], names[1]
	case 3:
		if names[2] == "Jr." || names[2] == "Sr." {
			return names[0], strings.Join(names[1:], " ")
		}

		return names[0], names[2]
	default:
		return names[0], strings.Join(names[2:], " ")
	}
}
