package stars

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

func (r *Repository) Create(ctx context.Context, star *StarDetails) error {
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

func (r *Repository) GetStarsPaginated(ctx context.Context, offset int, limit int) ([]*Star, int, error) {
	b := &pgx.Batch{}
	b.Queue("SELECT id, first_name, last_name, birth_date, death_date, created_at, deleted_at FROM stars WHERE deleted_at IS NULL ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	b.Queue("SELECT COUNT(*) FROM stars WHERE deleted_at IS NULL")
	br := r.db.SendBatch(ctx, b)
	defer br.Close()

	rows, err := br.Query()
	if err != nil {
		return nil, 0, apperrors.Internal(err)
	}
	defer rows.Close()

	var stars []*Star
	for rows.Next() {
		var star Star
		if err = rows.
			Scan(
				&star.ID,
				&star.FirstName,
				&star.LastName,
				&star.BirthDate,
				&star.DeathDate,
				&star.CreatedAt,
				&star.DeletedAt,
			); err != nil {
			return nil, 0, apperrors.Internal(err)
		}

		stars = append(stars, &star)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	var total int
	if err = br.QueryRow().Scan(&total); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	return stars, total, err
}

func (r *Repository) GetByID(ctx context.Context, starID int) (*StarDetails, error) {
	var star StarDetails

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, first_name, middle_name, last_name, birth_date, birth_place, death_date, bio, created_at
			FROM stars
			WHERE id = $1
			AND deleted_at IS NULL;`,
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

func (r *Repository) Update(ctx context.Context, star *StarDetails) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE stars 
			SET first_name = $1, 
			middle_name = $2, 
			last_name = $3, 
			birth_date = $4, 
			birth_place = $5, 
			death_date = $6, 
			bio = $7 
			WHERE id = $8`,
			star.FirstName,
			star.MiddleName,
			star.LastName,
			star.BirthDate,
			star.BirthPlace,
			star.DeathDate,
			star.Bio,
			star.ID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return errStarWithNotFound(star.ID)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, starID int) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE stars
			SET deleted_at = $1
			WHERE id = $2
			AND deleted_at IS NULL`,
			time.Now(), starID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return errStarWithNotFound(starID)
	}

	return nil
}

func errStarWithNotFound(starID int) error {
	return apperrors.NotFound("star", "id", starID)
}
