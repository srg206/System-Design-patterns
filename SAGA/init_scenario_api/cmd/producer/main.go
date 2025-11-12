package main

// import (
// 	"context"
// 	"log"
// 	"os"
// 	"time"

// 	"init_scenario_api/internal/application"
// 	"init_scenario_api/pkg/common"

// 	"go.uber.org/zap"
// )

// func main() {
// 	os.Exit(run())
// }

// func run() int {
// 	app, err := application.NewApp()
// 	if err != nil {
// 		log.Fatalf("failed to initialize application: %v", err)
// 		return common.FailExitCode
// 	}

// 	app.Logger.Info("producer service starting")

// 	// Создаем контекст для управления жизненным циклом продюсера
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	// Добавляем отмену контекста в closer
// 	app.Closer.Add(func() error {
// 		app.Logger.Info("cancelling producer context...")
// 		cancel()
// 		return nil
// 	})

// 	// Запускаем основную логику продюсера в горутине
// 	go runProducer(ctx, app.Logger)

// 	// Ожидаем сигналов для graceful shutdown
// 	app.Closer.Wait()

// 	app.Logger.Info("producer service stopped")
// 	return common.SuccessExitCode
// }

// func runProducer(ctx context.Context, lg *zap.Logger) {
// 	ticker := time.NewTicker(10 * time.Second)
// 	defer ticker.Stop()

// 	lg.Info("producer worker started")

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			lg.Info("producer worker stopping...")
// 			return
// 		case <-ticker.C:
// 			lg.Info("producer processing tick",
// 				zap.Time("timestamp", time.Now()),
// 				zap.String("status", "processing"),
// 			)
// 			// Здесь будет логика обработки сообщений
// 		}
// 	}
// }
