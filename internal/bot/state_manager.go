package bot

import (
	"context"

	"budget-bot/internal/repository"
	"go.uber.org/zap"
)

type StateManager struct {
	stateRepo repository.DialogStateRepository
	logger   *zap.Logger
}

func NewStateManager(repo repository.DialogStateRepository, logger *zap.Logger) *StateManager {
	return &StateManager{stateRepo: repo, logger: logger}
}

func (s *StateManager) SetState(ctx context.Context, telegramID int64, state repository.DialogState, context map[string]any) error {
	return s.stateRepo.SetState(ctx, telegramID, state, context, nil)
}

func (s *StateManager) GetState(ctx context.Context, telegramID int64) (*repository.DialogStateRecord, error) {
	return s.stateRepo.GetState(ctx, telegramID)
}

func (s *StateManager) ClearState(ctx context.Context, telegramID int64) error {
	return s.stateRepo.ClearState(ctx, telegramID)
}


