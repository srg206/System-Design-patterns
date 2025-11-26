package scheduler_processor

import "context"

type Repository interface {
	GetPendingOutboxScenarios(ctx context.Context, limit int32) ([]outbox.OutboxScenario, error)
}
