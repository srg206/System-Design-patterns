package main

import (
	"context"
	"log"
	"os"
	"time"

	"init_scenario_api/internal/application"
	"init_scenario_api/internal/infastructure/kafka"
	kafkaModels "init_scenario_api/internal/models/kafka"
	"init_scenario_api/internal/usecase/outbox_scenario_processor"
	"init_scenario_api/pkg/common"

	"go.uber.org/zap"
)

func main() {
	os.Exit(run())
}

func run() int {
	app, err := application.NewApp()
	if err != nil {
		log.Fatalf("failed to initialize application: %w", err)
		return common.FailExitCode
	}

	app.Logger.Info("producer service starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.Closer.Add(func() error {
		app.Logger.Info("cancelling producer context...")
		cancel()
		return nil
	})
	outboxScenarioUsecase := outbox_scenario_processor.NewUseCase(app.PostgresRepo, app.KafkaProducer)

	prepapeKafka(ctx, app.Logger, app.Config.Kafka.Brokers)

	go runProducer(ctx, app.Logger, outboxScenarioUsecase)

	app.Closer.Wait()

	app.Logger.Info("producer service stopped")
	return common.SuccessExitCode
}

func runProducer(ctx context.Context, lg *zap.Logger, outboxScenarioUsecase *outbox_scenario_processor.UseCase) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	lg.Info("producer worker started")

	for {
		select {
		case <-ctx.Done():
			lg.Info("producer worker stopping...")
			return
		case <-ticker.C:
			lg.Info("producer processing tick",
				zap.Time("timestamp", time.Now()),
				zap.String("status", "processing"),
			)
			outboxScenarioUsecase.ProcessScenarioOutboxMessages(ctx, kafkaModels.OutboxScenarioTopic, 35)
			lg.Info("producer processed tick",
				zap.Time("timestamp", time.Now()),
				zap.String("status", "processed"),
			)
		}
	}
}

func prepapeKafka(ctx context.Context, lg *zap.Logger, brokers []string) {

	err := kafka.EnsureTopic(ctx, brokers, kafkaModels.OutboxScenarioTopic, 3, 3)
	if err != nil {
		lg.Error("failed to ensure topic exists", zap.Error(err))
		return
	}

	lg.Info("kafka topic prepared", zap.String("topic", kafkaModels.OutboxScenarioTopic))
}
