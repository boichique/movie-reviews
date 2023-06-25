package reviews

import (
	"context"
	"fmt"

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

func (r *Repository) Create(ctx context.Context, review *Review) error {
	err := r.db.
		QueryRow(
			ctx,
			`INSERT INTO reviews (movie_id, user_id, title, content, rating) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id, created_at;`,
			review.MovieID,
			review.UserID,
			review.Title,
			review.Content,
			review.Rating,
		).
		Scan(
			&review.ID,
			&review.CreatedAt,
		)

	switch {
	case dbx.IsUniqueViolation(err, ""):
		return apperrors.AlreadyExists("review", "(movie_id,user_id)", fmt.Sprintf("(%d,%d)", review.MovieID, review.UserID))
	case err != nil:
		return apperrors.Internal(err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, reviewID int) (*Review, error) {
	var review Review

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, movie_id, user_id, title, content, rating, created_at
			FROM reviews
			WHERE deleted_at IS NULL 
			AND id = $1;`,
			reviewID,
		).
		Scan(
			&review.ID,
			&review.MovieID,
			&review.UserID,
			&review.Title,
			&review.Content,
			&review.Rating,
			&review.CreatedAt,
		)
	switch {
	case dbx.IsNoRows(err):
		return nil, apperrors.NotFound("review", "id", reviewID)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &review, nil
}

func (r *Repository) GetReviewsPaginated(ctx context.Context, movieID, userID *int, offset int, limit int) ([]*Review, int, error) {
	selectQuery := dbx.StatementBuilder.
		Select("id", "movie_id", "user_id", "title", "content", "rating", "created_at").
		From("reviews").
		Where("deleted_at is null").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	countQuery := dbx.StatementBuilder.
		Select("count(*)").
		From("reviews").
		Where("deleted_at is null")

	if movieID != nil {
		selectQuery = selectQuery.Where("movie_id = ?", *movieID)
		countQuery = countQuery.Where("movie_id = ?", *movieID)
	}

	if userID != nil {
		selectQuery = selectQuery.Where("user_id = ?", *userID)
		countQuery = countQuery.Where("user_id = ?", *userID)
	}

	b := &pgx.Batch{}
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

	var reviews []*Review
	for rows.Next() {
		var review Review
		if err = rows.Scan(
			&review.ID,
			&review.MovieID,
			&review.UserID,
			&review.Title,
			&review.Content,
			&review.Rating,
			&review.CreatedAt,
		); err != nil {
			return nil, 0, apperrors.Internal(err)
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	var total int
	if err = br.QueryRow().Scan(&total); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	return reviews, total, nil
}

func (r *Repository) Update(ctx context.Context, reviewID, userID int, title, content string, rating int) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE reviews
			SET title = $1, content = $2, rating = $3 
			WHERE deleted_at IS NULL
			AND id = $4 
			AND user_id = $5`,
			title,
			content,
			rating,
			reviewID,
			userID)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return r.specifyModificationError(ctx, reviewID, userID)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, reviewID, userID int) error {
	n, err := r.db.Exec(
		ctx,
		`UPDATE reviews
		SET deleted_at = NOW()
		WHERE deleted_at IS NULL
		AND id = $1
		AND user_id = $2;`,
		reviewID, userID)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return r.specifyModificationError(ctx, reviewID, userID)
	}

	return nil
}

func (r *Repository) specifyModificationError(ctx context.Context, reviewID, userID int) error {
	review, err := r.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	if review.UserID != userID {
		return apperrors.Forbidden(fmt.Sprintf("review with id %d is not owned by user with id %d", reviewID, userID))
	}

	return apperrors.Internal(fmt.Errorf("unexpected error creating/updating review with id %d", reviewID))
}
