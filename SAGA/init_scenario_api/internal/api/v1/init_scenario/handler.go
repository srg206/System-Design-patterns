package init_scenario

import (
	"init_scenario_api/internal/api/response"
	"init_scenario_api/internal/models/dto"
	"init_scenario_api/pkg/logger"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	useCase InitScenarioUseCase
}

func NewHandler(useCase InitScenarioUseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// InitScenario godoc
// @Summary      Инициализировать сценарий
// @Description  Создает новый сценарий и отправляет событие в Kafka
// @Tags         scenario
// @Accept       json
// @Produce      json
// @Param        request body dto.InitScenarioRequest true "Данные для создания сценария"
// @Success      201 {object} dto.InitScenarioResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /scenario/init [post]
func (h *Handler) InitScenario(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.InitScenarioRequest

	if err := req.Decode(r.Body); err != nil {
		log.Error("failed to decode request", zap.Error(err))
		response.Error(w, log, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.CameraID <= 0 {
		response.Error(w, log, http.StatusBadRequest, "Invalid camera_id", "camera_id must be positive")
		return
	}

	output, err := h.useCase.InitScenario(ctx, req)
	if err != nil {
		log.Error("failed to init scenario", zap.Error(err))
		response.Error(w, log, http.StatusInternalServerError, "Failed to init scenario", err.Error())
		return
	}

	resp := dto.InitScenarioResponse{
		ScenarioUUID: output.ScenarioUUID,
		Status:       output.Status,
	}

	response.JSON(w, log, http.StatusCreated, resp)
}
