package obtain_frame_worker

import (
	"context"
	inferencepb "runner/proto/client/inference/v1"
	"time"
)

type S3Storage interface {
	UploadFile(ctx context.Context, data []byte, filename string, contentType string, key *string) (string, error)
	DownloadFile(ctx context.Context, key string) ([]byte, error)
	PresignDownload(ctx context.Context, key string, expiresIn time.Duration) (string, error)
}

type InferenceService interface {
	Detect(ctx context.Context, imageData []byte) (*inferencepb.DetectResponse, error)
}

type DBStorage interface {
}
