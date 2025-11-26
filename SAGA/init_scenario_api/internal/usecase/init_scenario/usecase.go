package init_scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"init_scenario_api/internal/infastructure/repository/queries/outbox"
	"init_scenario_api/internal/infastructure/repository/queries/scenario"
	"init_scenario_api/internal/models/convert"
	"init_scenario_api/internal/models/dto"
	"init_scenario_api/internal/models/entity"
	"init_scenario_api/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UseCase struct {
	repo Repository
}

func NewUseCase(repo Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

func (uc *UseCase) InitScenario(ctx context.Context, input dto.InitScenarioRequest) (*dto.InitScenarioResponse, error) {
	log := logger.FromContext(ctx)

	log.Info("starting purchase",
		zap.Int32("camera_id", input.CameraID),
		zap.String("url", input.URL),
	)

	var result *dto.InitScenarioResponse
	err := uc.repo.WithinTransaction(ctx, func(txCtx context.Context) error {

		createdScenarioDB, err := uc.repo.CreateScenario(txCtx, scenario.CreateScenarioParams{
			Uuid:     uuidToUUIDV7(uuid.New()),
			CameraID: input.CameraID,
			Url:      input.URL,
		})
		if err != nil {
			log.Error("failed to create scenario", zap.Error(err))
			return fmt.Errorf("create scenario: %w", err)
		}

		scenarioEntity := convert.ScenarioFromDB(createdScenarioDB)

		payload := entity.NewInitScenarioPayload(createdScenarioDB.Uuid, input.CameraID, input.URL)
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Error("failed to marshal payload", zap.Error(err))
			return fmt.Errorf("marshal payload: %w", err)
		}

		createdOutboxDB, err := uc.repo.CreateOutboxScenario(txCtx, outbox.CreateOutboxScenarioParams{
			OutboxUuid:   uuidToUUIDV7(uuid.New()),
			ScenarioUuid: createdScenarioDB.Uuid,
			Payload:      payloadBytes,
		})

		if err != nil {
			log.Error("failed to create outbox scenario", zap.Error(err))
			return fmt.Errorf("create outbox scenario: %w", err)
		}

		outboxEntity := convert.OutboxScenarioFromDB(createdOutboxDB)

		log.Info("scenario created", zap.String("scenario_uuid", scenarioEntity.UUID.String()))
		log.Info("outbox scenario created", zap.String("outbox_uuid", outboxEntity.OutboxUUID.String()))

		result = convert.ScenarioToDTO(scenarioEntity)

		return nil
	})

	if err != nil {
		log.Error("transaction failed", zap.Error(err))
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	log.Info("scenario initialized successfully", zap.String("scenario_uuid", result.ScenarioUUID))
	return result, nil
}

func uuidToUUIDV7(u interface{ String() string }) pgtype.UUID {
	var pgUUID pgtype.UUID
	_ = pgUUID.Scan(u.String())
	return pgUUID
}
