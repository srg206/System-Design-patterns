package mocks

import (
	"context"
	"runner_scheduler/internal/infrastructure/repository/queries/worker"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type RepositoryMock struct {
	mu          sync.RWMutex
	workers     map[int32]worker.Worker
	nodes       map[string]worker.GetLeastLoadedNodesRow
	nodeWorkers []worker.NodeWorker
	nextID      int32

	// error injection
	GetWorkersErr         error
	GetNodesErr           error
	CreateNodeWorkerErr   error
	UpdateWorkerStatusErr error
}

func NewRepositoryMock() *RepositoryMock {
	return &RepositoryMock{
		workers:     make(map[int32]worker.Worker),
		nodes:       make(map[string]worker.GetLeastLoadedNodesRow),
		nodeWorkers: make([]worker.NodeWorker, 0),
		nextID:      1,
	}
}

func (r *RepositoryMock) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	return tFunc(ctx)
}

func (r *RepositoryMock) GetOldestWorkersByStatus(ctx context.Context, status string, limit int32) ([]worker.Worker, error) {
	if r.GetWorkersErr != nil {
		return nil, r.GetWorkersErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []worker.Worker
	for _, w := range r.workers {
		if w.Status == status {
			result = append(result, w)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Time.Before(result[j].CreatedAt.Time)
	})
	if int32(len(result)) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *RepositoryMock) GetLeastLoadedNodes(ctx context.Context, limit int32) ([]worker.GetLeastLoadedNodesRow, error) {
	if r.GetNodesErr != nil {
		return nil, r.GetNodesErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []worker.GetLeastLoadedNodesRow
	for _, n := range r.nodes {
		result = append(result, n)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].WorkerCount < result[j].WorkerCount
	})
	if int32(len(result)) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *RepositoryMock) AddWorker(w worker.Worker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workers[w.ID] = w
}

func (r *RepositoryMock) AddNode(n worker.GetLeastLoadedNodesRow) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[n.NodeID] = n
}

func (r *RepositoryMock) CreateNodeWorker(ctx context.Context, arg worker.CreateNodeWorkerParams) (worker.NodeWorker, error) {
	if r.CreateNodeWorkerErr != nil {
		return worker.NodeWorker{}, r.CreateNodeWorkerErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	nw := worker.NodeWorker{
		ID:       r.nextID,
		NodeID:   arg.NodeID,
		WorkerID: arg.WorkerID,
		AssignedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
	}
	r.nextID++
	r.nodeWorkers = append(r.nodeWorkers, nw)

	if node, ok := r.nodes[arg.NodeID]; ok {
		node.WorkerCount++
		r.nodes[arg.NodeID] = node
	}
	return nw, nil
}

func (r *RepositoryMock) GetNodeWorkers() []worker.NodeWorker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodeWorkers
}

func (r *RepositoryMock) GetNodeWorkerCount(nodeID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, nw := range r.nodeWorkers {
		if nw.NodeID == nodeID {
			count++
		}
	}
	return count
}

func (r *RepositoryMock) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodeWorkers = make([]worker.NodeWorker, 0)
	r.nextID = 1
}

func (r *RepositoryMock) UpdateWorkerStatus(ctx context.Context, arg worker.UpdateWorkerStatusParams) (worker.Worker, error) {
	if r.UpdateWorkerStatusErr != nil {
		return worker.Worker{}, r.UpdateWorkerStatusErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	w, ok := r.workers[arg.ID]
	if !ok {
		return worker.Worker{}, nil
	}
	w.Status = arg.Status
	w.UpdatedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
	r.workers[arg.ID] = w
	return w, nil
}

func (r *RepositoryMock) GetWorker(id int32) (worker.Worker, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.workers[id]
	return w, ok
}
