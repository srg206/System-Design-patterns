package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"runner_scheduler/internal/config"
	"runner_scheduler/internal/infrastructure/repository"
	"runner_scheduler/internal/infrastructure/repository/queries/inbox_start_scenario"
	modelerror "runner_scheduler/internal/models/error"
	"runner_scheduler/pkg/closer"
	"runner_scheduler/pkg/database"
	"runner_scheduler/pkg/logger"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	log, err := logger.InitLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		return 1
	}
	defer log.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", zap.Error(err))
		return 1
	}

	cls := closer.New(10 * time.Second)

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

	repo := repository.NewRepository(dbPool)

	go func() {
		for {
			time.Sleep(time.Millisecond * 500)

			if cls.IsShutdown() {
				log.Info("shutdown signal received, stopping message processing")
				break
			}
			log.Info("message processing")
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				log.Error("failed to read message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			log.Info("message read successfully")

			var payload struct {
				CameraID     int32  `json:"camera_id"`
				ScenarioUUID string `json:"scenario_uuid"`
				URL          string `json:"url"`
			}

			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Error("failed to unmarshal message", zap.Error(err))
				continue
			}

			outboxUUIDBytes, ok := msg.Headers["outbox_uuid"]
			if !ok {
				log.Error("outbox_uuid header not found")
				continue
			}
			outboxUUIDStr := string(outboxUUIDBytes)

			log.Debug("parsed message payload",
				zap.String("outbox_uuid", outboxUUIDStr),
				zap.Int32("camera_id", payload.CameraID),
				zap.String("scenario_uuid", payload.ScenarioUUID),
				zap.String("url", payload.URL),
				zap.String("raw_message", string(msg.Value)),
			)

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
				Url:          payload.URL,
			})
			if err != nil {
				if errors.Is(err, modelerror.ErrDuplicateKey) {
					log.Warn("message already processed (duplicate key), skipping",
						zap.String("outbox_uuid", outboxUUIDStr),
						zap.Int32("camera_id", payload.CameraID),
						zap.String("scenario_uuid", payload.ScenarioUUID),
						zap.String("url", payload.URL),
					)
				} else {
					log.Error("failed to save message to db", zap.Error(err))
					continue
				}
			} else {
				log.Info("message saved to db successfully")
			}

			if err := consumer.CommitMessages(ctx, msg); err != nil {
				log.Error("failed to commit message", zap.Error(err))
				continue
			}
			log.Info("message committed successfully")
			log.Info("message processed and committed successfully",
				zap.String("outbox_uuid", outboxUUIDStr),
				zap.Int32("camera_id", payload.CameraID),
				zap.String("scenario_uuid", payload.ScenarioUUID),
				zap.String("url", payload.URL),
			)
		}
	}()

	log.Info("consumer started successfully")

	cls.Wait()

	return 0
}
