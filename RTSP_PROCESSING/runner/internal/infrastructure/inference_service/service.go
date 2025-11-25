package inference_service

import (
	"context"
	"fmt"
	"os"
	"time"

	inferencepb "runner/proto/client/inference/v1"

	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InferenceService struct {
	client inferencepb.InferenceServiceClient
	conn   *grpc.ClientConn
	addr   string
	cb     *gobreaker.CircuitBreaker[*inferencepb.DetectResponse]
}

type Config struct {
	Address            string
	Timeout            time.Duration
	MaxRetries         int
	RetryDelay         time.Duration
	CircuitBreakerName string
}

func New(cfg Config) (*InferenceService, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = time.Second
	}
	if cfg.CircuitBreakerName == "" {
		cfg.CircuitBreakerName = "inference-service"
	}

	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inference service: %w", err)
	}

	client := inferencepb.NewInferenceServiceClient(conn)

	cbSettings := gobreaker.Settings{
		Name:        cfg.CircuitBreakerName,
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	}

	return &InferenceService{
		client: client,
		conn:   conn,
		addr:   cfg.Address,
		cb:     gobreaker.NewCircuitBreaker[*inferencepb.DetectResponse](cbSettings),
	}, nil
}

func (s *InferenceService) DetectFromFile(ctx context.Context, filePath string) (*inferencepb.DetectResponse, error) {
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file %s: %w", filePath, err)
	}

	return s.Detect(ctx, imageData)
}

func (s *InferenceService) Detect(ctx context.Context, imageData []byte) (*inferencepb.DetectResponse, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	return s.cb.Execute(func() (*inferencepb.DetectResponse, error) {
		return s.detectWithRetry(ctx, imageData, 3, time.Second)
	})
}

func (s *InferenceService) detectWithRetry(ctx context.Context, imageData []byte, maxRetries int, retryDelay time.Duration) (*inferencepb.DetectResponse, error) {
	req := &inferencepb.DetectRequest{
		Image: imageData,
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		resp, err := s.client.Detect(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed to detect objects after %d retries: %w", maxRetries, lastErr)
}

func (s *InferenceService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
