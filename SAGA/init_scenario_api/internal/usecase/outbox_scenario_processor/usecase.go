package outbox_scenario_processor

import (
	"context"
	"fmt"
	"init_scenario_api/internal/infastructure/kafka"
	"init_scenario_api/internal/infastructure/repository/queries/outbox"
	"init_scenario_api/internal/infastructure/repository/queries/scenario"
	"init_scenario_api/internal/models/entity"
	"init_scenario_api/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UseCase struct {
	repo     Repository
	producer KafkaProducer
}

func NewUseCase(repo Repository, producer KafkaProducer) *UseCase {
	return &UseCase{
		repo:     repo,
		producer: producer,
	}
}

// ProcessScenarioOutboxMessages вычитывает n записей из outbox и отправляет их в Kafka батчем
// topic - название топика для отправки сообщений
// batchSize - количество записей для обработки за один раз
func (uc *UseCase) ProcessScenarioOutboxMessages(ctx context.Context, topic string, batchSize int32) error {
	log := logger.FromContext(ctx)

	log.Info("starting scenario outbox processing",
		zap.String("topic", topic),
		zap.Int32("batch_size", batchSize),
	)

	var outboxRecords []outbox.OutboxScenario

	// Получаем записи из outbox и блокируем их на время (1 минуту)
	err := uc.repo.WithinTransaction(ctx, func(txCtx context.Context) error {
		records, err := uc.repo.GetPendingOutboxScenarios(txCtx, batchSize)
		if err != nil {
			return fmt.Errorf("get pending outbox scenarios: %w", err)
		}

		if len(records) == 0 {
			log.Debug("no pending outbox records found")
			return nil
		}

		outboxRecords = records

		uuids := make([]pgtype.UUID, len(records))
		for i, record := range records {
			uuids[i] = record.OutboxUuid
		}

		if err := uc.repo.LockOutboxScenariosBatch(txCtx, uuids); err != nil {
			return fmt.Errorf("lock outbox scenarios: %w", err)
		}

		log.Info("locked outbox records", zap.Int("count", len(records)))
		return nil
	})

	if err != nil {
		log.Error("failed to get and lock outbox records", zap.Error(err))
		return fmt.Errorf("get and lock records: %w", err)
	}

	if len(outboxRecords) == 0 {
		return nil
	}

	messages := make([]*kafka.Message, 0, len(outboxRecords))
	for _, record := range outboxRecords {
		headers := make(map[string][]byte)
		headers["outbox_uuid"] = []byte(uuidToString(record.OutboxUuid))

		msg := kafka.NewMessage(
			topic,
			nil, // key
			record.Payload,
			headers,
		)
		messages = append(messages, msg)
	}

	log.Info("sending batch to kafka", zap.Int("messages_count", len(messages)))
	if err := uc.producer.SendMessages(ctx, messages); err != nil {
		log.Error("failed to send messages batch to kafka", zap.Error(err))
		return fmt.Errorf("send messages to kafka: %w", err)
	}

	uuids := make([]pgtype.UUID, len(outboxRecords))
	for i, record := range outboxRecords {
		uuids[i] = record.OutboxUuid
	}

	// Помечаем outbox записи как sent и снимаем блокировку
	err = uc.repo.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repo.MarkOutboxScenariosAsSentBatch(txCtx, uuids); err != nil {
			return fmt.Errorf("mark outbox as sent: %w", err)
		}

		scenarioUUIDs := make([]pgtype.UUID, len(outboxRecords))
		for i, record := range outboxRecords {
			scenarioUUIDs[i] = record.ScenarioUuid
		}

		status := entity.StatusInStartupProcessing
		if err := uc.repo.UpdateScenarioStatusBatch(txCtx, scenario.UpdateScenarioStatusBatchParams{
			Column1: scenarioUUIDs,
			Status:  &status,
		}); err != nil {
			return fmt.Errorf("update scenario statuses: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Error("failed to update outbox and scenario statuses", zap.Error(err))
		return fmt.Errorf("update statuses: %w", err)
	}

	log.Info("outbox batch processing completed successfully",
		zap.Int("processed_count", len(outboxRecords)),
		zap.String("topic", topic),
	)

	return nil
}

func uuidToString(pgUUID pgtype.UUID) string {
	if !pgUUID.Valid {
		return ""
	}
	u := uuid.UUID(pgUUID.Bytes)
	return u.String()
}
