package movies

import (
	"context"
	"time"

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/dbx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, movie *MovieDetails) error {
	err := r.db.
		QueryRow(ctx,
			`INSERT INTO movies (title, description, release_date)
			VALUES ($1, $2, $3) 
			RETURNING id, created_at`,
			movie.Title, movie.Description, movie.ReleaseDate).
		Scan(
			&movie.ID,
			&movie.CreatedAt,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	return nil
}

func (r *Repository) GetMoviesPaginated(ctx context.Context, offset int, limit int) ([]*Movie, int, error) {
	b := &pgx.Batch{}
	b.Queue(
		`SELECT id, title,  release_date, created_at 
		FROM movies 
		WHERE deleted_at IS NULL 
		ORDER BY id 
		LIMIT $1 OFFSET $2;`,
		limit, offset,
	)
	b.Queue(
		`SELECT count(*) 
		FROM movies 
		WHERE deleted_at IS NULL`,
	)
	br := r.db.SendBatch(ctx, b)
	defer br.Close()

	rows, err := br.Query()
	if err != nil {
		return nil, 0, apperrors.Internal(err)
	}
	defer rows.Close()

	var movies []*Movie
	for rows.Next() {
		var movie Movie
		if err = rows.
			Scan(&movie.ID,
				&movie.Title,
				&movie.ReleaseDate,
				&movie.CreatedAt,
			); err != nil {
			return nil, 0, apperrors.Internal(err)
		}
		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	var total int
	if err = br.QueryRow().Scan(&total); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	return movies, total, err
}

func (r *Repository) GetByID(ctx context.Context, id int) (*MovieDetails, error) {
	var movie MovieDetails

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, version ,title, description, release_date, created_at 
			FROM movies 
			WHERE id = $1 
			AND deleted_at IS NULL;`,
			id,
		).
		Scan(
			&movie.ID,
			&movie.Version,
			&movie.Title,
			&movie.Description,
			&movie.ReleaseDate,
			&movie.CreatedAt,
		)
	switch {
	case dbx.IsNoRows(err):
		return nil, errMovieWithNotFound(id)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &movie, nil
}

func (r *Repository) Update(ctx context.Context, movie *MovieDetails) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE movies 
			SET version = version + 1, 
			title = $1,
			description = $2, 
			release_date = $3 
			WHERE id = $4 
			AND version = $5;`,
			movie.Title,
			movie.Description,
			movie.ReleaseDate,
			movie.ID,
			movie.Version,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		_, err := r.GetByID(ctx, movie.ID)
		if err != nil {
			return err
		}

		return apperrors.VersionMismatch("movie", "id", movie.ID, movie.Version)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, movieID int) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE movies 
			SET deleted_at = $1 
			WHERE id = $2 
			AND deleted_at IS NULL;`,
			time.Now(), movieID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return errMovieWithNotFound(movieID)
	}

	return nil
}

func errMovieWithNotFound(movieID int) error {
	return apperrors.NotFound("movie", "id", movieID)
}
