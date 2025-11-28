package scheduler_processor

import (
	"context"
	"runner_scheduler/internal/infrastructure/repository/queries/worker"
)

type Repository interface {
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
	GetOldestWorkersByStatus(ctx context.Context, status string, limit int32) ([]worker.Worker, error)
	GetLeastLoadedNodes(ctx context.Context, limit int32) ([]worker.GetLeastLoadedNodesRow, error)
	CreateNodeWorker(ctx context.Context, arg worker.CreateNodeWorkerParams) (worker.NodeWorker, error)
	UpdateWorkerStatus(ctx context.Context, arg worker.UpdateWorkerStatusParams) (worker.Worker, error)
}

type RunnerClient interface {
	StartWorker(ctx context.Context, cameraID int32, url string) error
	RemoveWorker(ctx context.Context, cameraID int32, url string) error
}
