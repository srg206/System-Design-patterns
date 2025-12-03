package error_handler_processor

import (
	"context"
	"encoding/json"
	"fmt"

	kafkainfra "runner_scheduler/internal/infrastructure/kafka"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"go.uber.org/zap"
)

const errorStatus = "error"

type Processor struct {
	repo     Repository
	producer Producer
	topic    string
	log      *zap.Logger
}

func NewProcessor(repo Repository, producer Producer, topic string, log *zap.Logger) *Processor {
	return &Processor{repo: repo, producer: producer, topic: topic, log: log}
}

type inboxPayload struct {
	CameraID     int32  `json:"camera_id"`
	ScenarioUUID string `json:"scenario_uuid"`
	URL          string `json:"url"`
}

func (p *Processor) RestartWorkers(ctx context.Context) (int, error) {
	workers, err := p.repo.GetWorkersByStatus(ctx, errorStatus)
	if err != nil {
		return 0, fmt.Errorf("get workers by status: %w", err)
	}

	if len(workers) == 0 {
		p.log.Info("no workers with error status")
		return 0, nil
	}

	p.log.Info("error workers found", zap.Int("count", len(workers)))

	uniqueScenarios := make(map[[16]byte]pgtype.UUID, len(workers))
	workerIDs := make([]int32, 0, len(workers))
	for _, w := range workers {
		workerIDs = append(workerIDs, w.ID)

		if !w.ScenarioUuid.Valid {
			return 0, fmt.Errorf("worker %d has invalid scenario uuid", w.ID)
		}

		key := w.ScenarioUuid.Bytes
		if _, ok := uniqueScenarios[key]; !ok {
			uniqueScenarios[key] = w.ScenarioUuid
		}
	}

	scenarioUUIDs := make([]pgtype.UUID, 0, len(uniqueScenarios))
	for _, uuidValue := range uniqueScenarios {
		scenarioUUIDs = append(scenarioUUIDs, uuidValue)
	}

	inboxRecords, err := p.repo.GetInboxStartScenariosByScenarioUUIDs(ctx, scenarioUUIDs)
	if err != nil {
		return 0, fmt.Errorf("get inbox_start_scenarios: %w", err)
	}

	messages := make([]*kafkainfra.Message, 0, len(inboxRecords))
	for _, record := range inboxRecords {
		payloadBytes, err := json.Marshal(inboxPayload{
			CameraID:     record.CameraID,
			ScenarioUUID: uuidToString(record.ScenarioUuid),
			URL:          record.Url,
		})
		if err != nil {
			return 0, fmt.Errorf("marshal payload for inbox %s: %w", uuidToString(record.OutboxUuid), err)
		}

		headers := map[string][]byte{
			"outbox_uuid": []byte(uuidToString(record.OutboxUuid)),
		}

		messages = append(messages, kafkainfra.NewMessage(p.topic, nil, payloadBytes, headers))
	}

	if err := p.repo.DeleteNodeWorkersByWorkerIDs(ctx, workerIDs); err != nil {
		return 0, fmt.Errorf("delete node_workers: %w", err)
	}

	if err := p.repo.DeleteInboxStartScenarios(ctx, scenarioUUIDs); err != nil {
		return 0, fmt.Errorf("delete inbox_start_scenarios: %w", err)
	}

	if len(messages) > 0 {
		if err := p.producer.SendMessages(ctx, messages); err != nil {
			return 0, fmt.Errorf("send kafka messages: %w", err)
		}
		p.log.Info("requeued scenarios", zap.Int("count", len(messages)))
	}

	if err := p.repo.DeleteWorkersByIDs(ctx, workerIDs); err != nil {
		return 0, fmt.Errorf("delete workers: %w", err)
	}

	return len(workerIDs), nil
}

func uuidToString(pgUUID pgtype.UUID) string {
	if !pgUUID.Valid {
		return ""
	}
	return uuid.UUID(pgUUID.Bytes).String()
}
