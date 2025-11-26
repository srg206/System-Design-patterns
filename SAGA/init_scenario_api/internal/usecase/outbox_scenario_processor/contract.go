package outbox_scenario_processor

import (
	"context"
	"init_scenario_api/internal/infastructure/kafka"
	"init_scenario_api/internal/infastructure/repository/queries/outbox"
	"init_scenario_api/internal/infastructure/repository/queries/scenario"

	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	GetPendingOutboxScenarios(ctx context.Context, limit int32) ([]outbox.OutboxScenario, error)
	LockOutboxScenariosBatch(ctx context.Context, outboxUUIDs []pgtype.UUID) error
	MarkOutboxScenariosAsSentBatch(ctx context.Context, outboxUUIDs []pgtype.UUID) error
	UpdateScenarioStatusBatch(ctx context.Context, arg scenario.UpdateScenarioStatusBatchParams) error
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}

type KafkaProducer interface {
	SendMessage(ctx context.Context, msg *kafka.Message) error
	SendMessages(ctx context.Context, msgs []*kafka.Message) error
}
