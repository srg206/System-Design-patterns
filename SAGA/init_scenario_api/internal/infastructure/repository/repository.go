package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"init_scenario_api/internal/infastructure/repository/queries"
)

type Repository struct {
	dbPool  *pgxpool.Pool
	queries *queries.Queries
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{
		dbPool:  dbPool,
		queries: queries.New(dbPool),
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

func (r *Repository) Queries(ctx context.Context) *queries.Queries {
	tx := extractTx(ctx)
	if tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *Repository) CreateOutboxScenario(ctx context.Context, arg queries.CreateOutboxScenarioParams) (queries.OutboxScenario, error) {
	return r.Queries(ctx).CreateOutboxScenario(ctx, arg)
}

func (r *Repository) CreateScenario(ctx context.Context, arg queries.CreateScenarioParams) (queries.Scenario, error) {
	return r.Queries(ctx).CreateScenario(ctx, arg)
}

func (r *Repository) UpdateOutboxScenarioState(ctx context.Context, arg queries.UpdateOutboxScenarioStateParams) error {
	return r.Queries(ctx).UpdateOutboxScenarioState(ctx, arg)
}

func (r *Repository) UpdateScenarioPredictByUUID(ctx context.Context, arg queries.UpdateScenarioPredictByUUIDParams) error {
	return r.Queries(ctx).UpdateScenarioPredictByUUID(ctx, arg)
}

func (r *Repository) UpdateScenarioStatusByUUID(ctx context.Context, arg queries.UpdateScenarioStatusByUUIDParams) error {
	return r.Queries(ctx).UpdateScenarioStatusByUUID(ctx, arg)
}

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

func (r *Repository) GetDBPool() *pgxpool.Pool {
	return r.dbPool
}
