package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"runner_scheduler/internal/config"
	"runner_scheduler/internal/infrastructure/kafka"
	"runner_scheduler/internal/infrastructure/repository"
	"runner_scheduler/internal/infrastructure/repository/queries/heartbeat"
	modelKafka "runner_scheduler/internal/models/kafka"
	"runner_scheduler/pkg/closer"
	"runner_scheduler/pkg/database"
	"runner_scheduler/pkg/logger"

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

	kafkaCfg := kafka.DefaultConfig(cfg.Consumer.KafkaBrokers...)
	kafkaCfg.ConsumerGroup = modelKafka.HeartbeatConsumerGroup

	consumer, err := kafka.NewKafkaConsumer(
		kafkaCfg,
		[]string{modelKafka.HeartbeatTopic},
		modelKafka.HeartbeatConsumerGroup,
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

	go func() {
		for {
			if cls.IsShutdown() {
				log.Info("shutdown signal received, stopping message processing")
				break
			}

			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					break
				}
				log.Error("failed to read message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			// var payload struct {
			// 	NodeID    int32         `json:"node_id"`
			// 	CameraIDs []interface{} `json:"camera_ids"`
			// 	UpdatedAt time.Time     `json:"updated_at"`
			// }
			var payload struct {
				NodeID    string  `json:"NodeId"`
				CameraIDs []int32 `json:"camera_ids"`
			}

			fmt.Println(msg.Value)
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Error("failed to unmarshal message", zap.Error(err))
				continue
			}
			fmt.Println(payload)
			nodeID, err := strconv.ParseInt(payload.NodeID, 10, 32)
			if err != nil {
				log.Error("failed to parse node_id", zap.Error(err))
				continue
			}
			if len(payload.CameraIDs) == 0 {
				err = repo.UpsertHeartbeat(ctx, heartbeat.UpsertHeartbeatParams{
					NodeID:   int32(nodeID),
					CameraID: "",
				})

				if err != nil {
					log.Error("failed to save heartbeat to db",
						zap.Error(err),
						zap.Int32("node_id", int32(nodeID)),
						zap.String("camera_id", ""))
				}

				log.Info("heartbeat processed with empty camera_id",
					zap.Int32("node_id", int32(nodeID)))

				if err := consumer.CommitMessages(ctx, msg); err != nil {
					log.Error("failed to commit message", zap.Error(err))
				}
				continue
			}

			successCount := 0
			for _, cameraIDRaw := range payload.CameraIDs {
				var cameraIDStr string
				cameraIDStr = strconv.Itoa(int(cameraIDRaw))

				err = repo.UpsertHeartbeat(ctx, heartbeat.UpsertHeartbeatParams{
					NodeID:   int32(nodeID),
					CameraID: cameraIDStr,
				})

				if err != nil {
					log.Error("failed to save heartbeat to db",
						zap.Error(err),
						zap.Int32("node_id", int32(nodeID)),
						zap.String("camera_id", cameraIDStr))
				} else {
					successCount++
				}
			}

			log.Info("heartbeats processed",
				zap.Int32("node_id", int32(nodeID)),
				zap.Int("total", len(payload.CameraIDs)),
				zap.Int("success", successCount))

			if err := consumer.CommitMessages(ctx, msg); err != nil {
				log.Error("failed to commit message", zap.Error(err))
			}
		}
	}()

	log.Info("heartbeat consumer started successfully")

	cls.Wait()

	return 0
}
