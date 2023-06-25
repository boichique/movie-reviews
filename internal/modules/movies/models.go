package movies

import (
	"time"

	"github.com/boichique/movie-reviews/internal/modules/genres"
	"github.com/boichique/movie-reviews/internal/modules/stars"
)

type Movie struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	ReleaseDate time.Time  `json:"release_date"`
	AvgRating   *float64   `json:"avg_rating,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type MovieDetails struct {
	Movie
	Description string               `json:"description"`
	Version     int                  `json:"version"`
	Genres      []*genres.Genre      `json:"genres"`
	Cast        []*stars.MovieCredit `json:"cast"`
}
