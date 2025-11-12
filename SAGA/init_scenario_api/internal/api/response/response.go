package response

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func JSON(w http.ResponseWriter, log *zap.Logger, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Error("failed to encode response", zap.Error(err))
	}
}

func Error(w http.ResponseWriter, log *zap.Logger, code int, error, message string) {
	JSON(w, log, code, ErrorResponse{
		Error:   error,
		Message: message,
	})
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
