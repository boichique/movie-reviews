package collectors

import (
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/boichique/movie-reviews/internal/maps"
	"github.com/boichique/movie-reviews/scrapper/models"
	"github.com/gocolly/colly/v2"
	"golang.org/x/exp/slog"
)

type MovieCollector struct {
	c *colly.Collector
	l *slog.Logger

	moviesMap map[string]*models.Movie
	allGenres map[string]bool
	mx        sync.RWMutex
}

func NewMovieCollector(c *colly.Collector, castCollector *CastCollector, logger *slog.Logger) *MovieCollector {
	collector := &MovieCollector{
		c:         c,
		l:         logger.With("collector", "movie"),
		moviesMap: make(map[string]*models.Movie),
		allGenres: make(map[string]bool),
	}

	c.OnHTML("html", func(e *colly.HTMLElement) {
		movieID := getMovieID(e.Request.URL)

		movie := collector.getOrCreateMovie(movieID, e.Request.URL.String())

		type movieInfo struct {
			URL           string   `json:"url"`
			Name          string   `json:"name"`
			Image         string   `json:"image"`
			Description   string   `json:"description"`
			Genre         []string `json:"genre"`
			DatePublished string   `json:"datePublished"`
		}

		var info movieInfo
		err := json.Unmarshal([]byte(e.ChildText("script[type='application/ld+json']")), &info)
		if err != nil {
			collector.l.
				With("movie_id", movieID).
				With("err", err).
				Error("error unmarshalling movie info")
			return
		}

		movie.Title = info.Name
		movie.Description = info.Description
		movie.Genres = info.Genre
		movie.ReleaseDate = mustParseDate(info.DatePublished)

		collector.toAllGenres(movie.Genres)

		collector.l.
			With("movie_id", movieID).
			With("title", movie.Title).
			Debug("movie collected")

		creditsLink, _ := url.JoinPath(info.URL, "/fullcredits")
		castCollector.Visit(creditsLink)
	})

	return collector
}

func (c *MovieCollector) Visit(link string) {
	visit(c.c, link, c.l)
}

func (c *MovieCollector) Wait() {
	c.c.Wait()
}

func (c *MovieCollector) Movies() map[string]*models.Movie {
	return c.moviesMap
}

func (c *MovieCollector) Genres() []string {
	genres := make([]string, 0, len(c.allGenres))
	for genre := range c.allGenres {
		genres = append(genres, genre)
	}
	return genres
}

func (c *MovieCollector) getOrCreateMovie(movieID string, link string) *models.Movie {
	movie, _, _ := maps.GetOrCreateLocked(c.moviesMap, movieID, &c.mx, func(key string) (*models.Movie, error) {
		return &models.Movie{
			ID:   movieID,
			Link: removeQueryPart(link),
		}, nil
	})

	return movie
}

func (c *MovieCollector) toAllGenres(genres []string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	for _, genre := range genres {
		c.allGenres[genre] = true
	}
}

func getMovieID(url *url.URL) string {
	id := strings.Split(url.Path, "/")[2]
	return id
}

func mustParseDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}
	}

	return t
}
