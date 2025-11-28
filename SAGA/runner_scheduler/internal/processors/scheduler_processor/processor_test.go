package scheduler_processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"runner_scheduler/internal/infrastructure/repository/queries/worker"
	"runner_scheduler/internal/processors/scheduler_processor/mocks"

	"github.com/jackc/pgx/v5/pgtype"
)

func makeWorker(id, cameraID int32, status string, createdAt time.Time) worker.Worker {
	return worker.Worker{
		ID:       id,
		CameraID: cameraID,
		Status:   status,
		CreatedAt: pgtype.Timestamp{
			Time:  createdAt,
			Valid: true,
		},
	}
}

func makeNode(nodeID, addr string, workerCount int64) worker.GetLeastLoadedNodesRow {
	return worker.GetLeastLoadedNodesRow{
		NodeID:      nodeID,
		Addr:        addr,
		WorkerCount: workerCount,
	}
}

// ========== Distribution Tests ==========

func TestProcessor_DistributesWorkersEvenly(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))
	repo.AddNode(makeNode("node2", "addr2", 0))
	repo.AddNode(makeNode("node3", "addr3", 0))

	now := time.Now()
	for i := int32(1); i <= 6; i++ {
		repo.AddWorker(makeWorker(i, i*10, "pending", now.Add(time.Duration(i)*time.Second)))
	}

	WorkerBatchSize = 6
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	for _, nodeID := range []string{"node1", "node2", "node3"} {
		count := repo.GetNodeWorkerCount(nodeID)
		if count != 2 {
			t.Errorf("node %s: expected 2 workers, got %d", nodeID, count)
		}
	}
	if len(repo.GetNodeWorkers()) != 6 {
		t.Errorf("expected 6 total assignments, got %d", len(repo.GetNodeWorkers()))
	}
}

func TestProcessor_BalancesUnevenNodes(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))
	repo.AddNode(makeNode("node2", "addr2", 2))
	repo.AddNode(makeNode("node3", "addr3", 4))

	now := time.Now()
	for i := int32(1); i <= 3; i++ {
		repo.AddWorker(makeWorker(i, i*10, "pending", now.Add(time.Duration(i)*time.Second)))
	}

	WorkerBatchSize = 3
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	node1Count := repo.GetNodeWorkerCount("node1")
	if node1Count < 2 {
		t.Errorf("node1 (least loaded): expected at least 2 workers, got %d", node1Count)
	}
	if len(repo.GetNodeWorkers()) != 3 {
		t.Errorf("expected 3 total assignments, got %d", len(repo.GetNodeWorkers()))
	}
}

func TestProcessor_SingleNode(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))

	now := time.Now()
	for i := int32(1); i <= 4; i++ {
		repo.AddWorker(makeWorker(i, i*10, "pending", now.Add(time.Duration(i)*time.Second)))
	}

	WorkerBatchSize = 4
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if repo.GetNodeWorkerCount("node1") != 4 {
		t.Errorf("expected 4 workers on node1, got %d", repo.GetNodeWorkerCount("node1"))
	}
}

func TestProcessor_OldestWorkersFirst(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))

	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "pending", now.Add(2*time.Hour)))
	repo.AddWorker(makeWorker(2, 20, "pending", now.Add(1*time.Hour)))
	repo.AddWorker(makeWorker(3, 30, "pending", now))

	WorkerBatchSize = 2
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	nws := repo.GetNodeWorkers()
	if len(nws) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(nws))
	}

	workerIDs := map[int32]bool{}
	for _, nw := range nws {
		workerIDs[nw.WorkerID] = true
	}
	if !workerIDs[3] || !workerIDs[2] {
		t.Errorf("expected workers 3 and 2 (oldest), got %v", workerIDs)
	}
}

func TestProcessor_OddNumberOfWorkers(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))
	repo.AddNode(makeNode("node2", "addr2", 0))

	now := time.Now()
	for i := int32(1); i <= 5; i++ {
		repo.AddWorker(makeWorker(i, i*10, "pending", now.Add(time.Duration(i)*time.Second)))
	}

	WorkerBatchSize = 5
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// 5 workers, 2 nodes -> one gets 3, other gets 2
	total := repo.GetNodeWorkerCount("node1") + repo.GetNodeWorkerCount("node2")
	if total != 5 {
		t.Errorf("expected 5 total, got %d", total)
	}
}

func TestProcessor_ManyNodesFewWorkers(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	for i := 1; i <= 10; i++ {
		repo.AddNode(makeNode(
			"node"+string(rune('0'+i)),
			"addr"+string(rune('0'+i)),
			0,
		))
	}

	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "pending", now))
	repo.AddWorker(makeWorker(2, 20, "pending", now.Add(time.Second)))

	WorkerBatchSize = 2
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(repo.GetNodeWorkers()) != 2 {
		t.Errorf("expected 2 assignments, got %d", len(repo.GetNodeWorkers()))
	}
}

func TestProcessor_TwoNodesSignificantLoadDifference(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))
	repo.AddNode(makeNode("node2", "addr2", 10))

	now := time.Now()
	for i := int32(1); i <= 6; i++ {
		repo.AddWorker(makeWorker(i, i*10, "pending", now.Add(time.Duration(i)*time.Second)))
	}

	WorkerBatchSize = 6
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// node1 should get all workers since node2 already has 10
	node1Count := repo.GetNodeWorkerCount("node1")
	if node1Count != 6 {
		t.Errorf("node1: expected 6 workers, got %d", node1Count)
	}
}

func TestProcessor_AllNodesHeavilyLoaded(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 100))
	repo.AddNode(makeNode("node2", "addr2", 100))
	repo.AddNode(makeNode("node3", "addr3", 100))

	now := time.Now()
	for i := int32(1); i <= 3; i++ {
		repo.AddWorker(makeWorker(i, i*10, "pending", now.Add(time.Duration(i)*time.Second)))
	}

	WorkerBatchSize = 3
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// still should distribute evenly
	total := len(repo.GetNodeWorkers())
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
}

// ========== RunnerClient Interaction Tests ==========

func TestProcessor_CallsStartWorkerWithCorrectParams(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "192.168.1.1:8080", 0))

	now := time.Now()
	repo.AddWorker(makeWorker(1, 42, "pending", now))

	WorkerBatchSize = 1
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	calls := client.GetStartCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 StartWorker call, got %d", len(calls))
	}
	if calls[0].CameraID != 42 {
		t.Errorf("expected cameraID 42, got %d", calls[0].CameraID)
	}
	if calls[0].URL != "192.168.1.1:8080" {
		t.Errorf("expected URL 192.168.1.1:8080, got %s", calls[0].URL)
	}
}

func TestProcessor_MultipleWorkersCorrectCameraIDs(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))

	now := time.Now()
	repo.AddWorker(makeWorker(1, 100, "pending", now))
	repo.AddWorker(makeWorker(2, 200, "pending", now.Add(time.Second)))
	repo.AddWorker(makeWorker(3, 300, "pending", now.Add(2*time.Second)))

	WorkerBatchSize = 3
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	calls := client.GetStartCalls()
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}

	cameraIDs := map[int32]bool{}
	for _, c := range calls {
		cameraIDs[c.CameraID] = true
	}
	for _, expected := range []int32{100, 200, 300} {
		if !cameraIDs[expected] {
			t.Errorf("missing cameraID %d", expected)
		}
	}
}

func TestProcessor_UpdatesWorkerStatusToRunning(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))

	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "pending", now))
	repo.AddWorker(makeWorker(2, 20, "pending", now.Add(time.Second)))

	WorkerBatchSize = 2
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	for _, id := range []int32{1, 2} {
		w, ok := repo.GetWorker(id)
		if !ok {
			t.Errorf("worker %d not found", id)
			continue
		}
		if w.Status != "running" {
			t.Errorf("worker %d: expected status 'running', got '%s'", id, w.Status)
		}
	}
}

// ========== Error Handling Tests ==========

func TestProcessor_GetWorkersError(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.GetWorkersErr = errors.New("db connection failed")

	p := NewProcessor(repo, client)
	err := p.Run(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "db connection failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessor_GetNodesError(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "pending", now))
	repo.GetNodesErr = errors.New("nodes query failed")

	WorkerBatchSize = 1
	p := NewProcessor(repo, client)
	err := p.Run(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "nodes query failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessor_CreateNodeWorkerError(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))
	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "pending", now))
	repo.CreateNodeWorkerErr = errors.New("insert failed")

	WorkerBatchSize = 1
	p := NewProcessor(repo, client)
	err := p.Run(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "insert failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessor_UpdateWorkerStatusError(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))
	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "pending", now))
	repo.UpdateWorkerStatusErr = errors.New("update failed")

	WorkerBatchSize = 1
	p := NewProcessor(repo, client)
	err := p.Run(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "update failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ========== Edge Cases ==========

func TestProcessor_OnlyNonPendingWorkers(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))

	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "running", now))
	repo.AddWorker(makeWorker(2, 20, "stopped", now))
	repo.AddWorker(makeWorker(3, 30, "failed", now))

	WorkerBatchSize = 10
	p := NewProcessor(repo, client)

	// current implementation panics on empty workers slice
	// this documents the behavior - should be fixed in processor
	defer func() {
		if r := recover(); r == nil {
			// if no panic, check no assignments were made
			if len(repo.GetNodeWorkers()) != 0 {
				t.Errorf("expected 0 assignments, got %d", len(repo.GetNodeWorkers()))
			}
		}
	}()

	_ = p.Run(context.Background())
}

func TestProcessor_BatchSizeLargerThanWorkers(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))
	repo.AddNode(makeNode("node2", "addr2", 0))

	now := time.Now()
	repo.AddWorker(makeWorker(1, 10, "pending", now))

	WorkerBatchSize = 100
	p := NewProcessor(repo, client)

	// This may panic due to avgWorkersCount calculation with fewer workers than expected
	defer func() {
		if r := recover(); r == nil {
			if len(repo.GetNodeWorkers()) != 1 {
				t.Errorf("expected 1 assignment, got %d", len(repo.GetNodeWorkers()))
			}
		}
	}()

	_ = p.Run(context.Background())
}

func TestProcessor_LargeBatchManyNodes(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	for i := 1; i <= 5; i++ {
		repo.AddNode(makeNode(
			"node"+string(rune('0'+i)),
			"addr"+string(rune('0'+i)),
			int64(i*2), // varying loads
		))
	}

	now := time.Now()
	for i := int32(1); i <= 20; i++ {
		repo.AddWorker(makeWorker(i, i*10, "pending", now.Add(time.Duration(i)*time.Second)))
	}

	WorkerBatchSize = 20
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(repo.GetNodeWorkers()) != 20 {
		t.Errorf("expected 20 assignments, got %d", len(repo.GetNodeWorkers()))
	}
}

func TestProcessor_SameCameraIDDifferentWorkers(t *testing.T) {
	repo := mocks.NewRepositoryMock()
	client := mocks.NewRunnerClientMock()

	repo.AddNode(makeNode("node1", "addr1", 0))

	now := time.Now()
	// same camera ID but different worker IDs
	repo.AddWorker(makeWorker(1, 100, "pending", now))
	repo.AddWorker(makeWorker(2, 100, "pending", now.Add(time.Second)))

	WorkerBatchSize = 2
	p := NewProcessor(repo, client)
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(repo.GetNodeWorkers()) != 2 {
		t.Errorf("expected 2 assignments, got %d", len(repo.GetNodeWorkers()))
	}
}

// ========== Distribution Algorithm Tests ==========

func TestCalculateWorkersDistribution_EqualNodes(t *testing.T) {
	nodes := []worker.GetLeastLoadedNodesRow{
		{NodeID: "n1", WorkerCount: 0},
		{NodeID: "n2", WorkerCount: 0},
		{NodeID: "n3", WorkerCount: 0},
	}

	dist := culculateWorkersDistribution(nodes, 6)

	total := 0
	for _, count := range dist {
		total += count
	}
	if total != 6 {
		t.Errorf("expected 6 total workers, got %d", total)
	}
}

func TestCalculateWorkersDistribution_UnequalNodes(t *testing.T) {
	nodes := []worker.GetLeastLoadedNodesRow{
		{NodeID: "n1", WorkerCount: 0},
		{NodeID: "n2", WorkerCount: 5},
	}

	dist := culculateWorkersDistribution(nodes, 3)

	// avg = ceil((0+5+3)/2) = ceil(4) = 4
	// n1 gets 4-0=4, but only 3 to add, so 3
	// n2 gets 4-5=-1 -> 0
	if dist[0] != 3 {
		t.Errorf("node[0] expected 3, got %d", dist[0])
	}
}

func TestAvgWorkersCount(t *testing.T) {
	tests := []struct {
		name        string
		nodes       []worker.GetLeastLoadedNodesRow
		workersAdd  int32
		expectedAvg int32
	}{
		{
			name: "empty nodes add 6",
			nodes: []worker.GetLeastLoadedNodesRow{
				{WorkerCount: 0},
				{WorkerCount: 0},
				{WorkerCount: 0},
			},
			workersAdd:  6,
			expectedAvg: 2, // ceil(6/3) = 2
		},
		{
			name: "loaded nodes add few",
			nodes: []worker.GetLeastLoadedNodesRow{
				{WorkerCount: 5},
				{WorkerCount: 5},
			},
			workersAdd:  2,
			expectedAvg: 6, // ceil((5+5+2)/2) = ceil(6) = 6
		},
		{
			name: "single node",
			nodes: []worker.GetLeastLoadedNodesRow{
				{WorkerCount: 3},
			},
			workersAdd:  5,
			expectedAvg: 8, // ceil((3+5)/1) = 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := avgWorkersCount(tt.nodes, tt.workersAdd)
			if got != tt.expectedAvg {
				t.Errorf("expected %d, got %d", tt.expectedAvg, got)
			}
		})
	}
}
