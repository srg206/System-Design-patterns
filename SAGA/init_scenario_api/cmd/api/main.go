package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"init_scenario_api/config"
	_ "init_scenario_api/docs"
	loggerMiddleware "init_scenario_api/internal/api/middleware"
	"init_scenario_api/internal/api/v1/health"
	"init_scenario_api/internal/api/v1/init_scenario"
	"init_scenario_api/internal/application"
	initScenarioUseCase "init_scenario_api/internal/usecase/init_scenario"
	"init_scenario_api/pkg/common"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @title           Init Scenario API
// @version         1.0
// @description     API для управления сценариями
// @BasePath        /api/v1
func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
		return common.FailExitCode
	}

	app, err := application.NewApp()
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
		return common.FailExitCode
	}

	app.Logger.Info("application initialized successfully")

	initScenarioUC := initScenarioUseCase.NewUseCase(app.PostgresRepo)
	initScenarioHandler := init_scenario.NewHandler(initScenarioUC)

	r := chi.NewRouter()

	r.Use(loggerMiddleware.New(app.Logger).Handle)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", health.HealthCheckHandler)

		r.Post("/scenario/init", initScenarioHandler.InitScenario)
	})

	addr := fmt.Sprintf(":%d", cfg.API.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	app.Closer.Add(func() error {
		app.Logger.Info("shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			app.Logger.Error("HTTP server shutdown error", zap.Error(err))
			return err
		}
		app.Logger.Info("HTTP server stopped")
		return nil
	})

	go func() {
		app.Logger.Info("application starting", zap.Int("port", cfg.API.Port), zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("ListenAndServe error", zap.Error(err))
		}
	}()

	app.Closer.Wait()

	app.Logger.Info("application stopped")
	return common.SuccessExitCode
}
