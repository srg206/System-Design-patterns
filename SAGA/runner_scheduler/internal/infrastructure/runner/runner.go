package runner

import (
	"context"
	"fmt"

	pb "runner_scheduler/proto/runner/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) StartWorker(ctx context.Context, cameraID int32, url string) error {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc dial: %w", err)
	}
	defer conn.Close()

	client := pb.NewRunnerServiceClient(conn)
	resp, err := client.StartWorker(ctx, &pb.StartWorkerRequest{
		CameraId: cameraID,
		Url:      url,
	})
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("start worker failed: %s", resp.Error)
	}
	return nil
}

func (c *Client) RemoveWorker(ctx context.Context, cameraID int32, url string) error {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc dial: %w", err)
	}
	defer conn.Close()

	client := pb.NewRunnerServiceClient(conn)
	resp, err := client.RemoveWorker(ctx, &pb.RemoveWorkerRequest{
		CameraId: cameraID,
	})
	if err != nil {
		return fmt.Errorf("remove worker: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("remove worker failed: %s", resp.Error)
	}
	return nil
}
