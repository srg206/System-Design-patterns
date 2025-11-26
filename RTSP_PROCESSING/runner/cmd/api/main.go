package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"runner/internal/env"
	"runner/internal/grpc_api/v1/global_handler"
	"runner/internal/infrastructure/inference_service"
	"runner/internal/infrastructure/s3"
	"runner/internal/infrastructure/worker_manager"
	pb "runner/proto/server/runner/v1"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	if err != nil {
		log.Printf("method=%s duration=%v error=%v", info.FullMethod, duration, err)
	} else {
		log.Printf("method=%s duration=%v", info.FullMethod, duration)
	}

	return resp, err
}

func main() {
	cfg := env.LoadEnv()

	inferenceService, err := inference_service.New(inference_service.Config{
		Address: cfg.Inference.Address,
		Timeout: cfg.Inference.Timeout,
	})
	if err != nil {
		log.Fatalf("failed to create inference service: %v", err)
	}
	defer inferenceService.Close()

	s3Client, err := s3.NewClient(s3.Config{
		Endpoint:    cfg.S3.Endpoint,
		Bucket:      cfg.S3.Bucket,
		AccessKeyID: cfg.S3.AccessKey,
		SecretKey:   cfg.S3.SecretKey,
	})
	if err != nil {
		log.Fatalf("failed to create s3 client: %v", err)
	}

	workerManager := worker_manager.NewWorkerManager()
	defer workerManager.Close()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))

	handler := global_handler.NewRunnerServiceHandler(workerManager, inferenceService, s3Client)
	pb.RegisterRunnerServiceServer(s, handler)

	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
