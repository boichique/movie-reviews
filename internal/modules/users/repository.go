package users

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

func (r *Repository) CreateUser(ctx context.Context, user *UserWithPassword) error {
	err := r.db.
		QueryRow(
			ctx,
			`INSERT INTO users (username, email, pass_hash, role) 
			VALUES ($1, $2, $3, $4) 
			RETURNING id, created_at;`,
			user.Username,
			user.Email,
			user.PasswordHash,
			user.Role,
		).
		Scan(
			&user.ID,
			&user.CreatedAt,
		)

	switch {
	case dbx.IsUniqueViolation(err, "email"):
		return apperrors.AlreadyExists("user", "email", user.Email)
	case dbx.IsUniqueViolation(err, "username"):
		return apperrors.AlreadyExists("user", "username", user.Username)
	case err != nil:
		return apperrors.Internal(err)
	}

	return nil
}

func (r *Repository) GetExistingUserWithPasswordByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	user := newUserWithPassword()

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, username, email, pass_hash, role, created_at, deleted_at, bio 
			FROM users 
			WHERE email = $1 
			AND deleted_at IS NULL;`,
			email,
		).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.DeletedAt,
			&user.Bio,
		)

	switch {
	case dbx.IsNoRows(err):
		return nil, apperrors.NotFound("user", "email", email)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return user, nil
}

func (r *Repository) GetExistingUserByID(ctx context.Context, userID int) (*User, error) {
	var user User

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, username, email, role, created_at, bio 
			FROM users 
			WHERE id = $1 
			AND deleted_at IS NULL;`,
			userID,
		).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.Bio,
		)

	switch {
	case dbx.IsNoRows(err):
		return nil, apperrors.NotFound("user", "id", userID)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &user, nil
}

func (r *Repository) GetExistingUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, username, email, role, created_at, bio 
			FROM users 
			WHERE username = $1 
			AND deleted_at IS NULL;`,
			username,
		).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.Bio,
		)

	switch {
	case dbx.IsNoRows(err):
		return nil, apperrors.NotFound("user", "username", username)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &user, nil
}

func (r *Repository) UpdateBio(ctx context.Context, userID int, bio string) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE users 
			SET bio = $1
			WHERE id = $2
			AND deleted_at IS NULL;`,
			bio,
			userID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return apperrors.NotFound("user", "id", userID)
	}

	return nil
}

func (r *Repository) UpdateRole(ctx context.Context, userID int, role string) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE users 
			SET role = $1
			WHERE id = $2
			AND deleted_at IS NULL;`,
			role,
			userID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return apperrors.NotFound("user", "id", userID)
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID int) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE users 
			SET deleted_at = NOW()
			WHERE id = $1
			AND deleted_at IS NULL;`,
			userID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return apperrors.NotFound("user", "id", userID)
	}

	return nil
}
