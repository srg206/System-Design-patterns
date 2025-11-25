package obtain_frame_worker

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	inferencepb "runner/proto/client/inference/v1"

	"gocv.io/x/gocv"
)

type ObtainFrameWorker struct {
	skipFrames      *int
	url             string
	videoCap        *gocv.VideoCapture
	inferenceClient InferenceService
	s3Client        S3Storage
	//dbClient               DBStorage
	lastDetections         []*inferencepb.Detection
	lastUploadedObjectKey  string
	lastDownloadedFileData []byte
}

func ObtainFrameWorkerNew(url string, skipFrames *int, inferenceClient InferenceService, s3Client S3Storage) *ObtainFrameWorker {
	return &ObtainFrameWorker{
		url:             url,
		skipFrames:      skipFrames,
		inferenceClient: inferenceClient,
		s3Client:        s3Client,
	}
}

func (w *ObtainFrameWorker) Init() {
	cap, err := gocv.OpenVideoCapture(w.url)
	if err != nil {
		log.Fatalf("failed to open stream: %v", err)
	}

	w.videoCap = cap

	fps := cap.Get(gocv.VideoCaptureFPS)

	if w.skipFrames == nil {
		skipFrames := int(fps)
		w.skipFrames = &skipFrames
	}

}

func (w *ObtainFrameWorker) Run() error {
	fmt.Println("Run TranscodeRTSPWorker")
	for {
		fmt.Println("Grabbing frame")
		if err := w.videoCap.Grab(1); err != nil {
			fmt.Println("stream closed or cannot grab frames:", err)
			return fmt.Errorf("stream closed or cannot grab frames: %w", err)
		}
		fmt.Println("Grabbed frame")

		nowFrame := gocv.NewMat()
		w.videoCap.Retrieve(&nowFrame)
		fmt.Println("Retrieved frame")

		w.ProcessInference(&nowFrame)

		if err := w.videoCap.Grab(*w.skipFrames); err != nil {
			fmt.Println("stream closed or cannot grab frames:", err)
			return err
		}
		fmt.Println("Skipped frames")
	}
}

func (w *ObtainFrameWorker) Close() error {
	if err := w.videoCap.Close(); err != nil {
		return err
	}
	return nil
}

func (w *ObtainFrameWorker) ProcessInference(frame *gocv.Mat) {
	detections, err := w.sendFrameToInference(frame)
	if err != nil {
		log.Printf("failed to send frame to inference: %v", err)
		return
	}
	if detections != nil {
		w.drawBoundingBoxes(frame, detections)
	}
	// Save the processed frame to a file
	filename := fmt.Sprintf("frame_%d.jpg", time.Now().UnixNano())
	if ok := gocv.IMWrite(filename, *frame); !ok {
		log.Printf("failed to write %s", filename)
	} else {
		log.Printf("saved %s", filename)
	}

}

func (w *ObtainFrameWorker) sendFrameToInference(frame *gocv.Mat) ([]*inferencepb.Detection, error) {
	if w.inferenceClient == nil {
		return nil, nil
	}

	buf, err := gocv.IMEncode(".jpg", *frame)
	if err != nil {
		log.Printf("failed to encode frame: %v", err)
		return nil, fmt.Errorf("failed to encode frame: %w", err)
	}
	defer buf.Close()

	ctx := context.Background()
	resp, err := w.inferenceClient.Detect(ctx, buf.GetBytes())
	if err != nil {
		log.Printf("inference failed: %v", err)
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	w.lastDetections = resp.GetDetections()
	return w.lastDetections, nil
}

func (w *ObtainFrameWorker) drawBoundingBoxes(frame *gocv.Mat, detections []*inferencepb.Detection) {
	if len(w.lastDetections) == 0 {
		return
	}

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	for _, detection := range w.lastDetections {
		rect := detection.GetRectangle()
		if rect == nil {
			continue
		}

		pt1 := image.Pt(int(rect.GetX0()), int(rect.GetY0()))
		pt2 := image.Pt(int(rect.GetX1()), int(rect.GetY1()))

		gocv.Rectangle(frame, image.Rectangle{Min: pt1, Max: pt2}, red, 2)

		className := detection.GetClassName()
		textPos := image.Pt(int(rect.GetX1()), int(rect.GetY0()))
		gocv.PutText(frame, className, textPos, gocv.FontHersheyPlain, 1.2, red, 2)
	}
}

func (w *ObtainFrameWorker) saveFrameToS3(frame *gocv.Mat) {
	if w.s3Client == nil {
		log.Printf("S3 client not set, skipping upload")
		return
	}

	buf, err := gocv.IMEncode(".jpg", *frame)
	if err != nil {
		log.Printf("failed to encode frame: %v", err)
		return
	}
	defer buf.Close()

	ctx := context.Background()
	filename := "frame.jpg"
	key, err := w.s3Client.UploadFile(ctx, buf.GetBytes(), filename, "image/jpeg", nil)
	if err != nil {
		log.Printf("upload file: %v", err)
		return
	}

	w.lastUploadedObjectKey = key
	log.Printf("uploaded frame, key: %s", key)
}

func (w *ObtainFrameWorker) signAndDownLoadFrame() error {
	if w.s3Client == nil {
		return fmt.Errorf("S3 client not set")
	}
	if w.lastUploadedObjectKey == "" {
		return fmt.Errorf("no uploaded key to sign")
	}

	ctx := context.Background()
	presignedURL, err := w.s3Client.PresignDownload(ctx, w.lastUploadedObjectKey, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("presign download: %w", err)
	}

	data, err := w.s3Client.DownloadFile(ctx, presignedURL)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}

	outFilename := "downloaded_frame.jpg"
	if err := os.WriteFile(outFilename, data, 0644); err != nil {
		log.Printf("failed to save downloaded frame to %s: %v", outFilename, err)
	} else {
		log.Printf("downloaded frame saved to %s", outFilename)
	}
	w.lastDownloadedFileData = data
	log.Printf("downloaded %d bytes", len(data))
	return nil
}

func (w *ObtainFrameWorker) saveFrameToDB(frame *gocv.Mat) {

}
