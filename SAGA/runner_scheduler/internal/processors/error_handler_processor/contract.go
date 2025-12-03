package error_handler_processor

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	kafkainfra "runner_scheduler/internal/infrastructure/kafka"
	"runner_scheduler/internal/infrastructure/repository/queries/inbox_start_scenario"
	"runner_scheduler/internal/infrastructure/repository/queries/worker"
)

type Repository interface {
	GetWorkersByStatus(ctx context.Context, status string) ([]worker.Worker, error)
	GetInboxStartScenariosByScenarioUUIDs(ctx context.Context, scenarioUUIDs []pgtype.UUID) ([]inbox_start_scenario.InboxStartScenario, error)
	DeleteNodeWorkersByWorkerIDs(ctx context.Context, workerIDs []int32) error
	DeleteWorkersByIDs(ctx context.Context, workerIDs []int32) error
	DeleteInboxStartScenarios(ctx context.Context, scenarioUUIDs []pgtype.UUID) error
}

type Producer interface {
	SendMessages(ctx context.Context, msgs []*kafkainfra.Message) error
}
