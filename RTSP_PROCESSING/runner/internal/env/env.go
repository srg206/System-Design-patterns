package env

import (
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type S3Env struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type InferenceEnv struct {
	Address string
	Timeout time.Duration
}

type Env struct {
	S3        S3Env
	Inference InferenceEnv
}

func LoadEnv() *Env {
	findAndLoadEnv()

	return &Env{
		S3: S3Env{
			Endpoint:  GetEnv("S3_ENDPOINT", "http://localhost:9000"),
			AccessKey: GetEnv("S3_ACCESS_KEY", "minioadmin"),
			SecretKey: GetEnv("S3_SECRET_KEY", "minioadmin"),
			Bucket:    GetEnv("S3_BUCKET", "frames"),
		},
		Inference: InferenceEnv{
			Address: GetEnv("INFERENCE_SERVICE_ADDRESS", "localhost:50051"),
			Timeout: func() time.Duration {
				d, err := time.ParseDuration(GetEnv("INFERENCE_SERVICE_TIMEOUT", "30s"))
				if err != nil {
					return 30 * time.Second
				}
				return d
			}(),
		},
	}
}

func findAndLoadEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
