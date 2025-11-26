package convert

import (
	"encoding/json"
	"init_scenario_api/internal/infastructure/repository/queries"
	"init_scenario_api/internal/models/dto"
	"init_scenario_api/internal/models/entity"

	"github.com/google/uuid"
)

func ScenarioFromDB(dbScenario queries.Scenario) *entity.Scenario {
	scenario := &entity.Scenario{
		CameraID: dbScenario.CameraID,
	}

	if dbScenario.Uuid.Valid {
		scenario.UUID = uuid.UUID(dbScenario.Uuid.Bytes)
	}

	if dbScenario.PredictID != nil {
		scenario.PredictID = *dbScenario.PredictID
	}

	if dbScenario.Status != nil {
		scenario.Status = *dbScenario.Status
	}

	if dbScenario.CreatedAt.Valid {
		scenario.CreatedAt = dbScenario.CreatedAt.Time
	}

	if dbScenario.UpdatedAt.Valid {
		scenario.UpdatedAt = &dbScenario.UpdatedAt.Time
	}

	return scenario
}

func OutboxScenarioFromDB(dbOutbox queries.OutboxScenario) *entity.OutboxScenario {
	outbox := &entity.OutboxScenario{}

	if dbOutbox.OutboxUuid.Valid {
		outbox.OutboxUUID = uuid.UUID(dbOutbox.OutboxUuid.Bytes)
	}

	if dbOutbox.ScenarioUuid.Valid {
		outbox.ScenarioUUID = uuid.UUID(dbOutbox.ScenarioUuid.Bytes)
	}

	if len(dbOutbox.Payload) > 0 {
		var payload map[string]interface{}
		if err := json.Unmarshal(dbOutbox.Payload, &payload); err == nil {
			outbox.Payload = payload
		}
	}

	if dbOutbox.State != nil {
		outbox.State = *dbOutbox.State
	}

	if dbOutbox.CreatedAt.Valid {
		outbox.CreatedAt = dbOutbox.CreatedAt.Time
	}

	if dbOutbox.UpdatedAt.Valid {
		outbox.UpdatedAt = &dbOutbox.UpdatedAt.Time
	}

	return outbox
}

func ScenarioToDTO(scenario *entity.Scenario) *dto.InitScenarioResponse {
	return &dto.InitScenarioResponse{
		ScenarioUUID: scenario.UUID.String(),
		Status:       scenario.Status,
	}
}
