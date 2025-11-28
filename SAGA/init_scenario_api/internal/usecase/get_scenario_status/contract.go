package get_scenario_status

import (
	"context"
	"init_scenario_api/internal/infastructure/repository/queries/scenario"

	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	GetScenarioStatusByUUID(ctx context.Context, uuid pgtype.UUID) (scenario.GetScenarioStatusByUUIDRow, error)
}
