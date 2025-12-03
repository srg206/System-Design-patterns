package scheduler_processor

import (
	"context"
	"fmt"
	"math"
	"strings"

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

		if len(workers) == 0 {
			return nil
		}

		nodes, err := p.repo.GetLeastLoadedNodes(ctx, int32(WorkerBatchSize))
		if err != nil {
			return err
		}

		if len(nodes) == 0 {
			return nil
		}

		workerDistribuition := culculateWorkersDistribution(nodes, len(workers))
		fmt.Println("workerDistribuition", workerDistribuition)
		fmt.Println("workers", workers)
		fmt.Println("nodes", nodes)
		workerId := 0
		for nodeId, workersToAdd := range workerDistribuition {
			for i := 0; i < workersToAdd; i++ {
				if workerId >= len(workers) {
					return nil
				}
				err := p.runnerClient.StartWorker(ctx, nodes[nodeId].Addr, workers[workerId].CameraID, workers[workerId].Url)
				if err != nil {
					if strings.Contains(err.Error(), "worker already exists") {
						fmt.Printf("worker %d already exists on node %s, updating status to running\n", workers[workerId].CameraID, nodes[nodeId].Addr)
						_, err = p.repo.CreateNodeWorker(ctx, worker.CreateNodeWorkerParams{
							NodeID:   nodes[nodeId].NodeID,
							WorkerID: workers[workerId].ID,
						})
						if err != nil && !strings.Contains(err.Error(), "duplicate key") {
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
						continue
					}
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
	if len(nodes) == 0 {
		return 0
	}
	totalWorkersCount := workersToAdd
	for _, node := range nodes {
		totalWorkersCount += int(node.WorkerCount)
	}
	return int(math.Ceil(float64(totalWorkersCount) / float64(len(nodes))))
}
