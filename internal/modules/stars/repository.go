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

func (r *Repository) GetStarsPaginated(ctx context.Context, movieID *int, offset int, limit int) ([]*Star, int, error) {
	b := &pgx.Batch{}
	selectQuery := dbx.StatementBuilder.
		Select("id, first_name, last_name, birth_date, death_date, created_at, deleted_at").
		From("stars").
		Where("deleted_at IS NULL").
		OrderBy("id").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	countQuery := dbx.StatementBuilder.
		Select("COUNT(*)").
		From("stars").
		Where("deleted_at IS NULL")

	if movieID != nil {
		selectQuery = selectQuery.
			Join("movie_stars ON stars.id = movie_stars.star_id").
			Where("movie_stars.movie_id = ?", movieID)

		countQuery = countQuery.
			Join("movie_stars ON stars.id = movie_stars.star_id").
			Where("movie_stars.movie_id = ?", movieID)
	}

	if err := dbx.QueueBatchSelect(b, selectQuery); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	if err := dbx.QueueBatchSelect(b, countQuery); err != nil {
		return nil, 0, apperrors.Internal(err)
	}
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

func (r *Repository) GetByMovieID(ctx context.Context, movieID int) ([]*MovieCredit, error) {
	rows, err := r.db.
		Query(
			ctx,
			`SELECT s.id, s.first_name, s.last_name, s.birth_date, s.death_date, s.created_at, ms.role, ms.details 
			FROM stars s
			INNER JOIN movie_stars ms ON ms.star_id = s.id
			WHERE ms.movie_id = $1
			ORDER BY ms.order_no`,
			movieID,
		)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	return scanStars(rows)
}

func (r *Repository) GetRelationByMovieID(ctx context.Context, movieID int) ([]*MovieStarRelation, error) {
	rows, err := dbx.FromContext(ctx, r.db).
		Query(ctx,
			`SELECT movie_id, star_id, role, details, order_no 
			FROM movie_stars 
			WHERE movie_id = $1
			ORDER BY order_no`,
			movieID,
		)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	var relations []*MovieStarRelation
	for rows.Next() {
		var relation MovieStarRelation
		if err = rows.
			Scan(
				&relation.MovieID,
				&relation.StarID,
				&relation.Role,
				&relation.Details,
				&relation.OrderNo,
			); err != nil {
			return nil, apperrors.Internal(err)
		}
		relations = append(relations, &relation)
	}
	return relations, nil
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

func scanStars(rows pgx.Rows) ([]*MovieCredit, error) {
	var stars []*MovieCredit
	for rows.Next() {
		var star MovieCredit
		if err := rows.Scan(
			&star.Star.ID, &star.Star.FirstName, &star.Star.LastName, &star.Star.BirthDate,
			&star.Star.DeathDate, &star.Star.CreatedAt, &star.Role, &star.Details); err != nil {
			return nil, apperrors.Internal(err)
		}
		stars = append(stars, &star)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Internal(err)
	}
	return stars, nil
}
