package entity

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// InitScenarioPayload представляет payload для инициализации сценария в outbox
type InitScenarioPayload struct {
	ScenarioUUID pgtype.UUID `json:"scenario_uuid"`
	CameraID     int32       `json:"camera_id"`
	URL          string      `json:"url"`
}

// NewInitScenarioPayload создает новый payload для инициализации сценария
func NewInitScenarioPayload(scenarioUUID pgtype.UUID, cameraID int32, url string) *InitScenarioPayload {
	return &InitScenarioPayload{
		ScenarioUUID: scenarioUUID,
		CameraID:     cameraID,
		URL:          url,
	}
}
