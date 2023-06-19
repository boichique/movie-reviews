package contracts

import (
	"strconv"
	"time"
)

type Movie struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	ReleaseDate time.Time  `json:"release_date"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type MovieDetails struct {
	Movie
	Description string         `json:"description"`
	Version     int            `json:"version"`
	Genres      []*Genre       `json:"genres"`
	Cast        []*MovieCredit `json:"cast"`
}

type MovieCredit struct {
	Star    Star    `json:"star"`
	Role    string  `json:"role"`
	Details *string `json:"details,omitempty"`
}

type MovieCreditInfo struct {
	StarID  int     `json:"starId"`
	Role    string  `json:"role"`
	Details *string `json:"details"`
}

type GetMovieRequest struct {
	MovieID int `param:"movieID" validate:"nonzero"`
}

type GetMoviesPaginatedRequest struct {
	PaginatedRequest
	StarID     *int    `query:"starID"`
	SearchTerm *string `query:"q"`
}

type CreateMovieRequest struct {
	Title       string             `json:"title" validate:"nonzero"`
	Description string             `json:"description" validate:"nonzero"`
	ReleaseDate time.Time          `json:"release_date" validate:"nonzero"`
	GenresID    []int              `json:"genresId"`
	Cast        []*MovieCreditInfo `json:"cast"`
}

type UpdateMovieRequest struct {
	MovieID     int                `param:"movieID" validate:"nonzero"`
	Version     int                `json:"version" validate:"min=0"`
	Title       string             `json:"title" validate:"nonzero"`
	Description string             `json:"description" validate:"nonzero"`
	ReleaseDate time.Time          `json:"release_date" validate:"nonzero"`
	GenresID    []int              `json:"genresId"`
	Cast        []*MovieCreditInfo `json:"cast"`
}

type DeleteMovieRequest struct {
	MovieID int `param:"movieID" validate:"nonzero"`
}

func (r *GetMoviesPaginatedRequest) ToQueryParams() map[string]string {
	param := r.PaginatedRequest.ToQueryParams()
	if r.StarID != nil {
		param["starID"] = strconv.Itoa(*r.StarID)
	}

	if r.SearchTerm != nil {
		param["q"] = *r.SearchTerm
	}

	return param
}
