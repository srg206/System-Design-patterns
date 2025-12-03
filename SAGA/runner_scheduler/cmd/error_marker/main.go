package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"runner_scheduler/internal/config"
	"runner_scheduler/internal/infrastructure/repository"
	"runner_scheduler/pkg/closer"
	"runner_scheduler/pkg/database"
	"runner_scheduler/pkg/logger"

	"go.uber.org/zap"
)

const (
	checkIntervalSeconds = 1
	staleTimeoutSeconds  = 30
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

	ticker := time.NewTicker(checkIntervalSeconds * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			if cls.IsShutdown() {
				log.Info("shutdown signal received, stopping error marker")
				break
			}

			select {
			case <-ticker.C:
				log.Info("checking for stale workers and nodes")
				checkAndMarkStaleWorkers(ctx, repo, log)
				checkAndMarkStaleNodes(ctx, repo, log)
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Info("error marker started successfully",
		zap.Int("check_interval_sec", checkIntervalSeconds),
		zap.Int("stale_timeout_sec", staleTimeoutSeconds))

	cls.Wait()

	return 0
}

func checkAndMarkStaleWorkers(ctx context.Context, repo *repository.Repository, log *zap.Logger) {
	staleWorkers, err := repo.GetStaleWorkers(ctx, staleTimeoutSeconds)
	if err != nil {
		log.Error("failed to get stale workers", zap.Error(err))
		return
	}
	log.Info("stale workers", zap.Any("stale_workers", staleWorkers))

	if len(staleWorkers) == 0 {
		return
	}

	for _, worker := range staleWorkers {
		if err := repo.MarkWorkerAsError(ctx, worker.ID); err != nil {
			log.Error("failed to mark worker as error",
				zap.Int32("worker_id", worker.ID),
				zap.Int32("camera_id", worker.CameraID),
				zap.Error(err))
			continue
		}

		log.Warn("marked worker as error due to stale heartbeat",
			zap.Int32("worker_id", worker.ID),
			zap.Int32("camera_id", worker.CameraID),
			zap.String("previous_status", worker.Status),
			zap.Time("last_heartbeat", worker.LastHeartbeat.Time))
	}

	log.Info("stale workers check completed",
		zap.Int("total_marked", len(staleWorkers)))
}

func checkAndMarkStaleNodes(ctx context.Context, repo *repository.Repository, log *zap.Logger) {
	staleNodes, err := repo.GetStaleNodes(ctx, staleTimeoutSeconds)
	if err != nil {
		log.Error("failed to get stale nodes", zap.Error(err))
		return
	}
	log.Info("stale nodes", zap.Any("stale_nodes", staleNodes))
	if len(staleNodes) == 0 {
		return
	}

	for _, node := range staleNodes {
		if err := repo.MarkNodeAsError(ctx, node.ID); err != nil {
			log.Error("failed to mark node as error",
				zap.Int32("node_id", node.ID),
				zap.String("addr", node.Addr),
				zap.Error(err))
			continue
		}

		log.Warn("marked node as error due to stale heartbeat",
			zap.Int32("node_id", node.ID),
			zap.String("addr", node.Addr),
			zap.String("previous_state", node.State),
			zap.Any("last_heartbeat", node.LastHeartbeat))
	}

	log.Info("stale nodes check completed",
		zap.Int("total_marked", len(staleNodes)))
}
