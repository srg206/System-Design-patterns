package convert

import (
	"encoding/json"
	"init_scenario_api/internal/infastructure/repository/queries/outbox"
	"init_scenario_api/internal/infastructure/repository/queries/scenario"
	"init_scenario_api/internal/models/dto"
	"init_scenario_api/internal/models/entity"

	"github.com/google/uuid"
)

func ScenarioFromDB(dbScenario scenario.Scenario) *entity.Scenario {
	result := &entity.Scenario{
		CameraID: dbScenario.CameraID,
	}

	if dbScenario.Uuid.Valid {
		result.UUID = uuid.UUID(dbScenario.Uuid.Bytes)
	}

	if dbScenario.PredictID != nil {
		result.PredictID = *dbScenario.PredictID
	}

	if dbScenario.Status != nil {
		result.Status = *dbScenario.Status
	}

	if dbScenario.CreatedAt.Valid {
		result.CreatedAt = dbScenario.CreatedAt.Time
	}

	if dbScenario.UpdatedAt.Valid {
		result.UpdatedAt = &dbScenario.UpdatedAt.Time
	}

	return result
}

func OutboxScenarioFromDB(dbOutbox outbox.OutboxScenario) *entity.OutboxScenario {
	result := &entity.OutboxScenario{}

	if dbOutbox.OutboxUuid.Valid {
		result.OutboxUUID = uuid.UUID(dbOutbox.OutboxUuid.Bytes)
	}

	if dbOutbox.ScenarioUuid.Valid {
		result.ScenarioUUID = uuid.UUID(dbOutbox.ScenarioUuid.Bytes)
	}

	if len(dbOutbox.Payload) > 0 {
		var payload map[string]interface{}
		if err := json.Unmarshal(dbOutbox.Payload, &payload); err == nil {
			result.Payload = payload
		}
	}

	if dbOutbox.State != nil {
		result.State = *dbOutbox.State
	}

	if dbOutbox.CreatedAt.Valid {
		result.CreatedAt = dbOutbox.CreatedAt.Time
	}

	if dbOutbox.UpdatedAt.Valid {
		result.UpdatedAt = &dbOutbox.UpdatedAt.Time
	}

	return result
}

func ScenarioToDTO(scenario *entity.Scenario) *dto.InitScenarioResponse {
	return &dto.InitScenarioResponse{
		ScenarioUUID: scenario.UUID.String(),
		Status:       scenario.Status,
	}
}
