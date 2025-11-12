package dto

import (
	"encoding/json"
	"io"
)

// InitScenarioRequest представляет запрос на создание нового сценария
type InitScenarioRequest struct {
	CameraID int32 `json:"camera_id" validate:"required,gt=0"`
}

// Decode декодирует JSON из reader в структуру
func (r *InitScenarioRequest) Decode(reader io.Reader) error {
	return json.NewDecoder(reader).Decode(r)
}

// InitScenarioResponse представляет ответ после создания сценария
type InitScenarioResponse struct {
	ScenarioUUID string `json:"scenario_uuid"`
	Status       string `json:"status"`
}

// UpdateScenarioStatusRequest представляет запрос на обновление статуса сценария
type UpdateScenarioStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// UpdateScenarioPredictRequest представляет запрос на обновление predict_id сценария
type UpdateScenarioPredictRequest struct {
	PredictID int32 `json:"predict_id" validate:"required,gt=0"`
}

// CreateOutboxScenarioRequest представляет запрос на создание outbox сообщения
type CreateOutboxScenarioRequest struct {
	ScenarioUUID string                 `json:"scenario_uuid" validate:"required"`
	Payload      map[string]interface{} `json:"payload" validate:"required"`
}

// UpdateOutboxScenarioStateRequest представляет запрос на обновление состояния outbox сообщения
type UpdateOutboxScenarioStateRequest struct {
	State string `json:"state" validate:"required,oneof=pending sent failed"`
}
