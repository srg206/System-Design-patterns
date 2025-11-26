package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"inference_scheduler/internal/config"
	"inference_scheduler/internal/infrastructure/kafka"
	"inference_scheduler/internal/infrastructure/repository"
	"inference_scheduler/internal/infrastructure/repository/queries/inbox_start_scenario"
	modelerror "inference_scheduler/internal/models/error"
	"inference_scheduler/pkg/closer"
	"inference_scheduler/pkg/database"
	"inference_scheduler/pkg/logger"

	modelKafka "inference_scheduler/internal/models/kafka"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	// Инициализация логгера
	log, err := logger.InitLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		return 1
	}
	defer log.Sync()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", zap.Error(err))
		return 1
	}

	// Создание closer
	cls := closer.New(10 * time.Second)

	// Подключение к БД
	dbPool, err := database.NewPool(ctx, cfg.Database, cfg.Pool)
	if err != nil {
		log.Error("failed to create db pool", zap.Error(err))
		return 1
	}
	cls.Add(func() error {
		log.Info("closing database connection")
		database.Close(dbPool)
		return nil
	})

	// Создание репозитория
	repo := repository.NewRepository(dbPool)

	// Создание Kafka consumer
	kafkaCfg := kafka.DefaultConfig(cfg.Consumer.KafkaBrokers...)
	kafkaCfg.ConsumerGroup = cfg.Consumer.KafkaConsumerGroup

	consumer, err := kafka.NewKafkaConsumer(
		kafkaCfg,
		[]string{modelKafka.OutboxScenarioApi},
		modelKafka.KafkaConsumerGroup,
		log,
	)
	if err != nil {
		log.Error("failed to create kafka consumer", zap.Error(err))
		return 1
	}
	cls.Add(func() error {
		log.Info("closing kafka consumer")
		return consumer.Close()
	})

	// Запуск обработки сообщений в горутине
	go func() {
		for {
			time.Sleep(time.Millisecond * 500)

			// Проверка на shutdown
			if cls.IsShutdown() {
				log.Info("shutdown signal received, stopping message processing")
				break
			}
			log.Info("message processing")
			// Чтение сообщения из Kafka
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				log.Error("failed to read message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			log.Info("message read successfully")

			// Парсинг сообщения
			var payload struct {
				CameraID     int32  `json:"camera_id"`
				ScenarioUUID string `json:"scenario_uuid"`
			}

			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Error("failed to unmarshal message", zap.Error(err))
				continue
			}

			// Читаем outbox_uuid из headers
			outboxUUIDBytes, ok := msg.Headers["outbox_uuid"]
			if !ok {
				log.Error("outbox_uuid header not found")
				continue
			}
			outboxUUIDStr := string(outboxUUIDBytes)

			// Debug: логируем распарсенный payload
			log.Debug("parsed message payload",
				zap.String("outbox_uuid", outboxUUIDStr),
				zap.Int32("camera_id", payload.CameraID),
				zap.String("scenario_uuid", payload.ScenarioUUID),
				zap.String("raw_message", string(msg.Value)),
			)

			// Сохранение в БД
			outboxUUID := pgtype.UUID{}
			if err := outboxUUID.Scan(outboxUUIDStr); err != nil {
				log.Error("failed to parse outbox_uuid",
					zap.Error(err),
					zap.String("outbox_uuid_string", outboxUUIDStr))
				continue
			}

			scenarioUUID := pgtype.UUID{}
			if err := scenarioUUID.Scan(payload.ScenarioUUID); err != nil {
				log.Error("failed to parse scenario_uuid", zap.Error(err))
				continue
			}

			_, err = repo.CreateInboxStartScenario(ctx, inbox_start_scenario.CreateInboxStartScenarioParams{
				OutboxUuid:   outboxUUID,
				CameraID:     payload.CameraID,
				ScenarioUuid: scenarioUUID,
			})
			if err != nil {
				// Проверяем, является ли это ошибкой дублирования ключа
				if errors.Is(err, modelerror.ErrDuplicateKey) {
					// Это нормальная ситуация при повторной обработке сообщения (идемпотентность)
					log.Warn("message already processed (duplicate key), skipping",
						zap.String("outbox_uuid", outboxUUIDStr),
						zap.Int32("camera_id", payload.CameraID),
						zap.String("scenario_uuid", payload.ScenarioUUID),
					)
					// Продолжаем выполнение для commit'а сообщения
				} else {
					// Это действительно ошибка - не делаем commit
					log.Error("failed to save message to db", zap.Error(err))
					continue
				}
			} else {
				log.Info("message saved to db successfully")
			}

			// Commit сообщения в Kafka после успешного сохранения в БД
			if err := consumer.CommitMessages(ctx, msg); err != nil {
				log.Error("failed to commit message", zap.Error(err))
				continue
			}
			log.Info("message committed successfully")
			log.Info("message processed and committed successfully",
				zap.String("outbox_uuid", outboxUUIDStr),
				zap.Int32("camera_id", payload.CameraID),
				zap.String("scenario_uuid", payload.ScenarioUUID),
			)
		}
	}()

	log.Info("consumer started successfully")

	// Ожидание сигнала завершения
	cls.Wait()

	return 0
}
