package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

type Config struct {
	Endpoint    string
	Bucket      string
	AccessKeyID string
	SecretKey   string
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("s3 config incomplete")
	}

	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &Client{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, data []byte, filename string, contentType string, key *string) (string, error) {
	if key == nil {
		k := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
		key = &k
	}
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &c.bucket,
		Key:         key,
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return *key, nil
}

func (c *Client) PresignDownload(ctx context.Context, key string, expires time.Duration) (string, error) {
	presigner := s3.NewPresignClient(c.client)
	ps, err := presigner.PresignGetObject(
		ctx,
		&s3.GetObjectInput{Bucket: &c.bucket, Key: &key},
		s3.WithPresignExpires(expires),
	)
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}

	return ps.URL, nil
}

func (c *Client) DownloadFile(ctx context.Context, presignedURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, presignedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return data, nil
}
