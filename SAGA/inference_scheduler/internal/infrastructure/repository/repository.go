package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"inference_scheduler/internal/infrastructure/repository/queries/inbox_start_scenario"
	modelerror "inference_scheduler/internal/models/error"
)

// PostgreSQL error codes
// Полный список: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	// pgErrCodeUniqueViolation - нарушение уникального ограничения (duplicate key)
	pgErrCodeUniqueViolation = "23505"
)

type Repository struct {
	dbPool                    *pgxpool.Pool
	inboxStartScenarioQueries *inbox_start_scenario.Queries
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{
		dbPool:                    dbPool,
		inboxStartScenarioQueries: inbox_start_scenario.New(dbPool),
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

// getInboxStartScenarioQueries returns inbox_start_scenario queries with transaction if exists in context
func (r *Repository) getInboxStartScenarioQueries(ctx context.Context) inbox_start_scenario.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.inboxStartScenarioQueries.WithTx(tx)
	}
	return r.inboxStartScenarioQueries
}

// CreateInboxStartScenario creates a new inbox_start_scenario record
func (r *Repository) CreateInboxStartScenario(ctx context.Context, arg inbox_start_scenario.CreateInboxStartScenarioParams) (inbox_start_scenario.InboxStartScenario, error) {
	result, err := r.getInboxStartScenarioQueries(ctx).CreateInboxStartScenario(ctx, arg)
	if err != nil {
		// Проверяем, является ли ошибка нарушением уникального ключа (duplicate key)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation {
			return result, modelerror.ErrDuplicateKey
		}
		return result, err
	}
	return result, nil
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
