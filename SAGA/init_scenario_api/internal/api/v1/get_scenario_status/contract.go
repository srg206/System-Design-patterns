package get_scenario_status

import (
	"context"
	"init_scenario_api/internal/models/dto"
)

type GetScenarioStatusUseCase interface {
	GetScenarioStatus(ctx context.Context, scenarioUUID string) (*dto.GetScenarioStatusResponse, error)
}
