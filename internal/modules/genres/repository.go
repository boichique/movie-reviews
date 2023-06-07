package genres

import (
	"context"

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/dbx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, name string) (*Genre, error) {
	var genre Genre

	err := r.db.
		QueryRow(
			ctx,
			`INSERT INTO genres (name)
			VALUES ($1) 
			RETURNING id, name;`,
			name,
		).
		Scan(
			&genre.ID,
			&genre.Name,
		)

	switch {
	case dbx.IsUniqueViolation(err, "name"):
		return nil, apperrors.AlreadyExists("genre", "name", name)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &genre, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]*Genre, error) {
	rows, err := r.db.
		Query(
			ctx,
			`SELECT id, name
			FROM genres;`,
		)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	var genres []*Genre
	for rows.Next() {
		var genre Genre
		if err = rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, apperrors.Internal(err)
		}

		genres = append(genres, &genre)
	}

	return genres, nil
}

func (r *Repository) GetByID(ctx context.Context, genreID int) (*Genre, error) {
	var genre Genre

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, name
			FROM genres
			WHERE id = $1;`,
			genreID,
		).
		Scan(
			&genre.ID,
			&genre.Name,
		)

	switch {
	case dbx.IsNoRows(err):
		return nil, errGenreWithNotFound(genreID)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &genre, nil
}

func (r *Repository) Update(ctx context.Context, genreID int, name string) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE genres
			SET name = $1
			WHERE id = $2;`,
			name,
			genreID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return errGenreWithNotFound(genreID)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, genreID int) error {
	n, err := r.db.
		Exec(
			ctx,
			`DELETE FROM genres
			WHERE id = $1;`,
			genreID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return errGenreWithNotFound(genreID)
	}

	return nil
}

func errGenreWithNotFound(id int) error {
	return apperrors.NotFound("genre", "id", id)
}
