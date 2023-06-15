package dbx

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var contextKey = struct{}{}

func WithTransaction(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, contextKey, tx)
}

func FromContext(ctx context.Context, def Queryable) Queryable {
	if tx, ok := ctx.Value(contextKey).(pgx.Tx); ok {
		return tx
	}

	return def
}

func InTransaction(ctx context.Context, db *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction:  %w", err)
	}

	ctx = WithTransaction(ctx, tx)
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				err = errors.Join(err, fmt.Errorf("rollback transaction: %w", err))
			}
		} else {
			cerr := tx.Commit(ctx)
			if cerr != nil {
				err = fmt.Errorf("commit transaction:  %w", err)
			}
		}
	}()

	return fn(ctx, tx)
}
