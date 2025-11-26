package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"inference_scheduler/internal/application"
	infraKafka "inference_scheduler/internal/infrastructure/kafka"
	"inference_scheduler/pkg/common"

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

	app.Logger.Info("consumer service starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	app.Closer.Add(func() error {
		app.Logger.Info("cancelling consumer context...")
		cancel()
		return nil
	})

	// Initialize Kafka consumer
	kafkaConfig := &infraKafka.Config{
		Brokers:       app.Config.Consumer.KafkaBrokers,
		ConsumerGroup: app.Config.Consumer.KafkaConsumerGroup,
		ReadTimeout:   10,
		WriteTimeout:  10,
	}

	topics := []string{app.Config.Consumer.KafkaInboxInferenceTopic}
	consumer, err := infraKafka.NewKafkaConsumer(kafkaConfig, topics, app.Logger)
	if err != nil {
		app.Logger.Error("failed to create kafka consumer", zap.Error(err))
		return common.FailExitCode
	}

	app.Closer.Add(func() error {
		app.Logger.Info("closing kafka consumer...")
		return consumer.Close()
	})

	app.Logger.Info("consumer initialized",
		zap.String("consumer_group", app.Config.Consumer.KafkaConsumerGroup),
		zap.Strings("topics", topics),
	)

	// Run consumer
	go runConsumer(ctx, app.Logger, consumer)

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		app.Logger.Info("received shutdown signal", zap.String("signal", sig.String()))
	case <-ctx.Done():
		app.Logger.Info("context cancelled")
	}

	app.Closer.Wait()

	app.Logger.Info("consumer service stopped")
	return common.SuccessExitCode
}

func runConsumer(ctx context.Context, lg *zap.Logger, consumer infraKafka.Consumer) {
	lg.Info("consumer worker started")

	for {
		select {
		case <-ctx.Done():
			lg.Info("consumer worker stopping...")
			return
		default:
			// Read single message
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				lg.Error("failed to read message", zap.Error(err))
				continue
			}

			// Process message
			processMessage(lg, msg)

			// Example: Read multiple messages at once (uncomment to use batch processing)
			// messages, err := consumer.ReadMessages(ctx, 10)
			// if err != nil {
			// 	if ctx.Err() != nil {
			// 		return
			// 	}
			// 	lg.Error("failed to read messages batch", zap.Error(err))
			// 	continue
			// }
			//
			// for _, msg := range messages {
			// 	processMessage(lg, msg)
			// }
		}
	}
}

func processMessage(lg *zap.Logger, msg *infraKafka.Message) {
	logFields := []zap.Field{
		zap.String("topic", msg.Topic),
	}
	if msg.Key != nil {
		logFields = append(logFields, zap.String("key", *msg.Key))
	}

	lg.Info("processing message", logFields...)

	// Parse message value as JSON for debugging
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		lg.Warn("failed to parse message as JSON, treating as raw data",
			zap.Error(err),
			zap.String("raw_value", string(msg.Value)),
		)
	} else {
		lg.Info("message payload", zap.Any("payload", payload))
	}

	// TODO: Add your business logic here
	// Example: save to database, trigger other processes, etc.

	lg.Info("message processed successfully", logFields...)
}
