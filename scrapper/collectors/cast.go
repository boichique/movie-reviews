package collectors

import (
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/boichique/movie-reviews/internal/maps"
	"github.com/boichique/movie-reviews/scrapper/models"
	"github.com/gocolly/colly/v2"
	"golang.org/x/exp/slog"
)

type CastCollector struct {
	c *colly.Collector
	l *slog.Logger

	castMap   map[string]*models.Cast
	starLinks map[string]bool
	mx        sync.RWMutex
}

func NewCastCollector(c *colly.Collector, starCollector *StarCollector, logger *slog.Logger) *CastCollector {
	collector := &CastCollector{
		c:         c,
		l:         logger.With("collector", "cast"),
		castMap:   make(map[string]*models.Cast),
		starLinks: make(map[string]bool),
	}

	// note about movie roles: 'actor', 'voice actor', 'writer', 'producer', 'director', 'composer'
	c.OnHTML("html", func(e *colly.HTMLElement) {
		movieID := getMovieID(e.Request.URL)

		cast := collector.getOrCreateCast(movieID, e.Request.URL.String())

		directorHeader := e.DOM.Find("h4#director")
		if directorHeader.Nodes != nil {
			collector.addCastFromSimpleTable(cast, "director", directorHeader.Next())
		}

		castHeader := e.DOM.Find("h4#cast")
		if castHeader.Nodes != nil {
			collector.addCastFromCastTable(cast, castHeader.Next(), 15)
		}

		writerHeader := e.DOM.Find("h4#writer")
		if writerHeader.Nodes != nil {
			collector.addCastFromSimpleTable(cast, "writer", writerHeader.Next())
		}

		producerHeader := e.DOM.Find("h4#producer")
		if producerHeader.Nodes != nil {
			collector.addCastFromSimpleTable(cast, "producer", producerHeader.Next())
		}

		composerHeader := e.DOM.Find("h4#composer")
		if composerHeader.Nodes != nil {
			collector.addCastFromSimpleTable(cast, "composer", composerHeader.Next())
		}

		collector.l.
			With("movie_id", movieID).
			Debug("cast collected")

		for _, credit := range cast.Cast {
			if collector.isNewStarLink(credit.StarLink) {
				starCollector.Visit(credit.StarLink)
			}
		}
	})

	return collector
}

func (c *CastCollector) Visit(link string) {
	visit(c.c, link, c.l)
}

func (c *CastCollector) Wait() {
	c.c.Wait()
}

func (c *CastCollector) Cast() map[string]*models.Cast {
	return c.castMap
}

func (c *CastCollector) getOrCreateCast(movieID string, link string) *models.Cast {
	cast, _, _ := maps.GetOrCreateLocked(c.castMap, movieID, &c.mx, func(key string) (*models.Cast, error) {
		return &models.Cast{
			MovieID: key,
			Link:    link,
		}, nil
	})

	return cast
}

func (c *CastCollector) isNewStarLink(link string) bool {
	c.mx.Lock()
	defer c.mx.Unlock()

	if _, ok := c.starLinks[link]; ok {
		return false
	}

	c.starLinks[link] = true
	return true
}

func (c *CastCollector) addCastFromSimpleTable(cast *models.Cast, role string, table *goquery.Selection) {
	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		starLink := row.Find("td.name a")
		if starLink.Nodes == nil {
			// it's just an empty row to separate the cast
			return
		}

		href := starLink.AttrOr("href", "")
		link, _ := url.JoinPath("https://www.imdb.com", href)
		link = removeQueryPart(link)
		details := row.Find("td.credit").Text()

		credit := &models.Credit{
			Role:     role,
			Details:  strings.TrimSpace(details),
			StarName: strings.TrimSpace(starLink.Text()),
			StarLink: link,
			StarID:   getStarID(link),
		}

		cast.Cast = append(cast.Cast, credit)
	})
}

func (c *CastCollector) addCastFromCastTable(cast *models.Cast, table *goquery.Selection, max int) {
	var added int
	table.Find("tr").EachWithBreak(func(i int, row *goquery.Selection) bool {
		// manage special rows
		castListLabel := row.Find("td.castlist_label")
		if castListLabel.Nodes != nil {
			// break if we reached the "rest of cast" section, otherwise - skip
			return !strings.Contains(castListLabel.Text(), "Rest of cast")
		}

		starLink := row.Find("td.primary_photo a")
		if starLink.Nodes == nil {
			c.l.With("movie_id", cast.MovieID).Warn("no star link found for actor")
			return true
		}

		href := starLink.AttrOr("href", "")
		link, _ := url.JoinPath("https://www.imdb.com", href)
		link = removeQueryPart(link)
		details := wordsSanitizer(row.Find("td.character").Text())
		name := row.Find("td:nth-child(2) > a").Text()

		role := "actor"
		if strings.Contains(details, "(voice)") {
			role = "voice actor"
		}

		credit := &models.Credit{
			Role:     role,
			Details:  details,
			StarID:   getStarID(link),
			StarName: strings.TrimSpace(name),
			StarLink: link,
		}

		cast.Cast = append(cast.Cast, credit)
		added++

		// break if we reached the max
		if max > 0 && added >= max {
			return false
		}
		return true
	})
}

func getStarID(link string) string {
	parts := strings.Split(link, "/")
	return parts[4]
}

func wordsSanitizer(s string) string {
	words := strings.Fields(s)
	return strings.Join(words, " ")
}
