package get_scenario_status

import (
	"init_scenario_api/internal/api/response"
	"init_scenario_api/pkg/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	useCase GetScenarioStatusUseCase
}

func NewHandler(useCase GetScenarioStatusUseCase) *Handler {
	return &Handler{useCase: useCase}
}

// GetScenarioStatus godoc
// @Summary      Получить статус сценария
// @Description  Возвращает статус сценария по UUID
// @Tags         scenario
// @Produce      json
// @Param        uuid path string true "UUID сценария"
// @Success      200 {object} dto.GetScenarioStatusResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /scenario/{uuid}/status [get]
func (h *Handler) GetScenarioStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	scenarioUUID := chi.URLParam(r, "uuid")
	if scenarioUUID == "" {
		response.Error(w, log, http.StatusBadRequest, "Invalid uuid", "uuid is required")
		return
	}

	result, err := h.useCase.GetScenarioStatus(ctx, scenarioUUID)
	if err != nil {
		log.Error("failed to get scenario status", zap.Error(err))
		response.Error(w, log, http.StatusInternalServerError, "Failed to get scenario status", err.Error())
		return
	}

	response.JSON(w, log, http.StatusOK, result)
}
