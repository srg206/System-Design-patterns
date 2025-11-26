package health

import (
	"net/http"

	"init_scenario_api/pkg/logger"

	"go.uber.org/zap"
)

// HealthCheckHandler проверяет состояние API
// @Summary      Проверка здоровья сервиса
// @Description  Возвращает статус OK если сервис работает
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	lg := logger.FromContext(r.Context())

	lg.Info("health check requested", zap.String("endpoint", "/health"))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
