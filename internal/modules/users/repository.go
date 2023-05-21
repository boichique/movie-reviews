package users

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *UserWithPassword) error {
	err := r.db.
		QueryRow(
			ctx,
			"INSERT INTO users (username, email, pass_hash) VALUES ($1, $2, $3) returning id, created_at",
			user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)

	return err
}

func (r *Repository) GetExistingUserWithPasswordByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	user := newUserWithPassword()

	err := r.db.
		QueryRow(
			ctx,
			"SELECT id, username, email, pass_hash, role, created_at, deleted_at, bio FROM users WHERE email = $1",
			email).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.DeletedAt, &user.Bio)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}

	return user, nil
}
