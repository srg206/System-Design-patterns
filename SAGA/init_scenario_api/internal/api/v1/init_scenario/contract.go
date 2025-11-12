package init_scenario

import (
	"context"
	"init_scenario_api/internal/models/dto"
)

// InitScenarioUseCase определяет интерфейс use case для создания покупки
type InitScenarioUseCase interface {
	InitScenario(ctx context.Context, input dto.InitScenarioRequest) (*dto.InitScenarioResponse, error)
}
