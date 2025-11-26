package entity

import (
	"time"

	"github.com/google/uuid"
)

// Scenario представляет сущность сценария из таблицы scenario
type Scenario struct {
	UUID      uuid.UUID  `json:"uuid" db:"uuid"`
	CameraID  int32      `json:"camera_id" db:"camera_id"`
	PredictID int32      `json:"predict_id" db:"predict_id"`
	Status    string     `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// ScenarioStatus представляет возможные статусы сценария
const (
	StatusInitStartup          = "init_startup"
	StatusInStartupProcessing  = "in_startup_processing"
	StatusActive               = "active"
	StatusInitShutdown         = "init_shutdown"
	StatusInShutdownProcessing = "in_shutdown_processing"
	StatusInactive             = "inactive"
)
