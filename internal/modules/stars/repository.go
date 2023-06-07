package stars

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

func (r *Repository) Create(ctx context.Context, star *Star) error {
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO stars (first_name, middle_name, last_name, birth_date, birth_place, death_date, bio)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;`,
		star.FirstName,
		star.MiddleName,
		star.LastName,
		star.BirthDate,
		star.BirthPlace,
		star.DeathDate,
		star.Bio,
	).
		Scan(&star.ID, &star.CreatedAt)
	if err != nil {
		return apperrors.Internal(err)
	}

	return nil
}

func (r *Repository) GetStars(ctx context.Context) ([]*Star, error) {
	rows, err := r.db.
		Query(
			ctx,
			`SELECT id, first_name, middle_name, last_name, birth_date, birth_place, death_date, bio, created_at
			FROM stars;`,
		)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	var stars []*Star
	for rows.Next() {
		var star Star
		if err := rows.
			Scan(
				&star.ID,
				&star.FirstName,
				&star.MiddleName,
				&star.LastName,
				&star.BirthDate,
				&star.BirthPlace,
				&star.DeathDate,
				&star.Bio,
				&star.CreatedAt,
			); err != nil {
			return nil, apperrors.Internal(err)
		}

		stars = append(stars, &star)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Internal(err)
	}

	return stars, nil
}

func (r *Repository) GetByID(ctx context.Context, starID int) (*Star, error) {
	var star Star

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, first_name, middle_name, last_name, birth_date, birth_place, death_date, bio, created_at
			FROM stars
			WHERE id = $1;`,
			starID,
		).
		Scan(
			&star.ID,
			&star.FirstName,
			&star.MiddleName,
			&star.LastName,
			&star.BirthDate,
			&star.BirthPlace,
			&star.DeathDate,
			&star.Bio,
			&star.CreatedAt,
		)
	switch {
	case dbx.IsNoRows(err):
		return nil, errStarWithNotFound(starID)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &star, nil
}

func (r *Repository) Delete(ctx context.Context, starID int) error {
	n, err := r.db.
		Exec(
			ctx,
			`DELETE FROM stars
			WHERE id = $1`,
			starID,
		)
	if err != nil {
		return err
	}

	if n.RowsAffected() == 0 {
		return errStarWithNotFound(starID)
	}

	return nil
}

func errStarWithNotFound(starID int) error {
	return apperrors.NotFound("star", "id", starID)
}
