package remove_worker

import (
	"context"

	pb "runner/proto/server/runner/v1"
)

type Handler struct {
	pb.UnimplementedRunnerServiceServer
}

func (h *Handler) RemoveWorker(ctx context.Context, req *pb.RemoveWorkerRequest) (*pb.RemoveWorkerResponse, error) {
	return &pb.RemoveWorkerResponse{
		Success: true,
	}, nil
}
