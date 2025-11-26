package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	dbPool *pgxpool.Pool
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{
		dbPool: dbPool,
	}
}

type txKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

// WithinTransaction executes a function within a database transaction
func (r *Repository) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	// If already in transaction, just execute the function
	tx := extractTx(ctx)
	if tx != nil {
		return tFunc(ctx)
	}

	// Start new transaction
	tx, err := r.dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Handle panic and rollback
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw panic after rollback
		}
	}()

	// Execute function with transaction in context
	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetDBPool returns the database connection pool
func (r *Repository) GetDBPool() *pgxpool.Pool {
	return r.dbPool
}
