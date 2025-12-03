package global_handler

import (
	"context"
	"log"

	"runner/internal/infrastructure/worker_manager"
	"runner/internal/infrastructure/workers/obtain_frame_worker"
	pb "runner/proto/server/runner/v1"
)

type RunnerServiceHandler struct {
	pb.UnimplementedRunnerServiceServer
	workerManager   *worker_manager.WorkerManager
	inferenceClient obtain_frame_worker.InferenceService
	s3Client        obtain_frame_worker.S3Storage
}

func NewRunnerServiceHandler(
	wm *worker_manager.WorkerManager,
	inferenceClient obtain_frame_worker.InferenceService,
	s3Client obtain_frame_worker.S3Storage,
) *RunnerServiceHandler {
	return &RunnerServiceHandler{
		workerManager:   wm,
		inferenceClient: inferenceClient,
		s3Client:        s3Client,
	}
}

func (h *RunnerServiceHandler) StartWorker(ctx context.Context, req *pb.StartWorkerRequest) (*pb.StartWorkerResponse, error) {
	cameraID := int(req.CameraId)
	log.Printf("starting worker for camera %d (url: %s)", cameraID, req.Url)
	worker := obtain_frame_worker.ObtainFrameWorkerNew(req.Url, nil, h.inferenceClient, h.s3Client)
	worker.CameraID = cameraID

	if err := h.workerManager.AddWorker(worker); err != nil {
		log.Printf("failed to add worker for camera %d (url: %s): %v", cameraID, req.Url, err)
		return &pb.StartWorkerResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	log.Printf("worker for camera %d started successfully", cameraID)
	return &pb.StartWorkerResponse{
		Success: true,
	}, nil
}

func (h *RunnerServiceHandler) RemoveWorker(ctx context.Context, req *pb.RemoveWorkerRequest) (*pb.RemoveWorkerResponse, error) {
	if err := h.workerManager.RemoveWorker(int(req.CameraId)); err != nil {
		return &pb.RemoveWorkerResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.RemoveWorkerResponse{
		Success: true,
	}, nil
}
