package add_worker

import (
	"context"

	pb "runner/proto/server/runner/v1"
)

type Handler struct {
	pb.UnimplementedRunnerServiceServer
}

func (h *Handler) StartWorker(ctx context.Context, req *pb.StartWorkerRequest) (*pb.StartWorkerResponse, error) {
	return &pb.StartWorkerResponse{
		Success: true,
	}, nil
}
