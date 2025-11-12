package init_scenario

import (
	"context"
	"init_scenario_api/internal/infastructure/repository/queries"
)

type Repository interface {
	queries.Querier
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}
