package application

import (
	"context"
	"fmt"
	"init_scenario_api/config"
	"init_scenario_api/internal/infastructure/repository"
	"init_scenario_api/pkg/closer"
	"init_scenario_api/pkg/database"
	"init_scenario_api/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	Logger       *zap.Logger
	Closer       *closer.Closer
	PostgresRepo *repository.Repository
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

	ctx := context.Background()
	dbPool, err := InitDB(ctx)
	if err != nil {
		log.Error("can not initialize db", zap.Error(err))
		return app, err
	}
	app.PostgresRepo = repository.NewRepository(dbPool)

	app.Closer.Add(func() error {
		database.Close(dbPool)
		return nil
	})

	return app, nil
}

func InitDB(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	dbPool, err := database.NewPool(ctx, cfg.Database, cfg.Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	return dbPool, nil
}
