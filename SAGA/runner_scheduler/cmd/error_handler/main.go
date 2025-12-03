package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"runner_scheduler/internal/config"
	kafkainfra "runner_scheduler/internal/infrastructure/kafka"
	"runner_scheduler/internal/infrastructure/repository"
	modelKafka "runner_scheduler/internal/models/kafka"
	"runner_scheduler/internal/processors/error_handler_processor"
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

	dbPool, err := database.NewPool(ctx, cfg.Database, cfg.Pool)
	if err != nil {
		log.Error("failed to init database pool", zap.Error(err))
		return 1
	}
	defer database.Close(dbPool)

	producer, err := kafkainfra.NewKafkaProducer(kafkainfra.DefaultConfig(cfg.Kafka.Brokers...), log)
	if err != nil {
		log.Error("failed to init kafka producer", zap.Error(err))
		return 1
	}
	defer func() {
		if closeErr := producer.Close(); closeErr != nil {
			log.Error("failed to close kafka producer", zap.Error(closeErr))
		}
	}()

	repo := repository.NewRepository(dbPool)
	processor := error_handler_processor.NewProcessor(repo, producer, modelKafka.OutboxScenarioApi, log)

	for {
		time.Sleep(3 * time.Second)
		if _, err := processor.RestartWorkers(ctx); err != nil {
			log.Error("restart workers failed", zap.Error(err))
			return 1
		}
	}
}
