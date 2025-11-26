package entity

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// InitScenarioPayload представляет payload для инициализации сценария в outbox
type InitScenarioPayload struct {
	ScenarioUUID pgtype.UUID `json:"scenario_uuid"`
	CameraID     int32       `json:"camera_id"`
}

// NewInitScenarioPayload создает новый payload для инициализации сценария
func NewInitScenarioPayload(scenarioUUID pgtype.UUID, cameraID int32) *InitScenarioPayload {
	return &InitScenarioPayload{
		ScenarioUUID: scenarioUUID,
		CameraID:     cameraID,
	}
}
