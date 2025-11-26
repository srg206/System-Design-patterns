package middleware

import (
	"net/http"
	"strings"

	"init_scenario_api/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Middleware struct {
	logger *zap.Logger
}

func New(log *zap.Logger) Middleware {
	return Middleware{
		logger: log,
	}
}

func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.URL.Path, "/swagger") {
			next.ServeHTTP(w, r)
			return
		}
		requestUID := uuid.New()

		log := m.logger.With(
			zap.String("route", r.URL.Path),
			zap.String("request_uid", requestUID.String()),
		)

		log.Info("Http request started")
		ctx := logger.WithContext(r.Context(), log)

		next.ServeHTTP(w, r.WithContext(ctx))

		log.Info("Http request finished")
	})
}
