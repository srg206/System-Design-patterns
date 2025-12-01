package scheduler_processor

import (
	"context"
	"fmt"
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
		workerDistribuition := culculateWorkersDistribution(nodes, len(workers))
		fmt.Println("workerDistribuition", workerDistribuition)
		fmt.Println("workers", workers)
		fmt.Println("nodes", nodes)
		workerId := 0
		for nodeId, workersToAdd := range workerDistribuition {
			for i := 0; i < workersToAdd; i++ {
				err := p.runnerClient.StartWorker(ctx, nodes[nodeId].Addr, workers[workerId].CameraID, workers[workerId].Url)
				if err != nil {
					return err
				}
				fmt.Println("StartWorker", nodes[nodeId].Addr, workers[workerId].CameraID, workers[workerId].Url)
				_, err = p.repo.CreateNodeWorker(ctx, worker.CreateNodeWorkerParams{
					NodeID:   nodes[nodeId].NodeID,
					WorkerID: workers[workerId].ID,
				})
				fmt.Println("CreateNodeWorker", nodes[nodeId].NodeID, workers[workerId].ID)
				if err != nil {
					return err
				}
				_, err = p.repo.UpdateWorkerStatus(ctx, worker.UpdateWorkerStatusParams{
					Status: "running",
					ID:     workers[workerId].ID,
				})
				fmt.Println("UpdateWorkerStatus", workers[workerId].ID)
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

func culculateWorkersDistribution(nodes []worker.GetLeastLoadedNodesRow, workersToAdd int) map[int]int {

	avgWorkersCount := avgWorkersCount(nodes, workersToAdd)
	workerDistribuition := make(map[int]int)

	restWorkersToAdd := workersToAdd
	for id, node := range nodes {
		addWorkers := int(avgWorkersCount) - int(node.WorkerCount)
		if addWorkers > int(restWorkersToAdd) {
			addWorkers = int(restWorkersToAdd)
		}
		workerDistribuition[id] = addWorkers
		restWorkersToAdd -= addWorkers
	}
	return workerDistribuition
}

func avgWorkersCount(nodes []worker.GetLeastLoadedNodesRow, workersToAdd int) int {
	totalWorkersCount := workersToAdd
	for _, node := range nodes {
		totalWorkersCount += int(node.WorkerCount)
	}
	return int(math.Ceil(float64(totalWorkersCount) / float64(len(nodes))))
}
