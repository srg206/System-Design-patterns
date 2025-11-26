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

func (r *Repository) getScenarioQueries(ctx context.Context) scenario.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.scenarioQueries.WithTx(tx)
	}
	return r.scenarioQueries
}

func (r *Repository) getOutboxQueries(ctx context.Context) outbox.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.outboxQueries.WithTx(tx)
	}
	return r.outboxQueries
}

func (r *Repository) CreateScenario(ctx context.Context, arg scenario.CreateScenarioParams) (scenario.Scenario, error) {
	return r.getScenarioQueries(ctx).CreateScenario(ctx, arg)
}

func (r *Repository) UpdateScenarioStatusByUUID(ctx context.Context, arg scenario.UpdateScenarioStatusByUUIDParams) error {
	return r.getScenarioQueries(ctx).UpdateScenarioStatusByUUID(ctx, arg)
}

func (r *Repository) UpdateScenarioStatusBatch(ctx context.Context, arg scenario.UpdateScenarioStatusBatchParams) error {
	return r.getScenarioQueries(ctx).UpdateScenarioStatusBatch(ctx, arg)
}

func (r *Repository) UpdateScenarioPredictByUUID(ctx context.Context, arg scenario.UpdateScenarioPredictByUUIDParams) error {
	return r.getScenarioQueries(ctx).UpdateScenarioPredictByUUID(ctx, arg)
}

func (r *Repository) CreateOutboxScenario(ctx context.Context, arg outbox.CreateOutboxScenarioParams) (outbox.OutboxScenario, error) {
	return r.getOutboxQueries(ctx).CreateOutboxScenario(ctx, arg)
}

func (r *Repository) UpdateOutboxScenarioState(ctx context.Context, arg outbox.UpdateOutboxScenarioStateParams) error {
	return r.getOutboxQueries(ctx).UpdateOutboxScenarioState(ctx, arg)
}

func (r *Repository) LockOutboxScenario(ctx context.Context, arg outbox.LockOutboxScenarioParams) error {
	return r.getOutboxQueries(ctx).LockOutboxScenario(ctx, arg)
}

func (r *Repository) GetPendingOutboxScenarios(ctx context.Context, limit int32) ([]outbox.OutboxScenario, error) {
	return r.getOutboxQueries(ctx).GetPendingOutboxScenarios(ctx, limit)
}

func (r *Repository) LockOutboxScenariosBatch(ctx context.Context, outboxUUIDs []pgtype.UUID) error {
	return r.getOutboxQueries(ctx).LockOutboxScenariosBatch(ctx, outboxUUIDs)
}

func (r *Repository) MarkOutboxScenariosAsSentBatch(ctx context.Context, outboxUUIDs []pgtype.UUID) error {
	return r.getOutboxQueries(ctx).MarkOutboxScenariosAsSentBatch(ctx, outboxUUIDs)
}

func (r *Repository) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
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
			panic(p)
		}
	}()

	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) GetDBPool() *pgxpool.Pool {
	return r.dbPool
}
