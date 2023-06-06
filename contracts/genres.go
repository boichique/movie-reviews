package contracts

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GetOrDeleteGenreRequest struct {
	GenreID int `param:"genreID" validate:"nonzero"`
}

type CreateGenreRequest struct {
	Name string `json:"name" validate:"min=3,max=32"`
}

type UpdateGenreRequest struct {
	GenreID int    `param:"genreID" validate:"nonzero"`
	Name    string `json:"name" validate:"min=3,max=32"`
}
