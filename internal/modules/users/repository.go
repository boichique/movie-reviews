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

func (r *Repository) CreateUser(ctx context.Context, user *UserWithPassword) error {
	err := r.db.
		QueryRow(
			ctx,
			`INSERT INTO users (username, email, pass_hash) 
			VALUES ($1, $2, $3) 
			RETURNING id, created_at`,
			user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)

	return err
}

func (r *Repository) GetExistingUserWithPasswordByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	user := newUserWithPassword()

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, username, email, pass_hash, role, created_at, deleted_at, bio 
			FROM users 
			WHERE email = $1 
			AND deleted_at IS NULL`,
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

func (r *Repository) GetExistingUserByID(ctx context.Context, userID int) (*User, error) {
	var user User

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, username, email, role, created_at ,bio 
			FROM users 
			WHERE id = $1 
			AND deleted_at IS NULL`,
			userID).
		Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt, &user.Bio)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, userID int, bio string) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE users 
			SET bio = $1
			WHERE id = $2
			AND deleted_at IS NULL`,
			bio, userID)
	if err != nil {
		return err
	}

	if n.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
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
		AND deleted_at IS NULL`,
			userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if n.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
