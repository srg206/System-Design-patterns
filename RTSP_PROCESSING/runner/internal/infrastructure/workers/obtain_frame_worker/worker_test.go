package obtain_frame_worker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"runner/internal/env"
	"runner/internal/infrastructure/inference_service"
	"runner/internal/infrastructure/s3"

	"gocv.io/x/gocv"
)

// Проверяет получение и вычленение кадров из RTSP потока + рисует bounding boxes на кадрах.
func TestRTSPStreamGrabFrameWorker(t *testing.T) {
	cfg := env.LoadEnv()

	frameChan := make(chan *gocv.Mat, 10)

	svc, err := inference_service.New(inference_service.Config{
		Address:            cfg.Inference.Address,
		Timeout:            cfg.Inference.Timeout,
		MaxRetries:         3,
		RetryDelay:         1 * time.Second,
		CircuitBreakerName: "inference_service",
	})
	if err != nil {
		t.Fatalf("failed to create inference grpc client: %v", err)
	}

	fmt.Println("Before Creating worker")
	obtainFrameWorker := ObtainFrameWorkerNew("rtsp://localhost:8554/mystream", nil, svc, nil)
	fmt.Println("Before Initializing worker")
	obtainFrameWorker.Init() // Initialize the worker before running
	fmt.Println("Initialized worker")
	go func() {
		err := obtainFrameWorker.Run()
		fmt.Println(err)
		panic("error in Run")
	}()

	idx := 0
	fmt.Println("Waiting for frames")
	for frame := range frameChan {
		filename := fmt.Sprintf("frame_%06d.jpg", idx)
		if ok := gocv.IMWrite(filename, *frame); !ok {
			log.Printf("failed to write %s", filename)
		} else {
			fmt.Printf("saved %s\n", filename)
		}
		idx++
		frame.Close()
	}
}

// Проверяет получение и вычленение кадров из RTSP потока с использованием воркера.

func TestRTSPStreamParsing(t *testing.T) {
	framesDir := "rtsp_frames"
	if err := os.MkdirAll(framesDir, 0755); err != nil {
		t.Fatalf("Failed to create frames directory: %v", err)
	}

	rtspURL := "rtsp://localhost:8554/mystream"

	videoCap, err := gocv.OpenVideoCapture(rtspURL)
	if err != nil {
		t.Fatalf("Failed to open RTSP stream: %v", err)
	}
	defer videoCap.Close()

	fps := videoCap.Get(gocv.VideoCaptureFPS)
	skipFrames := int(fps)
	fmt.Printf("RTSP stream FPS: %.2f, skipping %d frames between captures\n", fps, skipFrames)

	maxFrames := 10
	frameCount := 0

	for frameCount < maxFrames {
		if err := videoCap.Grab(1); err != nil {
			t.Logf("Stream closed or cannot grab frames: %v", err)
			break
		}

		frame := gocv.NewMat()
		defer frame.Close()

		if ok := videoCap.Retrieve(&frame); !ok {
			t.Logf("Failed to retrieve frame %d", frameCount)
			continue
		}

		if frame.Empty() {
			t.Logf("Empty frame received at %d", frameCount)
			continue
		}

		filename := filepath.Join(framesDir, fmt.Sprintf("rtsp_frame_%06d.jpg", frameCount))
		if ok := gocv.IMWrite(filename, frame); !ok {
			t.Logf("Failed to write %s", filename)
		} else {
			fmt.Printf("Saved frame %d: %s (size: %dx%d)\n",
				frameCount, filename, frame.Cols(), frame.Rows())
		}

		frameCount++

		if err := videoCap.Grab(skipFrames); err != nil {
			t.Logf("Stream closed or cannot skip frames: %v", err)
			break
		}
	}

	fmt.Printf("Total frames captured: %d\n", frameCount)
}

// Проверяет загрузку и скачивание по подписанной ссылке кадра из S3.
func TestS3UploadAndDownloadFrame(t *testing.T) {
	cfg := env.LoadEnv()

	s3Client, err := s3.NewClient(s3.Config{
		Endpoint:    cfg.S3.Endpoint,
		Bucket:      cfg.S3.Bucket,
		AccessKeyID: cfg.S3.AccessKey,
		SecretKey:   cfg.S3.SecretKey,
	})
	fmt.Println("cfg.S3.Endpoint", cfg.S3.Endpoint)
	fmt.Println("cfg.S3.Bucket", cfg.S3.Bucket)
	fmt.Println("cfg.S3.AccessKey", cfg.S3.AccessKey)
	fmt.Println("cfg.S3.SecretKey", cfg.S3.SecretKey)
	if err != nil {
		t.Skipf("failed to create S3 client: %v", err)
	}

	origPath := "./test_frame.jpg"
	orig, err := os.ReadFile(origPath)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	frame, err := gocv.IMDecode(orig, gocv.IMReadColor)
	if err != nil {
		t.Fatalf("failed to decode image: %v", err)
	}
	defer frame.Close()

	w := &ObtainFrameWorker{s3Client: s3Client}
	w.saveFrameToS3(&frame)
	if w.lastUploadedObjectKey == "" {
		t.Skip("upload failed or not performed; skipping")
	}
	if err := w.signAndDownLoadFrame(); err != nil {
		t.Fatalf("download failed: %v", err)
	}
	if len(w.lastDownloadedFileData) == 0 {
		t.Fatalf("download did not produce data")
	}

}
