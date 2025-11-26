package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"init_scenario_api/internal/infastructure/repository/queries/outbox"
	"init_scenario_api/internal/infastructure/repository/queries/scenario"
)

type Repository struct {
	dbPool          *pgxpool.Pool
	scenarioQueries *scenario.Queries
	outboxQueries   *outbox.Queries
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{
		dbPool:          dbPool,
		scenarioQueries: scenario.New(dbPool),
		outboxQueries:   outbox.New(dbPool),
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

// getScenarioQueries returns scenario queries with transaction if exists in context
func (r *Repository) getScenarioQueries(ctx context.Context) scenario.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.scenarioQueries.WithTx(tx)
	}
	return r.scenarioQueries
}

// getOutboxQueries returns outbox queries with transaction if exists in context
func (r *Repository) getOutboxQueries(ctx context.Context) outbox.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.outboxQueries.WithTx(tx)
	}
	return r.outboxQueries
}

// CreateScenario creates a new scenario record
func (r *Repository) CreateScenario(ctx context.Context, arg scenario.CreateScenarioParams) (scenario.Scenario, error) {
	return r.getScenarioQueries(ctx).CreateScenario(ctx, arg)
}

// UpdateScenarioStatusByUUID updates scenario status by UUID
func (r *Repository) UpdateScenarioStatusByUUID(ctx context.Context, arg scenario.UpdateScenarioStatusByUUIDParams) error {
	return r.getScenarioQueries(ctx).UpdateScenarioStatusByUUID(ctx, arg)
}

// UpdateScenarioStatusBatch updates scenario status for multiple UUIDs
func (r *Repository) UpdateScenarioStatusBatch(ctx context.Context, arg scenario.UpdateScenarioStatusBatchParams) error {
	return r.getScenarioQueries(ctx).UpdateScenarioStatusBatch(ctx, arg)
}

// UpdateScenarioPredictByUUID updates scenario predict_id by UUID
func (r *Repository) UpdateScenarioPredictByUUID(ctx context.Context, arg scenario.UpdateScenarioPredictByUUIDParams) error {
	return r.getScenarioQueries(ctx).UpdateScenarioPredictByUUID(ctx, arg)
}

// CreateOutboxScenario creates a new outbox scenario record
func (r *Repository) CreateOutboxScenario(ctx context.Context, arg outbox.CreateOutboxScenarioParams) (outbox.OutboxScenario, error) {
	return r.getOutboxQueries(ctx).CreateOutboxScenario(ctx, arg)
}

// UpdateOutboxScenarioState updates outbox scenario state by scenario UUID
func (r *Repository) UpdateOutboxScenarioState(ctx context.Context, arg outbox.UpdateOutboxScenarioStateParams) error {
	return r.getOutboxQueries(ctx).UpdateOutboxScenarioState(ctx, arg)
}

// LockOutboxScenario locks outbox scenario until specified time
func (r *Repository) LockOutboxScenario(ctx context.Context, arg outbox.LockOutboxScenarioParams) error {
	return r.getOutboxQueries(ctx).LockOutboxScenario(ctx, arg)
}

// GetPendingOutboxScenarios retrieves pending outbox records with row-level lock
func (r *Repository) GetPendingOutboxScenarios(ctx context.Context, limit int32) ([]outbox.OutboxScenario, error) {
	return r.getOutboxQueries(ctx).GetPendingOutboxScenarios(ctx, limit)
}

// LockOutboxScenariosBatch locks multiple outbox scenarios for 1 minute
func (r *Repository) LockOutboxScenariosBatch(ctx context.Context, outboxUUIDs []pgtype.UUID) error {
	return r.getOutboxQueries(ctx).LockOutboxScenariosBatch(ctx, outboxUUIDs)
}

// MarkOutboxScenariosAsSentBatch marks multiple outbox scenarios as sent and unlocks them
func (r *Repository) MarkOutboxScenariosAsSentBatch(ctx context.Context, outboxUUIDs []pgtype.UUID) error {
	return r.getOutboxQueries(ctx).MarkOutboxScenariosAsSentBatch(ctx, outboxUUIDs)
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
