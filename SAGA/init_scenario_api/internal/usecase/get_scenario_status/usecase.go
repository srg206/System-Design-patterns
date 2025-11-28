package get_scenario_status

import (
	"context"
	"fmt"
	"init_scenario_api/internal/models/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UseCase struct {
	repo Repository
}

func NewUseCase(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) GetScenarioStatus(ctx context.Context, scenarioUUID string) (*dto.GetScenarioStatusResponse, error) {
	parsedUUID, err := uuid.Parse(scenarioUUID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	var pgUUID pgtype.UUID
	_ = pgUUID.Scan(parsedUUID.String())

	row, err := uc.repo.GetScenarioStatusByUUID(ctx, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("get scenario status: %w", err)
	}

	status := ""
	if row.Status != nil {
		status = *row.Status
	}

	return &dto.GetScenarioStatusResponse{
		ScenarioUUID: scenarioUUID,
		Status:       status,
	}, nil
}
