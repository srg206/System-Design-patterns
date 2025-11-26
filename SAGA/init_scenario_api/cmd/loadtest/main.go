package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	targetURL      = "http://localhost:3000/api/v1/scenario/init"
	totalRequests  = 3
	workerCount    = 10
	requestTimeout = 10 * time.Second
)

type payload struct {
	CameraID int `json:"camera_id"`
}

type result struct {
	cameraID   int
	success    bool
	statusCode int
	duration   time.Duration
	err        error
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int)
	results := make(chan result, workerCount)

	var workers sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			client := &http.Client{Timeout: requestTimeout}

			for cameraID := range jobs {
				start := time.Now()
				res := result{cameraID: cameraID}

				body, err := json.Marshal(payload{CameraID: cameraID})
				if err != nil {
					res.err = fmt.Errorf("marshal payload: %w", err)
					results <- res
					continue
				}

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
				if err != nil {
					res.err = fmt.Errorf("build request: %w", err)
					results <- res
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					res.err = fmt.Errorf("send request: %w", err)
					results <- res
					continue
				}

				res.statusCode = resp.StatusCode
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				res.success = resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
				if !res.success && res.err == nil {
					res.err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				}

				res.duration = time.Since(start)
				results <- res
			}
		}()
	}

	go func() {
		for i := 1; i <= totalRequests; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	printStatistics(results)
}

func printStatistics(results <-chan result) {
	start := time.Now()

	var successCount, failureCount int
	var successDuration time.Duration

	for res := range results {
		if res.success {
			successCount++
			successDuration += res.duration
			fmt.Printf("[OK]   camera_id=%4d status=%3d duration=%s\n",
				res.cameraID, res.statusCode, res.duration)
			continue
		}

		failureCount++
		fmt.Printf("[FAIL] camera_id=%4d error=%v\n", res.cameraID, res.err)
	}

	totalDuration := time.Since(start)

	fmt.Println("----- summary -----")
	fmt.Printf("total requests : %d\n", totalRequests)
	fmt.Printf("success        : %d\n", successCount)
	fmt.Printf("failed         : %d\n", failureCount)
	fmt.Printf("total duration : %s\n", totalDuration)

	if successCount > 0 {
		fmt.Printf("avg success    : %s\n", successDuration/time.Duration(successCount))
	}
}
