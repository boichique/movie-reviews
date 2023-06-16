package contracts

import "time"

type Movie struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	ReleaseDate time.Time  `json:"release_date"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type MovieDetails struct {
	Movie
	Description string   `json:"description"`
	Version     int      `json:"version"`
	Genres      []*Genre `json:"genres"`
}

type GetMovieRequest struct {
	ID int `param:"movieId" validate:"nonzero"`
}

type GetMoviesRequest struct {
	PaginatedRequest
}

type CreateMovieRequest struct {
	Title       string    `json:"title" validate:"nonzero"`
	Description string    `json:"description" validate:"nonzero"`
	ReleaseDate time.Time `json:"release_date" validate:"nonzero"`
	GenresID    []int     `json:"genresId"`
}

type UpdateMovieRequest struct {
	ID          int       `json:"id"`
	Version     int       `json:"version" validate:"min=0"`
	Title       string    `json:"title" validate:"nonzero"`
	Description string    `json:"description" validate:"nonzero"`
	ReleaseDate time.Time `json:"release_date" validate:"nonzero"`
	GenresID    []int     `json:"genresId"`
}

type DeleteMovieRequest struct {
	ID int `param:"movieId" validate:"nonzero"`
}
