// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"

	"budget-bot/internal/repository"
	"go.uber.org/zap"
)

// StateManager manages dialog states.
type StateManager struct {
	stateRepo repository.DialogStateRepository
	logger   *zap.Logger
}

// NewStateManager constructs a StateManager.
func NewStateManager(repo repository.DialogStateRepository, logger *zap.Logger) *StateManager {
	return &StateManager{stateRepo: repo, logger: logger}
}

// SetState sets a dialog state with optional context.
func (s *StateManager) SetState(ctx context.Context, telegramID int64, state repository.DialogState, context map[string]any) error {
	return s.stateRepo.SetState(ctx, telegramID, state, context, nil)
}

// GetState returns current dialog state.
func (s *StateManager) GetState(ctx context.Context, telegramID int64) (*repository.DialogStateRecord, error) {
	return s.stateRepo.GetState(ctx, telegramID)
}

// ClearState clears dialog state for a user.
func (s *StateManager) ClearState(ctx context.Context, telegramID int64) error {
	return s.stateRepo.ClearState(ctx, telegramID)
}


