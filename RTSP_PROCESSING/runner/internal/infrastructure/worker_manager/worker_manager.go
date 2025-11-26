package worker_manager

import (
	"fmt"
	"runner/internal/infrastructure/workers/obtain_frame_worker"
	"sync"
)

type WorkerManager struct {
	mu      sync.Mutex
	workers map[int]*obtain_frame_worker.ObtainFrameWorker
}

func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		workers: make(map[int]*obtain_frame_worker.ObtainFrameWorker),
	}
}

func (wm *WorkerManager) AddWorker(worker *obtain_frame_worker.ObtainFrameWorker) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	_, ok := wm.workers[worker.CameraID]
	if ok {
		return fmt.Errorf("worker already exists")
	}

	wm.workers[worker.CameraID] = worker
	worker.Init()
	go worker.Run()

	return nil
}

func (wm *WorkerManager) GetWorker(cameraID int) (*obtain_frame_worker.ObtainFrameWorker, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	worker, ok := wm.workers[cameraID]
	if !ok {
		return nil, fmt.Errorf("worker not found")
	}
	return worker, nil
}

func (wm *WorkerManager) RemoveWorker(cameraID int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	worker, ok := wm.workers[cameraID]
	if !ok {
		return fmt.Errorf("worker not found")
	}
	if worker != nil {
		worker.Close()
		delete(wm.workers, cameraID)
	}
	return nil
}

func (wm *WorkerManager) Close() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for cameraID, worker := range wm.workers {
		worker.Close()
		delete(wm.workers, cameraID)
	}
}
