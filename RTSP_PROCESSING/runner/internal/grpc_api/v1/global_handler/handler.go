package global_handler

import (
	"context"

	pb "runner/proto/server/runner/v1"
)

type RunnerServiceHandler struct {
	pb.UnimplementedRunnerServiceServer
}

func (h *RunnerServiceHandler) StartWorker(ctx context.Context, req *pb.StartWorkerRequest) (*pb.StartWorkerResponse, error) {
	// Логика из add_worker
	return &pb.StartWorkerResponse{
		Success: true,
	}, nil
}

func (h *RunnerServiceHandler) RemoveWorker(ctx context.Context, req *pb.RemoveWorkerRequest) (*pb.RemoveWorkerResponse, error) {
	// Логика из remove_worker
	return &pb.RemoveWorkerResponse{
		Success: true,
	}, nil
}
