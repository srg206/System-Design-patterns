package application

import (
	"context"
	"fmt"
	"init_scenario_api/config"
	"init_scenario_api/internal/infastructure/kafka"
	"init_scenario_api/internal/infastructure/repository"
	"init_scenario_api/pkg/closer"
	"init_scenario_api/pkg/database"
	"init_scenario_api/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	Logger        *zap.Logger
	Closer        *closer.Closer
	PostgresRepo  *repository.Repository
	KafkaProducer kafka.Producer
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

	// Инициализация Kafka producer
	kafkaConfig := kafka.DefaultConfig(cfg.Producer.KafkaBrokers...)
	kafkaProducer, err := kafka.NewKafkaProducer(kafkaConfig, log)
	if err != nil {
		log.Error("failed to initialize kafka producer", zap.Error(err))
		return app, err
	}
	app.KafkaProducer = kafkaProducer

	app.Closer.Add(func() error {
		return kafkaProducer.Close()
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
