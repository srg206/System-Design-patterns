package scheduler_processor

import (
	"context"
	"math"

	"runner_scheduler/internal/infrastructure/repository/queries/worker"
)

type Processor struct {
	repo         Repository
	runnerClient RunnerClient
}

var WorkerBatchSize = 10

func NewProcessor(repo Repository, runnerClient RunnerClient) *Processor {
	return &Processor{
		repo:         repo,
		runnerClient: runnerClient,
	}
}

func (p *Processor) Run(ctx context.Context) error {

	err := p.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		workers, err := p.repo.GetOldestWorkersByStatus(ctx, "pending", int32(WorkerBatchSize))
		if err != nil {
			return err
		}

		nodes, err := p.repo.GetLeastLoadedNodes(ctx, int32(WorkerBatchSize))
		if err != nil {
			return err
		}
		workerDistribuition := culculateWorkersDistribution(nodes, int32(WorkerBatchSize))

		workerId := 0
		for nodeId, workersToAdd := range workerDistribuition {
			for i := 0; i < workersToAdd; i++ {
				err := p.runnerClient.StartWorker(ctx, workers[workerId].CameraID, nodes[nodeId].Addr)
				if err != nil {
					return err
				}
				_, err = p.repo.CreateNodeWorker(ctx, worker.CreateNodeWorkerParams{
					NodeID:   nodes[nodeId].NodeID,
					WorkerID: workers[workerId].ID,
				})
				if err != nil {
					return err
				}
				_, err = p.repo.UpdateWorkerStatus(ctx, worker.UpdateWorkerStatusParams{
					Status: "running",
					ID:     workers[workerId].ID,
				})
				if err != nil {
					return err
				}
				workerId++
			}
		}
		return nil
	})
	return err
}

func culculateWorkersDistribution(nodes []worker.GetLeastLoadedNodesRow, workersToAdd int32) map[int]int {

	avgWorkersCount := avgWorkersCount(nodes, workersToAdd)
	workerDistribuition := make(map[int]int)

	restWorkersToAdd := workersToAdd
	for id, node := range nodes {
		addWorkers := int(avgWorkersCount) - int(node.WorkerCount)
		if addWorkers > int(restWorkersToAdd) {
			addWorkers = int(restWorkersToAdd)
		}
		workerDistribuition[id] = addWorkers
		restWorkersToAdd -= int32(addWorkers)
	}
	return workerDistribuition
}

func avgWorkersCount(nodes []worker.GetLeastLoadedNodesRow, workersToAdd int32) int32 {
	totalWorkersCount := int(workersToAdd)
	for _, node := range nodes {
		totalWorkersCount += int(node.WorkerCount)
	}
	return int32(math.Ceil(float64(totalWorkersCount) / float64(len(nodes))))
}
