package contracts

import (
	"strconv"
	"time"
)

type Star struct {
	ID        int        `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	BirthDate time.Time  `json:"birth_date"`
	DeathDate *time.Time `json:"death_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type StarDetails struct {
	Star
	MiddleName *string `json:"middle_name,omitempty"`
	BirthPlace *string `json:"birth_place,omitempty"`
	Bio        *string `json:"bio,omitempty"`
}

type CreateStarRequest struct {
	FirstName  string     `json:"first_name" validate:"max=50"`
	MiddleName *string    `json:"middle_name,omitempty" validate:"max=50"`
	LastName   string     `json:"last_name" validate:"max=50"`
	BirthDate  time.Time  `json:"birth_date" validate:"nonzero"`
	BirthPlace *string    `json:"birth_place,omitempty" validate:"max=100"`
	DeathDate  *time.Time `json:"death_date,omitempty"`
	Bio        *string    `json:"bio,omitempty"`
}

type GetStarsPaginatedRequest struct {
	PaginatedRequest
	MovieID *int `query:"movieID"`
}

type GetStarRequest struct {
	StarID int `param:"starID" validate:"nonzero"`
}

type UpdateStarRequest struct {
	StarID     int        `json:"star_id"`
	FirstName  string     `json:"first_name"`
	MiddleName *string    `json:"middle_name,omitempty"`
	LastName   string     `json:"last_name"`
	BirthDate  time.Time  `json:"birth_date"`
	BirthPlace *string    `json:"birth_place,omitempty"`
	DeathDate  *time.Time `json:"death_date,omitempty"`
	Bio        *string    `json:"bio,omitempty"`
}

type DeleteStarRequest struct {
	StarID int `param:"starID" validate:"nonzero"`
}

func (r *GetStarsPaginatedRequest) ToQueryParams() map[string]string {
	param := r.PaginatedRequest.ToQueryParams()
	if r.MovieID != nil {
		param["movieID"] = strconv.Itoa(*r.MovieID)
	}

	return param
}
