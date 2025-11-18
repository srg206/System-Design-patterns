package application

import (
	"context"
	"fmt"
	"inference_scheduler/internal/config"
	"inference_scheduler/internal/infrastructure/kafka"
	"inference_scheduler/internal/infrastructure/repository"
	"inference_scheduler/pkg/closer"
	"inference_scheduler/pkg/database"
	"inference_scheduler/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	Logger        *zap.Logger
	Closer        *closer.Closer
	PostgresRepo  *repository.Repository
	KafkaConsumer kafka.Consumer
	Config        *config.Config
}

func NewApp() (App, error) {
	app := App{}

	app.Closer = closer.New(30 * time.Second)

	log, err := logger.InitLogger()
	if err != nil {
		fmt.Printf("error init logger: %s\n", err.Error())
		return app, err
	}
	app.Logger = log

	app.Closer.Add(func() error {
		_ = app.Logger.Sync()
		return nil
	})

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", zap.Error(err))
		return app, err
	}
	app.Config = cfg

	ctx := context.Background()
	dbPool, err := InitDB(ctx, cfg)
	if err != nil {
		log.Error("can not initialize db", zap.Error(err))
		return app, err
	}
	app.PostgresRepo = repository.NewRepository(dbPool)

	app.Closer.Add(func() error {
		database.Close(dbPool)
		return nil
	})

	// Инициализация Kafka consumer
	kafkaConfig := kafka.DefaultConfig(cfg.Consumer.KafkaBrokers...)
	kafkaConfig.ConsumerGroup = cfg.Consumer.KafkaConsumerGroup

	topics := []string{cfg.Consumer.KafkaInboxInferenceTopic}
	kafkaConsumer, err := kafka.NewKafkaConsumer(kafkaConfig, topics, log)
	if err != nil {
		log.Error("failed to initialize kafka consumer", zap.Error(err))
		return app, err
	}
	app.KafkaConsumer = kafkaConsumer

	app.Closer.Add(func() error {
		log.Info("closing kafka consumer...")
		return kafkaConsumer.Close()
	})

	return app, nil
}

func InitDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	dbPool, err := database.NewPool(ctx, cfg.Database, cfg.Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	return dbPool, nil
}
