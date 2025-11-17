package init_scenario

import (
	"context"
	"init_scenario_api/internal/infastructure/repository/queries/outbox"
	"init_scenario_api/internal/infastructure/repository/queries/scenario"
)

type Repository interface {
	CreateScenario(ctx context.Context, arg scenario.CreateScenarioParams) (scenario.Scenario, error)
	CreateOutboxScenario(ctx context.Context, arg outbox.CreateOutboxScenarioParams) (outbox.OutboxScenario, error)
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}
