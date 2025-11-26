package inference_service

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// TestDetectFromFile демонстрирует использование клиента для отправки изображения на inference
func TestDetectFromFile(t *testing.T) {

	cfg := Config{
		Address: "localhost:50051", // Замените на реальный адрес вашего inference сервера
		Timeout: 10 * time.Second,
	}

	service, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create inference service: %v", err)
	}
	defer service.Close()

	imagePath := filepath.Join("frame_000000.jpg")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := service.DetectFromFile(ctx, imagePath)
	if err != nil {
		t.Fatalf("Failed to detect objects: %v", err)
	}

	t.Logf("Detected %d objects", len(resp.Detections))
	for i, detection := range resp.Detections {
		t.Logf("Detection %d:", i+1)
		t.Logf("  Class: %s", detection.ClassName)
		if detection.Rectangle != nil {
			t.Logf("  Coordinates: (%.2f, %.2f) - (%.2f, %.2f)",
				detection.Rectangle.X0,
				detection.Rectangle.Y0,
				detection.Rectangle.X1,
				detection.Rectangle.Y1,
			)
		}
	}
}
