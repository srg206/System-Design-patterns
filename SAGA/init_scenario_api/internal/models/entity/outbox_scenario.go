package entity

import (
	"time"

	"github.com/google/uuid"
)

// OutboxScenario представляет сущность outbox сообщения из таблицы outbox_scenario
type OutboxScenario struct {
	OutboxUUID   uuid.UUID              `json:"outbox_uuid" db:"outbox_uuid"`
	ScenarioUUID uuid.UUID              `json:"scenario_uuid" db:"scenario_uuid"`
	Payload      map[string]interface{} `json:"payload" db:"payload"`
	State        string                 `json:"state" db:"state"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time             `json:"updated_at,omitempty" db:"updated_at"`
	LockedUntil  *time.Time             `json:"locked_until,omitempty" db:"locked_until"`
}

// OutboxState представляет возможные состояния outbox сообщения
const (
	OutboxStatePending = "pending"
	OutboxStateSent    = "sent"
	OutboxStateFailed  = "failed"
)
