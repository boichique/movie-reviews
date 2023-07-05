package collectors

import (
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/exp/slog"
)

type TopMoviesCollector struct {
	c *colly.Collector
	l *slog.Logger
}

func NewTopMoviesCollector(c *colly.Collector, movieCollector *MovieCollector, logger *slog.Logger) *TopMoviesCollector {
	collector := &TopMoviesCollector{
		c: c,
		l: logger.With("collector", "top_movies"),
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(i int, e *colly.HTMLElement) {
			link := e.Attr("href")
			text := strings.TrimSpace(e.Text)

			if text != "" && strings.HasPrefix(link, "/title/") {
				movieCollector.Visit(e.Request.AbsoluteURL(link))
			}
		})
	})

	return collector
}

func (c *TopMoviesCollector) Start() {
	startLink := "https://www.imdb.com/chart/top"
	visit(c.c, startLink, c.l)
}

func (c *TopMoviesCollector) Wait() {
	c.c.Wait()
}
