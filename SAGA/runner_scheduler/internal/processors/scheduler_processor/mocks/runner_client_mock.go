package mocks

import (
	"context"
	"sync"
)

type StartWorkerCall struct {
	CameraID int32
	URL      string
}

type RunnerClientMock struct {
	mu              sync.RWMutex
	StartWorkerErr  error
	RemoveWorkerErr error
	StartCalls      []StartWorkerCall
	RemoveCalls     []int32
}

func NewRunnerClientMock() *RunnerClientMock {
	return &RunnerClientMock{
		StartCalls:  make([]StartWorkerCall, 0),
		RemoveCalls: make([]int32, 0),
	}
}

func (r *RunnerClientMock) StartWorker(ctx context.Context, cameraID int32, url string) error {
	if r.StartWorkerErr != nil {
		return r.StartWorkerErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.StartCalls = append(r.StartCalls, StartWorkerCall{CameraID: cameraID, URL: url})
	return nil
}

func (r *RunnerClientMock) RemoveWorker(ctx context.Context, cameraID int32, url string) error {
	if r.RemoveWorkerErr != nil {
		return r.RemoveWorkerErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RemoveCalls = append(r.RemoveCalls, cameraID)
	return nil
}

func (r *RunnerClientMock) GetStartCalls() []StartWorkerCall {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.StartCalls
}

func (r *RunnerClientMock) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.StartCalls = make([]StartWorkerCall, 0)
	r.RemoveCalls = make([]int32, 0)
}
