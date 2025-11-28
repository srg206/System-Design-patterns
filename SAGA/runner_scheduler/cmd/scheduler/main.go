package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"runner_scheduler/internal/config"
	"runner_scheduler/internal/infrastructure/repository"
	"runner_scheduler/internal/processors/scheduler_processor"
	"runner_scheduler/pkg/closer"
	"runner_scheduler/pkg/database"
	"runner_scheduler/pkg/logger"

	"runner_scheduler/internal/infrastructure/runner"

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
	runnerClient := runner.New()

	processor := scheduler_processor.NewProcessor(repo, runnerClient)

	go func() {
		for {
			err := processor.Run(ctx)
			if err != nil {
				log.Error("failed to run processor", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			time.Sleep(1 * time.Second)
		}
	}()

	log.Info("consumer started successfully")

	cls.Wait()

	return 0
}
