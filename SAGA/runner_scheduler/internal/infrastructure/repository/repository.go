package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"runner_scheduler/internal/infrastructure/repository/queries/error_marker"
	"runner_scheduler/internal/infrastructure/repository/queries/heartbeat"
	"runner_scheduler/internal/infrastructure/repository/queries/inbox_start_scenario"
	"runner_scheduler/internal/infrastructure/repository/queries/worker"
	modelerror "runner_scheduler/internal/models/error"
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
	workerQueries             *worker.Queries
	heartbeatQueries          *heartbeat.Queries
	errorMarkerQueries        *error_marker.Queries
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{
		dbPool:                    dbPool,
		inboxStartScenarioQueries: inbox_start_scenario.New(dbPool),
		workerQueries:             worker.New(dbPool),
		heartbeatQueries:          heartbeat.New(dbPool),
		errorMarkerQueries:        error_marker.New(dbPool),
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

func (r *Repository) getInboxStartScenarioQueries(ctx context.Context) inbox_start_scenario.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.inboxStartScenarioQueries.WithTx(tx)
	}
	return r.inboxStartScenarioQueries
}

func (r *Repository) getWorkerQueries(ctx context.Context) worker.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.workerQueries.WithTx(tx)
	}
	return r.workerQueries
}

func (r *Repository) getHeartbeatQueries(ctx context.Context) heartbeat.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.heartbeatQueries.WithTx(tx)
	}
	return r.heartbeatQueries
}

func (r *Repository) getErrorMarkerQueries(ctx context.Context) error_marker.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return r.errorMarkerQueries.WithTx(tx)
	}
	return r.errorMarkerQueries
}

func (r *Repository) CreateInboxStartScenario(ctx context.Context, arg inbox_start_scenario.CreateInboxStartScenarioParams) (inbox_start_scenario.InboxStartScenario, error) {
	result, err := r.getInboxStartScenarioQueries(ctx).CreateInboxStartScenario(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation {
			return result, modelerror.ErrDuplicateKey
		}
		return result, err
	}
	return result, nil
}

func (r *Repository) GetInboxStartScenariosByScenarioUUIDs(ctx context.Context, scenarioUUIDs []pgtype.UUID) ([]inbox_start_scenario.InboxStartScenario, error) {
	if len(scenarioUUIDs) == 0 {
		return []inbox_start_scenario.InboxStartScenario{}, nil
	}

	return r.getInboxStartScenarioQueries(ctx).GetInboxStartScenariosByScenarioUUIDs(ctx, scenarioUUIDs)
}

func (r *Repository) DeleteInboxStartScenarios(ctx context.Context, scenarioUUIDs []pgtype.UUID) error {
	if len(scenarioUUIDs) == 0 {
		return nil
	}

	return r.getInboxStartScenarioQueries(ctx).DeleteInboxStartScenarios(ctx, scenarioUUIDs)
}

func (r *Repository) CreateWorker(ctx context.Context, arg worker.CreateWorkerParams) (worker.Worker, error) {
	result, err := r.getWorkerQueries(ctx).CreateWorker(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation {
			return result, modelerror.ErrDuplicateKey
		}
		return result, err
	}
	return result, nil
}

func (r *Repository) CreateNodeWorker(ctx context.Context, arg worker.CreateNodeWorkerParams) (worker.NodeWorker, error) {
	return r.getWorkerQueries(ctx).CreateNodeWorker(ctx, arg)
}

func (r *Repository) GetWorkerByCameraID(ctx context.Context, cameraID int32) (worker.Worker, error) {
	return r.getWorkerQueries(ctx).GetWorkerByCameraID(ctx, cameraID)
}

func (r *Repository) UpdateWorkerStatus(ctx context.Context, arg worker.UpdateWorkerStatusParams) (worker.Worker, error) {
	return r.getWorkerQueries(ctx).UpdateWorkerStatus(ctx, arg)
}

func (r *Repository) DeleteWorker(ctx context.Context, id int32) error {
	return r.getWorkerQueries(ctx).DeleteWorker(ctx, id)
}

func (r *Repository) DeleteNodeWorkersByWorkerIDs(ctx context.Context, workerIDs []int32) error {
	if len(workerIDs) == 0 {
		return nil
	}

	return r.getWorkerQueries(ctx).DeleteNodeWorkersByWorkerIDs(ctx, workerIDs)
}

func (r *Repository) DeleteWorkersByIDs(ctx context.Context, workerIDs []int32) error {
	if len(workerIDs) == 0 {
		return nil
	}

	return r.getWorkerQueries(ctx).DeleteWorkersByIDs(ctx, workerIDs)
}

func (r *Repository) GetOldestWorkersByStatus(ctx context.Context, status string, limit int32) ([]worker.Worker, error) {
	return r.getWorkerQueries(ctx).GetOldestWorkersByStatus(ctx, worker.GetOldestWorkersByStatusParams{
		Status: status,
		Limit:  limit,
	})
}

func (r *Repository) GetWorkersByStatus(ctx context.Context, status string) ([]worker.Worker, error) {
	return r.getWorkerQueries(ctx).GetWorkersByStatus(ctx, status)
}

func (r *Repository) GetLeastLoadedNodes(ctx context.Context, limit int32) ([]worker.GetLeastLoadedNodesRow, error) {
	return r.getWorkerQueries(ctx).GetLeastLoadedNodes(ctx, limit)
}

func (r *Repository) UpsertHeartbeat(ctx context.Context, arg heartbeat.UpsertHeartbeatParams) error {
	return r.getHeartbeatQueries(ctx).UpsertHeartbeat(ctx, arg)
}

func (r *Repository) GetStaleWorkers(ctx context.Context, intervalSeconds int32) ([]error_marker.GetStaleWorkersRow, error) {
	return r.getErrorMarkerQueries(ctx).GetStaleWorkers(ctx, float64(intervalSeconds))
}

func (r *Repository) MarkWorkerAsError(ctx context.Context, id int32) error {
	return r.getErrorMarkerQueries(ctx).MarkWorkerAsError(ctx, id)
}

func (r *Repository) GetStaleNodes(ctx context.Context, intervalSeconds int32) ([]error_marker.GetStaleNodesRow, error) {
	return r.getErrorMarkerQueries(ctx).GetStaleNodes(ctx, float64(intervalSeconds))
}

func (r *Repository) MarkNodeAsError(ctx context.Context, id int32) error {
	return r.getErrorMarkerQueries(ctx).MarkNodeAsError(ctx, id)
}

func (r *Repository) GetWorkersWithoutHeartbeat(ctx context.Context) ([]error_marker.GetWorkersWithoutHeartbeatRow, error) {
	return r.getErrorMarkerQueries(ctx).GetWorkersWithoutHeartbeat(ctx)
}

// WithinTransaction executes a function within a database transaction
func (r *Repository) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx := extractTx(ctx)
	if tx != nil {
		return tFunc(ctx)
	}

	tx, err := r.dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		// Rollback on error
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
